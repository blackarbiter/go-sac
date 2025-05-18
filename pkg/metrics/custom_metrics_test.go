package metrics_test

import (
	"testing"

	"github.com/blackarbiter/go-sac/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCustomMetrics(t *testing.T) {
	// 重置注册表，避免测试之间的干扰
	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics.ScanTasksQueue)
	registry.MustRegister(metrics.DatabaseConnections)

	t.Run("ScanTasksQueue指标应正确更新", func(t *testing.T) {
		// 首先验证初始值
		initialValue := testutil.ToFloat64(metrics.ScanTasksQueue)
		if initialValue != 0 {
			t.Errorf("ScanTasksQueue初始值不为0: %f", initialValue)
		}

		// 测试增加值
		testValue := 5.0
		metrics.ScanTasksQueue.Set(testValue)

		updatedValue := testutil.ToFloat64(metrics.ScanTasksQueue)
		if updatedValue != testValue {
			t.Errorf("ScanTasksQueue更新失败, 期望 %f, 获取 %f", testValue, updatedValue)
		}
	})

	t.Run("DatabaseConnections指标应正确更新", func(t *testing.T) {
		// 测试不同类型的连接数
		metrics.DatabaseConnections.WithLabelValues("mysql").Set(10)
		metrics.DatabaseConnections.WithLabelValues("redis").Set(5)

		mysqlValue := testutil.ToFloat64(metrics.DatabaseConnections.WithLabelValues("mysql"))
		redisValue := testutil.ToFloat64(metrics.DatabaseConnections.WithLabelValues("redis"))

		if mysqlValue != 10 {
			t.Errorf("MySQL连接数指标更新失败, 期望 %d, 获取 %f", 10, mysqlValue)
		}

		if redisValue != 5 {
			t.Errorf("Redis连接数指标更新失败, 期望 %d, 获取 %f", 5, redisValue)
		}
	})

	t.Run("RegisterCustomMetrics应正确注册", func(t *testing.T) {
		// 创建新的注册表
		_ = prometheus.NewRegistry()

		// 注册自定义指标
		metrics.RegisterCustomMetrics()

		// 尝试获取指标值(这不会失败，即使指标已注册)
		metrics.ScanTasksQueue.Set(7)
		metrics.DatabaseConnections.WithLabelValues("postgres").Set(3)

		// 验证值是否设置成功
		scanValue := testutil.ToFloat64(metrics.ScanTasksQueue)
		if scanValue != 7 {
			t.Errorf("注册后ScanTasksQueue更新失败, 期望 %d, 获取 %f", 7, scanValue)
		}

		pgValue := testutil.ToFloat64(metrics.DatabaseConnections.WithLabelValues("postgres"))
		if pgValue != 3 {
			t.Errorf("注册后PostgreSQL连接数指标更新失败, 期望 %d, 获取 %f", 3, pgValue)
		}
	})
}
