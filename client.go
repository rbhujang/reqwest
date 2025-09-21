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
	retryConfig *RetryConfig
}

func (c *client) Get(ctx context.Context, url string) (*Response, error) {
	return c.execute(ctx, url, http.MethodGet, nil)
}

func (c *client) Post(ctx context.Context, url string, body []byte) (*Response, error) {
	return c.execute(ctx, url, http.MethodPost, bytes.NewBuffer(body))
}

func (c *client) execute(
	ctx context.Context,
	url,
	method string,
	body io.Reader) (*Response, error) {
	startTime := time.Now()
	var lastErr error
	var resp *Response

	// Cache body content for retries
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %v", err)
		}
	}
	maxAttempts := 1
	if c.retryConfig != nil {
		maxAttempts = c.retryConfig.maxRetries + 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := contextCancelled(ctx); err != nil {
			return nil, err
		}
		if err := c.applyBackoff(ctx, attempt); err != nil {
			return nil, err
		}
		resp, lastErr = c.executeOnce(ctx, url, method, bodyReaderFromByteSlice(bodyBytes))
		// If successful and no retry needed, return immediately
		if lastErr == nil && !c.shouldRetry(resp) {
			resp.retryAttempts = attempt
			resp.totalDuration = time.Since(startTime)
			return resp, nil
		}

		// Check if we should retry this error/response
		if attempt < maxAttempts-1 && (c.shouldRetryError(lastErr) || c.shouldRetry(resp)) {
			continue
		}

		// No more retries or not retryable, break out of loop
		break
	}

	// Return the final result (could be success or failure)
	if resp != nil {
		resp.retryAttempts = maxAttempts - 1
		resp.totalDuration = time.Since(startTime)
		return resp, lastErr
	}

	return nil, lastErr
}

func (c *client) executeOnce(ctx context.Context, url, method string, body io.Reader) (*Response, error) {
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

func (c *client) shouldRetry(resp *Response) bool {
	if c.retryConfig == nil || resp == nil {
		return false
	}

	statusCode := resp.StatusCode()
	for _, code := range c.retryConfig.retryableStatusCodes {
		if statusCode == code {
			return true
		}
	}

	return false
}

func (c *client) shouldRetryError(err error) bool {
	if c.retryConfig == nil || err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	for retryableErr := range c.retryConfig.retryableError {
		if strings.Contains(errStr, retryableErr) {
			return true
		}
	}

	return false
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

func (c *client) applyBackoff(ctx context.Context, attempt int) error {
	// Apply backoff delay for retry attempts (skip on first attempt)
	if attempt > 0 && c.retryConfig != nil {
		delay := c.retryConfig.backoffStrategy.Delay(attempt)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func contextCancelled(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func bodyReaderFromByteSlice(content []byte) io.Reader {
	var reader io.Reader
	if len(content) > 0 {
		reader = bytes.NewBuffer(content)
	}
	return reader
}
