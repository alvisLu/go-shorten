package health

type Service interface {
	Health() string
}

type serviceImpl struct{}

func NewService() *serviceImpl {
	return &serviceImpl{}
}

func (s *serviceImpl) Health() string {
	return "Hello, World!"
}
