package consulagent

import (
	"errors"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"net/url"
	"sync"
)

const (
	withProxyFormat    = "http://%s"
	withoutProxyFormat = "http://%s:%d"
)

type Services struct {
	list      map[string]*Service
	m         sync.RWMutex
	agent     *consul.Agent
	populated bool
}

func NewServices(agent *consul.Agent, services ...*Service) (*Services, error) {
	s := &Services{
		list:  make(map[string]*Service),
		agent: agent,
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

func (s *Services) Parse(env string, behindProxy bool) error {
	services, err := s.agent.Services()
	if err != nil {
		return err
	}
	s.m.Lock()
	defer s.m.Unlock()
	for _, serv := range services {
		if serv.Tags[0] == env {
			if entry, ok := s.list[serv.ID]; ok {
				entry.address = serv.Address
				entry.port = serv.Port

				url, err := url.Parse(prepareHost(entry, behindProxy))
				if err != nil {
					return err
				}

				entry.url = url
			}
		}
	}
	s.populated = true
	return nil
}

func (s *Services) Update(env string, behindProxy bool) error {
	if !s.populated {
		return errors.New("services must be populated before updating")
	}
	services, err := s.agent.Services()
	if err != nil {
		return err
	}
	s.m.Lock()
	defer s.m.Unlock()
	for _, serv := range services {
		if entry, ok := s.list[serv.ID]; ok {
			if entry.address != serv.Address || entry.port != serv.Port {
				entry.address = serv.Address
				entry.port = serv.Port

				url, err := url.Parse(prepareHost(entry, behindProxy))
				if err != nil {
					return err
				}

				entry.url = url
			}
		}
	}

	return nil
}

type Service struct {
	path    string
	name    string
	address string
	port    int
	url     *url.URL
}

func NewService(name, path string) *Service {
	return &Service{
		name: name,
		path: path,
	}
}

func (s *Service) Path() string {
	return s.path
}

func (s *Service) Host() string {
	return s.url.Host
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

func prepareHost(s *Service, behindProxy bool) string {
	if behindProxy {
		return fmt.Sprintf(withProxyFormat, s.address)
	}

	return fmt.Sprintf(withoutProxyFormat, s.address, s.port)
}
