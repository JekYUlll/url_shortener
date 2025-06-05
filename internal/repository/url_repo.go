package repository

import (
	"context"
	"errors"

	"github.com/jekyulll/url_shortener/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type URLRepository interface {
	CreateURL(ctx context.Context, url *model.URL) error
	GetURLByShortCode(ctx context.Context, shortCode string) (*model.URL, error)
	UpdateURL(ctx context.Context, url *model.URL) error
	UpsertURL(ctx context.Context, url *model.URL) error
	DeleteURLByID(ctx context.Context, id uint) error

	GetAllURLs(ctx context.Context) ([]model.URL, error)
	GetAllActiveURLs(ctx context.Context) ([]model.URL, error)

	DeleteAllExpired(ctx context.Context) error
}

type gormURLRepositoryImpl struct {
	db *gorm.DB
}

func New(db *gorm.DB) *gormURLRepositoryImpl {
	return &gormURLRepositoryImpl{
		db: db,
	}
}

func (r *gormURLRepositoryImpl) CreateURL(ctx context.Context, url *model.URL) error {
	if err := r.db.Create(url).Error; err != nil {
		return err
	}
	return nil
}

// WARNING 找不到的时候不会返回 error，而是直接返回空指针
// 过期的时候仍然会返回，由外部判断
func (r *gormURLRepositoryImpl) GetURLByShortCode(ctx context.Context, code string) (*model.URL, error) {
	var url model.URL
	err := r.db.Where("short_code = ?", code).
		First(&url).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &url, err
}

func (r *gormURLRepositoryImpl) GetAllURLs(ctx context.Context) ([]model.URL, error) {
	var urls []model.URL
	err := r.db.Find(&urls).Error
	return urls, err
}

func (r *gormURLRepositoryImpl) GetAllActiveURLs(ctx context.Context) ([]model.URL, error) {
	var urls []model.URL
	err := r.db.WithContext(ctx).
		Where("expired_at > NOW()").
		Find(&urls).
		Error
	return urls, err
}

// Deprecated:
// gorm 的 Save 方法会根据主键值判断执行插入或更新
// 然而此处 short_code 并不是主键
func (r *gormURLRepositoryImpl) UpdateURL(ctx context.Context, url *model.URL) error {
	return r.db.Save(url).Error
}

// 根据 short_code 是否存在执行插入或者更新
func (r *gormURLRepositoryImpl) UpsertURL(ctx context.Context, url *model.URL) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "short_code"}},
		DoUpdates: clause.AssignmentColumns([]string{"original_url", "expired_at", "is_custom"}),
	}).Create(url).Error
}

func (r *gormURLRepositoryImpl) DeleteURLByID(ctx context.Context, id uint) error {
	return r.db.Delete(&model.URL{}, id).Error
}

func (r *gormURLRepositoryImpl) DeleteAllExpired(ctx context.Context) error {
	return r.db.Where("expired_at < NOW()").Delete(&model.URL{}).Error
}

// TODO UpdateOriginalURL、ListRecent

func (r *gormURLRepositoryImpl) ExistsInDB(ctx context.Context, shortCode string) (bool, error) {
	var cnt int64
	if err := r.db.
		Model(&model.URL{}).
		Where("short_code = ?", shortCode).
		Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}
