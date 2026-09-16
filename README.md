# Scrapeless SDK Go

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)



The official Go SDK of [Scrapeless AI](https://scrapeless.com?utm_source=github&utm_medium=referral&utm_campaign=go_sdk_repo) - a powerful web scraping and browser automation platform that helps you extract data from any website at scale.

New to Scrapeless? [Sign up](https://app.scrapeless.com/passport/login?utm_source=github&utm_medium=referral&utm_campaign=go_sdk_repo) and get $5 in free credits.

## 📑 Table of Contents

- [🌟 Features](#-features)
- [📦 Installation](#-installation)
- [🚀 Quick Start](#-quick-start)
- [📖 Usage Examples](#-usage-examples)
- [🔧 API Reference](#-api-reference)
- [📚 Examples](#-examples)
- [🛠️ Contribution & Development Guide](#-contribution--development-guide)
- [📄 License](#-license)
- [📞 Support](#-support)
- [🏢 About Scrapeless](#-about-scrapeless)

## 🌟 Features

- **Browser**: Remote browser session management with configurable anti-detection capabilities (e.g., fingerprint spoofing, CAPTCHA solving) and extensible automation workflows.
- **Web Unlocker**: web interaction and data extraction with full browser capabilities. Execute JavaScript rendering, simulate user interactions (clicks, scrolls), bypass anti-scraping measures, and export structured data in formats.
- **Crawl**: Extract data from single pages or traverse entire domains, exporting in formats including Markdown, JSON, HTML, screenshots, and links.
- **Scraping API**: Direct data extraction APIs for websites (e.g., e-commerce, travel platforms). Retrieve structured product information, pricing, and reviews with pre-built connectors.
- **Google Search API**: Google SERP data extraction API. Fetch organic results, news, images, and more with customizable parameters and real-time updates.
- **AI Scraper**: Extract AI chat answers, citations, and brand mentions across supported models.
- **Proxies**: Geo-targeted proxy network with 195+ countries. Optimize requests for better success rates and regional data access.

## 📦 Installation

Install the SDK using `go get`:

```bash
go get -u github.com/scrapeless-ai/sdk-go
```

## 🚀 Quick Start

### Basic Setup

```go
package main

import (
	"github.com/scrapeless-ai/sdk-go/scrapeless"
)

func main() {
	// Initialize the client (uses SCRAPELESS_API_KEY)
	client := scrapeless.New(scrapeless.WithBrowser())
	defer client.Close()
}
```

### Environment Variables

You can also configure the SDK using environment variables:

```bash
# Required
SCRAPELESS_API_KEY=your-api-key

# Optional - Custom API endpoints
SCRAPELESS_BASE_API_URL=https://api.scrapeless.com
SCRAPELESS_BROWSER_API_URL=https://browser.scrapeless.com
SCRAPELESS_CRAWL_API_URL=https://api.scrapeless.com
```

## 📖 Usage Examples

### Browser

```go
package main

import (
	"context"
	"github.com/scrapeless-ai/sdk-go/scrapeless"
	"github.com/scrapeless-ai/sdk-go/scrapeless/log"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/browser"
)

func main() {
	client := scrapeless.New(scrapeless.WithBrowser())
	defer client.Close()

	browserInfo, err := client.Browser.Create(context.Background(), browser.Actor{
		Input:        browser.Input{SessionTtl: "180"},
		ProxyCountry: "US",
	})
	if err != nil {
		panic(err)
	}
	log.Infof("%+v", browserInfo)
}
```

### Browser Profile

```go
package main

import (
	"context"
	"fmt"
	"github.com/scrapeless-ai/sdk-go/scrapeless"
)

func main() {
	client := scrapeless.New(scrapeless.WithProfile())
	defer client.Close()

	result, err := client.Profile.CreateProfile(context.Background(), "My Profile")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", result)
}
```

### Scraping API

```go
package main

import (
	"context"
	"github.com/scrapeless-ai/sdk-go/scrapeless"
	"github.com/scrapeless-ai/sdk-go/scrapeless/log"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/scraping"
)

func main() {
	client := scrapeless.New(scrapeless.WithScraping())

	scrape, err := client.Scraping.Scrape(context.Background(), scraping.ScrapingTaskRequest{
		Actor: "scraper.google.search",
		Input: map[string]interface{}{
			"q": "nike site:www.nike.com",
		},
		ProxyCountry: "US",
	})
	if err != nil {
		log.Errorf("scraping create err:%v", err)
		return
	}
	log.Infof("%+v", scrape)
}
```

### Web Unlocker

Extract data from websites using Web Unlocker (exposed as `client.Universal`).

```go
package main

import (
	"context"
	"fmt"
	"github.com/scrapeless-ai/sdk-go/scrapeless"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/universal"
)

func main() {
	client := scrapeless.New(scrapeless.WithUniversal())
	defer client.Close()

	result, err := client.Universal.CreateTask(context.Background(), universal.UniversalTaskRequest{
		Actor: universal.ScraperUniversal,
		Input: map[string]any{
			"url":      "https://example.com",
			"method":   "GET",
			"redirect": false,
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", result)
}
```

### Crawl

```go
package main

import (
	"context"
	"github.com/scrapeless-ai/sdk-go/scrapeless"
	"github.com/scrapeless-ai/sdk-go/scrapeless/log"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/crawl"
)

func main() {
	client := scrapeless.New(scrapeless.WithCrawl())

	// Crawl
	response, err := client.Crawl.CrawlUrl(context.Background(), "https://redditinc.com/blog", crawl.CrawlParams{
		Limit: 10,
		ScrapeOptions: crawl.ScrapeOptions{
			Formats: []string{"links",
				"markdown",
				"html",
				"screenshot"},
		},
		BrowserOptions: crawl.ICreateBrowser{
			SessionName:      "Crawl",
			SessionTTL:       "900",
			SessionRecording: "true",
			ProxyCountry:     "ANY",
		},
	})
	if err != nil {
		panic(err)
	}
	log.Infof("Crawl response: %v", response)

	// scrape
	scrapeResponse, err := client.Crawl.ScrapeUrl(context.Background(), "https://docs.scrapeless.com/en/overview/", crawl.ScrapeOptions{
		BrowserOptions: crawl.ICreateBrowser{
			SessionName:      "Crawl",
			SessionTTL:       "900",
			SessionRecording: "true",
			ProxyCountry:     "ANY",
		},
	})
	if err != nil {
		panic(err)
	}
	log.Infof("Scrape response: %v", scrapeResponse)
}
```

### Proxy

```go
package main

import (
	"context"
	"fmt"
	"github.com/scrapeless-ai/sdk-go/scrapeless"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/proxies"
)

func main() {
	client := scrapeless.New(scrapeless.WithProxy())
	defer client.Close()

	result, err := client.Proxy.Proxy(context.Background(), proxies.ProxyActor{
		Country:         "US",
		SessionDuration: 30,
		SessionId:       "my-session",
		Gateway:         "your-proxy-gateway:port",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", result)
}
```

### AI Scraper

Extract AI chat content in bulk to monitor brand mentions, compare answers, and analyze competitive intelligence from the latest models. Retrieve URLs, prompts, Markdown answers, citations, and more through one integration.

Supported actors include `scraper.chatgpt`, `scraper.perplexity`, `scraper.copilot`, `scraper.gemini`, `scraper.aimode`, `scraper.overview`, `scraper.grok`, and `scraper.alexa`. The `input` JSON depends on the actor; see the [AI Scraper documentation](https://docs.scrapeless.com/en/llm-chat-scraper/quickstart/introduction/?utm_source=github&utm_medium=referral&utm_campaign=go_sdk_repo) for detailed parameters. The optional `webhook` JSON contains a callback `url`.

```go
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
```

Both methods return the API JSON unchanged. Creation returns `task_id`, `status`, and, when available, `task_result`. Result retrieval returns `status`, `task_result` when available, and `message` on failure. Status is `success`, `failed`, or `running`; the SDK does not poll automatically.

Responses are raw JSON bytes (`[]byte`), preserving every API field. Decode them with `encoding/json` as needed.

## 🔧 API Reference

### Available Services

The SDK provides the following services:

- `Client.Browser` - Browser session management.
- `Client.Scraping` - Web scraping and data extraction.
- `Client.DeepSerp` - the Google Search API feature: search engine (Google SERP) result extraction.
- `Client.Universal` - the Web Unlocker feature: universal data extraction.
- `Client.Proxy` - Proxy management.
- `Client.Profile` - Browser profile management.
- `Client.Crawl` - Site crawling and page extraction.
- `Client.AIScraper` - AI chat task creation and result retrieval (enable with `WithAIScraper()`).
- `Client.Server` - HTTP service.
- `Client.Router` - Route access.
- `Client.Captcha` - Captcha processing.

## 📚 Examples

Check the `example` directory for complete usage examples:

- [Google Search API](./example/deepserp/deepserp.go)
- [Scraping API](./example/scraping/scraping.go)
- [Browser Profile](./example/profile/profile.go)
- [Web Unlocker](./example/universal/universal.go)
- [AI Scraper](./example/ai_scraper/ai_scraper.go)
- [Browser Operation Example](./example/browser/browser.go)
- [Captcha Recognition Example](./example/captcha/captcha.go)
- [Proxy Management Example](./example/proxy/proxy.go)
- [Route Call](./example/router/router.go)
- [HTTP Service](./example/httpserver/httpserver.go)
- [Crawl Usage Example](./example/crawl/crawl.go)

## 🛠️ Contribution & Development Guide

All forms of contributions are welcome! For detailed information on how to submit issues, PRs, code specifications, local development, etc., please refer to the [Contribution & Development Guide](./CONTRIBUTING.md).

**Quick Start**:

```bash
git clone https://github.com/scrapeless-ai/sdk-go.git
cd sdk-go
go mod tidy
go run ./example/ai_scraper/ai_scraper.go
```

For more information on project structure, best practices, etc., please refer to [CONTRIBUTING.md](./CONTRIBUTING.md).

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Support

- 📖 **Documentation**: [https://docs.scrapeless.com](https://docs.scrapeless.com?utm_source=github&utm_medium=referral&utm_campaign=go_sdk_repo)
- 💬 **Community**: [Join our Discord](https://backend.scrapeless.com/app/api/v1/public/links/discord)
- 🐛 **Issues**: [GitHub Issues](https://github.com/scrapeless-ai/scrapeless-sdk-node/issues)
- 📧 **Email**: [support@scrapeless.com](mailto:support@scrapeless.com)

## 🏢 About Scrapeless

Scrapeless is a powerful web scraping and browser automation platform that helps enterprises extract data from any website at scale. Our platform provides:

- High-performance web scraping infrastructure.
- Global proxy network.
- Browser automation capabilities.
- Enterprise-level reliability and support.

Visit [scrapeless.com](https://scrapeless.com?utm_source=github&utm_medium=referral&utm_campaign=go_sdk_repo) to learn more and get started.

---

Made with ❤️ by the Scrapeless team

## Related Projects

- [Scrapeless Python SDK](https://github.com/scrapeless-ai/sdk-python)
- [Scrapeless Node.js SDK](https://github.com/scrapeless-ai/sdk-node)
- [Scrapeless Go SDK](https://github.com/scrapeless-ai/sdk-go)
