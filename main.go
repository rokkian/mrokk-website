package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	Slug    string   `yaml:"slug"`
	Title   string   `yaml:"title"`
	Date    string   `yaml:"date"`
	Summary string   `yaml:"summary"`
	Tags    []string `yaml:"tags"`
	Body    string   `yaml:"body"`
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
}

type Content struct {
	SiteName        string               `yaml:"site_name"`
	MetaDescription string               `yaml:"meta_description"`
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
	Year    int
	Content Content
	Page    string
	Post    *BlogPost
}

func main() {
	tmpl := template.Must(template.New("index.html").Funcs(template.FuncMap{
		"addOne":    func(number int) int { return number + 1 },
		"hasPrefix": func(value, prefix string) bool { return strings.HasPrefix(value, prefix) },
	}).ParseFS(websiteFiles, "templates/*.html"))

	content, err := loadContent("content.yaml")
	if err != nil {
		log.Fatalf("load content.yaml: %v", err)
	}

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
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{Year: time.Now().Year(), Content: content, Page: "landing"}
		if err := tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
			log.Printf("render homepage: %v", err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	})
	r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{Year: time.Now().Year(), Content: content, Page: "resume"}
		if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
			log.Printf("render resume page: %v", err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	})
	r.Get("/contacts", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{Year: time.Now().Year(), Content: content, Page: "contacts"}
		if err := tmpl.ExecuteTemplate(w, "contacts.html", data); err != nil {
			log.Printf("render contacts page: %v", err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	})
	rickrollHandler := func(w http.ResponseWriter, r *http.Request) {
		data := PageData{Year: time.Now().Year(), Content: content, Page: "mattrolled"}
		if err := tmpl.ExecuteTemplate(w, "mattrolled.html", data); err != nil {
			log.Printf("render rickrolled page: %v", err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	}
	r.Get("/rickrolled", rickrollHandler)
	r.Get("/mattrolled", rickrollHandler)

	r.Get("/blog", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{Year: time.Now().Year(), Content: content, Page: "blog"}
		if err := tmpl.ExecuteTemplate(w, "blog.html", data); err != nil {
			log.Printf("render blog page: %v", err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	})
	r.Get("/blog/{slug}", func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		var post *BlogPost
		for index := range content.Blog.Posts {
			if content.Blog.Posts[index].Slug == slug {
				post = &content.Blog.Posts[index]
				break
			}
		}
		if post == nil {
			http.NotFound(w, r)
			return
		}

		data := PageData{Year: time.Now().Year(), Content: content, Page: "post", Post: post}
		if err := tmpl.ExecuteTemplate(w, "blog-post.html", data); err != nil {
			log.Printf("render blog post %s: %v", slug, err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	})

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
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; img-src 'self' data:; media-src 'self'; script-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'")
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
	return content, nil
}
