package consulagent

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestServiceCreate(t *testing.T) {
	u, err := url.Parse("http://127.0.0.1:9000")
	assert.Nil(t, err)
	s := &Service{
		path:    "/",
		name:    "Service",
		address: "127.0.0.1",
		port:    9000,
		url:     u,
	}

	assert.Equal(t, "http://127.0.0.1:9000", s.HostString("http"))
	assert.Equal(t, "http://127.0.0.1:9000/suffix/", s.HostStringWithSuffix("http", "suffix"))
}
