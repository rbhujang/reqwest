package reqwest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_Get(t *testing.T) {
	t.Run("Successful GET request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Hello, World!"))
		}))
		defer server.Close()

		client := NewClientBuilder().Build()
		resp, err := client.Get(context.TODO(), server.URL)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if resp.StatusCode() != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode())
		}

		body, err := io.ReadAll(resp.Body())
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		defer resp.Body().Close()

		if string(body) != "Hello, World!" {
			t.Errorf("Expected body 'Hello, World!', got '%s'", string(body))
		}
	})

	t.Run("GET request with error", func(t *testing.T) {
		client := NewClientBuilder().Build()
		_, err := client.Get(context.TODO(), "http://localhost:1")

		if err == nil {
			t.Error("Expected error for unreachable URL, got nil")
		}
	})
}

func TestClient_Post(t *testing.T) {
	t.Run("Successful POST request", func(t *testing.T) {
		expectedBody := "test payload"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST method, got %s", r.Method)
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("Failed to read request body: %v", err)
			}

			if string(body) != expectedBody {
				t.Errorf("Expected request body '%s', got '%s'", expectedBody, string(body))
			}

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("Created"))
		}))
		defer server.Close()

		client := NewClientBuilder().Build()
		resp, err := client.Post(context.TODO(), server.URL, []byte(expectedBody))

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if resp.StatusCode() != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode())
		}

		responseBody, err := io.ReadAll(resp.Body())
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		defer resp.Body().Close()

		if string(responseBody) != "Created" {
			t.Errorf("Expected response body 'Created', got '%s'", string(responseBody))
		}
	})

	t.Run("POST request with empty body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("Failed to read request body: %v", err)
			}

			if len(body) != 0 {
				t.Errorf("Expected empty body, got '%s'", string(body))
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClientBuilder().Build()
		resp, err := client.Post(context.TODO(), server.URL, []byte{})

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if resp.StatusCode() != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode())
		}
	})
}

func TestClient_BuildURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		url      string
		expected string
	}{
		{
			name:     "No base URL with absolute URL",
			baseURL:  "",
			url:      "https://api.example.com/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "Base URL with relative path",
			baseURL:  "https://api.example.com",
			url:      "/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "Base URL with relative path without leading slash",
			baseURL:  "https://api.example.com",
			url:      "users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "Base URL ignored for absolute URL",
			baseURL:  "https://api.example.com",
			url:      "https://other.example.com/data",
			expected: "https://other.example.com/data",
		},
		{
			name:     "Base URL ignored for HTTP URL",
			baseURL:  "https://api.example.com",
			url:      "http://other.example.com/data",
			expected: "http://other.example.com/data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c *client
			if tt.baseURL != "" {
				c = NewClientBuilder().WithBaseURL(tt.baseURL).Build().(*client)
			} else {
				c = NewClientBuilder().Build().(*client)
			}

			result := c.buildURL(tt.url)
			if result != tt.expected {
				t.Errorf("buildURL(%q) with baseURL %q = %q, want %q", tt.url, tt.baseURL, result, tt.expected)
			}
		})
	}
}

func TestClient_WithBaseURL(t *testing.T) {
	t.Run("Base URL functionality", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/users" {
				t.Errorf("Expected path '/api/users', got '%s'", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClientBuilder().
			WithBaseURL(server.URL + "/api").
			Build()

		resp, err := client.Get(context.TODO(), "/users")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		defer resp.Body().Close()

		if resp.StatusCode() != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode())
		}
	})
}

func TestClient_Timeout(t *testing.T) {
	t.Run("Request should timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 + time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClientBuilder().Build()
		ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Millisecond)
		defer cancel()
		_, err := client.Get(ctx, server.URL)

		if err == nil {
			t.Error("Expected timeout error, got nil")
		}

		if !strings.Contains(err.Error(), "context deadline exceeded") {
			t.Errorf("Expected timeout error, got: %v", err)
		}
	})
}

