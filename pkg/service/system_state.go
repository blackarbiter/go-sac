package service

import (
	"sync"
	"time"

	"github.com/blackarbiter/go-sac/pkg/logger"
	"go.uber.org/zap"
)

// SystemState manages the global system state for backpressure control
type SystemState struct {
	mu               sync.RWMutex
	globalQueueFull  bool
	lastFullTime     time.Time
	consumingStopped bool
}

// NewSystemState creates a new instance of SystemState
func NewSystemState() *SystemState {
	return &SystemState{}
}

// TriggerBackpressure triggers the backpressure state (idempotent operation)
func (s *SystemState) TriggerBackpressure() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.consumingStopped {
		s.globalQueueFull = true
		s.consumingStopped = true
		s.lastFullTime = time.Now()
		logger.Logger.Info("Backpressure triggered",
			zap.Time("lastFullTime", s.lastFullTime))
	}
}

// ReleaseBackpressure releases the backpressure state (idempotent operation)
func (s *SystemState) ReleaseBackpressure() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.consumingStopped {
		s.globalQueueFull = false
		s.consumingStopped = false
		logger.Logger.Info("Backpressure released",
			zap.Duration("duration", time.Since(s.lastFullTime)))
	}
}

// ShouldStopProcessing checks if processing should be stopped
func (s *SystemState) ShouldStopProcessing() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.consumingStopped
}

// ShouldStopConsuming checks if message consumption should be stopped
func (s *SystemState) ShouldStopConsuming() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.consumingStopped
}

// IsBackpressureActive checks if backpressure is currently active
func (s *SystemState) IsBackpressureActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.consumingStopped
}

// GetLastFullTime returns the last time the queue was full
func (s *SystemState) GetLastFullTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastFullTime
}
