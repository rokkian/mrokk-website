package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadOnlyRequestsAllowsGetAndHead(t *testing.T) {
	handler := readOnlyRequests(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, method := range []string{http.MethodGet, http.MethodHead} {
		request := httptest.NewRequest(method, "/", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusNoContent {
			t.Fatalf("%s request returned status %d, want %d", method, response.Code, http.StatusNoContent)
		}
	}
}

func TestReadOnlyRequestsRejectsMutatingMethods(t *testing.T) {
	handler := readOnlyRequests(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler was called for a mutating request")
	}))

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("payload"))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST request returned status %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if got := response.Header().Get("Allow"); got != "GET, HEAD" {
		t.Fatalf("Allow header = %q, want %q", got, "GET, HEAD")
	}
}

func TestReadOnlyRequestsRejectsLongRequestURI(t *testing.T) {
	handler := readOnlyRequests(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler was called for an oversized URI")
	}))

	request := httptest.NewRequest(http.MethodGet, "/"+strings.Repeat("a", 8<<10), nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusRequestURITooLong {
		t.Fatalf("oversized URI returned status %d, want %d", response.Code, http.StatusRequestURITooLong)
	}
}

func TestReadOnlyRequestsLimitsRequestBody(t *testing.T) {
	handler := readOnlyRequests(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err == nil {
			t.Error("reading oversized body succeeded")
		}
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	}))

	request := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(strings.Repeat("a", 1<<20+1)))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized body returned status %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	wantHeaders := map[string]string{
		"Content-Security-Policy":   "default-src 'self'; style-src 'self' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; media-src 'self'; script-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Permissions-Policy":        "camera=(), microphone=(), geolocation=(), payment=()",
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
	}
	for name, want := range wantHeaders {
		if got := response.Header().Get(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestRenderMarkdown(t *testing.T) {
	rendered, err := renderMarkdown("# Hello\n\nThis is **bold**.")
	if err != nil {
		t.Fatalf("renderMarkdown returned error: %v", err)
	}

	output := string(rendered)
	for _, expected := range []string{"<h1>Hello</h1>", "<strong>bold</strong>"} {
		if !strings.Contains(output, expected) {
			t.Errorf("rendered Markdown does not contain %q: %s", expected, output)
		}
	}
}

func TestRenderMarkdownEscapesRawHTML(t *testing.T) {
	rendered, err := renderMarkdown("<script>alert('xss')</script>")
	if err != nil {
		t.Fatalf("renderMarkdown returned error: %v", err)
	}

	if strings.Contains(string(rendered), "<script>") {
		t.Fatal("rendered Markdown contains executable raw HTML")
	}
}

func TestLoadContentRendersMarkdownPosts(t *testing.T) {
	content, err := loadContent("content.yaml")
	if err != nil {
		t.Fatalf("loadContent returned error: %v", err)
	}
	if len(content.Blog.Posts) == 0 {
		t.Fatal("expected at least one blog post")
	}
	if !strings.Contains(string(content.Blog.Posts[0].HTML), "<h1>Lorem Ipsum</h1>") {
		t.Fatalf("blog post Markdown was not rendered: %s", content.Blog.Posts[0].HTML)
	}
}

func TestCloudRunHealthEndpoint(t *testing.T) {
	router := setupRouter(nil, nil, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))

	if response.Code != http.StatusOK || response.Body.String() != "ok\n" {
		t.Fatalf("GET /health returned %d %q, want 200 %q", response.Code, response.Body.String(), "ok\n")
	}
}

func TestHealthHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()
	healthHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("health endpoint returned status %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != "ok\n" {
		t.Fatalf("health endpoint body = %q, want %q", response.Body.String(), "ok\n")
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("health endpoint must not be cached")
	}
}

