package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/yuin/goldmark"
	"gopkg.in/yaml.v3"
)

//go:embed templates/*.html static
var websiteFiles embed.FS

type Project struct {
	Name        string
	Description string
	Tags        []string
	Link        string
}

type BlogPost struct {
	Slug    string        `yaml:"slug"`
	Title   string        `yaml:"title"`
	Date    string        `yaml:"date"`
	Summary string        `yaml:"summary"`
	Tags    []string      `yaml:"tags"`
	Source  string        `yaml:"source"`
	Body    string        `yaml:"body"` // Inline Markdown fallback for small posts.
	HTML    template.HTML `yaml:"-"`
}

type Experience struct {
	Period      string
	Role        string
	Company     string
	Location    string
	Description string
}

type Education struct {
	Name        string `yaml:"name"`
	Institution string `yaml:"institution"`
	Location    string `yaml:"location"`
	Description string `yaml:"description"`
	Duration    string `yaml:"duration"`
	Grade       string `yaml:"grade"`
}

type Certification struct {
	Name   string `yaml:"name"`
	Issuer string `yaml:"issuer"`
	Issued string `yaml:"issued"`
	Link   string `yaml:"link"`
	Icon   string `yaml:"icon"`
}

type Content struct {
	Language        string               `yaml:"language"`
	UI              map[string]string    `yaml:"ui"`
	SiteName        string               `yaml:"site_name"`
	BaseURL         string               `yaml:"base_url"`
	MetaDescription string               `yaml:"meta_description"`
	Homepage        Homepage             `yaml:"homepage"`
	Navigation      Navigation           `yaml:"navigation"`
	Hero            Hero                 `yaml:"hero"`
	About           About                `yaml:"about"`
	Experience      ExperienceSection    `yaml:"experience"`
	Certifications  CertificationSection `yaml:"certifications"`
	Work            Work                 `yaml:"work"`
	Education       EducationSection     `yaml:"education"`
	Skills          []string             `yaml:"skills"`
	Blog            Blog                 `yaml:"blog"`
	Mattrolled      Mattrolled           `yaml:"mattrolled"`
	Contact         Contact              `yaml:"contact"`
	Footer          Footer               `yaml:"footer"`
	Theme           Theme                `yaml:"theme"`
}

type Homepage struct {
	Name         string `yaml:"name"`
	Role         string `yaml:"role"`
	Intro        string `yaml:"intro"`
	Evidence     string `yaml:"evidence"`
	CVLabel      string `yaml:"cv_label"`
	ContactLabel string `yaml:"contact_label"`
}

