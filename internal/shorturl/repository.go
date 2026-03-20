package shorturl

import "gorm.io/gorm"

type Repository interface {
	Create(url *URL) error
	GetURLByCode(code string) (*URL, error)
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repositoryImpl {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) Create(url *URL) error {
	return r.db.Create(url).Error
}

func (r *repositoryImpl) GetURLByCode(code string) (*URL, error) {
	var url URL
	if err := r.db.Where("code = ?", code).First(&url).Error; err != nil {
		return nil, err
	}
	return &url, nil
}