func TestPublicPagesRender(t *testing.T) {
	content, err := loadContent("content.yaml")
	if err != nil {
		t.Fatal(err)
	}
	feedXML, err := generateRSSFeed(content)
	if err != nil {
		t.Fatal(err)
	}
	router := setupRouter(map[string]Content{"en": content}, parseTemplates(), map[string][]byte{"en": feedXML})
	for _, path := range []string{"/", "/me", "/contacts", "/blog", "/blog/lorem", "/mattrolled"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("GET %s returned %d", path, response.Code)
			}
			body := response.Body.String()
			if !strings.Contains(body, "</html>") || strings.Contains(body, "Something went wrong") {
				t.Fatalf("GET %s did not render a complete page", path)
			}
			if path == "/" {
				for _, expected := range []string{content.Homepage.Role, content.Homepage.CVLabel, content.Homepage.ContactLabel} {
					if !strings.Contains(body, template.HTMLEscapeString(expected)) {
						t.Errorf("homepage is missing configured content %q", expected)
					}
				}
			}
		})
	}
}

func TestGenerateRSSFeed(t *testing.T) {
	content := Content{
		SiteName: "Test Site",
		BaseURL:  "https://example.com",
		Blog: Blog{
			Intro: "Test Blog Intro",
			Posts: []BlogPost{
				{
					Slug:    "test-post",
					Title:   "Test Post Title",
					Date:    "August 2026",
					Summary: "Test summary",
					Tags:    []string{"Go", "Testing"},
					HTML:    template.HTML("<p>Test content</p>"),
				},
			},
		},
	}

	feedBytes, err := generateRSSFeed(content)
	if err != nil {
		t.Fatalf("generateRSSFeed returned error: %v", err)
	}

	// 1. Verify XML is syntactically well-formed
	decoder := xml.NewDecoder(bytes.NewReader(feedBytes))
	for {
		_, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("generated RSS is not well-formed XML: %v", err)
		}
	}

	xmlStr := string(feedBytes)
	expectedSubstrings := []string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:content="http://purl.org/rss/1.0/modules/content/">`,
		`<title>Test Site — Blog</title>`,
		`<link>https://example.com/blog</link>`,
		`<description>Test Blog Intro</description>`,
		`<atom:link href="https://example.com/blog/feed.xml" rel="self" type="application/rss+xml"></atom:link>`,
		`<title>Test Post Title</title>`,
		`<link>https://example.com/blog/test-post</link>`,
		`<guid>https://example.com/blog/test-post</guid>`,
		`<pubDate>Sat, 01 Aug 2026 00:00:00 +0000</pubDate>`,
		`<description>Test summary</description>`,
		`<content:encoded><![CDATA[<p>Test content</p>]]></content:encoded>`,
		`<category>Go</category>`,
		`<category>Testing</category>`,
	}

	for _, expected := range expectedSubstrings {
		if !strings.Contains(xmlStr, expected) {
			t.Errorf("generated RSS missing %q\nFull XML:\n%s", expected, xmlStr)
		}
	}
}