type RSSFeed struct {
	XMLName      xml.Name   `xml:"rss"`
	Version      string     `xml:"version,attr"`
	XMLNSAtom    string     `xml:"xmlns:atom,attr"`
	XMLNSContent string     `xml:"xmlns:content,attr"`
	Channel      RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language,omitempty"`
	LastBuildDate string    `xml:"lastBuildDate"`
	AtomLink      AtomLink  `xml:"atom:link"`
	Items         []RSSItem `xml:"item"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type RSSItem struct {
	Title       string     `xml:"title"`
	Link        string     `xml:"link"`
	GUID        string     `xml:"guid"`
	PubDate     string     `xml:"pubDate"`
	Description string     `xml:"description"`
	Content     *CDATAText `xml:"content:encoded,omitempty"`
	Categories  []string   `xml:"category,omitempty"`
}

type CDATAText struct {
	Value string `xml:",cdata"`
}

type Navigation struct {
	About      string `yaml:"about"`
	Experience string `yaml:"experience"`
	Work       string `yaml:"work"`
	Blog       string `yaml:"blog"`
	Contact    string `yaml:"contact"`
}

type Theme struct {
	LightLabel string `yaml:"light_label"`
	DarkLabel  string `yaml:"dark_label"`
}

type Hero struct {
	Eyebrow  string `yaml:"eyebrow"`
	Title    string `yaml:"title"`
	Intro    string `yaml:"intro"`
	Cta      string `yaml:"cta"`
	Image    string `yaml:"image"`
	ImageAlt string `yaml:"image_alt"`
}

type About struct {
	Label string `yaml:"label"`
	Title string `yaml:"title"`
	Text  string `yaml:"text"`
}

type ExperienceSection struct {
	Label   string       `yaml:"label"`
	Intro   string       `yaml:"intro"`
	Entries []Experience `yaml:"entries"`
}

type CertificationSection struct {
	Label        string          `yaml:"label"`
	Intro        string          `yaml:"intro"`
	Certificates []Certification `yaml:"certificates"`
}

type Work struct {
	Label    string    `yaml:"label"`
	Intro    string    `yaml:"intro"`
	Projects []Project `yaml:"projects"`
}

type EducationSection struct {
	Label   string      `yaml:"label"`
	Intro   string      `yaml:"intro"`
	Degrees []Education `yaml:"degrees"`
}

type Blog struct {
	Label string     `yaml:"label"`
	Intro string     `yaml:"intro"`
	Posts []BlogPost `yaml:"posts"`
}

type Mattrolled struct {
	Eyebrow  string `yaml:"eyebrow"`
	Title    string `yaml:"title"`
	Message  string `yaml:"message"`
	Gif      string `yaml:"gif"`
	GifAlt   string `yaml:"gif_alt"`
	Video    string `yaml:"video"`
	Telegram string `yaml:"telegram"`
	LinkedIn string `yaml:"linkedin"`
}

type Contact struct {
	Label string        `yaml:"label"`
	Title string        `yaml:"title"`
	Cta   string        `yaml:"cta"`
	Email string        `yaml:"email"`
	Links []ContactLink `yaml:"links"`
}

type ContactLink struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	Description string `yaml:"description"`
}

type Footer struct {
	Tagline string `yaml:"tagline"`
}

type PageData struct {
	Language string
	Path     string
	Year     int
	Content  Content
	Page     string
	Post     *BlogPost
}

func parseTemplates() *template.Template {
	return template.Must(template.New("index.html").Funcs(template.FuncMap{
		"hasPrefix": func(value, prefix string) bool { return strings.HasPrefix(value, prefix) },
		"markdown": func(value string) (template.HTML, error) {
			rendered, err := renderMarkdown(value)
			return template.HTML(rendered), err
		},
	}).ParseFS(websiteFiles, "templates/*.html"))
}

var dateLayouts = []string{
	"January 2006",
	"2006-01-02",
	"02 Jan 2006",
	time.RFC3339,
	time.RFC822,
	time.RFC1123,
	time.RFC1123Z,
}

func parsePostDate(dateStr string) time.Time {
	clean := strings.TrimSpace(dateStr)
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, clean); err == nil {
			return t
		}
	}
	return time.Now().UTC()
}

func generateRSSFeed(content Content) ([]byte, error) {
	baseURL := publicBaseURL(content)

	channelTitle := content.SiteName + " — Blog"
	channelLink := baseURL + localizedPath(content.Language, "/blog")
	feedURL := baseURL + localizedPath(content.Language, "/blog/feed.xml")

	items := make([]RSSItem, 0, len(content.Blog.Posts))
	for _, post := range content.Blog.Posts {
		postURL := baseURL + localizedPath(content.Language, "/blog/"+post.Slug)
		pubDate := parsePostDate(post.Date).Format(time.RFC1123Z)

		var encodedContent *CDATAText
		if post.HTML != "" {
			encodedContent = &CDATAText{Value: string(post.HTML)}
		}

		items = append(items, RSSItem{
			Title:       post.Title,
			Link:        postURL,
			GUID:        postURL,
			PubDate:     pubDate,
			Description: post.Summary,
			Content:     encodedContent,
			Categories:  post.Tags,
		})
	}

	feed := RSSFeed{
		Version:      "2.0",
		XMLNSAtom:    "http://www.w3.org/2005/Atom",
		XMLNSContent: "http://purl.org/rss/1.0/modules/content/",
		Channel: RSSChannel{
			Title:         channelTitle,
			Link:          channelLink,
			Description:   content.Blog.Intro,
			Language:      languageOrEnglish(content.Language),
			LastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
			AtomLink: AtomLink{
				Href: feedURL,
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: items,
		},
	}

	output, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), output...), nil
}

func setupRouter(contents map[string]Content, tmpl *template.Template, feeds map[string][]byte) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Throttle(100))
	r.Use(readOnlyRequests)
	r.Use(securityHeaders)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Mount("/static", http.FileServer(http.FS(websiteFiles)))
	r.Get("/healthz", healthHandler)

	r.Get("/ifconfig", ifconfigHandler)
	r.Get("/ip", ifconfigHandler)
	for _, language := range []string{"en", "it"} {
		if content, ok := contents[language]; ok {
			content.Language = language
			registerLocalizedPages(r, content, tmpl, feeds[language])
		}
	}
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		language := "en"
		if strings.HasPrefix(r.URL.Path, "/it/") || r.URL.Path == "/it" {
			language = "it"
		}
		w.Header().Set("Content-Language", language)
		http.Error(w, contents[language].UI["not_found"], http.StatusNotFound)
	})

	return r
}

func main() {
	tmpl := parseTemplates()

	contents, err := loadSiteContent(".")
	if err != nil {
		log.Fatalf("load site content: %v", err)
	}
	feeds := make(map[string][]byte, len(contents))
	for language, content := range contents {
		feeds[language], err = generateRSSFeed(content)
		if err != nil {
			log.Fatalf("generate %s RSS feed: %v", language, err)
		}
	}
	r := setupRouter(contents, tmpl, feeds)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}

	log.Printf("personal website listening on port %s", port)
	log.Fatal(server.ListenAndServe())
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func getClientIP(r *http.Request) string {
	// 1. Cloudflare authoritative connecting IP
	if cfIP := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); cfIP != "" {
		if net.ParseIP(cfIP) != nil {
			return cfIP
		}
	}

	// 2. Standard X-Forwarded-For header (leftmost IP is the original client)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			first := strings.TrimSpace(parts[0])
			if net.ParseIP(first) != nil {
				return first
			}
		}
	}

	// 3. X-Real-IP header
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		if net.ParseIP(realIP) != nil {
			return realIP
		}
	}

	// 4. Fallback to RemoteAddr (strip port if present)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		if net.ParseIP(host) != nil {
			return host
		}
	}
	cleanRemote := strings.TrimSpace(r.RemoteAddr)
	if net.ParseIP(cleanRemote) != nil {
		return cleanRemote
	}

	return r.RemoteAddr
}

func ifconfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
	w.Header().Set("Pragma", "no-cache")

	ip := getClientIP(r)

	// Content negotiation: return JSON if Accept header asks for application/json or format=json query parameter
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") || r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"ip": ip})
		return
	}

	// Default plain text with newline (ifconfig.me style for CLI / curl)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "%s\n", ip)
}

func readOnlyRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if len(r.RequestURI) > 8<<10 {
			http.Error(w, "request URI too long", http.StatusRequestURITooLong)
			return
		}

		// There are no request bodies in this application. Limit one anyway so a
		// future handler cannot accidentally accept an unbounded upload.
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; media-src 'self'; script-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

func loadContent(path string) (Content, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Content{}, err
	}

	var content Content
	if err := yaml.Unmarshal(contents, &content); err != nil {
		return Content{}, err
	}
	if content.SiteName == "" {
		return Content{}, fmt.Errorf("site_name is required")
	}

	for index := range content.Blog.Posts {
		post := &content.Blog.Posts[index]
		markdown := post.Body
		if post.Source != "" {
			sourcePath := filepath.Join(filepath.Dir(path), "blog", post.Source)
			blogDir, err := filepath.Abs(filepath.Join(filepath.Dir(path), "blog"))
			if err != nil {
				return Content{}, fmt.Errorf("resolve blog directory: %w", err)
			}
			resolvedSource, err := filepath.Abs(sourcePath)
			if err != nil {
				return Content{}, fmt.Errorf("resolve blog post %q: %w", post.Source, err)
			}
			relativeSource, err := filepath.Rel(blogDir, resolvedSource)
			if err != nil || relativeSource == ".." || strings.HasPrefix(relativeSource, ".."+string(filepath.Separator)) {
				return Content{}, fmt.Errorf("blog post source escapes blog directory: %q", post.Source)
			}
			contents, err := os.ReadFile(resolvedSource)
			if err != nil {
				return Content{}, fmt.Errorf("read blog post %q: %w", post.Source, err)
			}
			markdown = string(contents)
		}

		rendered, err := renderMarkdown(markdown)
		if err != nil {
			return Content{}, fmt.Errorf("render blog post %q: %w", post.Slug, err)
		}
		post.HTML = template.HTML(rendered)
	}
	return content, nil
}

func renderMarkdown(markdown string) ([]byte, error) {
	var rendered bytes.Buffer
	if err := goldmark.New().Convert([]byte(markdown), &rendered); err != nil {
		return nil, err
	}
	return rendered.Bytes(), nil
}
