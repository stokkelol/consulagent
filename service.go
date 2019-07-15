package consulagent

import (
	"fmt"
	"net/url"
	"sync"
)

const (
	hostFormat = "http://%s:%d"
)

type Services struct {
	list map[string]*Service
	m    sync.RWMutex
}

func NewServices(services ...*Service) (*Services, error) {
	s := &Services{
		list: make(map[string]*Service),
	}

	for _, serv := range services {
		if err := s.Add(serv); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Services) Get(name string) *Service {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.list[name]
}

func (s *Services) Add(service *Service) error {
	if s.Has(service.name) {
		return nil
	}

	s.m.Lock()
	defer s.m.Unlock()
	s.list[service.name] = service
	return nil
}

func (s *Services) Delete(name string) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.list, name)
}

func (s *Services) Has(name string) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	if _, ok := s.list[name]; ok {
		return true
	}

	return false
}

type Service struct {
	path    string
	name    string
	address string
	port    int
	url     *url.URL
}

func NewService(name, address string, port int) (*Service, error) {
	u, err := url.Parse(fmt.Sprintf(hostFormat, address, port))
	if err != nil {
		return nil, err
	}

	return &Service{
		address: address,
		port:    port,
		name:    name,
		url:     u,
	}, nil
}

func (s *Service) Path() string {
	return s.path
}

func (s *Service) Host() string {
	return fmt.Sprintf(hostFormat, s.address, s.port)
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) Address() string {
	return s.address
}

func (s *Service) Port() int {
	return s.port
}

func (s *Service) Url() *url.URL {
	return s.url
}
