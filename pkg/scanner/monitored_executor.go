package scanner

import (
	"context"
	"time"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/metrics"
)

// MonitoredExecutor 是一个装饰器，用于为扫描器添加监控和熔断功能
type MonitoredExecutor struct {
	executor       TaskExecutor
	metrics        *metrics.ScannerMetrics
	circuitBreaker *CircuitBreaker
	scanType       domain.ScanType
}

// NewMonitoredExecutor 创建一个新的监控装饰器
func NewMonitoredExecutor(executor TaskExecutor, metrics *metrics.ScannerMetrics, circuitBreaker *CircuitBreaker, scanType domain.ScanType) *MonitoredExecutor {
	return &MonitoredExecutor{
		executor:       executor,
		metrics:        metrics,
		circuitBreaker: circuitBreaker,
		scanType:       scanType,
	}
}

// AsyncExecute 实现 TaskExecutor 接口
func (m *MonitoredExecutor) AsyncExecute(ctx context.Context, task *domain.ScanTaskPayload) (string, error) {
	start := time.Now()
	handle, err := m.executor.AsyncExecute(ctx, task)
	duration := time.Since(start)

	if err != nil {
		m.metrics.RecordExecutorFailure(m.executor.Meta().Type)
		m.circuitBreaker.RecordFailure()
	} else {
		m.circuitBreaker.RecordSuccess()
	}

	m.metrics.RecordExecutionTime(m.scanType, duration)
	return handle, err
}

// Cancel 实现 TaskExecutor 接口
func (m *MonitoredExecutor) Cancel(handle string) error {
	return m.executor.Cancel(handle)
}

// GetStatus 实现 TaskExecutor 接口
func (m *MonitoredExecutor) GetStatus(handle string) (domain.TaskStatus, error) {
	return m.executor.GetStatus(handle)
}

// HealthCheck 实现 TaskExecutor 接口
func (m *MonitoredExecutor) HealthCheck() error {
	return m.executor.HealthCheck()
}

// Meta 实现 TaskExecutor 接口
func (m *MonitoredExecutor) Meta() ExecutorMeta {
	return m.executor.Meta()
}
