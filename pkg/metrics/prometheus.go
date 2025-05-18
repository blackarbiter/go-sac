package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	ResponseTimeHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_time_seconds",
			Help:    "Histogram of response time",
			Buckets: []float64{0.1, 0.5, 1, 2, 5},
		},
		[]string{"method", "path"},
	)
)

func InitPrometheus() {
	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(ResponseTimeHistogram)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":9090", nil)
	}()
}

// 中间件示例
func MonitorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timer := prometheus.NewTimer(ResponseTimeHistogram.WithLabelValues(r.Method, r.URL.Path))
		defer timer.ObserveDuration()

		rw := NewResponseWrapper(w)
		next.ServeHTTP(rw, r)

		HttpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, rw.Status()).Inc()
	})
}

type ResponseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWrapper(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{w, http.StatusOK}
}

func (rw *ResponseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWrapper) Status() string {
	return strconv.Itoa(rw.statusCode)
}
