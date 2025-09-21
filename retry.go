package reqwest

import (
	"math"
	"math/rand"
	"time"
)

// Default retry configuration constants
const (
	DefaultMaxRetries    = 3
	DefaultJitterPercent = 0.25
)

// Default backoff timing constants
const (
	DefaultExponentialBaseDelay  = 100 * time.Millisecond
	DefaultExponentialMultiplier = 2.0
	DefaultExponentialMaxDelay   = 10 * time.Second
	DefaultFixedBackoffDelay     = 1 * time.Second
)

// Jitter calculation constants
const (
	JitterCenterPoint = 0.5 // Center point for jitter calculation
	JitterMultiplier  = 2.0 // Multiplier for jitter range
)

// Default retryable HTTP status codes
var DefaultRetryableStatusCodes = []int{429, 500, 502, 503, 504}

// Default retryable error strings
var DefaultRetryableErrors = map[string]bool{
	"connection refused": true,
	"timeout":            true,
	"temporary failure":  true,
	"no such host":       true,
}

type RetryConfig struct {
	maxRetries           int
	retryableStatusCodes []int
	retryableError       map[string]bool
	backoffStrategy      BackoffStrategy
}

type retryConfigBuilder struct {
	config *RetryConfig
}

// NewRetryConfigBuilder creates a new builder for RetryConfig.
// The builder is not safe for concurrent use. Each goroutine should create
// its own builder instance.
func NewRetryConfigBuilder() *retryConfigBuilder {
	return &retryConfigBuilder{
		config: &RetryConfig{},
	}
}

func (r *retryConfigBuilder) WithMaxRetries(maxRetries int) *retryConfigBuilder {
	r.config.maxRetries = maxRetries
	return r
}

func (r *retryConfigBuilder) WithBackoffStrategy(strategy BackoffStrategy) *retryConfigBuilder {
	r.config.backoffStrategy = strategy
	return r
}

func (r *retryConfigBuilder) WithRetryableStatusCodes(codes []int) *retryConfigBuilder {
	r.config.retryableStatusCodes = codes
	return r
}

func (r *retryConfigBuilder) WithRetryableErrors(errors map[string]bool) *retryConfigBuilder {
	r.config.retryableError = errors
	return r
}

func (r *retryConfigBuilder) Build() *RetryConfig {
	// Default to 3 retries
	if r.config.maxRetries <= 0 {
		r.config.maxRetries = DefaultMaxRetries
	}
	// Default to exponential backoff
	if r.config.backoffStrategy == nil {
		r.config.backoffStrategy = NewExponentialBackoffBuilder().Build()
	}
	// Default to common HTTP retriable codes
	if len(r.config.retryableStatusCodes) == 0 {
		r.config.retryableStatusCodes = DefaultRetryableStatusCodes
	}
	// Default to probably network error strings
	if len(r.config.retryableError) == 0 {
		r.config.retryableError = DefaultRetryableErrors
	}
	return r.config
}

type BackoffStrategy interface {
	Delay(count int) time.Duration
}

type ExponentialBackoff struct {
	baseDelay  time.Duration
	multiplier float64
	maxDelay   time.Duration
	jitter     bool
}

type exponentialBackoffBuilder struct {
	backoff *ExponentialBackoff
}

// NewExponentialBackoffBuilder creates a new builder for ExponentialBackoff.
// The builder is not safe for concurrent use. Each goroutine should create
// its own builder instance.
func NewExponentialBackoffBuilder() *exponentialBackoffBuilder {
	return &exponentialBackoffBuilder{
		backoff: &ExponentialBackoff{},
	}
}

func (b *exponentialBackoffBuilder) WithBaseDelay(delay time.Duration) *exponentialBackoffBuilder {
	b.backoff.baseDelay = delay
	return b
}

func (b *exponentialBackoffBuilder) WithMultiplier(multiplier float64) *exponentialBackoffBuilder {
	b.backoff.multiplier = multiplier
	return b
}

func (b *exponentialBackoffBuilder) WithMaxDelay(delay time.Duration) *exponentialBackoffBuilder {
	b.backoff.maxDelay = delay
	return b
}

func (b *exponentialBackoffBuilder) WithJitter(jitter bool) *exponentialBackoffBuilder {
	b.backoff.jitter = jitter
	return b
}

func (b *exponentialBackoffBuilder) Build() BackoffStrategy {
	// Default to 100 milliseconds
	if b.backoff.baseDelay <= 0 {
		b.backoff.baseDelay = DefaultExponentialBaseDelay
	}
	// Default multiplier of 2
	if b.backoff.multiplier <= 0 {
		b.backoff.multiplier = DefaultExponentialMultiplier
	}
	// Default maxDelay to 10 seconds
	if b.backoff.maxDelay <= 0 {
		b.backoff.maxDelay = DefaultExponentialMaxDelay
	}
	return b.backoff
}

func (e *ExponentialBackoff) Delay(count int) time.Duration {
	multiplier := math.Pow(e.multiplier, float64(count))
	delay := time.Duration(float64(e.baseDelay) * multiplier)
	if delay > e.maxDelay {
		delay = e.maxDelay
	}

	// Check if we should add a jitter
	if e.jitter {
		delay = time.Duration(float64(delay) + jitter25(delay))
		if delay < 0 {
			delay = e.baseDelay
		}
	}
	return delay
}

type FixedBackoff struct {
	delay  time.Duration
	jitter bool
}

func (f *FixedBackoff) Delay(count int) time.Duration {
	delay := f.delay
	if f.jitter {
		delay = time.Duration(float64(delay) + jitter25(delay))
		if delay < 0 {
			delay = f.delay
		}
	}
	return delay
}

type fixedBackoffBuilder struct {
	backoff *FixedBackoff
}

// NewFixedBackoffBuilder creates a new builder for FixedBackoff.
// The builder is not safe for concurrent use. Each goroutine should create
// its own builder instance.
func NewFixedBackoffBuilder() *fixedBackoffBuilder {
	return &fixedBackoffBuilder{
		backoff: &FixedBackoff{},
	}
}

func (b *fixedBackoffBuilder) WithDelay(delay time.Duration) *fixedBackoffBuilder {
	b.backoff.delay = delay
	return b
}

func (b *fixedBackoffBuilder) WithJitter(jitter bool) *fixedBackoffBuilder {
	b.backoff.jitter = jitter
	return b
}

func (b *fixedBackoffBuilder) Build() BackoffStrategy {
	// Default delay to 1 second
	if b.backoff.delay <= 0 {
		b.backoff.delay = DefaultFixedBackoffDelay
	}
	return b.backoff
}

func jitter25(delay time.Duration) float64 {
	window := float64(delay) * DefaultJitterPercent
	// #nosec G404 - Using math/rand for jitter is acceptable, crypto/rand not needed
	return (rand.Float64() - JitterCenterPoint) * JitterMultiplier * window
}