func TestClient_ContextCancel(t *testing.T) {
	t.Run("Should cancel", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClientBuilder().Build()

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		_, err := client.Get(ctx, server.URL)

		if err == nil {
			t.Error("Expected cancellation error, got nil")
		}

		if !strings.Contains(err.Error(), "context canceled") {
			t.Errorf("Expected context canceled error, got: %v", err)
		}
	})
}
func TestClient_NoTimeout(t *testing.T) {
	t.Run("Request without timeout should succeed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond) // Short delay
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClientBuilder().Build()

		// No timeout set
		_, err := client.Get(context.Background(), server.URL)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestClient_RetryOnStatusCode(t *testing.T) {
	t.Run("retries on 500 status code", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Success"))
		}))
		defer server.Close()

		retryConfig := NewRetryConfigBuilder().
			WithMaxRetries(3).
			WithBackoffStrategy(NewFixedBackoffBuilder().
				WithDelay(10 * time.Millisecond).
				WithJitter(false).
				Build()).
			Build()

		client := NewClientBuilder().
			WithRetryConfig(retryConfig).
			Build()

		resp, err := client.Get(context.TODO(), server.URL)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if resp.StatusCode() != http.StatusOK {
			t.Errorf("Expected final status code %d, got %d", http.StatusOK, resp.StatusCode())
		}

		if callCount != 3 {
			t.Errorf("Expected 3 calls (2 retries + 1 success), got %d", callCount)
		}

		if resp.RetryAttempts() != 2 {
			t.Errorf("Expected 2 retry attempts, got %d", resp.RetryAttempts())
		}

		if resp.TotalDuration() <= 0 {
			t.Error("Expected positive total duration")
		}
	})
}

func TestClient_RetryExhaustion(t *testing.T) {
	t.Run("stops after max retries", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		retryConfig := NewRetryConfigBuilder().
			WithMaxRetries(2).
			WithBackoffStrategy(NewFixedBackoffBuilder().
				WithDelay(10 * time.Millisecond).
				WithJitter(false).
				Build()).
			Build()

		client := NewClientBuilder().
			WithRetryConfig(retryConfig).
			Build()

		resp, err := client.Get(context.TODO(), server.URL)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if resp.StatusCode() != http.StatusInternalServerError {
			t.Errorf("Expected final status code %d, got %d", http.StatusInternalServerError,
				resp.StatusCode())
		}

		expectedCalls := 3 // 1 initial + 2 retries
		if callCount != expectedCalls {
			t.Errorf("Expected %d calls, got %d", expectedCalls, callCount)
		}

		if resp.RetryAttempts() != 2 {
			t.Errorf("Expected 2 retry attempts, got %d", resp.RetryAttempts())
		}
	})

	t.Run("no retries without retry config", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewClientBuilder().Build() // No retry config

		resp, err := client.Get(context.TODO(), server.URL)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if callCount != 1 {
			t.Errorf("Expected 1 call (no retries), got %d", callCount)
		}

		if resp.RetryAttempts() != 0 {
			t.Errorf("Expected 0 retry attempts, got %d", resp.RetryAttempts())
		}
	})
}

func TestClient_NonRetryableStatusCodes(t *testing.T) {
	t.Run("does not retry on 404", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		retryConfig := NewRetryConfigBuilder().
			WithMaxRetries(3).
			WithBackoffStrategy(NewFixedBackoffBuilder().
				WithDelay(10 * time.Millisecond).
				Build()).
			Build()

		client := NewClientBuilder().
			WithRetryConfig(retryConfig).
			Build()

		resp, err := client.Get(context.TODO(), server.URL)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if resp.StatusCode() != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode())
		}

		if callCount != 1 {
			t.Errorf("Expected 1 call (no retries for 404), got %d", callCount)
		}

		if resp.RetryAttempts() != 0 {
			t.Errorf("Expected 0 retry attempts, got %d", resp.RetryAttempts())
		}
	})

	t.Run("does not retry on 2xx success", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Success"))
		}))
		defer server.Close()

		retryConfig := NewRetryConfigBuilder().
			WithMaxRetries(3).
			Build()

		client := NewClientBuilder().
			WithRetryConfig(retryConfig).
			Build()

		resp, err := client.Get(context.TODO(), server.URL)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if resp.StatusCode() != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode())
		}

		if callCount != 1 {
			t.Errorf("Expected 1 call (no retries for success), got %d", callCount)
		}

		if resp.RetryAttempts() != 0 {
			t.Errorf("Expected 0 retry attempts, got %d", resp.RetryAttempts())
		}
	})
}

