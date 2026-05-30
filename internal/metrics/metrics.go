package metrics

import (
	"sync"

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

	// QualityCorrectTotal counts the number of times the model's response was correct
	QualityCorrectTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poker_quality_correct_total",
		Help: "TNumber of ARC-Challenge questions answered correctly",
	})

	// QualityWrongTotal counts the number of times the model's response was incorrect
	QualityWrongTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poker_quality_wrong_total",
		Help: "Number of ARC-Challenge questions answered incorrectly or timed out",
	})

	// QualityTimeoutTotal counts the number of times the model timed out
	QualityTimeoutTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poker_quality_timeout_total",
		Help: "Number of ARC-Challenge questions timed out",
	})

	// QualityAccuracy tracks the model's accuracy over time
	QualityAccuracy = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "poker_quality_accuracy",
		Help: "Rolling accuracy over the last 20 Ask probes (0.0–1.0)",
	})

	// AskDuration tracks request latency
	AskDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "poker_ask_duration_seconds",
		Help:    "Latency of each ARC-Challenge Ask probe",
		Buckets: []float64{1, 5, 10, 20, 30, 45, 60, 90, 120, 150, 170},
	})
)

// AccuracyWindow tracks the rolling accuracy of the model
type AccuracyWindow struct {
	window []bool
	size   int
	pos    int
	count  int
	mutex  sync.Mutex
}

// NewAccuracyWindow creates a new AccuracyWindow with the given size
func NewAccuracyWindow(size int) *AccuracyWindow {
	return &AccuracyWindow{
		window: make([]bool, size),
		size:   size,
	}
}

func (win *AccuracyWindow) Record(correct bool) float64 {
	win.mutex.Lock()
	defer win.mutex.Unlock()

	win.window[win.pos] = correct
	win.pos = (win.pos + 1) % win.size
	if win.count < win.size {
		win.count++
	}

	var sum int
	for i := 0; i < win.count; i++ {
		if win.window[(i+win.pos)%win.size] {
			sum++
		}
	}
	return float64(sum) / float64(win.count)
}
