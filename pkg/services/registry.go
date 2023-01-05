package services

import "github.com/srand/go-init/pkg/config"

type Registry struct {
	Services []*Service
}

func (r *Registry) AddService(svc *Service) {
	r.Services = append(r.Services, svc)
}

func NewRegistry(config *config.ConfigFile) (*Registry, error) {
	registry := &Registry{}

	for _, svcConfig := range config.Services {
		svc, err := NewService(svcConfig)
		if err != nil {
			return nil, err
		}
		registry.AddService(svc)
	}

	return registry, nil
}
