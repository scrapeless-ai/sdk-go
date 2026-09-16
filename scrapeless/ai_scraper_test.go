package scrapeless

import "testing"

func TestWithAIScraper(t *testing.T) {
	t.Chdir(t.TempDir())
	client := New(WithAIScraper(), WithScraping(), WithActor(), WithStorage())
	defer client.Close()
	if client.AIScraper == nil || client.Scraping == nil || client.Actor == nil || client.Storage == nil {
		t.Fatal("AI Scraper and legacy services must coexist")
	}
}
