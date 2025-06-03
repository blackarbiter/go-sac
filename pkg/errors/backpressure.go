package errors

import (
	"fmt"
	"time"
)

// BackpressureError represents a system backpressure error
type BackpressureError struct {
	QueueSize   int
	RequestedAt time.Time
}

// NewBackpressureError creates a new BackpressureError
func NewBackpressureError(queueSize int) error {
	return &BackpressureError{
		QueueSize:   queueSize,
		RequestedAt: time.Now(),
	}
}

// Error implements the error interface
func (e *BackpressureError) Error() string {
	return fmt.Sprintf("system backpressure (queue_size=%d, requested_at=%s)",
		e.QueueSize, e.RequestedAt.Format(time.RFC3339))
}

// IsBackpressureError checks if an error is a BackpressureError
func IsBackpressureError(err error) bool {
	_, ok := err.(*BackpressureError)
	return ok
}
