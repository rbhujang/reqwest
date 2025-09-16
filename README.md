# reqwest

[![Go Version](https://img.shields.io/badge/Go-1.23.4-blue.svg)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/Coverage-97.1%25-green.svg)](coverage.html)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A simple and elegant HTTP client library for Go

> **⚠️ Work in Progress**: This library is currently under active development. APIs may change and features may be incomplete.

## Features

- Simple and intuitive API
- Builder pattern for client configuration
- Built-in timeout handling (10 seconds default)
- Base URL support for API clients
- GET and POST methods
- Context-aware requests

## Installation

```bash
go get github.com/rbhujang/reqwest
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "io"

    "github.com/rbhujang/reqwest"
)

func main() {
    // Create a client
    client := reqwest.NewClientBuilder().Build()

    // Make a GET request
    resp, err := client.Get("https://api.github.com/users/octocat")
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
resp, err := client.Get("/users/octocat")
// This will request: https://api.github.com/users/octocat
```

### POST Requests

```go
client := reqwest.NewClientBuilder().Build()

jsonData := []byte(`{"key": "value"}`)
resp, err := client.Post("https://httpbin.org/post", jsonData)
if err != nil {
    panic(err)
}
defer resp.Body().Close()

fmt.Printf("Status: %d\n", resp.StatusCode())
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

#### `Get(url string) (*Response, error)`

Performs a GET request to the specified URL.

#### `Post(url string, body []byte) (*Response, error)`

Performs a POST request to the specified URL with the given body.

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

All requests have a default timeout of 10 seconds. This helps prevent hanging requests and ensures your application remains responsive.

## Requirements

- Go 1.23.4 or later

## License

This project is open source. See the license file for details.
