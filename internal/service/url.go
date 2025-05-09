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
)

var (
	ErrShortCodeTaken = errors.New("short code already taken")
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

// TODO 此处传入 filter 是否合理？
func New(repo repository.URLRepository, filter filter.BloomFilter, generator ShortCodeGenerator, duration time.Duration, cache Cacher, baseURL string) *URLService {
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
		// 如果用户定制，先验证它是否可用
		ok, err := s.IsShortCodeAvailable(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("check custom shortcode: %w", err)
		}
		if !ok {
			return nil, ErrShortCodeTaken
		}
	}
	// 2. 写入布隆过滤器 --已测试
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
	if err := s.repo.CreateURL(ctx, u); err != nil {
		return nil, fmt.Errorf("create url record: %w", err)
	}
	// 6. 异步写缓存
	go func() {
		if err := s.cache.SetURL(ctx, *u); err != nil {
			// 此处记录日志而不中断流程
			log.Printf("failed to set cache: %v", err)
		}
	}()
	// 6. 返回给上层
	return &dto.CreateURLResponse{
		ShortUrl:  s.bashURL + "/" + u.ShortCode,
		ExpiredAt: u.ExpiredAt,
	}, nil
}

func (s *URLService) GetURL(ctx context.Context, shortCode string) (string, error) {
	// 访问缓存
	url, err := s.cache.GetURL(ctx, shortCode)
	if err != nil {
		return "", err
	}
	if url != nil {
		return url.OriginalURL, nil
	}
	// 缓存中不存在，访问数据库
	url, err = s.repo.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}
	if url == nil { // 不是查询出错，而是数据库没该数据
		return "", nil
	}
	// 存入缓存
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
	ok, err := s.IsShortCodeAvailable(ctx, code)
	if err != nil {
		return "", err
	}
	if ok {
		return code, nil
	}
	// 递归调用
	return s.getShortCode(ctx, n+1)
}

func (s *URLService) DeleteExpired(ctx context.Context) error {
	return s.repo.DeleteExpired(ctx)
}

// NEW: 将 IsShortCodeAvailable 从 repo 层提到 service, 与布隆过滤器的判断整合
// TODO 增加缓存的逻辑
// TODO 如果存在于数据库、但是过期了，实际上是可以生成的 —— 生成新链接，写入数据库覆盖之前的
func (s *URLService) IsShortCodeAvailable(ctx context.Context, shortCode string) (bool, error) {
	if !s.filter.Exists(shortCode) { // 布隆过滤器中不存在，则一定不存在。 --已测试
		return true, nil
	}
	// 布隆过滤器中存在，仍然可能不存在。判断是否在数据库中
	isInDB, err := s.repo.ExistsInDB(ctx, shortCode)
	if err != nil {
		return false, err
	}
	return !isInDB, nil
}
