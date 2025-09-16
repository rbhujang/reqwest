package reqwest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClientBuilder(t *testing.T) {
	builder := NewClientBuilder()
	if builder == nil {
		t.Fatal("NewClientBuilder() returned nil")
	}
	if builder.baseURL != "" {
		t.Errorf("Expected empty baseURL, got %q", builder.baseURL)
	}
}

func TestClientBuilder_WithBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL without trailing slash",
			input:    "https://api.example.com",
			expected: "https://api.example.com",
		},
		{
			name:     "URL with single trailing slash",
			input:    "https://api.example.com/",
			expected: "https://api.example.com",
		},
		{
			name:     "URL with multiple trailing slashes",
			input:    "https://api.example.com///",
			expected: "https://api.example.com",
		},
		{
			name:     "Empty URL",
			input:    "",
			expected: "",
		},
		{
			name:     "URL with path and trailing slash",
			input:    "https://api.example.com/v1/",
			expected: "https://api.example.com/v1",
		},
		{
			name:     "URL with query parameters and trailing slash",
			input:    "https://api.example.com/v1?key=value/",
			expected: "https://api.example.com/v1?key=value",
		},
		{
			name:     "Localhost URL with port",
			input:    "http://localhost:8080/",
			expected: "http://localhost:8080",
		},
		{
			name:     "URL with only slashes",
			input:    "///",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewClientBuilder().WithBaseURL(tt.input)
			if builder.baseURL != tt.expected {
				t.Errorf("WithBaseURL(%q) = %q, want %q", tt.input, builder.baseURL, tt.expected)
			}
		})
	}
}

func TestClientBuilder_WithBaseURL_Chaining(t *testing.T) {
	builder := NewClientBuilder()
	result := builder.WithBaseURL("https://api.example.com")

	if result != builder {
		t.Error("WithBaseURL() should return the same builder instance for chaining")
	}
}

func TestClientBuilder_WithBaseURL_MultipleChaining(t *testing.T) {
	builder := NewClientBuilder()

	result := builder.
		WithBaseURL("https://first.com/").
		WithBaseURL("https://second.com/v1/").
		WithBaseURL("https://final.com/api/")

	if result != builder {
		t.Error("Multiple WithBaseURL() calls should return the same builder instance")
	}

	if builder.baseURL != "https://final.com/api" {
		t.Errorf("Expected final baseURL to be 'https://final.com/api', got %q", builder.baseURL)
	}
}

func TestClientBuilder_Build(t *testing.T) {

	t.Run("Build preserves builder state", func(t *testing.T) {
		builder := NewClientBuilder().WithBaseURL("https://api.example.com")

		client1 := builder.Build()
		client2 := builder.Build()

		// Both clients should have the same configuration
		impl1 := client1.(*client)
		impl2 := client2.(*client)

		if impl1.baseURL != impl2.baseURL {
			t.Error("Multiple Build() calls should produce clients with same configuration")
		}

		// But they should be different instances
		if impl1 == impl2 {
			t.Error("Multiple Build() calls should produce different client instances")
		}
	})
}

func TestClientBuilder_Integration(t *testing.T) {
	t.Run("Empty builder produces working client", func(t *testing.T) {
		client := NewClientBuilder().Build()

		// Should be able to call interface methods without panic
		if client == nil {
			t.Fatal("Build() returned nil client")
		}

		// Verify it implements the Client interface
		var _ Client = client
	})
}

func TestClientBuilder_EdgeCases(t *testing.T) {
	t.Run("Nil safety", func(t *testing.T) {
		builder := NewClientBuilder()
		if builder == nil {
			t.Fatal("NewClientBuilder() returned nil")
		}

		// These operations should not panic
		builder.WithBaseURL("")
		builder.WithBaseURL("test")
		client := builder.Build()

		if client == nil {
			t.Error("Build() returned nil after edge case operations")
		}
	})

	t.Run("Unicode URL handling", func(t *testing.T) {
		unicodeURL := "https://例え.テスト/"
		expected := "https://例え.テスト"

		builder := NewClientBuilder().WithBaseURL(unicodeURL)
		if builder.baseURL != expected {
			t.Errorf("Unicode URL handling failed: got %q, want %q", builder.baseURL, expected)
		}
	})
}

