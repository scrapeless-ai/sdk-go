package scrapeless

import (
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/actor"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/aiscraper"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/browser"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/captcha"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/crawl"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/deepserp"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/httpserver"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/profile"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/proxies"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/router"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/scraping"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/storage"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/universal"
)

type Client struct {
	Browser   *browser.Browser
	Proxy     *proxies.Proxy
	Captcha   *captcha.Captcha
	Storage   *storage.Storage
	Server    *httpserver.Server
	Router    *router.Router
	DeepSerp  *deepserp.DeepSerp
	Scraping  *scraping.Scraping
	Universal *universal.Universal
	Actor     *actor.ActorService
	Crawl     *crawl.Crawl
	Profile   *profile.Profile
	AIScraper *aiscraper.AIScraper
	CloseFun  []func() error
}

func New(opts ...Option) *Client {
	var client = new(Client)
	for _, opt := range opts {
		opt.Apply(client)
	}
	client.Router = router.New(typeHttp)
	return client
}

// Close closes the Client.
func (c *Client) Close() {
	for _, f := range c.CloseFun {
		_ = f()
	}
}

const (
	typeHttp = "http"
	typeGrpc = "grpc"
)

type Option interface {
	Apply(*Client)
}

type BrowserOption struct {
	tp string
}

func (o *BrowserOption) Apply(c *Client) {
	c.Browser = browser.NewBrowser(o.tp)
	c.CloseFun = append(c.CloseFun, c.Browser.Close)
}

// WithBrowser choose browser type.
func WithBrowser(tp ...string) Option {
	if len(tp) == 0 {
		tp = append(tp, typeHttp)
	}
	return &BrowserOption{
		tp: tp[0],
	}
}

type ProxyOption struct {
	tp string
}

func (o *ProxyOption) Apply(a *Client) {
	a.Proxy = proxies.NewProxy(o.tp)
	a.CloseFun = append(a.CloseFun, a.Proxy.Close)
}

// WithProxy choose proxies type.
func WithProxy(tp ...string) Option {
	if len(tp) == 0 {
		tp = append(tp, typeHttp)
	}
	return &ProxyOption{
		tp: tp[0],
	}
}

type CaptchaOption struct {
	tp string
}

func (o *CaptchaOption) Apply(a *Client) {
	a.Captcha = captcha.NewCaptcha(o.tp)
	a.CloseFun = append(a.CloseFun, a.Captcha.Close)
}

// WithCaptcha choose captcha type.
func WithCaptcha(tp ...string) Option {
	if len(tp) == 0 {
		tp = append(tp, typeHttp)
	}
	return &CaptchaOption{
		tp: tp[0],
	}
}

type StorageOption struct {
	tp string
}

func (o *StorageOption) Apply(a *Client) {
	a.Storage = storage.NewStorage(o.tp)
}

// WithStorage choose storage type.
func WithStorage(tp ...string) Option {
	if len(tp) == 0 {
		tp = append(tp, typeHttp)
	}
	return &StorageOption{
		tp: tp[0],
	}
}

type ServerOption struct {
	mode httpserver.ServerMode
}

func (s *ServerOption) Apply(a *Client) {
	a.Server = httpserver.New(s.mode)
}

// WithServer choose server mode.
func WithServer(mode ...httpserver.ServerMode) Option {
	if len(mode) == 0 {
		mode = append(mode, httpserver.ReleaseMode)
	}
	return &ServerOption{mode: mode[0]}
}

type DeepSerpOption struct {
	tp string
}

func (d *DeepSerpOption) Apply(c *Client) {
	c.DeepSerp = deepserp.NewDeepSerp(d.tp)
	c.CloseFun = append(c.CloseFun, c.DeepSerp.Close)
}

// WithDeepSerp choose DeepSerp type.
func WithDeepSerp(tp ...string) Option {
	if len(tp) == 0 {
		tp = append(tp, typeHttp)
	}
	return &DeepSerpOption{
		tp: tp[0],
	}
}

type ScrapingOption struct {
	tp string
}

func (s *ScrapingOption) Apply(c *Client) {
	c.Scraping = scraping.New(s.tp)
	c.CloseFun = append(c.CloseFun, c.Scraping.Close)
}

// WithScraping choose scraping type.
func WithScraping(tp ...string) Option {
	if len(tp) == 0 {
		tp = append(tp, typeHttp)
	}
	return &ScrapingOption{
		tp: tp[0],
	}
}

type UniversalOption struct {
	tp string
}

func (s *UniversalOption) Apply(c *Client) {
	c.Universal = universal.New(s.tp)
	c.CloseFun = append(c.CloseFun, c.Universal.Close)
}

// WithUniversal choose universal type.
func WithUniversal(tp ...string) Option {
	if len(tp) == 0 {
		tp = append(tp, typeHttp)
	}
	return &UniversalOption{
		tp: tp[0],
	}
}

type ActorOption struct {
	tp string
}

func (s *ActorOption) Apply(c *Client) {
	c.Actor = actor.NewActor(s.tp)
	c.CloseFun = append(c.CloseFun, c.Actor.Close)
}

// WithActor choose Actor type.
func WithActor(tp ...string) Option {
	if len(tp) == 0 {
		tp = append(tp, typeHttp)
	}
	return &ActorOption{
		tp: tp[0],
	}
}

type CrawlOption struct {
	tp string
}

func (o *CrawlOption) Apply(c *Client) {
	c.Crawl = crawl.New()
	c.CloseFun = append(c.CloseFun, c.Crawl.Close)
}

// WithCrawl choose crawl type.
func WithCrawl(tp ...string) Option {
	if len(tp) == 0 {
		tp = append(tp, typeHttp)
	}
	return &CrawlOption{
		tp: tp[0],
	}
}

type ProfileOption struct {
	tp string
}

func (o *ProfileOption) Apply(c *Client) {
	c.Profile = profile.New()
	c.CloseFun = append(c.CloseFun, c.Profile.Close)
}

// WithProfile choose profile type.
func WithProfile(tp ...string) Option {
	if len(tp) == 0 {
		tp = append(tp, typeHttp)
	}
	return &ProfileOption{
		tp: tp[0],
	}
}

type AIScraperOption struct{}

func (o *AIScraperOption) Apply(c *Client) {
	c.AIScraper = aiscraper.New()
	c.CloseFun = append(c.CloseFun, c.AIScraper.Close)
}

// WithAIScraper enables AI chat task creation and result retrieval.
func WithAIScraper() Option {
	return &AIScraperOption{}
}
