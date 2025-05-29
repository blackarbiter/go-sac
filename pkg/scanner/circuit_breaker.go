package scanner

import (
	"sync"
	"sync/atomic"
	"time"
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	failures     uint32
	lastFailure  time.Time
	threshold    uint32
	resetTimeout time.Duration
	mu           sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold uint32, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:    threshold,
		resetTimeout: resetTimeout,
	}
}

// IsOpen checks if the circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if atomic.LoadUint32(&cb.failures) >= cb.threshold {
		// Check if reset timeout has elapsed
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.reset()
			return false
		}
		return true
	}
	return false
}

// RecordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	atomic.AddUint32(&cb.failures, 1)
	cb.lastFailure = time.Now()
}

// RecordSuccess records a success and potentially closes the circuit
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	atomic.StoreUint32(&cb.failures, 0)
}

// reset resets the circuit breaker state
func (cb *CircuitBreaker) reset() {
	atomic.StoreUint32(&cb.failures, 0)
	cb.lastFailure = time.Time{}
}

// GetFailureCount returns the current failure count
func (cb *CircuitBreaker) GetFailureCount() uint32 {
	return atomic.LoadUint32(&cb.failures)
}

// GetLastFailureTime returns the time of the last failure
func (cb *CircuitBreaker) GetLastFailureTime() time.Time {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.lastFailure
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if atomic.LoadUint32(&cb.failures) >= cb.threshold {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			return StateHalfOpen
		}
		return StateOpen
	}
	return StateClosed
}

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// String returns the string representation of the state
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}
