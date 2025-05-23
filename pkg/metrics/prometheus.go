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

	// RabbitMQ 相关指标
	DeadLetterCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_dead_letter_total",
			Help: "Total number of messages processed by dead letter handler",
		},
		[]string{"queue", "retry_count", "reason"},
	)

	MessageProcessingCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_message_processing_total",
			Help: "Total number of messages processed",
		},
		[]string{"queue", "status"}, // status: success, failure
	)

	BatchProcessingCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_batch_processing_total",
			Help: "Total number of message batches processed",
		},
		[]string{"queue", "status"}, // status: success, failure
	)

	BatchMessageCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_batch_messages_total",
			Help: "Total number of messages in batches",
		},
		[]string{"queue", "status"}, // status: success, failure
	)

	MessageProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rabbitmq_message_processing_duration_seconds",
			Help:    "Duration of message processing",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"queue", "status"},
	)
)

func InitPrometheus() {
	// 注册HTTP指标
	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(ResponseTimeHistogram)

	// 注册RabbitMQ指标
	prometheus.MustRegister(DeadLetterCounter)
	prometheus.MustRegister(MessageProcessingCounter)
	prometheus.MustRegister(BatchProcessingCounter)
	prometheus.MustRegister(BatchMessageCounter)
	prometheus.MustRegister(MessageProcessingDuration)

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
