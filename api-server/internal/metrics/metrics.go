package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	buildsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ritmo_builds_total",
			Help: "Total number of builds",
		},
		[]string{"status"},
	)

	buildsQueued = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ritmo_builds_queued",
			Help: "Number of builds currently queued",
		},
	)

	buildsRunning = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ritmo_builds_running",
			Help: "Number of builds currently running",
		},
	)

	buildDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ritmo_build_duration_seconds",
			Help:    "Build duration in seconds",
			Buckets: prometheus.ExponentialBuckets(10, 2, 10), // 10s to ~2.5 hours
		},
		[]string{"job_name", "status"},
	)

	workersTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ritmo_workers_total",
			Help: "Total number of workers by status",
		},
		[]string{"status"},
	)

	workerUtilization = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ritmo_worker_utilization",
			Help: "Worker utilization (current_builds / max_concurrent_builds)",
		},
		[]string{"worker_name"},
	)

	deploymentsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ritmo_deployments_total",
			Help: "Total number of deployments",
		},
		[]string{"environment", "status"},
	)

	apiRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ritmo_api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	apiRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ritmo_api_request_duration_seconds",
			Help:    "API request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	prometheus.MustRegister(buildsTotal)
	prometheus.MustRegister(buildsQueued)
	prometheus.MustRegister(buildsRunning)
	prometheus.MustRegister(buildDuration)
	prometheus.MustRegister(workersTotal)
	prometheus.MustRegister(workerUtilization)
	prometheus.MustRegister(deploymentsTotal)
	prometheus.MustRegister(apiRequestsTotal)
	prometheus.MustRegister(apiRequestDuration)
}

// Collector provides methods to record metrics
type Collector struct{}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{}
}

// RecordBuildScheduled increments the scheduled builds counter
func (c *Collector) RecordBuildScheduled() {
	buildsTotal.WithLabelValues("scheduled").Inc()
}

// RecordBuildCompleted records a completed build
func (c *Collector) RecordBuildCompleted(jobName, status string, duration float64) {
	buildsTotal.WithLabelValues(status).Inc()
	buildDuration.WithLabelValues(jobName, status).Observe(duration)
}

// RecordWorkerCount updates the worker count metric
func (c *Collector) RecordWorkerCount(status string, count int) {
	workersTotal.WithLabelValues(status).Set(float64(count))
}

// RecordDeployment records a deployment
func (c *Collector) RecordDeployment(environment, status string) {
	deploymentsTotal.WithLabelValues(environment, status).Inc()
}

// RecordAPIRequest records an API request
func (c *Collector) RecordAPIRequest(method, endpoint, status string, duration float64) {
	apiRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	apiRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// Handler returns the Prometheus HTTP handler
func Handler() http.Handler {
	return promhttp.Handler()
}
