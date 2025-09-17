package reqwest

import (
	"context"
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