func TestClient_RetryWithPOSTBody(t *testing.T) {
	t.Run("retries POST with body correctly", func(t *testing.T) {
		callCount := 0
		expectedBody := "test payload data"
		var receivedBodies []string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("Failed to read request body: %v", err)
			}
			receivedBodies = append(receivedBodies, string(body))

			if callCount < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Success"))
		}))
		defer server.Close()

		retryConfig := NewRetryConfigBuilder().
			WithMaxRetries(3).
			WithBackoffStrategy(NewFixedBackoffBuilder().
				WithDelay(10 * time.Millisecond).
				WithJitter(false).
				Build()).
			Build()

		client := NewClientBuilder().
			WithRetryConfig(retryConfig).
			Build()

		resp, err := client.Post(context.TODO(), server.URL, []byte(expectedBody))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if resp.StatusCode() != http.StatusOK {
			t.Errorf("Expected final status code %d, got %d", http.StatusOK, resp.StatusCode())
		}

		if callCount != 3 {
			t.Errorf("Expected 3 calls, got %d", callCount)
		}

		if resp.RetryAttempts() != 2 {
			t.Errorf("Expected 2 retry attempts, got %d", resp.RetryAttempts())
		}

		// Verify body was sent correctly in all attempts
		if len(receivedBodies) != 3 {
			t.Errorf("Expected 3 received bodies, got %d", len(receivedBodies))
		}

		for i, body := range receivedBodies {
			if body != expectedBody {
				t.Errorf("Body mismatch in attempt %d: expected %q, got %q", i+1, expectedBody,
					body)
			}
		}
	})

	t.Run("retries POST with empty body", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("Failed to read request body: %v", err)
			}

			if len(body) != 0 {
				t.Errorf("Expected empty body, got %q", string(body))
			}

			if callCount < 2 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		retryConfig := NewRetryConfigBuilder().
			WithMaxRetries(2).
			WithBackoffStrategy(NewFixedBackoffBuilder().
				WithDelay(5 * time.Millisecond).
				Build()).
			Build()

		client := NewClientBuilder().
			WithRetryConfig(retryConfig).
			Build()

		resp, err := client.Post(context.TODO(), server.URL, []byte{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if callCount != 2 {
			t.Errorf("Expected 2 calls, got %d", callCount)
		}

		if resp.RetryAttempts() != 1 {
			t.Errorf("Expected 1 retry attempt, got %d", resp.RetryAttempts())
		}
	})
}

func TestClient_RetryContextCancellation(t *testing.T) {
	t.Run("cancels during retry backoff", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		retryConfig := NewRetryConfigBuilder().
			WithMaxRetries(5).
			WithBackoffStrategy(NewFixedBackoffBuilder().
				WithDelay(500 * time.Millisecond). // Long delay for cancellation test
				WithJitter(false).
				Build()).
			Build()

		client := NewClientBuilder().
			WithRetryConfig(retryConfig).
			Build()

		ctx, cancel := context.WithCancel(context.Background())
		// Cancel context after first failure, during backoff delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		_, err := client.Get(ctx, server.URL)

		if err == nil {
			t.Error("Expected context cancellation error, got nil")
		}
		if !strings.Contains(err.Error(), "context canceled") {
			t.Errorf("Expected context canceled error, got: %v", err)
		}

		// Should have made at least 1 call but not all retries
		if callCount == 0 {
			t.Error("Expected at least 1 call before cancellation")
		}

		if callCount > 2 {
			t.Errorf("Expected cancellation to stop retries, but got %d calls", callCount)
		}
	})

	t.Run("respects timeout during retries", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		retryConfig := NewRetryConfigBuilder().
			WithMaxRetries(10).
			WithBackoffStrategy(NewFixedBackoffBuilder().
				WithDelay(200 * time.Millisecond).
				WithJitter(false).
				Build()).
			Build()

		client := NewClientBuilder().
			WithRetryConfig(retryConfig).
			Build()

		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		_, err := client.Get(ctx, server.URL)

		if err == nil {
			t.Error("Expected timeout error, got nil")
		}

		if !strings.Contains(err.Error(), "context deadline exceeded") {
			t.Errorf("Expected timeout error, got: %v", err)
		}

		// Should have made some calls but not all 10 retries due to timeout
		if callCount == 0 {
			t.Error("Expected at least 1 call before timeout")
		}

		if callCount > 5 {
			t.Errorf("Expected timeout to limit calls, but got %d calls", callCount)
		}
	})
}

