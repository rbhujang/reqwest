package reqwest

import (
	"testing"
	"time"
)

func TestRetryConfigBuilder(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		config := NewRetryConfigBuilder().Build()

		if config.maxRetries != DefaultMaxRetries {
			t.Errorf("Expected maxRetries %v, got %v", DefaultMaxRetries, config.maxRetries)
		}

		if config.backoffStrategy == nil {
			t.Error("Expected default backoff strategy, got nil")
		}

		if len(config.retryableStatusCodes) != len(DefaultRetryableStatusCodes) {
			t.Errorf("Expected %d retryable status codes, got %d", len(DefaultRetryableStatusCodes),
				len(config.retryableStatusCodes))
		}

		if len(config.retryableError) != len(DefaultRetryableErrors) {
			t.Errorf("Expected %d retryable errors, got %d", len(DefaultRetryableErrors),
				len(config.retryableError))
		}
	})

	t.Run("custom values", func(t *testing.T) {
		customBackoff := NewFixedBackoffBuilder().WithDelay(1 * time.Second).Build()
		customCodes := []int{500, 503}
		customErrors := map[string]bool{"timeout": true}

		config := NewRetryConfigBuilder().
			WithMaxRetries(5).
			WithBackoffStrategy(customBackoff).
			WithRetryableStatusCodes(customCodes).
			WithRetryableErrors(customErrors).
			Build()

		if config.maxRetries != 5 {
			t.Errorf("Expected maxRetries 5, got %v", config.maxRetries)
		}

		if config.backoffStrategy != customBackoff {
			t.Error("Expected custom backoff strategy")
		}

		if len(config.retryableStatusCodes) != 2 {
			t.Errorf("Expected 2 retryable status codes, got %d", len(config.retryableStatusCodes))
		}

		if len(config.retryableError) != 1 {
			t.Errorf("Expected 1 retryable error, got %d", len(config.retryableError))
		}
	})

	t.Run("zero and negative maxRetries gets default", func(t *testing.T) {
		config := NewRetryConfigBuilder().
			WithMaxRetries(0).
			Build()

		if config.maxRetries != DefaultMaxRetries {
			t.Errorf("Expected zero maxRetries to use default %v, got %v",
				DefaultMaxRetries, config.maxRetries)
		}

		config2 := NewRetryConfigBuilder().
			WithMaxRetries(-5).
			Build()

		if config2.maxRetries != DefaultMaxRetries {
			t.Errorf("Expected negative maxRetries to use default %v, got %v",
				DefaultMaxRetries, config2.maxRetries)
		}
	})

	t.Run("empty slices and maps get defaults", func(t *testing.T) {
		config := NewRetryConfigBuilder().
			WithRetryableStatusCodes([]int{}).
			WithRetryableErrors(map[string]bool{}).
			Build()

		if len(config.retryableStatusCodes) != len(DefaultRetryableStatusCodes) {
			t.Errorf("Expected empty status codes to use defaults, got %d codes",
				len(config.retryableStatusCodes))
		}

		if len(config.retryableError) != len(DefaultRetryableErrors) {
			t.Errorf("Expected empty errors to use defaults, got %d errors",
				len(config.retryableError))
		}
	})
}

func TestBackoffStrategy_EdgeCases(t *testing.T) {
	t.Run("ExponentialBackoff with zero attempt", func(t *testing.T) {
		backoff := &ExponentialBackoff{
			baseDelay:  100 * time.Millisecond,
			multiplier: 2.0,
			maxDelay:   1 * time.Second,
			jitter:     false,
		}

		delay := backoff.Delay(0)
		expected := 100 * time.Millisecond // 100ms * 2^0 = 100ms
		if delay != expected {
			t.Errorf("Delay(0) = %v, want %v", delay, expected)
		}
	})

	t.Run("FixedBackoff with zero attempt", func(t *testing.T) {
		backoff := &FixedBackoff{
			delay:  500 * time.Millisecond,
			jitter: false,
		}

		delay := backoff.Delay(0)
		if delay != 500*time.Millisecond {
			t.Errorf("Delay(0) = %v, want %v", delay, 500*time.Millisecond)
		}
	})

	t.Run("ExponentialBackoff with negative jitter result", func(t *testing.T) {
		backoff := &ExponentialBackoff{
			baseDelay:  10 * time.Millisecond,
			multiplier: 2.0,
			maxDelay:   1 * time.Second,
			jitter:     true,
		}

		delay := backoff.Delay(1)
		if delay < 0 {
			t.Errorf("Delay should never be negative, got %v", delay)
		}
	})
}

