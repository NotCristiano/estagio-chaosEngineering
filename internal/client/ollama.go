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
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

type GenerateOptions struct {
	Temperature float64 `json:"temperature"`
	NumCtx      int     `json:"num_ctx"`
	NumPredict  int     `json:"num_predict"`
}

// GenerateRequest /api/generate
type GenerateRequest struct {
	Model   string          `json:"model"`
	Prompt  string          `json:"prompt"`
	Stream  bool            `json:"stream"`
	Options GenerateOptions `json:"options"`
}

// GenerateResponse holds what we care about from /api/generate
type GenerateResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
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

// NewOllamaClient creates a new OllamaClient
func NewOllamaClient(baseURL, model string, timeout time.Duration) *OllamaClient {

	customTransport := &http.Transport{
		DisableKeepAlives: true, // we force a new connection for every request
	}

	return &OllamaClient{
		BaseURL: baseURL,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout:   timeout,
			Transport: customTransport,
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
			NumCtx:      2048,
			NumPredict:  5,
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
