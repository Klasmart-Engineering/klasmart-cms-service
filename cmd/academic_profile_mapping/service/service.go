package service

import "sync"

type Service interface {
	Do(mapper Mapper) error
}

var (
	serviceOnce sync.Once
	services    = []Service{}
)

func GetServices() []Service {
	serviceOnce.Do(func() {
		services = []Service{}
	})

	return services
}
