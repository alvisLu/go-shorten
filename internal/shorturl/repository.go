package repository

import (
	"github.com/alvisLu/go-shorten/internal/model"
	"gorm.io/gorm"
)

type URLRepository interface {
	Create(url *model.URL) error
	GetURLByCode(code string) (*model.URL, error)
}

type URLRepositoryImpl struct {
	db *gorm.DB
}

func NewURLRepository(db *gorm.DB) *URLRepositoryImpl {
	return &URLRepositoryImpl{db: db}
}

func (r *URLRepositoryImpl) Create(url *model.URL) error {
	return r.db.Create(url).Error
}

func (r *URLRepositoryImpl) GetURLByCode(code string) (*model.URL, error) {
	var url model.URL
	if err := r.db.Where("code = ?", code).First(&url).Error; err != nil {
		return nil, err
	}
	return &url, nil
}
