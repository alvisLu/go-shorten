package service

type Service interface {
	Health() string
}

type ServiceImpl struct {
}

func NewService() *ServiceImpl {
	return &ServiceImpl{}
;}

func (s *ServiceImpl) Health() string {
	return "Hello, World!"
}
