package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jekyulll/url_shortener/internal/dto"
	"github.com/jekyulll/url_shortener/internal/model"
	"github.com/jekyulll/url_shortener/internal/repository"
	"github.com/jekyulll/url_shortener/pkg/filter"
	"gorm.io/gorm"
)

type CodeStatus int

const (
	CodeAvailable CodeStatus = iota // 全新可用
	CodeExpired                     // 存在但已过期
	CodeInUse                       // 存在且有效
)

type Cacher interface {
	SetURL(ctx context.Context, url model.URL) error
	GetURL(ctx context.Context, shortCode string) (*model.URL, error)
}

type ShortCodeGenerator interface {
	GenerateShortCode() string
}

type URLService struct {
	repo               repository.URLRepository
	filter             filter.BloomFilter
	shortCodeGenerator ShortCodeGenerator
	defaultDuration    time.Duration
	cache              Cacher
	bashURL            string
}

func New(repo repository.URLRepository, filter filter.BloomFilter, generator ShortCodeGenerator, duration time.Duration, cache Cacher, baseURL string) *URLService {
	// 启动时加载所有有效短码到过滤器
	if urls, err := repo.GetAllActiveURLs(context.Background()); err == nil {
		for _, url := range urls {
			filter.Add(url.ShortCode)
		}
	}
	return &URLService{
		repo:               repo,
		shortCodeGenerator: generator,
		defaultDuration:    duration,
		cache:              cache,
		bashURL:            baseURL,
		filter:             filter,
	}
}

// 如出错返回 err，如短链接已存在，返回预定义错误 ErrShortCodeTaken
func (s *URLService) CreateURL(ctx context.Context, req dto.CreateURLRequest) (*dto.CreateURLResponse, error) {
	// 1. 决定要用的短码：优先用用户自己的，其次自动生成
	code := req.CustomeCode
	var err error
	if code == "" {
		code, err = s.getShortCode(ctx, 0)
		if err != nil {
			return nil, err
		}
	} else {
		// 用户定制，先验证它是否可用
		status, err := s.CheckShortCode(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("check custom shortcode: %w", err)
		}
		if status == CodeInUse {
			return nil, ErrShortCodeTaken
		}
	}
	// 2. 写入布隆过滤器
	s.filter.Add(code)
	// 3. 组装要入库的 URL 实体
	u := &model.URL{
		OriginalURL: req.OriginalURL,
		ShortCode:   code,
		IsCustom:    req.CustomeCode != "",
	}
	// 4. 处理过期时间：不传的话使用默认有效期
	if req.Duration == nil {
		u.ExpiredAt = time.Now().Add(s.defaultDuration)
	} else {
		u.ExpiredAt = time.Now().Add(time.Duration(*req.Duration) * time.Hour)
	}
	// 5. 写库
	if err := s.repo.UpsertURL(ctx, u); err != nil {
		return nil, fmt.Errorf("create url record: %w", err)
	}
	// 6. 异步写缓存
	go func() {
		if err := s.cache.SetURL(ctx, *u); err != nil {
			log.Printf("failed to set cache: %v", err)
		}
	}()
	// 6. 返回给上层
	return &dto.CreateURLResponse{
		ShortUrl: s.bashURL + "/" + u.ShortCode,
		//ExpiredAt: u.ExpiredAt,
	}, nil
}

func (s *URLService) GetURL(ctx context.Context, shortCode string) (string, error) {
	// 1. 查找布隆过滤器
	if !s.filter.Exists(shortCode) { // 布隆过滤器中不存在，则一定不存在。
		return "", nil
	}
	// 2. 访问缓存
	url, err := s.cache.GetURL(ctx, shortCode)
	if err != nil {
		return "", err
	}
	if url != nil {
		return url.OriginalURL, nil
	}
	// 3. 缓存中不存在，访问数据库
	url, err = s.repo.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}
	if url == nil { // 不是查询出错，而是数据库没该数据
		return "", nil
	}
	// 4. 存入缓存
	go func() {
		if err := s.cache.SetURL(ctx, *url); err != nil {
			log.Printf("failed to set cache: %v", err)
		}
	}()
	return url.OriginalURL, nil
}

// @pragma n:重试次数
func (s *URLService) getShortCode(ctx context.Context, n int) (string, error) {
	if n > 5 {
		return "", errors.New("retry too many times")
	}
	code := s.shortCodeGenerator.GenerateShortCode()
	status, err := s.CheckShortCode(ctx, code)
	if err != nil {
		return "", err
	}
	if status == CodeAvailable || status == CodeExpired {
		return code, nil
	}
	// 递归调用
	return s.getShortCode(ctx, n+1)
}

func (s *URLService) DeleteAllExpired(ctx context.Context) error {
	return s.repo.DeleteAllExpired(ctx)
}

// TODO 短链接会过期，需要定时重建布隆过滤器（扫描整个数据库，很麻烦）

// CheckShortCode NEW: 从 repo 层提到 service, 与布隆过滤器的判断整合
// TODO 增加缓存的逻辑
// 如果存在于数据库、但是过期了，也视为合法 —— 生成新链接，写入数据库覆盖之前的
func (s *URLService) CheckShortCode(ctx context.Context, shortCode string) (CodeStatus, error) {
	if !s.filter.Exists(shortCode) { // 布隆过滤器中不存在，则一定不存在。
		return CodeAvailable, nil
	}
	// 布隆过滤器中存在，仍然可能不存在。判断是否在数据库中
	url, err := s.repo.GetURLByShortCode(ctx, shortCode)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CodeAvailable, nil // 不存在
	}
	if err != nil {
		return CodeInUse, err
	}
	if url.IsExpired() {
		return CodeExpired, nil
	}
	return CodeInUse, nil
}
