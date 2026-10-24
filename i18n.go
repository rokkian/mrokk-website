package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// English keeps the original URLs; Italian pages have an explicit /it prefix.
// Links carry the selected language without cookies or client-side redirects.
func localizedPath(language, path string) string {
	if language == "it" && strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "//") {
		return "/it" + path
	}
	return path
}

func publicBaseURL(content Content) string {
	if content.BaseURL != "" {
		return strings.TrimRight(content.BaseURL, "/")
	}
	return "https://mrokk.com"
}

func languageOrEnglish(language string) string {
	if language == "" {
		return "en"
	}
	return language
}

func (p PageData) URL(path string) string { return localizedPath(p.Language, path) }

func (p PageData) OtherLanguage() string {
	if p.Language == "it" {
		return "en"
	}
	return "it"
}

func (p PageData) LanguageURL() string   { return localizedPath(p.OtherLanguage(), p.Path) }
func (p PageData) LanguageLabel() string { return strings.ToUpper(p.OtherLanguage()) }
func (p PageData) LanguageTitle() string {
	if p.Language == "it" {
		return "Read in English"
	}
	return "Leggi in italiano"
}

func (p PageData) CanonicalURL() string { return publicBaseURL(p.Content) + p.URL(p.Path) }
func (p PageData) EnglishURL() string   { return publicBaseURL(p.Content) + p.Path }
func (p PageData) ItalianURL() string   { return publicBaseURL(p.Content) + localizedPath("it", p.Path) }

func (p PageData) Date(value string) string {
	if p.Language != "it" {
		return value
	}
	return italianMonths.Replace(value)
}

var italianMonths = strings.NewReplacer(
	"January", "gennaio", "February", "febbraio", "March", "marzo",
	"April", "aprile", "May", "maggio", "June", "giugno",
	"July", "luglio", "August", "agosto", "September", "settembre",
	"October", "ottobre", "November", "novembre", "December", "dicembre",
)

func loadSiteContent(directory string) (map[string]Content, error) {
	contents := make(map[string]Content, 2)
	for _, locale := range []struct{ language, file string }{{"en", "content.yaml"}, {"it", "content.it.yaml"}} {
		content, err := loadContent(filepath.Join(directory, locale.file))
		if err != nil {
			return nil, fmt.Errorf("load %s content: %w", locale.language, err)
		}
		content.Language = locale.language
		contents[locale.language] = content
	}
	for key := range contents["en"].UI {
		if strings.TrimSpace(contents["it"].UI[key]) == "" {
			return nil, fmt.Errorf("missing Italian UI translation: %s", key)
		}
	}
	italian := contents["it"]
	if italian.BaseURL == "" {
		italian.BaseURL = contents["en"].BaseURL
		contents["it"] = italian
	}
	return contents, nil
}

func registerLocalizedPages(r chi.Router, content Content, tmpl *template.Template, feedXML []byte) {
	prefix := ""
	if content.Language == "it" {
		prefix = "/it"
		r.Get(prefix, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, prefix+"/", http.StatusMovedPermanently)
		})
	}

	r.Get(prefix+"/blog/feed.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.Header().Set("Content-Language", content.Language)
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = w.Write(feedXML)
	})
	r.Get(prefix+"/feed.xml", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, prefix+"/blog/feed.xml", http.StatusMovedPermanently)
	})

	render := func(w http.ResponseWriter, r *http.Request, page, name string, post *BlogPost) {
		if tmpl == nil {
			return
		}
		w.Header().Set("Content-Language", content.Language)
		data := PageData{Year: time.Now().Year(), Content: content, Page: page, Post: post,
			Language: content.Language, Path: strings.TrimPrefix(r.URL.Path, prefix)}
		var body bytes.Buffer
		if err := tmpl.ExecuteTemplate(&body, name, data); err != nil {
			log.Printf("render %s %s: %v", content.Language, page, err)
			http.Error(w, content.UI["site_error"], http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(body.Bytes())
	}

	for _, route := range []struct{ path, page, template string }{
		{"/", "landing", "home.html"},
		{"/me", "resume", "index.html"},
		{"/contacts", "contacts", "contacts.html"},
		{"/blog", "blog", "blog.html"},
		{"/mattrolled", "mattrolled", "mattrolled.html"},
		{"/rickrolled", "mattrolled", "mattrolled.html"},
	} {
		r.Get(prefix+route.path, func(w http.ResponseWriter, r *http.Request) {
			render(w, r, route.page, route.template, nil)
		})
	}

	r.Get(prefix+"/blog/{slug}", func(w http.ResponseWriter, r *http.Request) {
		for index := range content.Blog.Posts {
			if content.Blog.Posts[index].Slug == chi.URLParam(r, "slug") {
				render(w, r, "post", "blog-post.html", &content.Blog.Posts[index])
				return
			}
		}
		w.Header().Set("Content-Language", content.Language)
		http.Error(w, content.UI["not_found"], http.StatusNotFound)
	})
}
