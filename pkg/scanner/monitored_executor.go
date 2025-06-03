package scanner

import (
	"context"
	"runtime"
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
	// 记录队列等待开始时间
	queueStart := time.Now()

	// 记录任务开始时的内存使用情况
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	startMemUsage := memStats.Alloc

	// 执行任务
	start := time.Now()
	handle, err := m.executor.AsyncExecute(ctx, task)
	duration := time.Since(start)
	queueWaitTime := time.Since(queueStart)

	// 记录任务结束时的内存使用情况
	runtime.ReadMemStats(&memStats)
	endMemUsage := memStats.Alloc
	memoryDelta := endMemUsage - startMemUsage

	// 记录基本指标
	tags := map[string]string{
		"scanner_type": m.scanType.String(),
		"task_id":      task.TaskID,
		"asset_type":   string(task.AssetType),
		"asset_id":     task.AssetID,
		"error_type":   "none",
	}

	if err != nil {
		tags["error_type"] = "execution_error"
		m.metrics.RecordExecutorFailure(m.executor.Meta().Type)
		m.circuitBreaker.RecordFailure(TransientError)
	} else {
		m.circuitBreaker.RecordSuccess()
	}

	// 记录队列等待时间
	m.metrics.Record("queue_wait_seconds", queueWaitTime.Seconds(), tags)

	// 记录执行时间
	m.metrics.Record("execution_time_seconds", duration.Seconds(), tags)

	// 记录内存使用情况
	m.metrics.Gauge("memory_usage_bytes", float64(memoryDelta), tags)

	// 记录任务状态
	if err != nil {
		m.metrics.Record("task_errors", 1, tags)
	} else {
		m.metrics.Record("task_success", 1, tags)
	}

	// 记录熔断器状态
	circuitState := m.circuitBreaker.GetState()
	m.metrics.Gauge("circuit_breaker_state", float64(circuitState), tags)

	// 记录执行器健康状态
	if healthErr := m.executor.HealthCheck(); healthErr != nil {
		m.metrics.Record("executor_health_errors", 1, tags)
	} else {
		m.metrics.Record("executor_health_success", 1, tags)
	}

	return handle, err
}

// Cancel 实现 TaskExecutor 接口
func (m *MonitoredExecutor) Cancel(handle string) error {
	start := time.Now()
	err := m.executor.Cancel(handle)
	duration := time.Since(start)

	tags := map[string]string{
		"scanner_type": m.scanType.String(),
		"handle":       handle,
	}

	// 记录取消操作时间
	m.metrics.Record("cancel_operation_seconds", duration.Seconds(), tags)

	// 记录取消操作状态
	if err != nil {
		m.metrics.Record("cancel_errors", 1, tags)
	} else {
		m.metrics.Record("cancel_success", 1, tags)
	}

	return err
}

// GetStatus 实现 TaskExecutor 接口
func (m *MonitoredExecutor) GetStatus(handle string) (domain.TaskStatus, error) {
	start := time.Now()
	status, err := m.executor.GetStatus(handle)
	duration := time.Since(start)

	tags := map[string]string{
		"scanner_type": m.scanType.String(),
		"handle":       handle,
		"status":       string(status),
	}

	// 记录状态查询时间
	m.metrics.Record("status_query_seconds", duration.Seconds(), tags)

	// 记录状态查询结果
	if err != nil {
		m.metrics.Record("status_query_errors", 1, tags)
	} else {
		m.metrics.Record("status_query_success", 1, tags)
	}

	return status, err
}

// HealthCheck 实现 TaskExecutor 接口
func (m *MonitoredExecutor) HealthCheck() error {
	start := time.Now()
	err := m.executor.HealthCheck()
	duration := time.Since(start)

	tags := map[string]string{
		"scanner_type": m.scanType.String(),
	}

	// 记录健康检查时间
	m.metrics.Record("health_check_seconds", duration.Seconds(), tags)

	// 记录健康检查结果
	if err != nil {
		m.metrics.Record("health_check_errors", 1, tags)
	} else {
		m.metrics.Record("health_check_success", 1, tags)
	}

	return err
}

// Meta 实现 TaskExecutor 接口
func (m *MonitoredExecutor) Meta() ExecutorMeta {
	return m.executor.Meta()
}

// SyncExecute 实现 TaskExecutor 接口
func (m *MonitoredExecutor) SyncExecute(ctx context.Context, task *domain.ScanTaskPayload) (*domain.ScanResult, error) {
	// 记录队列等待开始时间
	queueStart := time.Now()

	// 记录任务开始时的内存使用情况
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	startMemUsage := memStats.Alloc

	// 执行任务
	start := time.Now()
	result, err := m.executor.SyncExecute(ctx, task)
	duration := time.Since(start)
	queueWaitTime := time.Since(queueStart)

	// 记录任务结束时的内存使用情况
	runtime.ReadMemStats(&memStats)
	endMemUsage := memStats.Alloc
	memoryDelta := endMemUsage - startMemUsage

	// 记录基本指标
	tags := map[string]string{
		"scanner_type": m.scanType.String(),
		"task_id":      task.TaskID,
		"asset_type":   string(task.AssetType),
		"asset_id":     task.AssetID,
		"error_type":   "none",
	}

	if err != nil {
		tags["error_type"] = "execution_error"
		m.metrics.RecordExecutorFailure(m.executor.Meta().Type)
		m.circuitBreaker.RecordFailure(TransientError)
	} else {
		m.circuitBreaker.RecordSuccess()
	}

	// 记录队列等待时间
	m.metrics.Record("queue_wait_seconds", queueWaitTime.Seconds(), tags)

	// 记录执行时间
	m.metrics.Record("execution_time_seconds", duration.Seconds(), tags)

	// 记录内存使用情况
	m.metrics.Gauge("memory_usage_bytes", float64(memoryDelta), tags)

	// 记录任务状态
	if err != nil {
		m.metrics.Record("task_errors", 1, tags)
	} else {
		m.metrics.Record("task_success", 1, tags)
	}

	// 记录熔断器状态
	circuitState := m.circuitBreaker.GetState()
	m.metrics.Gauge("circuit_breaker_state", float64(circuitState), tags)

	// 记录执行器健康状态
	if healthErr := m.executor.HealthCheck(); healthErr != nil {
		m.metrics.Record("executor_health_errors", 1, tags)
	} else {
		m.metrics.Record("executor_health_success", 1, tags)
	}

	return result, err
}
