// pkg/storage/mysql/monitor.go
package mysql

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	// 全局指标注册，确保在应用启动时注册
	defaultQueryHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "mysql_query_duration_seconds",
		Help:    "MySQL查询执行时间直方图（秒）",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
	})

	defaultQueryCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mysql_query_total",
		Help: "MySQL查询总数",
	}, []string{"status"})

	defaultQueryTypeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mysql_query_type_total",
		Help: "按类型统计的MySQL查询数量",
	}, []string{"type"})
)

type MonitorConfig struct {
	SlowQueryThreshold time.Duration // 慢查询阈值
	EnableMetrics      bool          // 是否启用指标收集
	MetricsPrefix      string        // 指标名称前缀
}

type QueryMonitor struct {
	logger           *zap.Logger
	config           MonitorConfig
	slowQueryMetric  prometheus.Histogram
	queryCounter     *prometheus.CounterVec
	queryTypeCounter *prometheus.CounterVec
	connectionGauge  prometheus.Gauge
}

// 查询类型
const (
	QueryTypeSelect = "select"
	QueryTypeInsert = "insert"
	QueryTypeUpdate = "update"
	QueryTypeDelete = "delete"
	QueryTypeOther  = "other"
)

// 查询状态
const (
	QueryStatusSuccess = "success"
	QueryStatusError   = "error"
	QueryStatusSlow    = "slow"
)

func NewMonitor(cfg MonitorConfig, logger *zap.Logger) *QueryMonitor {
	prefix := cfg.MetricsPrefix
	if prefix != "" {
		prefix += "_"
	}

	var slowQueryMetric prometheus.Histogram
	var queryCounter *prometheus.CounterVec
	var queryTypeCounter *prometheus.CounterVec
	var connectionGauge prometheus.Gauge

	if cfg.EnableMetrics {
		// 如果启用度量，创建自定义的指标
		slowQueryMetric = promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    prefix + "mysql_slow_queries_seconds",
			Help:    "MySQL慢查询统计（秒）",
			Buckets: []float64{0.1, 0.5, 1, 2, 5},
		})

		queryCounter = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: prefix + "mysql_query_total",
			Help: "MySQL查询总数",
		}, []string{"status"})

		queryTypeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: prefix + "mysql_query_type_total",
			Help: "按类型统计的MySQL查询数量",
		}, []string{"type"})

		connectionGauge = promauto.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "mysql_connections",
			Help: "MySQL当前连接数",
		})
	} else {
		// 如果未启用自定义度量，使用默认全局指标
		slowQueryMetric = defaultQueryHistogram
		queryCounter = defaultQueryCounter
		queryTypeCounter = defaultQueryTypeCounter
		connectionGauge = prometheus.NewGauge(prometheus.GaugeOpts{Name: "unused_gauge"}) // 未使用的空指标
	}

	return &QueryMonitor{
		logger:           logger.Named("mysql.monitor"),
		config:           cfg,
		slowQueryMetric:  slowQueryMetric,
		queryCounter:     queryCounter,
		queryTypeCounter: queryTypeCounter,
		connectionGauge:  connectionGauge,
	}
}

func (m *QueryMonitor) MonitorQuery(ctx context.Context, query string, args []interface{}, execTime time.Duration, err error) {
	// 统计查询类型
	queryType := detectQueryType(query)
	m.queryTypeCounter.WithLabelValues(queryType).Inc()

	// 记录执行时间
	m.slowQueryMetric.Observe(execTime.Seconds())

	// 检查是否为慢查询
	if execTime > m.config.SlowQueryThreshold {
		m.logger.Warn("检测到慢查询",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Duration("exec_time", execTime))

		m.queryCounter.WithLabelValues(QueryStatusSlow).Inc()
	}

	// 记录查询状态
	if err != nil {
		m.queryCounter.WithLabelValues(QueryStatusError).Inc()
		m.logger.Error("查询执行失败",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
	} else {
		m.queryCounter.WithLabelValues(QueryStatusSuccess).Inc()
	}
}

func (m *QueryMonitor) WrapQuery(ctx context.Context, query string, args []interface{}, execFunc func() error) error {
	startTime := time.Now()
	err := execFunc()
	execTime := time.Since(startTime)

	go m.MonitorQuery(ctx, query, args, execTime, err)
	return err
}

// SetConnectionCount 设置当前连接数
func (m *QueryMonitor) SetConnectionCount(count float64) {
	if m.config.EnableMetrics {
		m.connectionGauge.Set(count)
	}
}

// 检测SQL查询类型
func detectQueryType(query string) string {
	// 简单的SQL类型检测，可以根据需要进行扩展
	if len(query) < 7 {
		return QueryTypeOther
	}

	firstWord := query[0:6]
	switch firstWord {
	case "SELECT", "select":
		return QueryTypeSelect
	case "INSERT", "insert":
		return QueryTypeInsert
	case "UPDATE", "update":
		return QueryTypeUpdate
	case "DELETE", "delete":
		return QueryTypeDelete
	default:
		return QueryTypeOther
	}
}