func TestExponentialBackoff_Delay(t *testing.T) {
	t.Run("Without Jitter", func(t *testing.T) {
		backoff := &ExponentialBackoff{
			baseDelay:  100 * time.Millisecond,
			multiplier: 2.0,
			maxDelay:   1 * time.Second,
			jitter:     false,
		}

		tests := []struct {
			name     string
			attempt  int
			expected time.Duration
		}{
			{"first attempt", 1, 200 * time.Millisecond},  // 100ms * 2^1
			{"second attempt", 2, 400 * time.Millisecond}, // 100ms * 2^2
			{"third attempt", 3, 800 * time.Millisecond},  // 100ms * 2^3
			{"max delay reached", 4, 1 * time.Second},     // Should cap at maxDelay
			{"beyond max", 10, 1 * time.Second},           // Should stay at maxDelay
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				actual := backoff.Delay(tt.attempt)
				if actual != tt.expected {
					t.Errorf("Delay(%d) = %v, want %v", tt.attempt, actual, tt.expected)
				}
			})
		}
	})

	t.Run("With Jitter", func(t *testing.T) {
		backoff := &ExponentialBackoff{
			baseDelay:  100 * time.Millisecond,
			multiplier: 2.0,
			maxDelay:   1 * time.Second,
			jitter:     true,
		}

		attempt := 2
		expectedBase := 400 * time.Millisecond // 100ms * 2^2

		// Test multiple times to ensure jitter is working
		delays := make([]time.Duration, 10)
		for i := 0; i < 10; i++ {
			delays[i] = backoff.Delay(attempt)
		}

		// They all should not be the same
		allSame := true
		for i := 1; i < len(delays); i++ {
			if delays[i] != delays[0] {
				allSame = false
				break
			}
		}
		if allSame {
			t.Error("Expected jitter to produce varying delays, but all delays were the same")
		}

		// Delay should be in range
		minExpected := time.Duration(float64(expectedBase) * 0.75)
		maxExpected := time.Duration(float64(expectedBase) * 1.25)

		for i, delay := range delays {
			if delay < minExpected || delay > maxExpected {
				t.Errorf("Delay[%d] = %v, expected between %v and %v", i, delay, minExpected, maxExpected)
			}
		}
	})
}

func TestFixedBackoff_Delay(t *testing.T) {
	tests := []struct {
		name       string
		backoff    *FixedBackoff
		attempt    int
		want       time.Duration
		testJitter bool
	}{
		{
			name: "fixed delay without jitter",
			backoff: &FixedBackoff{
				delay:  500 * time.Millisecond,
				jitter: false,
			},
			attempt:    1,
			want:       500 * time.Millisecond,
			testJitter: false,
		},
		{
			name: "fixed delay same for different attempts",
			backoff: &FixedBackoff{
				delay:  300 * time.Millisecond,
				jitter: false,
			},
			attempt:    5,
			want:       300 * time.Millisecond,
			testJitter: false,
		},
		{
			name: "fixed delay with jitter",
			backoff: &FixedBackoff{
				delay:  500 * time.Millisecond,
				jitter: true,
			},
			attempt:    1,
			want:       500 * time.Millisecond, // Base value for jitter range
			testJitter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.testJitter {
				got := tt.backoff.Delay(tt.attempt)
				if got != tt.want {
					t.Errorf("FixedBackoff.Delay() = %v, want %v", got, tt.want)
				}
			} else {
				// Test jitter behavior
				delays := make([]time.Duration, 10)
				for i := 0; i < 10; i++ {
					delays[i] = tt.backoff.Delay(tt.attempt)
				}

				// Check that delays vary
				allSame := true
				for i := 1; i < len(delays); i++ {
					if delays[i] != delays[0] {
						allSame = false
						break
					}
				}
				if allSame {
					t.Error("Expected jitter to produce varying delays")
				}

				// Check range (±25% of base)
				minExpected := time.Duration(float64(tt.want) * 0.75)
				maxExpected := time.Duration(float64(tt.want) * 1.25)

				for i, delay := range delays {
					if delay < minExpected || delay > maxExpected {
						t.Errorf("Delay[%d] = %v, expected between %v and %v",
							i, delay, minExpected, maxExpected)
					}
				}
			}
		})
	}
}

