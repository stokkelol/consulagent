package consulagent

import (
	"errors"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"os"
	"time"
)

var (
	errServiceName        = errors.New("service name is not provided")
	errServiceAddr        = errors.New("service address is not provided")
	errConsulAddr         = errors.New("consul address is not provided")
	errConfigNotValidated = errors.New("consul agent config has not been validated")
)

type Config struct {
	ServiceName   string
	ContainerPort int
	Address       string
	TTL           time.Duration
	Env           string
	ConsulAddress string
	AgentPort     int
	PassPhrase    string
	FailPhrase    string

	validated bool
}

type Agent struct {
	Config  *Config
	Agent   *consul.Agent
	Catalog *consul.Catalog
	KV      *consul.KV
	Client  *consul.Client
}

type CheckFunc func() bool

func (c *Config) Validate() error {
	if c.ServiceName == "" {
		return errServiceName
	}

	if c.ContainerPort == 0 {
		c.ContainerPort = 9000
	}

	if c.Address == "" {
		return errServiceAddr
	}

	if c.TTL == 0 {
		c.TTL = time.Duration(time.Second * 15)
	}

	if c.Env == "" {
		c.Env = "dev"
	}

	if c.ConsulAddress == "" {
		return errConsulAddr
	}

	if c.AgentPort == 0 {
		c.AgentPort = 8500
	}

	if c.PassPhrase == "" {
		c.PassPhrase = "Service alive and reachable."
	}

	if c.FailPhrase == "" {
		c.FailPhrase = "Service unreachable."
	}
	c.validated = true
	return nil
}

func NewAgent(config *Config) (*Agent, error) {
	if !config.validated {
		return nil, errConfigNotValidated
	}

	s := &Agent{Config: config}
	err := s.newClient()
	if err != nil {
		return nil, err
	}

	serviceDef := &consul.AgentServiceRegistration{
		Name: s.Config.ServiceName,
		Check: &consul.AgentServiceCheck{
			TTL: s.Config.TTL.String(),
		},
		Port:    s.Config.ContainerPort,
		Address: s.Config.Address,
		Tags:    []string{s.Config.Env},
	}

	if err := s.Agent.ServiceRegister(serviceDef); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Agent) UpdateTTL(check CheckFunc) {
	ticker := time.NewTicker(s.Config.TTL / 2)
	for range ticker.C {
		if err := s.update(check); err != nil {
			return
		}
	}
}

func (s *Agent) update(check CheckFunc) error {
	if !check() {
		if err := s.Agent.UpdateTTL("Service: "+s.Config.ServiceName, s.Config.FailPhrase, "fail"); err != nil {
			return err
		}
	}

	return s.Agent.UpdateTTL("Service: "+s.Config.ServiceName, s.Config.PassPhrase, "pass")
}

func (s *Agent) newClient() error {
	client, err := consul.NewClient(&consul.Config{
		Address: fmt.Sprintf("%s:%d", os.Getenv("CONSUL_SERVER"), s.Config.AgentPort),
	})
	if err != nil {
		return err
	}
	s.Client = client
	s.Catalog = client.Catalog()
	s.Agent = client.Agent()
	s.KV = client.KV()
	return nil
}