func TestClientBuilder_WithMiddleware(t *testing.T) {
	t.Run("One middleware", func(t *testing.T) {
		cameInsideMiddleware := false
		middleware := func(req *http.Request) error {
			cameInsideMiddleware = true
			req.Header.Add("Testing-Key", "Testing-Value")
			return nil
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Testing-Key") != "Testing-Value" {
				t.Errorf("Expected header Testing-Key=Testing-Value, got %s",
					r.Header.Get("Testing-Key"))
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClientBuilder().WithMiddleware(middleware).Build()

		_, err := client.Get(server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !cameInsideMiddleware {
			t.Error("Middleware not executed, when it should have been")
		}
	})

	t.Run("Multiple middlewares", func(t *testing.T) {
		var execution []string

		middleware1 := func(req *http.Request) error {
			execution = append(execution, "middleware1")
			req.Header.Set("M1", "M1-Value")
			return nil
		}

		middleware2 := func(req *http.Request) error {
			execution = append(execution, "middleware2")
			req.Header.Set("M2", "M2-Value")
			return nil
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("M1") != "M1-Value" {
				t.Errorf("Expected header M1=M1-Value, got %s", r.Header.Get("M1"))
			}
			if r.Header.Get("M2") != "M2-Value" {
				t.Errorf("Expected header M2=M2-Value, got %s",
					r.Header.Get("M2"))
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClientBuilder().
			WithMiddleware(middleware1).
			WithMiddleware(middleware2).
			Build()

		_, err := client.Get(server.URL)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedExecution := []string{"middleware1", "middleware2"}
		if len(execution) != len(expectedExecution) {
			t.Fatalf("Expected %d middleware executions, got %d",
				len(expectedExecution), len(execution))
		}
		for i, expected := range expectedExecution {
			if execution[i] != expected {
				t.Errorf("Expected execution[%d]=%s, got %s", i, expected, execution[i])
			}
		}
	})

	t.Run("Middleware with error", func(t *testing.T) {
		errorMiddleware := func(req *http.Request) error {
			return fmt.Errorf("middleware failed")
		}

		client := NewClientBuilder().
			WithMiddleware(errorMiddleware).
			Build()

		_, err := client.Get("http://example.com")
		if err == nil {
			t.Error("Expected error from middleware, got nil")
		}

		if !strings.Contains(err.Error(), "middleware error") {
			t.Errorf("Expected error to contain 'middleware error', got: %v", err)
		}
	})

	t.Run("Multiple Clients", func(t *testing.T) {
		var client1Headers, client2Headers []string

		middleware1 := func(req *http.Request) error {
			client1Headers = append(client1Headers, "client1")
			req.Header.Set("X-Client", "client1")
			return nil
		}

		middleware2 := func(req *http.Request) error {
			client2Headers = append(client2Headers, "client2")
			req.Header.Set("X-Client", "client2")
			return nil
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client1 := NewClientBuilder().WithMiddleware(middleware1).Build()
		client2 := NewClientBuilder().WithMiddleware(middleware2).Build()

		_, err1 := client1.Get(server.URL)
		_, err2 := client2.Get(server.URL)

		if err1 != nil || err2 != nil {
			t.Fatalf("Unexpected errors: %v, %v", err1, err2)
		}

		if len(client1Headers) != 1 || client1Headers[0] != "client1" {
			t.Errorf("Client1 middleware not executed correctly: %v", client1Headers)
		}

		if len(client2Headers) != 1 || client2Headers[0] != "client2" {
			t.Errorf("Client2 middleware not executed correctly: %v", client2Headers)
		}
	})
}