func TestClient_WithRetriesConvenience(t *testing.T) {
	t.Run("default retry behavior works", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Success"))
		}))
		defer server.Close()

		client := NewClientBuilder().
			WithRetries(). // Using convenience method
			Build()

		resp, err := client.Get(context.TODO(), server.URL)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if resp.StatusCode() != http.StatusOK {
			t.Errorf("Expected final status code %d, got %d", http.StatusOK, resp.StatusCode())
		}

		if callCount != 3 {
			t.Errorf("Expected 3 calls, got %d", callCount)
		}

		if resp.RetryAttempts() != 2 {
			t.Errorf("Expected 2 retry attempts, got %d", resp.RetryAttempts())
		}
	})

	t.Run("retries on default status codes", func(t *testing.T) {
		testCases := []struct {
			name        string
			statusCode  int
			shouldRetry bool
		}{
			{"retries on 429", http.StatusTooManyRequests, true},
			{"retries on 500", http.StatusInternalServerError, true},
			{"retries on 502", http.StatusBadGateway, true},
			{"retries on 503", http.StatusServiceUnavailable, true},
			{"retries on 504", http.StatusGatewayTimeout, true},
			{"no retry on 400", http.StatusBadRequest, false},
			{"no retry on 404", http.StatusNotFound, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				callCount := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					callCount++
					w.WriteHeader(tc.statusCode)
				}))
				defer server.Close()

				client := NewClientBuilder().WithRetries().Build()

				resp, err := client.Get(context.TODO(), server.URL)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				expectedCalls := 1
				if tc.shouldRetry {
					expectedCalls = 4 // 1 initial + 3 default retries
				}

				if callCount != expectedCalls {
					t.Errorf("Expected %d calls, got %d", expectedCalls, callCount)
				}

				expectedRetries := 0
				if tc.shouldRetry {
					expectedRetries = 3
				}

				if resp.RetryAttempts() != expectedRetries {
					t.Errorf("Expected %d retry attempts, got %d", expectedRetries,
						resp.RetryAttempts())
				}
			})
		}
	})
}

func TestBodyReaderFromByteSlice(t *testing.T) {
	t.Run("nil slice returns nil reader", func(t *testing.T) {
		reader := bodyReaderFromByteSlice(nil)
		if reader != nil {
			t.Errorf("Expected nil reader for nil slice, got %T", reader)
		}
	})

	t.Run("empty slice returns nil reader", func(t *testing.T) {
		reader := bodyReaderFromByteSlice([]byte{})
		if reader != nil {
			t.Errorf("Expected nil reader for empty slice, got %T", reader)
		}
	})

	t.Run("non-empty slice returns buffer", func(t *testing.T) {
		data := []byte("test data")
		reader := bodyReaderFromByteSlice(data)

		if reader == nil {
			t.Fatal("Expected non-nil reader for non-empty slice")
		}

		// Verify we can read the data back
		result, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Failed to read from reader: %v", err)
		}

		if string(result) != "test data" {
			t.Errorf("Expected 'test data', got '%s'", string(result))
		}
	})

	t.Run("multiple calls create independent readers", func(t *testing.T) {
		data := []byte("shared data")

		reader1 := bodyReaderFromByteSlice(data)
		reader2 := bodyReaderFromByteSlice(data)

		if reader1 == nil || reader2 == nil {
			t.Fatal("Expected non-nil readers")
		}

		// Read from first reader
		result1, err := io.ReadAll(reader1)
		if err != nil {
			t.Fatalf("Failed to read from first reader: %v", err)
		}

		// Second reader should still have all data
		result2, err := io.ReadAll(reader2)
		if err != nil {
			t.Fatalf("Failed to read from second reader: %v", err)
		}

		if string(result1) != "shared data" || string(result2) != "shared data" {
			t.Errorf("Both readers should return same data")
		}
	})
}

