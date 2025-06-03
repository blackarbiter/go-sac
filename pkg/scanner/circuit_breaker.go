package scanner

import (
	"sync"
	"sync/atomic"
	"time"
)

// ErrorType 定义错误类型
type ErrorType int

const (
	TransientError ErrorType = iota // 临时错误，如网络超时
	CriticalError                   // 严重错误，如系统错误
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	transientFailures uint32        // 临时错误计数，如网络超时、临时连接失败等
	criticalFailures  uint32        // 严重错误计数，如系统错误、权限错误等
	lastFailure       time.Time     // 最后一次失败的时间戳
	threshold         uint32        // 总错误阈值，超过此值将触发熔断
	criticalThreshold uint32        // 严重错误阈值，超过此值将立即触发熔断
	resetTimeout      time.Duration // 熔断器重置超时时间，超过此时间后尝试恢复
	mu                sync.RWMutex  // 并发控制锁
}

// NewCircuitBreaker creates a new circuit breaker
// threshold: 总错误阈值，当 transientFailures + criticalFailures >= threshold 时触发熔断
// criticalThreshold: 严重错误阈值，当 criticalFailures >= criticalThreshold 时立即触发熔断
// resetTimeout: 熔断器重置超时时间，超过此时间后尝试恢复服务
func NewCircuitBreaker(threshold, criticalThreshold uint32, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:         threshold,
		criticalThreshold: criticalThreshold,
		resetTimeout:      resetTimeout,
	}
}

// IsOpen checks if the circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	/**
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// 检查严重错误是否超过阈值
	if atomic.LoadUint32(&cb.criticalFailures) >= cb.criticalThreshold {
		return true
	}

	// 检查总错误数是否超过阈值
	totalFailures := atomic.LoadUint32(&cb.transientFailures) + atomic.LoadUint32(&cb.criticalFailures)
	if totalFailures >= cb.threshold {
		// 检查重置超时是否已过
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.reset()
			return false
		}
		return true
	}*/
	return false
}

// RecordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) RecordFailure(errType ErrorType) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch errType {
	case TransientError:
		atomic.AddUint32(&cb.transientFailures, 1)
	case CriticalError:
		atomic.AddUint32(&cb.criticalFailures, 1)
		cb.lastFailure = time.Now()
	}
}

// RecordSuccess records a success and potentially closes the circuit
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	atomic.StoreUint32(&cb.transientFailures, 0)
	atomic.StoreUint32(&cb.criticalFailures, 0)
}

// reset resets the circuit breaker state
func (cb *CircuitBreaker) reset() {
	atomic.StoreUint32(&cb.transientFailures, 0)
	atomic.StoreUint32(&cb.criticalFailures, 0)
	cb.lastFailure = time.Time{}
}

// GetFailureCount returns the current failure counts
func (cb *CircuitBreaker) GetFailureCount() (transient, critical uint32) {
	return atomic.LoadUint32(&cb.transientFailures), atomic.LoadUint32(&cb.criticalFailures)
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

	if atomic.LoadUint32(&cb.criticalFailures) >= cb.criticalThreshold {
		return StateOpen
	}

	totalFailures := atomic.LoadUint32(&cb.transientFailures) + atomic.LoadUint32(&cb.criticalFailures)
	if totalFailures >= cb.threshold {
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
