package reqwest

import (
	"io"
	"net/http"
	"time"
)

type Response struct {
	statusCode    int
	body          io.ReadCloser
	retryAttempts int
	totalDuration time.Duration
}

func fromHTTPResponse(resp *http.Response) *Response {
	return &Response{
		statusCode: resp.StatusCode,
		body:       resp.Body,
	}
}

func (r *Response) StatusCode() int {
	return r.statusCode
}

func (r *Response) Body() io.ReadCloser {
	return r.body
}

func (r *Response) RetryAttempts() int {
	return r.retryAttempts
}

func (r *Response) TotalDuration() time.Duration {
	return r.totalDuration
}
