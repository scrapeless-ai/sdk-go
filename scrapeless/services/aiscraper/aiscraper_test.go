package aiscraper

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/scrapeless-ai/sdk-go/env"
)

func testService(t *testing.T, handler http.HandlerFunc) *AIScraper {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	oldURL, oldKey := env.Env.ScrapelessBaseApiUrl, env.GetActorEnv().ApiKey
	env.Env.ScrapelessBaseApiUrl, env.GetActorEnv().ApiKey = server.URL, "test-key"
	t.Cleanup(func() {
		env.Env.ScrapelessBaseApiUrl, env.GetActorEnv().ApiKey = oldURL, oldKey
	})
	service := New()
	t.Cleanup(func() { _ = service.Close() })
	return service
}

func TestCreateTask(t *testing.T) {
	for _, webhook := range []map[string]any{nil, {"url": "https://callback.example.com", "extra": true}} {
		t.Run(fmtWebhook(webhook), func(t *testing.T) {
			input := map[string]any{"prompt": "test", "nested": map[string]any{"models": []any{"a", "b"}}}
			response := `{"task_id":"task-1","status":"running","data":{"retained":true},"extra":42}`
			calls := 0
			service := testService(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodPost || r.URL.Path != "/api/v2/scraper/request" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
				}
				if r.Header.Get("x-api-token") != "test-key" || r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("unexpected headers: %v", r.Header)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				want := map[string]any{"actor": "scraper.future-model", "input": input}
				if webhook != nil {
					want["webhook"] = webhook
				}
				if !reflect.DeepEqual(body, want) {
					t.Errorf("body = %#v, want %#v", body, want)
				}
				w.WriteHeader(http.StatusCreated)
				_, _ = io.WriteString(w, response)
			})
			result, err := service.CreateTask(context.Background(), TaskRequest{Actor: "scraper.future-model", Input: input, Webhook: webhook})
			if err != nil || string(result) != response || calls != 1 {
				t.Fatalf("result = %s, err = %v, calls = %d", result, err, calls)
			}
		})
	}
}

func fmtWebhook(webhook map[string]any) string {
	if webhook == nil {
		return "without webhook"
	}
	return "with webhook"
}

func TestGetTaskResult(t *testing.T) {
	for _, response := range []string{
		`{"status":"running"}`,
		`{"status":"success","task_result":{"markdown":"answer","citations":[{"url":"https://example.com"}]}}`,
		`{"status":"failed","message":"Model unavailable"}`,
	} {
		t.Run(response, func(t *testing.T) {
			service := testService(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.EscapedPath() != "/api/v2/scraper/result/task%20%2F%3F%23" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
				}
				if r.Header.Get("x-api-token") != "test-key" {
					t.Error("missing API token")
				}
				body, _ := io.ReadAll(r.Body)
				if len(body) != 0 {
					t.Errorf("unexpected GET body: %s", body)
				}
				_, _ = io.WriteString(w, response)
			})
			result, err := service.GetTaskResult(context.Background(), "task /?#")
			if err != nil || string(result) != response {
				t.Fatalf("result = %s, err = %v", result, err)
			}
		})
	}
}

func TestHTTPError(t *testing.T) {
	response := `{"message":"Invalid token"}`
	service := testService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, response)
	})
	result, err := service.GetTaskResult(context.Background(), "task-1")
	if string(result) != response || err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("result = %s, err = %v", result, err)
	}
}

func TestInvalidInputAndCanceledContext(t *testing.T) {
	service := testService(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("invalid input or canceled context must not reach the server")
	})
	if _, err := service.CreateTask(context.Background(), TaskRequest{Input: map[string]any{"invalid": make(chan int)}}); err == nil {
		t.Fatal("expected JSON encoding error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.GetTaskResult(ctx, "task-1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
