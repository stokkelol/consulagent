package consulagent

import "sync"

type Credentials struct {
	mu sync.RWMutex

	ServiceName string
	Env         string
	Index       uint64
	List        []*Credential
}

type Credential struct {
	Name string
	Map  []KV
}

type KV struct {
	Index    uint64
	KeyValue map[string]string
}

func NewCredentials(name, env string, creds ...*Credential) *Credentials {
	c := &Credentials{
		ServiceName: name,
		Env:         env,
	}

	for _, cr := range creds {
		c.List = append(c.List, cr)
	}

	return c
}

func NewCredential(name string, creds ...string) *Credential {
	c := &Credential{
		Name: name,
	}

	for _, cr := range creds {
		m := map[string]string{cr: ""}
		c.Map = append(c.Map, KV{KeyValue: m, Index: 0})
	}

	return c
}
