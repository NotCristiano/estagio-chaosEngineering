package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ProbeTotal counts every probe attempt, labeled sucess/failure
	ProbeTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "poker_probe_total",
		Help: "Total number of probes sent to Ollama",
	}, []string{"result"}) // success or failure

	// ProbeDuration tracks request latency
	ProbeDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "poker_probe_duration_seconds",
		Help:    "Request latency in seconds",
		Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 5, 10, 30},
	}, []string{"result"})

	// ServiceUp 1 = last probe was successful / 0 = last probe failed
	ServiceUp = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "poker_service_up",
		Help: "1 = last probe was successful",
	})

	// RecoveryDuration tracks how long it took to recover from a failure
	RecoveryDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "poker_recovery_duration_seconds",
		Help:    "How long it took to recover from a failure",
		Buckets: []float64{1, 2, 5, 10, 20, 30, 60, 120},
	})

	// RetryTotal counts every retry attempt
	RetryTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poker_retry_total",
		Help: "Total number of retries due to failed probes",
	})
)
