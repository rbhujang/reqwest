package reqwest

import (
	"io"
	"net/http"
)

type Response struct {
	statusCode int
	body       io.ReadCloser
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
