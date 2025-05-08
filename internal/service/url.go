package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jekyulll/url_shortener/internal/dto"
	"github.com/jekyulll/url_shortener/internal/model"
	"github.com/jekyulll/url_shortener/internal/repository"
	"gorm.io/gorm"
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
	shortCodeGenerator ShortCodeGenerator
	defaultDuration    time.Duration
	cache              Cacher
	bashURL            string
}

func NewURLService(db *gorm.DB, generator ShortCodeGenerator, duration time.Duration, cache Cacher, baseURL string) *URLService {
	return &URLService{
		repo:               repository.NewURLRepository(db),
		shortCodeGenerator: generator,
		defaultDuration:    duration,
		cache:              cache,
		bashURL:            baseURL,
	}
}

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
		ok, err := s.repo.IsShortCodeAvailable(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("check custom shortcode: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("custom shortcode %q already taken", code)
		}
	}
	// 2. 组装要入库的 URL 实体
	u := &model.URL{
		OriginalURL: req.OriginalURL,
		ShortCode:   code,
		IsCustom:    req.CustomeCode != "",
	}
	// 3. 处理过期时间：不传的话使用默认有效期
	if req.Duration == nil {
		u.ExpiredAt = time.Now().Add(s.defaultDuration)
	} else {
		u.ExpiredAt = time.Now().Add(time.Duration(*req.Duration) * time.Hour)
	}
	// 4. 写库
	if err := s.repo.CreateURL(ctx, u); err != nil {
		return nil, fmt.Errorf("create url record: %w", err)
	}
	// 5.  写缓存
	if err := s.cache.SetURL(ctx, *u); err != nil {
		return nil, err
	}
	// 6. 返回给上层
	//    假设短链的域名前缀是 https://short.ly/
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
	if url != nil {
		// 存入缓存
		if err := s.cache.SetURL(ctx, *url); err != nil {
			return url.OriginalURL, err
		}
	}
	return url.OriginalURL, nil
}

// getShortCode
// @pragma n:重试次数
func (s *URLService) getShortCode(ctx context.Context, n int) (string, error) {
	if n > 5 {
		return "", errors.New("retry too many times")
	}
	code := s.shortCodeGenerator.GenerateShortCode()
	ok, err := s.repo.IsShortCodeAvailable(ctx, code)
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