func TestFeedEndpoint(t *testing.T) {
	content, err := loadContent("content.yaml")
	if err != nil {
		t.Fatalf("loadContent failed: %v", err)
	}

	feedXML, err := generateRSSFeed(content)
	if err != nil {
		t.Fatalf("generateRSSFeed failed: %v", err)
	}

	router := setupRouter(map[string]Content{"en": content}, nil, map[string][]byte{"en": feedXML})

	// Test GET /blog/feed.xml
	req := httptest.NewRequest(http.MethodGet, "/blog/feed.xml", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /blog/feed.xml returned status %d, want %d", rec.Code, http.StatusOK)
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "xml") {
		t.Errorf("Content-Type = %q, want XML", contentType)
	}
	if cacheControl := rec.Header().Get("Cache-Control"); !strings.Contains(cacheControl, "public") {
		t.Errorf("Cache-Control = %q, want public caching", cacheControl)
	}

	// Test GET /feed.xml redirect
	reqRedirect := httptest.NewRequest(http.MethodGet, "/feed.xml", nil)
	recRedirect := httptest.NewRecorder()
	router.ServeHTTP(recRedirect, reqRedirect)

	if recRedirect.Code != http.StatusMovedPermanently {
		t.Fatalf("GET /feed.xml returned status %d, want %d", recRedirect.Code, http.StatusMovedPermanently)
	}
	if loc := recRedirect.Header().Get("Location"); loc != "/blog/feed.xml" {
		t.Fatalf("redirect Location = %q, want %q", loc, "/blog/feed.xml")
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		wantIP     string
	}{
		{
			name:       "CF-Connecting-IP takes highest priority",
			headers:    map[string]string{"CF-Connecting-IP": "203.0.113.1", "X-Forwarded-For": "198.51.100.2"},
			remoteAddr: "192.0.2.1:1234",
			wantIP:     "203.0.113.1",
		},
		{
			name:       "X-Forwarded-For first IP when CF header missing",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.5, 70.41.3.18, 150.172.238.178"},
			remoteAddr: "192.0.2.1:1234",
			wantIP:     "203.0.113.5",
		},
		{
			name:       "X-Real-IP used when other proxy headers missing",
			headers:    map[string]string{"X-Real-IP": "203.0.113.9"},
			remoteAddr: "192.0.2.1:1234",
			wantIP:     "203.0.113.9",
		},
		{
			name:       "Fallback to RemoteAddr stripping port",
			headers:    map[string]string{},
			remoteAddr: "198.51.100.42:54321",
			wantIP:     "198.51.100.42",
		},
		{
			name:       "IPv6 handling in headers",
			headers:    map[string]string{"CF-Connecting-IP": "2001:db8::1"},
			remoteAddr: "[::1]:1234",
			wantIP:     "2001:db8::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ifconfig", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			got := getClientIP(req)
			if got != tt.wantIP {
				t.Errorf("getClientIP() = %q, want %q", got, tt.wantIP)
			}
		})
	}
}

func TestIfconfigEndpoint(t *testing.T) {
	router := setupRouter(nil, nil, nil)

	// 1. Test plain text default on /ifconfig
	req := httptest.NewRequest(http.MethodGet, "/ifconfig", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.195")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /ifconfig returned %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "no-store") {
		t.Errorf("Cache-Control = %q, want no-store", cc)
	}
	if rec.Body.String() != "203.0.113.195\n" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "203.0.113.195\n")
	}

	// 2. Test alias /ip
	reqIP := httptest.NewRequest(http.MethodGet, "/ip", nil)
	reqIP.Header.Set("CF-Connecting-IP", "203.0.113.195")
	recIP := httptest.NewRecorder()
	router.ServeHTTP(recIP, reqIP)

	if recIP.Code != http.StatusOK {
		t.Fatalf("GET /ip returned %d, want %d", recIP.Code, http.StatusOK)
	}
	if recIP.Body.String() != "203.0.113.195\n" {
		t.Errorf("/ip body = %q, want %q", recIP.Body.String(), "203.0.113.195\n")
	}

	// 3. Test JSON content negotiation with Accept: application/json
	reqJSON := httptest.NewRequest(http.MethodGet, "/ifconfig", nil)
	reqJSON.Header.Set("Accept", "application/json")
	reqJSON.Header.Set("CF-Connecting-IP", "203.0.113.195")
	recJSON := httptest.NewRecorder()
	router.ServeHTTP(recJSON, reqJSON)

	if recJSON.Code != http.StatusOK {
		t.Fatalf("GET /ifconfig (JSON) returned %d, want %d", recJSON.Code, http.StatusOK)
	}
	if ct := recJSON.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var res map[string]string
	if err := json.Unmarshal(recJSON.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}
	if res["ip"] != "203.0.113.195" {
		t.Errorf("JSON ip = %q, want %q", res["ip"], "203.0.113.195")
	}
}
