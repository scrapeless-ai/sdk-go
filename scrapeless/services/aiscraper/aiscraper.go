// Package aiscraper provides access to the v2 AI Scraper API.
package aiscraper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scrapeless-ai/sdk-go/env"
)

// TaskRequest is forwarded to the API. Input fields depend on the actor.
// Actor is unrestricted to allow newly supported models without an SDK update.
type TaskRequest struct {
	Actor   string         `json:"actor"`
	Input   map[string]any `json:"input"`
	Webhook map[string]any `json:"webhook,omitempty"`
}

// AIScraper creates tasks and retrieves their status and results without polling.
type AIScraper struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

// New uses SCRAPELESS_API_KEY and SCRAPELESS_BASE_API_URL from the SDK environment.
// Use a context deadline to set a shorter timeout for individual calls.
func New() *AIScraper {
	return &AIScraper{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: strings.TrimRight(env.Env.ScrapelessBaseApiUrl, "/"),
		apiKey:  env.GetActorEnv().ApiKey,
	}
}

// CreateTask returns the complete API JSON, including task_id, status, and
// task_result when available. No fields or async flags are added to the request.
func (s *AIScraper) CreateTask(ctx context.Context, req TaskRequest) ([]byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return s.request(ctx, http.MethodPost, "/api/v2/scraper/request", body)
}

// GetTaskResult returns the complete API JSON. Task statuses (success, failed,
// running), failure messages, and actor-specific results are left unchanged.
func (s *AIScraper) GetTaskResult(ctx context.Context, taskID string) ([]byte, error) {
	return s.request(ctx, http.MethodGet, "/api/v2/scraper/result/"+url.PathEscape(taskID), nil)
}

func (s *AIScraper) request(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, s.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-token", s.apiKey)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return data, fmt.Errorf("AI Scraper request failed with status %d: %s", resp.StatusCode, data)
	}
	return data, nil
}

// Close releases idle HTTP connections.
func (s *AIScraper) Close() error {
	s.client.CloseIdleConnections()
	return nil
}
