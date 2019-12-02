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

	nameWithEnv = "%s-%s"
)

// Services list of all registered services
type Services struct {
	list      map[string]*Service
	m         sync.RWMutex
	catalog   *consul.Catalog
	populated bool
}

// NewServices returns new instance of Services object
func NewServices(catalog *consul.Catalog, services ...*Service) (*Services, error) {
	s := &Services{
		list:    make(map[string]*Service),
		catalog: catalog,
	}

	for _, serv := range services {
		if err := s.Add(serv); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Get returns service instance by name
func (s *Services) Get(name string) *Service {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.list[name]
}

// Add
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
	for _, entry := range s.list {
		entries, _, err := s.catalog.Service(entry.name, env, nil)
		if err != nil {
			return err
		}

		if err := s.updateService(entries, env, behindProxy); err != nil {
			return err
		}
	}

	s.populated = true
	return nil
}

func (s *Services) updateService(entries []*consul.CatalogService, env string, behindProxy bool) error {
	s.m.Lock()
	defer s.m.Unlock()
	for _, serv := range entries {
		if serv.ServiceTags[0] == env {
			if entry, ok := s.list[serv.ServiceName]; ok {
				if entry.index != serv.ModifyIndex {
					entry.address = serv.ServiceAddress
					entry.port = serv.ServicePort

					url, err := url.Parse(prepareHost(entry, behindProxy))
					if err != nil {
						return err
					}
					entry.index = serv.ModifyIndex
					entry.url = url
				}
			}
		}
	}

	return nil
}

func (s *Services) Update(env string, behindProxy bool) error {
	if !s.populated {
		return errors.New("services must be populated before updating")
	}

	for _, entry := range s.list {
		entries, _, err := s.catalog.Service(entry.name, env, nil)
		if err != nil {
			return err
		}
		if err := s.updateService(entries, env, behindProxy); err != nil {
			return err
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
	index   uint64
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

func (s *Service) HostString(protocol string) string {
	if s.url != nil {
		return fmt.Sprintf("%s://%s", protocol, s.url.Host)
	}

	return ""
}

func (s *Service) HostStringWithSuffix(protocol, suffix string) string {
	if s.url != nil {
		return fmt.Sprintf("%s://%s/%s/", protocol, s.url.Host, suffix)
	}

	return ""
}

func prepareHost(s *Service, behindProxy bool) string {
	if behindProxy {
		return fmt.Sprintf(withProxyFormat, s.address)
	}

	return fmt.Sprintf(withoutProxyFormat, s.address, s.port)
}

func PrepareServiceNameEnv(name, env string) string {
	return fmt.Sprintf(nameWithEnv, name, env)
}
