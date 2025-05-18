package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/blackarbiter/go-sac/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMonitorMiddleware(t *testing.T) {
	// 重置注册表，避免测试之间的干扰
	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics.HttpRequestsTotal)
	registry.MustRegister(metrics.ResponseTimeHistogram)

	// 创建测试处理器
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "error") {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	})

	// 创建监控中间件包装的处理器
	handler := metrics.MonitorMiddleware(testHandler)

	// 测试成功请求
	t.Run("成功请求应计入200状态指标", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 %d, 获取 %d", http.StatusOK, w.Code)
		}

		// 验证指标是否正确记录
		metricCount := testutil.ToFloat64(metrics.HttpRequestsTotal.WithLabelValues("GET", "/api/test", strconv.Itoa(http.StatusOK)))
		if metricCount == 0 {
			t.Error("请求指标未正确增加")
		}
	})

	// 测试错误请求
	t.Run("错误请求应计入500状态指标", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/error", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("期望状态码 %d, 获取 %d", http.StatusInternalServerError, w.Code)
		}

		// 验证指标是否正确记录
		metricCount := testutil.ToFloat64(metrics.HttpRequestsTotal.WithLabelValues("GET", "/api/error", strconv.Itoa(http.StatusInternalServerError)))
		if metricCount == 0 {
			t.Error("错误请求指标未正确增加")
		}
	})
}

func TestResponseWrapper(t *testing.T) {
	t.Run("响应包装器应保留原状态码", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapper := metrics.NewResponseWrapper(w)

		statusCode := http.StatusNotFound
		wrapper.WriteHeader(statusCode)

		if w.Code != statusCode {
			t.Errorf("期望状态码 %d, 获取 %d", statusCode, w.Code)
		}

		if wrapper.Status() != strconv.Itoa(statusCode) {
			t.Errorf("Status()方法应返回状态码的字符串形式: %s, 但获取 %s",
				strconv.Itoa(statusCode), wrapper.Status())
		}
	})

	t.Run("响应包装器应默认为200", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapper := metrics.NewResponseWrapper(w)

		if wrapper.Status() != strconv.Itoa(http.StatusOK) {
			t.Errorf("未设置状态码时默认应为 %s, 获取 %s",
				strconv.Itoa(http.StatusOK), wrapper.Status())
		}
	})
}
