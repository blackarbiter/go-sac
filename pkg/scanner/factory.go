package scanner

import (
	"fmt"
	"sync"
	"time"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/logger"
	"github.com/blackarbiter/go-sac/pkg/metrics"
	"go.uber.org/zap"
)

// ScannerFactory defines the interface for scanner factory
type ScannerFactory interface {
	RegisterExecutor(scanType domain.ScanType, executor TaskExecutor)
	GetScanner(scanType domain.ScanType) (TaskExecutor, error)
	GetMetrics() *metrics.ScannerMetrics
	GetCircuitBreaker() *CircuitBreaker
	ListSupportedTypes() []domain.ScanType
	HealthCheck() error
	Close() error
	GetAllScanners() map[domain.ScanType]TaskExecutor
}

// ScannerFactoryImpl creates and manages scanner instances
type ScannerFactoryImpl struct {
	scanners       map[domain.ScanType]TaskExecutor
	metrics        *metrics.ScannerMetrics
	circuitBreaker *CircuitBreaker
	logger         *zap.Logger
	mu             sync.RWMutex
}

// NewScannerFactory creates a new scanner factory
func NewScannerFactory(createScanners func() map[domain.ScanType]TaskExecutor) *ScannerFactoryImpl {
	metrics := metrics.NewScannerMetrics()
	metrics.Register()

	return &ScannerFactoryImpl{
		scanners:       createScanners(),
		metrics:        metrics,
		circuitBreaker: NewCircuitBreaker(5, 3, 30*time.Second),
		logger:         logger.Logger,
	}
}

// RegisterExecutor registers a new executor for a scan type
func (f *ScannerFactoryImpl) RegisterExecutor(scanType domain.ScanType, executor TaskExecutor) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.scanners[scanType] = executor
	f.logger.Info("registered executor",
		zap.String("scan_type", string(scanType)),
		zap.String("executor_type", executor.Meta().Type),
	)
}

// GetScanner returns a scanner for the given scan type
func (f *ScannerFactoryImpl) GetScanner(scanType domain.ScanType) (TaskExecutor, error) {
	f.mu.RLock()
	executor, exists := f.scanners[scanType]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no executor registered for scan type: %s", scanType)
	}

	// Check circuit breaker
	if f.circuitBreaker.IsOpen() {
		return nil, fmt.Errorf("circuit breaker open for scan type: %s", scanType)
	}

	// 包装Executor添加监控
	return NewMonitoredExecutor(executor, f.metrics, f.circuitBreaker, scanType), nil
}

// GetMetrics returns the metrics collector
func (f *ScannerFactoryImpl) GetMetrics() *metrics.ScannerMetrics {
	return f.metrics
}

// GetCircuitBreaker returns the circuit breaker
func (f *ScannerFactoryImpl) GetCircuitBreaker() *CircuitBreaker {
	return f.circuitBreaker
}

// ListSupportedTypes returns all supported scan types
func (f *ScannerFactoryImpl) ListSupportedTypes() []domain.ScanType {
	f.mu.RLock()
	defer f.mu.RUnlock()

	types := make([]domain.ScanType, 0, len(f.scanners))
	for t := range f.scanners {
		types = append(types, t)
	}
	return types
}

// HealthCheck performs health check on all executors
func (f *ScannerFactoryImpl) HealthCheck() error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var lastErr error
	for scanType, executor := range f.scanners {
		if err := executor.HealthCheck(); err != nil {
			f.logger.Error("executor health check failed",
				zap.String("scan_type", string(scanType)),
				zap.Error(err),
			)
			lastErr = err
		}
	}
	return lastErr
}

// Close closes all executors
func (f *ScannerFactoryImpl) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var lastErr error
	for scanType, executor := range f.scanners {
		f.logger.Info("closing executor",
			zap.String("scan_type", string(scanType)),
			zap.String("executor_type", executor.Meta().Type),
		)
	}
	return lastErr
}

// GetAllScanners returns all scanners
func (f *ScannerFactoryImpl) GetAllScanners() map[domain.ScanType]TaskExecutor {
	return f.scanners
}
