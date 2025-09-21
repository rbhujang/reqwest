# reqwest

[![Go Version](https://img.shields.io/badge/Go-1.23.4-blue.svg)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/Coverage-96.7%25-green.svg)](coverage.html)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A simple and elegant HTTP client library for Go

> **⚠️ Work in Progress**: This library is currently under active development. APIs may change and features may be incomplete.

## Features

- Simple and intuitive API
- Builder pattern for client configuration
- Context-aware requests with full timeout control
- Base URL support for API clients
- GET and POST methods
- Middleware support for request interception
- Automatic retries with configurable backoff strategies

## Installation

```bash
go get github.com/rbhujang/reqwest
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "io"
    "time"

    "github.com/rbhujang/reqwest"
)

func main() {
    // Create a client
    client := reqwest.NewClientBuilder().Build()

    // Make a GET request with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    resp, err := client.Get(ctx, "https://api.github.com/users/octocat")
    if err != nil {
        panic(err)
    }
    defer resp.Body().Close()

    // Read response
    body, _ := io.ReadAll(resp.Body())
    fmt.Printf("Status: %d\n", resp.StatusCode())
    fmt.Printf("Body: %s\n", body)
}
```

### Using Base URL

```go
// Create a client with base URL
client := reqwest.NewClientBuilder().
    WithBaseURL("https://api.github.com").
    Build()

// Make requests with relative paths
ctx := context.Background()
resp, err := client.Get(ctx, "/users/octocat")
// This will request: https://api.github.com/users/octocat
```

### POST Requests

```go
client := reqwest.NewClientBuilder().Build()

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

jsonData := []byte(`{"key": "value"}`)
resp, err := client.Post(ctx, "https://httpbin.org/post", jsonData)
if err != nil {
    panic(err)
}
defer resp.Body().Close()

fmt.Printf("Status: %d\n", resp.StatusCode())
```

### Requests with Retries

```go
// Create a client with automatic retries
client := reqwest.NewClientBuilder().
    WithBaseURL("https://api.github.com").
    WithRetries().  // Enable default retry configuration
    Build()

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// This request will automatically retry on transient failures
resp, err := client.Get(ctx, "/users/octocat")
if err != nil {
    panic(err)
}
defer resp.Body().Close()

fmt.Printf("Status: %d\n", resp.StatusCode())
fmt.Printf("Completed after %d retry attempts\n", resp.RetryAttempts())
```

## Retry Configuration

The library supports automatic retries with configurable backoff strategies for handling transient failures.

### Basic Retry Setup

```go
// Enable retries with default configuration
client := reqwest.NewClientBuilder().
    WithRetries().
    Build()

// Make requests - they will automatically retry on failure
resp, err := client.Get(ctx, "https://api.example.com/data")
```

### Custom Retry Configuration

```go
// Configure custom retry behavior
retryConfig := reqwest.NewRetryConfigBuilder().
    WithMaxRetries(5).
    WithRetryableStatusCodes([]int{429, 500, 502, 503, 504}).
    WithBackoffStrategy(
        reqwest.NewExponentialBackoffBuilder().
            WithBaseDelay(200 * time.Millisecond).
            WithMultiplier(2.0).
            WithMaxDelay(30 * time.Second).
            WithJitter(true).
            Build(),
    ).
    Build()

client := reqwest.NewClientBuilder().
    WithRetryConfig(retryConfig).
    Build()
```

### Backoff Strategies

#### Exponential Backoff (Default)
```go
backoff := reqwest.NewExponentialBackoffBuilder().
    WithBaseDelay(100 * time.Millisecond).  // Starting delay
    WithMultiplier(2.0).                     // Multiplier for each retry
    WithMaxDelay(10 * time.Second).         // Maximum delay cap
    WithJitter(true).                       // Add randomization
    Build()
```

#### Fixed Backoff
```go
backoff := reqwest.NewFixedBackoffBuilder().
    WithDelay(1 * time.Second).             // Fixed delay between retries
    WithJitter(true).                       // Add randomization
    Build()
```

### Retry Behavior

- **Default retryable status codes**: 429 (Too Many Requests), 500, 502, 503, 504
- **Default retryable errors**: Connection refused, timeouts, temporary failures, DNS resolution failures
- **Default max retries**: 3 attempts
- **Jitter**: Adds ±25% randomization to backoff delays to prevent thundering herd

### Checking Retry Attempts

```go
resp, err := client.Get(ctx, "https://api.example.com/data")
if err != nil {
    panic(err)
}

fmt.Printf("Request completed after %d retry attempts\n", resp.RetryAttempts())
```

## Context and Timeouts

All requests require a `context.Context` parameter, giving you full control over request lifecycle:

### Request Timeouts

```go
// 5 second timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
resp, err := client.Get(ctx, "/users")
```

### Request Cancellation

```go
// Cancel request programmatically
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(2 * time.Second)
    cancel() // Cancel the request
}()
resp, err := client.Get(ctx, "/users")
```

### No Timeout (Use with caution)

```go
// No timeout - request can hang indefinitely
resp, err := client.Get(context.Background(), "/users")
```

### Request Tracing

```go
// Add request ID for tracing
ctx := context.WithValue(context.Background(), "request-id", "abc-123")
resp, err := client.Get(ctx, "/users")
```

## API Reference

### ClientBuilder

#### `NewClientBuilder() *ClientBuilder`

Creates a new client builder.

#### `WithBaseURL(url string) *ClientBuilder`

Sets the base URL for the client. Trailing slashes are automatically trimmed.

#### `WithRetries() *ClientBuilder`

Enables retries with default configuration (3 max retries, exponential backoff).

#### `WithRetryConfig(config *RetryConfig) *ClientBuilder`

Sets a custom retry configuration for the client.

#### `Build() Client`

Builds and returns the configured client.

### Client

#### `Get(ctx context.Context, url string) (*Response, error)`

Performs a GET request to the specified URL with the provided context.

#### `Post(ctx context.Context, url string, body []byte) (*Response, error)`

Performs a POST request to the specified URL with the given body and context.

### Response

#### `StatusCode() int`

Returns the HTTP status code of the response.

#### `Body() io.ReadCloser`

Returns the response body as a ReadCloser. Remember to close it when done.

#### `RetryAttempts() int`

Returns the number of retry attempts made for this request.

### RetryConfig

#### `NewRetryConfigBuilder() *retryConfigBuilder`

Creates a new retry configuration builder.

#### `WithMaxRetries(maxRetries int) *retryConfigBuilder`

Sets the maximum number of retry attempts.

#### `WithRetryableStatusCodes(codes []int) *retryConfigBuilder`

Sets which HTTP status codes should trigger retries.

#### `WithRetryableErrors(errors map[string]bool) *retryConfigBuilder`

Sets which error strings should trigger retries.

#### `WithBackoffStrategy(strategy BackoffStrategy) *retryConfigBuilder`

Sets the backoff strategy for delays between retries.

### Exponential Backoff

#### `NewExponentialBackoffBuilder() *exponentialBackoffBuilder`

Creates a new exponential backoff strategy builder.

#### `WithBaseDelay(delay time.Duration) *exponentialBackoffBuilder`

Sets the initial delay for the first retry.

#### `WithMultiplier(multiplier float64) *exponentialBackoffBuilder`

Sets the multiplication factor for each subsequent retry.

#### `WithMaxDelay(delay time.Duration) *exponentialBackoffBuilder`

Sets the maximum delay cap for retries.

#### `WithJitter(jitter bool) *exponentialBackoffBuilder`

Enables or disables jitter (±25% randomization) in delays.

### Fixed Backoff

#### `NewFixedBackoffBuilder() *fixedBackoffBuilder`

Creates a new fixed backoff strategy builder.

#### `WithDelay(delay time.Duration) *fixedBackoffBuilder`

Sets the fixed delay between retries.

#### `WithJitter(jitter bool) *fixedBackoffBuilder`

Enables or disables jitter (±25% randomization) in delays.

## URL Handling

- If a base URL is configured, relative URLs will be appended to it
- Absolute URLs (starting with `http://` or `https://`) will be used as-is, ignoring the base URL
- Leading slashes in relative URLs are automatically handled

## Timeouts

Timeout handling is controlled entirely through the context parameter. If no timeout is specified in the context, requests can potentially hang indefinitely. It's recommended to always use `context.WithTimeout()` for production applications to ensure your application remains responsive.

## Requirements

- Go 1.23.4 or later

## License

This project is open source. See the license file for details.
