package service

import (
	"github.com/alvisLu/go-shorten/internal/repository"
)

type URLService interface {
	CreateShortURL(longURL string) (string, error)
	GetOriginalURL(shortURL string) (string, error)
}

type URLServiceImpl struct {
	repository repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLServiceImpl {
	return &URLServiceImpl{repository: repo}
}

func (s *URLServiceImpl) CreateShortURL(longURL string) (string, error) {
	url := "url"
	return url, nil
}

func (s *URLServiceImpl) GetOriginalURL(code string) (string, error) {

	url, err := s.repository.GetURLByCode(code)
	if err != nil {
		return "", err
	}

	return url.OriginalURL, nil
}
