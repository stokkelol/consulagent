package consulagent

import (
	"io/ioutil"
	"net/http"
)

const awsMagicUrl = "http://169.254.169.254/latest/meta-data/local-ipv4"

// GetPrivateIPV4 returns private ipv4 for EC2 instance
func GetPrivateIPV4() (string, error) {
	client := http.Client{}

	req, err := http.NewRequest("GET", awsMagicUrl, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
