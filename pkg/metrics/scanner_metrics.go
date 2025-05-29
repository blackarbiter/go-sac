package metrics

import (
	"time"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// ScannerExecutionTime 记录扫描器执行时间
	ScannerExecutionTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "scanner_execution_time_seconds",
			Help:    "Scanner execution time in seconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 10),
		},
		[]string{"scan_type"},
	)

	// ScannerExecutorFailures 记录扫描器执行器失败次数
	ScannerExecutorFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scanner_executor_failures_total",
			Help: "Total number of scanner executor failures",
		},
		[]string{"executor_type"},
	)

	// ScannerTimeouts 记录扫描器超时次数
	ScannerTimeouts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scanner_timeouts_total",
			Help: "Total number of scanner timeouts",
		},
		[]string{"scan_type", "severity"},
	)

	// ScannerCriticalTimeouts 记录扫描器严重超时次数
	ScannerCriticalTimeouts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scanner_critical_timeouts_total",
			Help: "Total number of critical scanner timeouts",
		},
		[]string{"scan_type"},
	)
)

// ScannerMetrics 实现扫描器指标收集
type ScannerMetrics struct{}

// NewScannerMetrics 创建新的扫描器指标收集器
func NewScannerMetrics() *ScannerMetrics {
	return &ScannerMetrics{}
}

// Register 注册指标
func (m *ScannerMetrics) Register() {
	prometheus.MustRegister(ScannerExecutionTime)
	prometheus.MustRegister(ScannerExecutorFailures)
	prometheus.MustRegister(ScannerTimeouts)
	prometheus.MustRegister(ScannerCriticalTimeouts)
}

// Record 记录指标
func (m *ScannerMetrics) Record(name string, value float64, tags map[string]string) {
	switch name {
	case "command_duration_seconds":
		ScannerExecutionTime.WithLabelValues(tags["scanner_type"]).Observe(value)
	case "command_errors":
		ScannerExecutorFailures.WithLabelValues(tags["scanner_type"]).Inc()
	}
}

// Gauge 设置仪表盘指标
func (m *ScannerMetrics) Gauge(name string, value float64, tags map[string]string) {
	// 目前没有仪表盘指标需要实现
}

// RecordExecutionTime 记录执行时间
func (m *ScannerMetrics) RecordExecutionTime(scanType domain.ScanType, duration time.Duration) {
	ScannerExecutionTime.WithLabelValues(string(scanType)).Observe(duration.Seconds())
}

// RecordExecutorFailure 记录执行器失败
func (m *ScannerMetrics) RecordExecutorFailure(executorType string) {
	ScannerExecutorFailures.WithLabelValues(executorType).Inc()
}

// RecordTimeout 记录超时事件
func (m *ScannerMetrics) RecordTimeout(scanType domain.ScanType, severity string, isHard bool) {
	ScannerTimeouts.WithLabelValues(string(scanType), severity).Inc()
}

// RecordCriticalTimeout 记录严重超时事件
func (m *ScannerMetrics) RecordCriticalTimeout(scanType domain.ScanType) {
	ScannerCriticalTimeouts.WithLabelValues(string(scanType)).Inc()
}
