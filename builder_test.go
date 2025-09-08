package reqwest

import (
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
