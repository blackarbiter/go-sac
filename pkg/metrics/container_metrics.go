package metrics

import (
	"time"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// ContainerCreateErrors 记录容器创建错误
	ContainerCreateErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "container_create_errors_total",
			Help: "Total container creation errors",
		},
		[]string{"scan_type"},
	)

	// ContainerStartErrors 记录容器启动错误
	ContainerStartErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "container_start_errors_total",
			Help: "Total container start errors",
		},
		[]string{"scan_type"},
	)

	// ContainerExecutionTime 记录容器执行时间
	ContainerExecutionTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "container_execution_time_seconds",
			Help:    "Container execution time distribution",
			Buckets: prometheus.ExponentialBuckets(1, 2, 10),
		},
		[]string{"scan_type"},
	)
)

// ContainerMetrics 实现容器指标收集
type ContainerMetrics struct{}

// NewContainerMetrics 创建新的容器指标收集器
func NewContainerMetrics() *ContainerMetrics {
	return &ContainerMetrics{}
}

// Register 注册指标
func (m *ContainerMetrics) Register() {
	prometheus.MustRegister(ContainerCreateErrors)
	prometheus.MustRegister(ContainerStartErrors)
	prometheus.MustRegister(ContainerExecutionTime)
}

// ContainerCreateError 记录容器创建错误
func (m *ContainerMetrics) ContainerCreateError(scanType domain.ScanType) {
	ContainerCreateErrors.WithLabelValues(string(scanType)).Inc()
}

// ContainerStartError 记录容器启动错误
func (m *ContainerMetrics) ContainerStartError(scanType domain.ScanType) {
	ContainerStartErrors.WithLabelValues(string(scanType)).Inc()
}

// ContainerExecutionTime 记录容器执行时间
func (m *ContainerMetrics) RecordExecutionTime(scanType domain.ScanType, duration time.Duration) {
	ContainerExecutionTime.WithLabelValues(string(scanType)).Observe(duration.Seconds())
}
