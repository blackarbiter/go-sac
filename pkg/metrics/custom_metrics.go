package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	ScanTasksQueue = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "scan_tasks_queue_size",
			Help: "Current pending scan tasks in queue",
		},
	)

	DatabaseConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_total",
			Help: "Database connection pool status",
		},
		[]string{"type"},
	)
)

func RegisterCustomMetrics() {
	prometheus.MustRegister(ScanTasksQueue)
	prometheus.MustRegister(DatabaseConnections)
}
