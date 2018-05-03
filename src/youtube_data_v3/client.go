package youtube

import (
	"net/http"
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type impleClient struct {
	client *http.Client
}

func (c *impleClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
