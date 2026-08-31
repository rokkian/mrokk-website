package main

import (
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
		"Content-Security-Policy":   "default-src 'self'; style-src 'self' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; img-src 'self' data:; media-src 'self'; script-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'",
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
