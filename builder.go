package reqwest

import (
	"net/http"
	"strings"
)

type ClientBuilder struct {
	baseURL string
}

func NewClientBuilder() *ClientBuilder {
	return &ClientBuilder{}
}

func (cb *ClientBuilder) WithBaseURL(url string) *ClientBuilder {
	cb.baseURL = strings.TrimRight(url, "/")
	return cb
}

func (cb *ClientBuilder) Build() Client {
	c := &client{
		httpClient: http.DefaultClient,
	}
	if cb.baseURL != "" {
		c.baseURL = cb.baseURL
	}

	return c
}
