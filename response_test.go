package reqwest

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFromHTTPResponse(t *testing.T) {
	t.Run("Create response from http.Response", func(t *testing.T) {
		body := io.NopCloser(strings.NewReader("test body"))
		httpResp := &http.Response{
			StatusCode: 200,
			Body:       body,
		}

		resp := fromHTTPResponse(httpResp)

		if resp == nil {
			t.Fatal("fromHTTPResponse returned nil")
		}

		if resp.statusCode != 200 {
			t.Errorf("Expected status code 200, got %d", resp.statusCode)
		}

		if resp.body != body {
			t.Error("Expected body to match original http.Response body")
		}
	})

	t.Run("Create response with different status codes", func(t *testing.T) {
		testCases := []int{200, 201, 400, 404, 500}

		for _, statusCode := range testCases {
			httpResp := &http.Response{
				StatusCode: statusCode,
				Body:       io.NopCloser(strings.NewReader("")),
			}

			resp := fromHTTPResponse(httpResp)

			if resp.statusCode != statusCode {
				t.Errorf("Expected status code %d, got %d", statusCode, resp.statusCode)
			}
		}
	})
}

func TestResponse_Integration(t *testing.T) {
	t.Run("Full response lifecycle", func(t *testing.T) {
		expectedStatusCode := 201
		expectedBody := "Created successfully"

		httpResp := &http.Response{
			StatusCode: expectedStatusCode,
			Body:       io.NopCloser(strings.NewReader(expectedBody)),
		}

		resp := fromHTTPResponse(httpResp)

		// Test status code
		if resp.StatusCode() != expectedStatusCode {
			t.Errorf("Expected status code %d, got %d", expectedStatusCode, resp.StatusCode())
		}

		// Test body reading
		body := resp.Body()
		defer body.Close()

		content, err := io.ReadAll(body)
		if err != nil {
			t.Fatalf("Failed to read body: %v", err)
		}

		if string(content) != expectedBody {
			t.Errorf("Expected body %q, got %q", expectedBody, string(content))
		}
	})
}
