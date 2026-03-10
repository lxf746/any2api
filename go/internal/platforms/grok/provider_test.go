package grok

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"any2api-go/internal/core"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestGrokProviderFlattensMessages(t *testing.T) {
	p := NewProviderWithConfig(core.GrokConfig{
		Request: core.RequestConfig{
			MaxInputLength:     core.DefaultCursorMaxInputLength,
			SystemPromptInject: "follow the rules",
		},
	}).(*grokProvider)

	got := p.flattenMessages(core.UnifiedRequest{
		Model:  "grok-4",
		System: "be precise",
		Messages: []core.Message{
			{Role: "assistant", Content: "earlier reply"},
			{Role: "user", Content: "latest question"},
		},
	})
	want := "system: be precise\nfollow the rules\n\nassistant: earlier reply\n\nlatest question"
	if got != want {
		t.Fatalf("unexpected flattened messages:\nwant=%q\n got=%q", want, got)
	}
}

func TestGrokStreamFilterHandlesSplitToolCard(t *testing.T) {
	filter := &grokStreamFilter{}
	first := filter.filter("hello<xai:tool_usage_card><xai:tool_name>web_search</xai:tool_name>")
	if first != "hello" {
		t.Fatalf("unexpected first filtered chunk: %q", first)
	}
	second := filter.filter("<xai:tool_args>{\"query\":\"q\"}</xai:tool_args></xai:tool_usage_card>world")
	if !strings.Contains(second, "[web_search] {\"query\":\"q\"}") {
		t.Fatalf("expected tool summary in second chunk, got %q", second)
	}
	if !strings.Contains(second, "world") {
		t.Fatalf("expected trailing text to remain, got %q", second)
	}
}

func TestGrokProviderUsesDefaultsForOptionalHeaders(t *testing.T) {
	p := NewProviderWithConfig(core.GrokConfig{
		CookieToken: "test-token",
		Request:     core.RequestConfig{MaxInputLength: core.DefaultCursorMaxInputLength},
	}).(*grokProvider)

	headers := p.headers()
	if got := headers["User-Agent"]; got != core.DefaultGrokUserAgent {
		t.Fatalf("expected default user-agent, got %q", got)
	}
	if got := headers["Origin"]; got != core.DefaultGrokOrigin {
		t.Fatalf("expected default origin, got %q", got)
	}
	if got := headers["Referer"]; got != core.DefaultGrokReferer {
		t.Fatalf("expected default referer, got %q", got)
	}
	if got := headers["Sec-Fetch-Site"]; got != "same-origin" {
		t.Fatalf("expected same-origin sec-fetch-site, got %q", got)
	}
	if got := headers["Cookie"]; got != "sso=test-token; sso-rw=test-token" {
		t.Fatalf("expected normalized sso cookie, got %q", got)
	}
}

func TestBuildGrokCookieHeaderMergesCloudflareCookies(t *testing.T) {
	got := buildGrokCookieHeader("test-token", "theme=dark; cf_clearance=old", "new")
	if !strings.Contains(got, "sso=test-token") {
		t.Fatalf("expected sso cookie, got %q", got)
	}
	if !strings.Contains(got, "theme=dark") {
		t.Fatalf("expected custom cf cookie fragment, got %q", got)
	}
	if !strings.Contains(got, "cf_clearance=new") {
		t.Fatalf("expected updated cf_clearance, got %q", got)
	}
	if strings.Contains(got, "cf_clearance=old") {
		t.Fatalf("expected old cf_clearance to be replaced, got %q", got)
	}
}

func TestGrokProviderUsesExplicitProxyURL(t *testing.T) {
	p := NewProviderWithConfig(core.GrokConfig{
		CookieToken: "test-token",
		ProxyURL:    "socks5://127.0.0.1:7890",
		Request:     core.RequestConfig{MaxInputLength: core.DefaultCursorMaxInputLength},
	}).(*grokProvider)

	transport, ok := p.client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected http transport, got %T", p.client.Transport)
	}
	req, err := http.NewRequest(http.MethodGet, "https://grok.com", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	proxyURL, err := transport.Proxy(req)
	if err != nil {
		t.Fatalf("resolve proxy: %v", err)
	}
	if proxyURL == nil || proxyURL.String() != "socks5h://127.0.0.1:7890" {
		t.Fatalf("expected normalized socks proxy, got %v", proxyURL)
	}
}

func TestGrokProviderRetriesTooManyRequests(t *testing.T) {
	attempts := 0
	p := NewProviderWithConfig(core.GrokConfig{
		CookieToken: "test-token",
		Request:     core.RequestConfig{MaxInputLength: core.DefaultCursorMaxInputLength},
	}).(*grokProvider)
	p.client = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": []string{"1"}},
				Body:       io.NopCloser(strings.NewReader("rate limited")),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{},
			Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
		}, nil
	})}
	var slept []time.Duration
	p.sleep = func(delay time.Duration) { slept = append(slept, delay) }

	resp, err := p.doRequest(context.Background(), core.UnifiedRequest{
		Model:    "grok-4",
		Messages: []core.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}
	defer resp.Body.Close()
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if len(slept) != 1 || slept[0] != time.Second {
		t.Fatalf("expected retry-after delay to be respected, got %#v", slept)
	}
}

func TestGrokProviderResetsClientOnForbidden(t *testing.T) {
	resetCalls := 0
	p := NewProviderWithConfig(core.GrokConfig{
		CookieToken: "test-token",
		Request:     core.RequestConfig{MaxInputLength: core.DefaultCursorMaxInputLength},
	}).(*grokProvider)
	p.client = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Header:     http.Header{"Retry-After": []string{"0"}},
			Body:       io.NopCloser(strings.NewReader("forbidden")),
		}, nil
	})}
	p.newClient = func() *http.Client {
		resetCalls++
		return &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
			}, nil
		})}
	}
	p.sleep = func(time.Duration) {}

	resp, err := p.doRequest(context.Background(), core.UnifiedRequest{
		Model:    "grok-4",
		Messages: []core.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("expected forbidden retry to succeed, got %v", err)
	}
	defer resp.Body.Close()
	if resetCalls != 1 {
		t.Fatalf("expected exactly one client reset, got %d", resetCalls)
	}
}