func TestContextCancelled(t *testing.T) {
	t.Run("returns nil for active context", func(t *testing.T) {
		ctx := context.Background()
		err := contextCancelled(ctx)
		if err != nil {
			t.Errorf("Expected nil for active context, got %v", err)
		}
	})

	t.Run("returns error for cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := contextCancelled(ctx)
		if err == nil {
			t.Error("Expected error for cancelled context, got nil")
		}

		if !strings.Contains(err.Error(), "context canceled") {
			t.Errorf("Expected context canceled error, got: %v", err)
		}
	})

	t.Run("returns error for deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(1 * time.Millisecond) // Ensure timeout

		err := contextCancelled(ctx)
		if err == nil {
			t.Error("Expected error for timed out context, got nil")
		}

		if !strings.Contains(err.Error(), "context deadline exceeded") {
			t.Errorf("Expected deadline exceeded error, got: %v", err)
		}
	})

	t.Run("returns immediately without blocking", func(t *testing.T) {
		ctx := context.Background()

		start := time.Now()
		err := contextCancelled(ctx)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Expected nil for background context, got %v", err)
		}

		if duration > 10*time.Millisecond {
			t.Errorf("contextCancelled took too long: %v", duration)
		}
	})
}

func TestClient_ApplyBackoff(t *testing.T) {
	t.Run("no delay on first attempt", func(t *testing.T) {
		retryConfig := NewRetryConfigBuilder().
			WithBackoffStrategy(NewFixedBackoffBuilder().
				WithDelay(100 * time.Millisecond).
				Build()).
			Build()

		client := &client{retryConfig: retryConfig}
		ctx := context.Background()

		start := time.Now()
		err := client.applyBackoff(ctx, 0) // First attempt
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if duration > 10*time.Millisecond {
			t.Errorf("Expected no delay for first attempt, but took %v", duration)
		}
	})

	t.Run("applies delay on retry attempts", func(t *testing.T) {
		retryConfig := NewRetryConfigBuilder().
			WithBackoffStrategy(NewFixedBackoffBuilder().
				WithDelay(50 * time.Millisecond).
				WithJitter(false).
				Build()).
			Build()

		client := &client{retryConfig: retryConfig}
		ctx := context.Background()

		start := time.Now()
		err := client.applyBackoff(ctx, 1) // First retry
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if duration < 40*time.Millisecond || duration > 60*time.Millisecond {
			t.Errorf("Expected ~50ms delay, got %v", duration)
		}
	})

	t.Run("respects context cancellation during backoff", func(t *testing.T) {
		retryConfig := NewRetryConfigBuilder().
			WithBackoffStrategy(NewFixedBackoffBuilder().
				WithDelay(500 * time.Millisecond).
				Build()).
			Build()

		client := &client{retryConfig: retryConfig}
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel after short delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		start := time.Now()
		err := client.applyBackoff(ctx, 1)
		duration := time.Since(start)

		if err == nil {
			t.Error("Expected context cancellation error, got nil")
		}

		if !strings.Contains(err.Error(), "context canceled") {
			t.Errorf("Expected context canceled error, got: %v", err)
		}

		if duration > 100*time.Millisecond {
			t.Errorf("Expected early cancellation, but took %v", duration)
		}
	})

	t.Run("no delay without retry config", func(t *testing.T) {
		client := &client{retryConfig: nil}
		ctx := context.Background()

		start := time.Now()
		err := client.applyBackoff(ctx, 1)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if duration > 10*time.Millisecond {
			t.Errorf("Expected no delay without retry config, but took %v", duration)
		}
	})
}

