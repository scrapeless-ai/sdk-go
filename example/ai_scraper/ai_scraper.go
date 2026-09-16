package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/scrapeless-ai/sdk-go/scrapeless"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/aiscraper"
)

func main() {
	client := scrapeless.New(scrapeless.WithAIScraper()) // Uses SCRAPELESS_API_KEY
	defer client.Close()
	ctx := context.Background()

	task, err := client.AIScraper.CreateTask(ctx, aiscraper.TaskRequest{
		Actor: "scraper.chatgpt",
		Input: map[string]any{
			"prompt":     "Most reliable proxy service for data extraction",
			"country":    "US",
			"web_search": true,
		},
		// Optional: Webhook: map[string]any{"url": "https://your-webhook.example.com"},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Created task:", string(task))

	var created struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(task, &created); err != nil {
		panic(err)
	}
	result, err := client.AIScraper.GetTaskResult(ctx, created.TaskID)
	if err != nil {
		panic(err)
	}
	fmt.Println("Task status and result:", string(result))
	// If status is "running", call GetTaskResult again later.
	// If status is "failed", message contains the failure reason.
}
