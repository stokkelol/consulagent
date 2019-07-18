package consulagent

import (
	"sync"
)

type Credentials struct {
	mu sync.RWMutex

	ServiceName string
	Env         string
	Index       uint64
	List        []*Credential
}

type Credential struct {
	Name string
	Map  []*KV
}

type KV struct {
	Index uint64
	Key   string
	Value string
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
		c.Map = append(c.Map, &KV{Key: cr, Index: 0, Value: ""})
	}

	return c
}
func (kv *KV) GetKV() (string, string) {
	return kv.Key, kv.Value
}

func (kv *KV) GetKey() string {
	return kv.Key
}

func (kv *KV) GetValue() string {
	return kv.Value
}
