package reqwest

import (
	"net/http"
	"strings"
)

type ClientBuilder struct {
	baseURL     string
	middlewares []Middleware
	retryConfig *RetryConfig
}

func NewClientBuilder() *ClientBuilder {
	return &ClientBuilder{
		middlewares: make([]Middleware, 0),
	}
}

func (cb *ClientBuilder) WithBaseURL(url string) *ClientBuilder {
	cb.baseURL = strings.TrimRight(url, "/")
	return cb
}

func (cb *ClientBuilder) WithMiddleware(middleware Middleware) *ClientBuilder {
	cb.middlewares = append(cb.middlewares, middleware)
	return cb
}

func (cb *ClientBuilder) WithRetryConfig(config *RetryConfig) *ClientBuilder {
	cb.retryConfig = config
	return cb
}

func (cb *ClientBuilder) WithRetries() *ClientBuilder {
	cb.retryConfig = NewRetryConfigBuilder().Build()
	return cb
}

func (cb *ClientBuilder) Build() Client {
	c := &client{
		httpClient:  http.DefaultClient,
		middlewares: make([]Middleware, len(cb.middlewares)),
		retryConfig: cb.retryConfig,
	}
	if cb.baseURL != "" {
		c.baseURL = cb.baseURL
	}
	copy(c.middlewares, cb.middlewares)
	return c
}
