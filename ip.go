package consulagent

import (
	"errors"
	"io/ioutil"
	"net"
	"net/http"
)

const awsMagicUrl = "http://169.254.169.254/latest/meta-data/local-ipv4"

// GetPrivateIPV4 returns private ipv4 for EC2 instance
func GetPrivateIPV4() (net.IP, error) {
	client := http.Client{}

	req, err := http.NewRequest("GET", awsMagicUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parsed := net.ParseIP(string(body))

	if parsed == nil {
		return nil, errors.New("malformed IPV4")
	}

	return parsed, nil
}
