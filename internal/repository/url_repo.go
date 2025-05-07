package repository

import (
	"context"

	"github.com/jekyulll/url_shortener/internal/model"
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

type uRLRepositoryImpl struct {
	db *gorm.DB
}

func NewURLRepository(db *gorm.DB) *uRLRepositoryImpl {
	return &uRLRepositoryImpl{db: db}
}

func (r *uRLRepositoryImpl) CreateURL(ctx context.Context, url *model.URL) error {
	return r.db.Create(url).Error
}

func (r *uRLRepositoryImpl) GetURLByShortCode(ctx context.Context, code string) (*model.URL, error) {
	var url model.URL
	err := r.db.Where("short_code = ? AND (expired_at IS NULL OR expired_at > ?)", code).
		First(&url).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 找不到或已经过期
			return nil, nil
		}
		return nil, err
	}
	return &url, err
}

// TODO 过期时间
func (r *uRLRepositoryImpl) GetAllURLs(ctx context.Context) ([]model.URL, error) {
	var urls []model.URL
	err := r.db.Find(&urls).Error
	return urls, err
}

func (r *uRLRepositoryImpl) UpdateURL(ctx context.Context, url *model.URL) error {
	// gorm 的 Save 方法会根据主键值判断执行插入或更新
	return r.db.Save(url).Error
}

func (r *uRLRepositoryImpl) DeleteURLByID(ctx context.Context, id uint) error {
	return r.db.Delete(&model.URL{}, id).Error
}

func (r *uRLRepositoryImpl) DeleteExpired(ctx context.Context) error {
	return r.db.Where("expired_at < NOW()").Delete(&model.URL{}).Error
}

// TODO UpdateOriginalURL、ListRecent

func (r *uRLRepositoryImpl) IsShortCodeAvailable(ctx context.Context, code string) (bool, error) {
	var cnt int64
	// 统计有多少行和 short_code 匹配
	if err := r.db.
		Model(&model.URL{}).
		Where("short_code = ?", code).
		Count(&cnt).Error; err != nil {
		return false, err
	}
	// count == 0 表示不存在，短码可用
	return cnt == 0, nil
}
