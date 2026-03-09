package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"estagio-chaosEngineering/internal/client"
	"estagio-chaosEngineering/internal/reporter"
)

func main() {
	// Default flag values
	target := flag.String("target", "https://localhost:11434", "Ollama base URL")
	model := flag.String("model", "llama3.2", "Ollama model name")
	count := flag.Int("count", 20, "Number of sends (0 = inf)")
	interval := flag.Duration("interval", 2*time.Second, "Delay between sends")
	timeout := flag.Duration("timeout", 10*time.Second, "Per-request timeout")
	retries := flag.Int("retries", 3, "Max retries on failure")
	backoff := flag.Duration("backoff", 500*time.Microsecond, "Base backoff duration between retries")
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

	// Setup client
	c := client.NewOllamaClient(*target, *model, *timeout)

	reporter.PrintHeader(*target, *model, *count, *interval)

	// Waiting for Ollama to be ready
	fmt.Printf("	Checking Ollama health at %s/api/tags", *target)
	if err := waitForOllama(ctx, c, 30*time.Second); err != nil {
		fmt.Printf("	[ERROR] Ollama not reachable: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("	[OK] Ollama is ready - starting probes\n")

	// Probe loop
	var results []client.ProbeResult
	start := time.Now()

	for i := 1; ; i++ {
		select {
		case <-ctx.Done():
			goto done
		default:
		}

		result := probeWithRetry(ctx, c, i, *retries, *backoff)
		results = append(results, result)
		reporter.PrintResult(result)

		// Finite part when reaching requested count
		if *count != 0 && i >= *count {
			break
		}

		select {
		case <-ctx.Done():
			goto done
		case <-time.After(*interval):
		}
	}

done:
	summary := reporter.BuildSummary(results, time.Since(start))
	reporter.PrintSummary(summary)

	// Exit non-zero if +50% failed
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
