# reqwest

[![Go Version](https://img.shields.io/badge/Go-1.23.4-blue.svg)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/Coverage-96.8%25-green.svg)](coverage.html)
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
