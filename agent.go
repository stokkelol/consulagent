package consulagent

import (
	"errors"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"strings"
	"time"
)

const (
	keySeparator = "/"
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
	config  *Config
	agent   *consul.Agent
	catalog *consul.Catalog
	kv      *consul.KV
	client  *consul.Client
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

	s := &Agent{config: config}
	err := s.newClient()
	if err != nil {
		return nil, err
	}

	serviceDef := &consul.AgentServiceRegistration{
		Name: s.config.ServiceName,
		Check: &consul.AgentServiceCheck{
			TTL: s.config.TTL.String(),
		},
		Port:    s.config.ContainerPort,
		Address: s.config.Address,
		Tags:    []string{s.config.Env},
		ID:      s.config.ServiceName,
	}

	if err := s.agent.ServiceRegister(serviceDef); err != nil {
		return nil, err
	}

	return s, nil
}

func (a *Agent) KV() *consul.KV {
	return a.kv
}

func (a *Agent) Client() *consul.Client {
	return a.client
}

func (a *Agent) Agent() *consul.Agent {
	return a.agent
}

func (a *Agent) Catalog() *consul.Catalog {
	return a.catalog
}

func (a *Agent) UpdateTTL(check CheckFunc) {
	ticker := time.NewTicker(a.config.TTL / 2)
	for range ticker.C {
		if err := a.update(check); err != nil {
			return
		}
	}
}

func (a *Agent) LoadKV() (uint64, consul.KVPairs, error) {
	kvPairs, meta, err := a.kv.List(a.formatPrefix(), nil)
	if err != nil {
		return 0, kvPairs, err
	}

	return meta.LastIndex, kvPairs, nil
}

func (a *Agent) IterateKV(kvPairs consul.KVPairs, creds map[string]string) {
	for _, kv := range kvPairs {
		k := a.replaceKey(kv.Key)
		if _, ok := creds[k]; ok {
			creds[k] = string(kv.Value)
		}
	}
}

func (a *Agent) update(check CheckFunc) error {
	if !check() {
		if err := a.agent.UpdateTTL(a.formatCheckID(), a.config.FailPhrase, "fail"); err != nil {
			return err
		}
	}

	return a.agent.UpdateTTL(a.formatCheckID(), a.config.PassPhrase, "pass")
}

func (a *Agent) newClient() error {
	client, err := consul.NewClient(&consul.Config{
		Address: fmt.Sprintf("%s:%d", a.config.ConsulAddress, a.config.AgentPort),
	})
	if err != nil {
		return err
	}
	a.client = client
	a.catalog = client.Catalog()
	a.agent = client.Agent()
	a.kv = client.KV()
	return nil
}

func (a *Agent) formatCheckID() string {
	return fmt.Sprintf("service:%s", a.config.ServiceName)
}

func (a *Agent) formatPrefix() string {
	return fmt.Sprintf("%s/%s/", a.config.ServiceName, a.config.Env)
}

func (a *Agent) formatCredential(cred string) string {
	return fmt.Sprintf("%s/%s/%s", a.config.ServiceName, a.config.Env, cred)
}

func (a *Agent) replaceKey(key string) string {
	parts := strings.Split(key, keySeparator)

	return parts[len(parts)-1]
}
