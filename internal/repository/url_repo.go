package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jekyulll/url_shortener/internal/model"
	"github.com/jekyulll/url_shortener/pkg/filter"
	"gorm.io/gorm"
)

type URLRepository interface {
	CreateURL(ctx context.Context, url *model.URL) error
	GetURLByShortCode(ctx context.Context, shortCode string) (*model.URL, error)
	GetAllURLs(ctx context.Context) ([]model.URL, error)
	UpdateURL(ctx context.Context, url *model.URL) error
	DeleteURLByID(ctx context.Context, id uint) error
	DeleteExpired(ctx context.Context) error
	IsShortCodeAvailable(ctx context.Context, code string) (bool, error)
}

type gormURLRepository struct {
	db     *gorm.DB
	filter *filter.BloomFilterImpl
}

func NewURLRepository(db *gorm.DB, filter *filter.BloomFilterImpl) *gormURLRepository {
	return &gormURLRepository{
		db:     db,
		filter: filter,
	}
}

func (r *gormURLRepository) CreateURL(ctx context.Context, url *model.URL) error {
	if err := r.db.Create(url).Error; err != nil {
		return err
	}
	// 加入布隆过滤器
	r.filter.Add(url.ShortCode)
	return nil
}

// WARNING 找不到或已经过期的时候不会返回 error，而是直接返回空指针
func (r *gormURLRepository) GetURLByShortCode(ctx context.Context, code string) (*model.URL, error) {
	var url model.URL
	err := r.db.Where("short_code = ? AND (expired_at IS NULL OR expired_at > ?)", code, time.Now()).
		First(&url).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &url, err
}

// TODO 过期时间
// 如果短链接只是用于临时分享，如一次性的文件分享、临时活动通知等，在规定的时间内完成分享和访问后，就没有必要延长其过期时间

func (r *gormURLRepository) GetAllURLs(ctx context.Context) ([]model.URL, error) {
	var urls []model.URL
	err := r.db.Find(&urls).Error
	return urls, err
}

func (r *gormURLRepository) UpdateURL(ctx context.Context, url *model.URL) error {
	// gorm 的 Save 方法会根据主键值判断执行插入或更新
	return r.db.Save(url).Error
}

func (r *gormURLRepository) DeleteURLByID(ctx context.Context, id uint) error {
	return r.db.Delete(&model.URL{}, id).Error
}

func (r *gormURLRepository) DeleteExpired(ctx context.Context) error {
	return r.db.Where("expired_at < NOW()").Delete(&model.URL{}).Error
}

// TODO UpdateOriginalURL、ListRecent

func (r *gormURLRepository) IsShortCodeAvailable(ctx context.Context, shortCode string) (bool, error) {
	if r.filter.Exists(shortCode) {
		return false, nil
	}
	// 过滤器里没有，查数据库
	var cnt int64
	// 统计有多少行和 short_code 匹配，0 行才合法
	if err := r.db.
		Model(&model.URL{}).
		Where("short_code = ?", shortCode).
		Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt == 0, nil
}

// // 重建过滤器
// func (r *gormURLRepository) RebuildBloomFilter(ctx context.Context) error {
// 	newFilter := filter.NewBloomFilter(r.filter.Capacity, r.filter.ErrorRate)
// 	err := r.loadExitingShortCode(ctx, newFilter)
// 	{
// 		r.filter.mu.Lock()
// 		r.filter = newFilter
// 		r.filter.mu.Unlock()
// 	}
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // 把数据库所有 short code 加载到 bloom
// func (r *gormURLRepository) loadExitingShortCode(ctx context.Context, filter *filter.BloomFilter) error {
// 	urls, err := r.GetAllURLs(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	for _, url := range urls {
// 		filter.Add(url.ShortCode)
// 	}
// 	return nil
// }
