package main

import (
	"context"
	"estagio-chaosEngineering/internal/dataset"
	"estagio-chaosEngineering/internal/metrics"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"estagio-chaosEngineering/internal/client"
	"estagio-chaosEngineering/internal/reporter"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Default flag values
	target := flag.String("target", "http://localhost:11434", "Ollama base URL")
	model := flag.String("model", "llama3.2", "Ollama model name")
	count := flag.Int("count", 20, "Number of sends (0 = inf)")
	interval := flag.Duration("interval", 2*time.Second, "Delay between sends")
	timeout := flag.Duration("timeout", 10*time.Second, "Per-request timeout")
	retries := flag.Int("retries", 3, "Max retries on failure")
	backoff := flag.Duration("backoff", 500*time.Millisecond, "Base backoff duration between retries")
	metricsAddr := flag.String("metric-addr", ":9090", "Prometheus metrics address, exposes /metrics")
	waitTimeout := flag.Duration("wait-timeout", 5*time.Minute, "How long to wait for Ollama to be ready")
	workers := flag.Int("workers", 1, "Number of concurrent workers")
	qualityInterval := flag.Duration("quality-interval", 45*time.Second, "Interval between quality checks (ARC-Challenge)")
	flag.Parse()

	// Signal handler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("Shutting down...")
		cancel()
	}()

	// Prometheus
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		fmt.Printf("  Metrics available at http://%s/metrics\n", *metricsAddr)
		if err := http.ListenAndServe(*metricsAddr, mux); err != nil {
			fmt.Printf("  [WARN] Error starting metrics server: %v\n", err)
		}
	}()

	// Load ARC Challange dataset for quality checks
	arc, err := dataset.Load()
	if err != nil {
		fmt.Printf("  [WARN] Failed to load ARC dataset: %v\n", err)
	} else {
		fmt.Printf("  [OK] Loaded ARC dataset: %d questions\n", arc.Len())
	}

	// Setup client
	c := client.NewOllamaClient(*target, *model, *timeout)

	reporter.PrintHeader(*target, *model, *count, *interval)

	// Waiting for Ollama to be ready
	fmt.Printf("	Checking Ollama health at %s/api/tags", *target)
	if err := waitForOllama(ctx, c, *waitTimeout); err != nil {
		fmt.Printf("	[ERROR] Ollama not reachable: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("	[OK] Ollama is ready - starting %d workers\n", *workers)

	// Concurrent workers setup
	var wg sync.WaitGroup
	resultCh := make(chan client.ProbeResult, *workers*10)

	start := time.Now()

	for w := 1; w <= *workers; w++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()
			var downtimeStart time.Time
			wasDown := false

			for i := 1; ; i++ {
				select {
				case <-ctx.Done():
					return
				default:
				}

				result := probeWithRetry(ctx, c, i, *retries, *backoff)

				resultCh <- result

				if result.OK() {
					metrics.ProbeTotal.WithLabelValues("success").Inc()
					metrics.ProbeDuration.WithLabelValues("success").Observe(result.Latency.Seconds())
					metrics.ServiceUp.Set(1)

					if wasDown {
						metrics.RecoveryDuration.Observe(time.Since(downtimeStart).Seconds())
						wasDown = false
					}

				} else {
					metrics.ProbeTotal.WithLabelValues("failure").Inc()
					metrics.ProbeDuration.WithLabelValues("failure").Observe(result.Latency.Seconds())
					metrics.ServiceUp.Set(0)

					if !wasDown {
						downtimeStart = time.Now()
						wasDown = true
					}
				}

				if *count != 0 && i >= *count {
					return
				}

				select {
				case <-ctx.Done():
					return
				case <-time.After(*interval):
				}
			}
		}(w)
	}

	if arc != nil {
		accuracy := metrics.NewAccuracyWindow(20)
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(*qualityInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					question := arc.Next()
					askCtx, askCancel := context.WithTimeout(ctx, 90*time.Second)
					result := c.Ask(askCtx, question.FormatPrompt())
					askCancel()

					metrics.AskDuration.Observe(result.Latency.Seconds())

					correct := result.Error == nil && result.Answer == question.Answer
					if correct {
						metrics.QualityCorrectTotal.Inc()
					} else {
						metrics.QualityWrongTotal.Inc()
					}

					acc := accuracy.Record(correct)
					metrics.QualityAccuracy.Set(acc)

					reporter.PrintQualityResult(question.ID, result.Answer, question.Answer, correct, result.Latency, result.Error)
				}

			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var results []client.ProbeResult
	for result := range resultCh {
		results = append(results, result)
		reporter.PrintResult(result)
	}

	summary := reporter.BuildSummary(results, time.Since(start))
	reporter.PrintSummary(summary)

	if summary.Failures > summary.Total/2 {
		os.Exit(1)
	}
}

// probeWithRetry performs retries with exponential backoff to send a ping request using the provided client
func probeWithRetry(ctx context.Context, c *client.OllamaClient, attempt, maxRetries int, base time.Duration) client.ProbeResult {
	var result client.ProbeResult

	for try := 0; try <= maxRetries; try++ {
		result = c.Ping(ctx, attempt)
		if result.OK() {
			return result
		}

		if try < maxRetries {
			metrics.RetryTotal.Inc()
			// Exponential backoff
			wait := time.Duration(float64(base) * math.Pow(2, float64(try)))
			select {
			case <-ctx.Done():
				return result
			case <-time.After(wait):
			}
		}
	}
	return result
}

// waitForOllama repeatedly checks the health of the Ollama service until it becomes ready or a deadline is reached
func waitForOllama(ctx context.Context, c *client.OllamaClient, deadline time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timed out after %s", deadline)
		default:
		}

		code, _, err := c.HealthCheck(timeoutCtx)
		if err == nil && code == http.StatusOK {
			return nil
		}
		fmt.Printf("Not ready yet (code=%d err=%s), retrying...\n", code, err)
		time.Sleep(2 * time.Second)
	}
}