func TestClient_ShouldRetry(t *testing.T) {
	retryConfig := NewRetryConfigBuilder().
		WithRetryableStatusCodes([]int{500, 502, 503}).
		Build()

	cli := &client{retryConfig: retryConfig}

	t.Run("retries on configured status codes", func(t *testing.T) {
		testCases := []struct {
			statusCode  int
			shouldRetry bool
		}{
			{500, true},
			{502, true},
			{503, true},
			{404, false},
			{200, false},
			{429, false}, // Not in our custom config
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("status_%d", tc.statusCode), func(t *testing.T) {
				resp := &Response{statusCode: tc.statusCode}
				result := cli.shouldRetry(resp)

				if result != tc.shouldRetry {
					t.Errorf("shouldRetry(%d) = %t, want %t", tc.statusCode, result, tc.shouldRetry)
				}
			})
		}
	})

	t.Run("returns false for nil response", func(t *testing.T) {
		result := cli.shouldRetry(nil)
		if result {
			t.Error("shouldRetry(nil) should return false")
		}
	})

	t.Run("returns false without retry config", func(t *testing.T) {
		clientNoRetry := &client{retryConfig: nil}
		resp := &Response{statusCode: 500}

		result := clientNoRetry.shouldRetry(resp)
		if result {
			t.Error("shouldRetry should return false without retry config")
		}
	})
}

func TestClient_ShouldRetryError(t *testing.T) {
	t.Run("retries on default retryable errors", func(t *testing.T) {
		retryConfig := NewRetryConfigBuilder().Build() // Uses default retryable errors
		cli := &client{retryConfig: retryConfig}

		testCases := []struct {
			name        string
			err         error
			shouldRetry bool
		}{
			{"connection refused error", fmt.Errorf("dial tcp: connection refused"), true},
			{"timeout error", fmt.Errorf("context deadline exceeded: timeout"), true},
			{"temporary failure error", fmt.Errorf("temporary failure in name resolution"), true},
			{"no such host error", fmt.Errorf("dial tcp: lookup example.com: no such host"), true},
			{"case insensitive matching", fmt.Errorf("TIMEOUT occurred"), true},
			{"non-retryable error", fmt.Errorf("invalid request format"), false},
			{"empty error message", fmt.Errorf(""), false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := cli.shouldRetryError(tc.err)
				if result != tc.shouldRetry {
					t.Errorf("shouldRetryError(%v) = %t, want %t", tc.err, result, tc.shouldRetry)
				}
			})
		}
	})

	t.Run("retries on custom retryable errors", func(t *testing.T) {
		customRetryableErrors := map[string]bool{
			"service unavailable": true,
			"rate limit":          true,
		}
		retryConfig := NewRetryConfigBuilder().
			WithRetryableErrors(customRetryableErrors).
			Build()
		cli := &client{retryConfig: retryConfig}

		testCases := []struct {
			name        string
			err         error
			shouldRetry bool
		}{
			{"custom service unavailable", fmt.Errorf("service unavailable"), true},
			{"custom rate limit", fmt.Errorf("rate limit exceeded"), true},
			{"case insensitive custom", fmt.Errorf("SERVICE UNAVAILABLE"), true},
			{"default error not in custom config", fmt.Errorf("connection refused"), false},
			{"non-retryable error", fmt.Errorf("bad request"), false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := cli.shouldRetryError(tc.err)
				if result != tc.shouldRetry {
					t.Errorf("shouldRetryError(%v) = %t, want %t", tc.err, result, tc.shouldRetry)
				}
			})
		}
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		retryConfig := NewRetryConfigBuilder().Build()
		cli := &client{retryConfig: retryConfig}

		result := cli.shouldRetryError(nil)
		if result {
			t.Error("shouldRetryError(nil) should return false")
		}
	})

	t.Run("returns false without retry config", func(t *testing.T) {
		cli := &client{retryConfig: nil}
		err := fmt.Errorf("connection refused")

		result := cli.shouldRetryError(err)
		if result {
			t.Error("shouldRetryError should return false without retry config")
		}
	})

	t.Run("partial string matching", func(t *testing.T) {
		retryConfig := NewRetryConfigBuilder().Build()
		cli := &client{retryConfig: retryConfig}

		testCases := []struct {
			name        string
			err         error
			shouldRetry bool
		}{
			{"timeout in middle of message", fmt.Errorf("request failed due to timeout while connecting"), true},
			{"connection refused with details", fmt.Errorf("failed to connect: connection refused by server"), true},
			{"no such host with context", fmt.Errorf("DNS lookup failed: no such host found"), true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := cli.shouldRetryError(tc.err)
				if result != tc.shouldRetry {
					t.Errorf("shouldRetryError(%v) = %t, want %t", tc.err, result, tc.shouldRetry)
				}
			})
		}
	})
}
