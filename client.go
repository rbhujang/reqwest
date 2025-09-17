package reqwest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultClientTimeout = 10 * time.Second
)

type Client interface {
	Get(ctx context.Context, url string) (*Response, error)
	Post(ctx context.Context, url string, body []byte) (*Response, error)
}

type client struct {
	baseURL     string
	httpClient  *http.Client
	middlewares []Middleware
}

func (c *client) Get(ctx context.Context, url string) (*Response, error) {
	return c.execute(ctx, url, http.MethodGet, nil)
}

func (c *client) Post(ctx context.Context, url string, body []byte) (*Response, error) {
	return c.execute(ctx, url, http.MethodPost, bytes.NewBuffer(body))
}

func (c *client) execute(ctx context.Context, url, method string, body io.Reader) (*Response, error) {
	fullURL := c.buildURL(url)
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to make http request: %v", err)
	}
	for _, middleware := range c.middlewares {
		if err := middleware(req); err != nil {
			return nil, fmt.Errorf("middleware error: %v", err)
		}
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do http request: %v", err)
	}

	return fromHTTPResponse(resp), nil
}

func (c *client) buildURL(url string) string {
	if c.baseURL == "" {
		return url
	}

	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		return url
	}

	return c.baseURL + "/" + strings.TrimLeft(url, "/")
}
