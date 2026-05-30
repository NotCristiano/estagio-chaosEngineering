package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaClient wrapping the Ollama API
type OllamaClient struct {
	BaseURL         string
	Model           string
	HTTPClient      *http.Client
	askHTTPClient   *http.Client
	ProbeNumPredict int
	ProbeNumCtx     int
	AskNumPredict   int
}

type GenerateOptions struct {
	Temperature float64 `json:"temperature"`
	NumCtx      int     `json:"num_ctx"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

// GenerateRequest /api/generate
type GenerateRequest struct {
	Model   string          `json:"model"`
	Prompt  string          `json:"prompt"`
	Stream  bool            `json:"stream"`
	Options GenerateOptions `json:"options"`
	Think   bool            `json:"think"`
}

// GenerateResponse holds what we care about from /api/generate
type GenerateResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Thinking bool   `json:"thinking"`
	Done     bool   `json:"done"`
}

// ProbeResult holds the results of a single probe (request)
type ProbeResult struct {
	Attempt    int
	StatusCode int
	Latency    time.Duration
	Response   string
	Error      error
	Timestamp  time.Time
}

func (r ProbeResult) OK() bool {
	return r.StatusCode == http.StatusOK && r.Error == nil
}

type AskResult struct {
	Answer   string
	Latency  time.Duration
	Error    error
	TimedOut bool
}

// NewOllamaClient creates a new OllamaClient
func NewOllamaClient(baseURL, model string, timeout time.Duration, probeNumPredict, probeNumCtx, askNumPredict int) *OllamaClient {

	newTransport := func() *http.Transport {
		return &http.Transport{
			DisableKeepAlives: true,
		}
	}

	return &OllamaClient{
		BaseURL:         baseURL,
		Model:           model,
		ProbeNumPredict: probeNumPredict,
		ProbeNumCtx:     probeNumCtx,
		AskNumPredict:   askNumPredict,
		HTTPClient: &http.Client{
			Timeout:   timeout,
			Transport: newTransport(),
		},
		askHTTPClient: &http.Client{
			Timeout:   10 * time.Minute, // context manages pratic timeout in main
			Transport: newTransport(),
		},
	}
}

// Ping sends a "Hello" request to the Ollama API and returns the response
func (c *OllamaClient) Ping(ctx context.Context, attempt int) ProbeResult {
	start := time.Now()
	result := ProbeResult{Attempt: attempt, Timestamp: start}

	body, err := json.Marshal(GenerateRequest{
		Model:  c.Model,
		Prompt: "Hello",
		Stream: false,
		Options: GenerateOptions{
			Temperature: 0.0, // Greedy
			NumCtx:      c.ProbeNumCtx,
			NumPredict:  c.ProbeNumPredict,
		},
	})

	if err != nil {
		result.Error = fmt.Errorf("marshal error: %w", err)
		result.Latency = time.Since(start)
		return result
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		result.Error = fmt.Errorf("request build error: %w", err)
		result.Latency = time.Since(start)
		return result
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	result.Latency = time.Since(start)

	if err != nil {
		result.Error = fmt.Errorf("http error: %w", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	if resp.StatusCode == http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		var gen GenerateResponse
		if json.Unmarshal(raw, &gen) == nil {
			result.Response = gen.Response
		}
	}
	return result
}

func (c *OllamaClient) Ask(ctx context.Context, prompt string) AskResult {
	start := time.Now()

	body, err := json.Marshal(GenerateRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
		Think:  false,
		Options: GenerateOptions{
			Temperature: 0.0, // Greedy, first token should be the answer
			NumCtx:      2048,
			NumPredict:  c.AskNumPredict,
		},
	})
	if err != nil {
		return AskResult{Error: fmt.Errorf("marshal error: %w", err)}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return AskResult{Error: fmt.Errorf("request build error: %w", err)}
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.askHTTPClient.Do(request)
	latency := time.Since(start)
	if err != nil {
		timedOut := ctx.Err() != nil
		return AskResult{Error: fmt.Errorf("http error: %w", err), Latency: latency, TimedOut: timedOut}
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return AskResult{Error: fmt.Errorf("bad status code: %d", response.StatusCode), Latency: latency}
	}

	raw, _ := io.ReadAll(response.Body)
	var gen GenerateResponse

	if err := json.Unmarshal(raw, &gen); err != nil {
		return AskResult{Error: fmt.Errorf("unmarshal error: %w", err), Latency: latency}
	}

	answer := extractAnswer(gen.Response)
	if answer == "" {
		return AskResult{Latency: latency, Error: fmt.Errorf("no answer found in response: %s", gen.Response)}
	}
	return AskResult{Answer: answer, Latency: time.Since(start), Error: nil}
}

func extractAnswer(text string) string {
	for _, ch := range text {
		switch ch {
		case 'A', 'B', 'C', 'D':
			return string(ch)
		}
	}
	return ""
}

func (c *OllamaClient) HealthCheck(ctx context.Context) (int, time.Duration, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/tags", nil)
	if err != nil {
		return 0, time.Since(start), err
	}

	resp, err := c.HTTPClient.Do(req)
	latency := time.Since(start)
	if err != nil {
		return 0, latency, err
	}
	resp.Body.Close()
	return resp.StatusCode, latency, nil
}
