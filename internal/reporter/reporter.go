package reporter

import (
	"estagio-chaosEngineering/internal/client"
	"fmt"
	"strings"
	"time"
)

// Summary holds the summary of the test run
type Summary struct {
	Total          int
	Successes      int
	Failures       int
	TotalTime      time.Duration
	MinLatency     time.Duration
	MaxLatency     time.Duration
	AverageLatency time.Duration
}

// PrintHeader prints the header of the test run
func PrintHeader(target, model string, count int, interval time.Duration) {
	fmt.Println()
	fmt.Printf("===== Chaos Ollama Poker =====\n")
	fmt.Printf("	Target: %s\n", target)
	fmt.Printf("	Model: %s\n", model)
	fmt.Printf("	Probes: %d\n", count)
	fmt.Printf("	Interval: %s\n", interval)
	fmt.Println(strings.Repeat("-", 30))
}

// PrintResult prints a single result
func PrintResult(r client.ProbeResult) {
	ts := r.Timestamp.Format("15:04:05.000")

	if r.OK() {
		preview := r.Response
		if len(preview) > 40 {
			preview = preview[:40] + "..."
		}
		fmt.Printf("	%s	%d	%s	%s\n", ts, r.Attempt, r.Latency.Round(time.Millisecond), preview)
	} else {
		fmt.Printf("	%s	%d	FAILED	%s\n", ts, r.Attempt, r.Error)
	}
}

// PrintSummary prints the summary of the test run
func PrintSummary(s Summary) {

	successRate := 0.0
	if s.Total > 0 {
		successRate = float64(s.Successes) / float64(s.Total) * 100
	}

	fmt.Println(strings.Repeat("-", 30))
	fmt.Printf("Summary\n")
	fmt.Printf("	Total probes: %d\n", s.Total)
	fmt.Printf("	Success rate: %.2f%%\n", successRate)
	fmt.Printf("	Failures: %d\n", s.Failures)
	fmt.Printf("	Average latency: %s\n", s.AverageLatency.Round(time.Millisecond))
	fmt.Printf("	Min latency: %s\n", s.MinLatency.Round(time.Millisecond))
	fmt.Printf("	Max latency: %s\n", s.MaxLatency.Round(time.Millisecond))
	fmt.Printf("	Total time: %s\n", s.TotalTime.Round(time.Millisecond))
	fmt.Println()
}

// BuildSummary builds the summary of the test run
func BuildSummary(results []client.ProbeResult, totalTime time.Duration) Summary {
	s := Summary{
		Total:      len(results),
		TotalTime:  totalTime,
		MinLatency: time.Duration(1<<63 - 1),
	}

	var latSum time.Duration
	for _, r := range results {
		if r.OK() {
			s.Successes++
			latSum += r.Latency
			if r.Latency < s.MinLatency {
				s.MinLatency = r.Latency
			}
			if r.Latency > s.MaxLatency {
				s.MaxLatency = r.Latency
			}
		} else {
			s.Failures++
		}
	}

	if s.Successes > 0 {
		s.AverageLatency = latSum / time.Duration(s.Successes)
	} else {
		s.MinLatency = 0
	}

	return s
}