func TestExponentialBackoffBuilder(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		backoff := NewExponentialBackoffBuilder().Build().(*ExponentialBackoff)
		if backoff.baseDelay != DefaultExponentialBaseDelay {
			t.Errorf("Expected baseDelay %v, got %v", DefaultExponentialBaseDelay, backoff.baseDelay)
		}
		if backoff.multiplier != DefaultExponentialMultiplier {
			t.Errorf("Expected multiplier %v, got %v", DefaultExponentialMultiplier, backoff.multiplier)
		}
		if backoff.maxDelay != DefaultExponentialMaxDelay {
			t.Errorf("Expected maxDelay %v, got %v", DefaultExponentialMaxDelay, backoff.maxDelay)
		}
		if backoff.jitter != false {
			t.Errorf("Expected jitter %v, got %v", false, backoff.jitter)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		backoff := NewExponentialBackoffBuilder().
			WithBaseDelay(200 * time.Millisecond).
			WithMultiplier(3.0).
			WithMaxDelay(5 * time.Second).
			WithJitter(true).
			Build().(*ExponentialBackoff)

		if backoff.baseDelay != 200*time.Millisecond {
			t.Errorf("Expected baseDelay %v, got %v", 200*time.Millisecond, backoff.baseDelay)
		}
		if backoff.multiplier != 3.0 {
			t.Errorf("Expected multiplier %v, got %v", 3.0, backoff.multiplier)
		}
		if backoff.maxDelay != 5*time.Second {
			t.Errorf("Expected maxDelay %v, got %v", 5*time.Second, backoff.maxDelay)
		}
		if backoff.jitter != true {
			t.Errorf("Expected jitter %v, got %v", true, backoff.jitter)
		}
	})

	t.Run("zero and negative values should default", func(t *testing.T) {
		backoff := NewExponentialBackoffBuilder().
			WithBaseDelay(0).
			WithMultiplier(-1.0).
			WithMaxDelay(-5 * time.Second).
			Build().(*ExponentialBackoff)

		if backoff.baseDelay != DefaultExponentialBaseDelay {
			t.Errorf("Expected zero baseDelay to use default %v, got %v",
				DefaultExponentialBaseDelay, backoff.baseDelay)
		}

		if backoff.multiplier != DefaultExponentialMultiplier {
			t.Errorf("Expected negative multiplier to use default %v, got %v",
				DefaultExponentialMultiplier, backoff.multiplier)
		}

		if backoff.maxDelay != DefaultExponentialMaxDelay {
			t.Errorf("Expected negative maxDelay to use default %v, got %v",
				DefaultExponentialMaxDelay, backoff.maxDelay)
		}
	})
}

func TestFixedBackoffBuilder(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		backoff := NewFixedBackoffBuilder().Build().(*FixedBackoff)

		if backoff.delay != DefaultFixedBackoffDelay {
			t.Errorf("Expected delay %v, got %v", DefaultFixedBackoffDelay, backoff.delay)
		}
		if backoff.jitter != false {
			t.Errorf("Expected jitter %v, got %v", false, backoff.jitter)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		backoff := NewFixedBackoffBuilder().
			WithDelay(750 * time.Millisecond).
			WithJitter(true).
			Build().(*FixedBackoff)

		if backoff.delay != 750*time.Millisecond {
			t.Errorf("Expected delay %v, got %v", 750*time.Millisecond, backoff.delay)
		}
		if backoff.jitter != true {
			t.Errorf("Expected jitter %v, got %v", true, backoff.jitter)
		}
	})

	t.Run("zero delay gets default", func(t *testing.T) {
		backoff := NewFixedBackoffBuilder().
			WithDelay(0).
			Build().(*FixedBackoff)

		if backoff.delay != DefaultFixedBackoffDelay {
			t.Errorf("Expected zero delay to use default %v, got %v", DefaultFixedBackoffDelay,
				backoff.delay)
		}
	})
}

func TestJitter25(t *testing.T) {
	delay := 1 * time.Second

	results := make([]float64, 100)
	for i := 0; i < 100; i++ {
		results[i] = jitter25(delay)
	}

	// They should be different
	allSame := true
	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("Expected jitter25 to produce varying results, but all were the same")
	}

	// Jitter should be in range
	expectedMin := -0.25 * float64(delay)
	expectedMax := 0.25 * float64(delay)

	for i, result := range results {
		if result < expectedMin || result > expectedMax {
			t.Errorf("jitter25[%d] = %v, expected between %v and %v",
				i, result, expectedMin, expectedMax)
		}
	}

	// Test with zero delay
	zeroResult := jitter25(0)
	if zeroResult != 0 {
		t.Errorf("Expected jitter25(0) = 0, got %v", zeroResult)
	}
}
