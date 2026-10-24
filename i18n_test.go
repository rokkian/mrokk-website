package main

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func bilingualTestRouter(t *testing.T) (map[string]Content, http.Handler) {
	t.Helper()
	contents, err := loadSiteContent(".")
	if err != nil {
		t.Fatal(err)
	}
	feeds := make(map[string][]byte)
	for language, content := range contents {
		feeds[language], err = generateRSSFeed(content)
		if err != nil {
			t.Fatal(err)
		}
	}
	return contents, setupRouter(contents, parseTemplates(), feeds)
}

func TestBilingualPages(t *testing.T) {
	contents, router := bilingualTestRouter(t)
	paths := []string{"/", "/me", "/contacts", "/blog", "/blog/lorem", "/mattrolled", "/rickrolled"}
	for _, language := range []string{"en", "it"} {
		for _, path := range paths {
			t.Run(language+path, func(t *testing.T) {
				url := localizedPath(language, path)
				response := httptest.NewRecorder()
				router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, url, nil))
				if response.Code != http.StatusOK {
					t.Fatalf("GET %s returned %d: %s", url, response.Code, response.Body.String())
				}
				if response.Header().Get("Content-Language") != language {
					t.Fatalf("wrong Content-Language: %s", response.Header().Get("Content-Language"))
				}
				body := response.Body.String()
				for _, expected := range []string{
					`<html lang="` + language + `"`,
					`rel="canonical" href="https://mrokk.com` + url + `"`,
					`hreflang="en" href="https://mrokk.com` + path + `"`,
					`hreflang="it" href="https://mrokk.com` + localizedPath("it", path) + `"`,
					`href="` + localizedPath(language, "/me") + `"`,
					`href="` + localizedPath(language, "/contacts") + `"`,
				} {
					if !strings.Contains(body, expected) {
						t.Errorf("missing %q", expected)
					}
				}
				if strings.Contains(body, "<no value>") || !strings.Contains(body, "</html>") {
					t.Fatal("page has missing translations or incomplete HTML")
				}
				other := "it"
				if language == "it" {
					other = "en"
				}
				if !strings.Contains(body, `class="language-switch" href="`+localizedPath(other, path)+`"`) {
					t.Fatal("language switch must keep the equivalent page")
				}
				if language == "it" {
					// Internal content links retain Italian; only the language switch leaves it.
					for _, match := range regexp.MustCompile(`href="([^"]+)"`).FindAllStringSubmatch(body, -1) {
						href := match[1]
						if strings.HasPrefix(href, "/") && !strings.HasPrefix(href, "/static/") && !strings.HasPrefix(href, "/it/") && href != path {
							t.Errorf("Italian page links outside its locale: %s", href)
						}
					}
				}
			})
		}
	}
	if contents["en"].Contact.Email != contents["it"].Contact.Email {
		t.Fatal("translations must preserve the contact address")
	}
	for index, english := range contents["en"].Certifications.Certificates {
		italian := contents["it"].Certifications.Certificates[index]
		if english.Name != italian.Name || english.Link != italian.Link || english.Icon != italian.Icon {
			t.Errorf("certification identity changed in translation: %s", english.Name)
		}
	}
}

func TestItalianContentAndControls(t *testing.T) {
	contents, router := bilingualTestRouter(t)
	for path, expected := range map[string][]string{
		"/it/":           {contents["it"].Homepage.Intro, "Guarda il CV", "Contattami", `data-pause-label="Ferma la pioggia"`, `data-resume-label="Riprendi la pioggia"`},
		"/it/me":         {"Esperienza", "Certificazioni", "Strumenti e competenze", "Progetti selezionati", "Giugno 2024", "Progettazione di un'architettura", `id="certifications" open`},
		"/it/contacts":   {"Restiamo in contatto", "Mandami un&#39;email", "Scrivimi su Telegram"},
		"/it/blog":       {"Segui via RSS", "agosto 2026", "Articolo di esempio"},
		"/it/blog/lorem": {"Tutti gli articoli", "agosto 2026", "<p>Articolo di esempio</p>"},
		"/it/mattrolled": {"Sei stato Mattrollato.", "Torna al sito"},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		for _, text := range expected {
			if !strings.Contains(response.Body.String(), text) {
				t.Errorf("%s missing %q", path, text)
			}
		}
	}
}

func TestLocalizedFeedsAndRedirects(t *testing.T) {
	_, router := bilingualTestRouter(t)
	var publicationDates []string
	for _, language := range []string{"en", "it"} {
		response := httptest.NewRecorder()
		path := localizedPath(language, "/blog/feed.xml")
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		var feed RSSFeed
		if err := xml.Unmarshal(response.Body.Bytes(), &feed); err != nil {
			t.Fatalf("invalid %s XML: %v", language, err)
		}
		if response.Code != http.StatusOK || feed.Channel.Language != language || response.Header().Get("Content-Language") != language {
			t.Fatalf("wrong %s feed response", language)
		}
		if !strings.Contains(response.Body.String(), "<link>https://mrokk.com"+localizedPath(language, "/blog")+"</link>") || feed.Channel.Items[0].Link != "https://mrokk.com"+localizedPath(language, "/blog/lorem") {
			t.Fatalf("wrong localized feed links: %+v", feed.Channel)
		}
		publicationDates = append(publicationDates, feed.Channel.Items[0].PubDate)
	}
	if publicationDates[0] != publicationDates[1] || publicationDates[0] != "Sat, 01 Aug 2026 00:00:00 +0000" {
		t.Fatalf("translated dates changed RSS publication time: %v", publicationDates)
	}
	for from, to := range map[string]string{"/it": "/it/", "/it/feed.xml": "/it/blog/feed.xml", "/feed.xml": "/blog/feed.xml"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, from, nil))
		if response.Code != http.StatusMovedPermanently || response.Header().Get("Location") != to {
			t.Errorf("%s redirects to %q (%d), want %s", from, response.Header().Get("Location"), response.Code, to)
		}
	}
}

func TestLocalizedNotFound(t *testing.T) {
	_, router := bilingualTestRouter(t)
	for _, path := range []string{"/it/blog/missing", "/it/missing"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound || response.Header().Get("Content-Language") != "it" || !strings.Contains(response.Body.String(), "Pagina non trovata") {
			t.Errorf("wrong localized 404 for %s: %d %s", path, response.Code, response.Body.String())
		}
	}
}

func TestLocalizedPathPreservesExternalAndFragmentLinks(t *testing.T) {
	for _, path := range []string{"#contact", "mailto:hello@mrokk.com", "https://example.com", "//example.com/icon"} {
		if got := localizedPath("it", path); got != path {
			t.Errorf("localizedPath(%q) = %q", path, got)
		}
	}
}
