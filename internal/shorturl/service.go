package shorturl

type Service interface {
	CreateShortURL(longURL string) (string, error)
	GetOriginalURL(code string) (string, error)
}

type serviceImpl struct {
	repository Repository
}

func NewService(repo Repository) *serviceImpl {
	return &serviceImpl{repository: repo}
}

func (s *serviceImpl) CreateShortURL(longURL string) (string, error) {
	url := "url"
	return url, nil
}

func (s *serviceImpl) GetOriginalURL(code string) (string, error) {
	url, err := s.repository.GetURLByCode(code)
	if err != nil {
		return "", err
	}
	return url.OriginalURL, nil
}
