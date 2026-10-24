# Personal website in Go

Small server-rendered personal website using Go, the `chi` router, and Go's standard library. English content lives in `content.yaml`, with the Italian version in `content.it.yaml`.

## Run it

```sh
go run .
```

Then open <http://localhost:8080>. Once deployed, the site can be served at <https://mrokk.com>.

## Pages

The site is organized into a landing page, a one-page CV, and a separate blog:

- `/` — intro landing page
- `/me` — one-page CV and portfolio
- `/blog` — blog index
- `/blog/{slug}` — individual blog post
- `/blog/feed.xml` — RSS 2.0 / Atom feed
- `/ifconfig` (or `/ip`) — caller IP network echo (`text/plain` or `application/json`)
- `/contacts` — public contact links
- `/mattrolled` — event NFC joke page with configurable video and contact links
- `/healthz` — lightweight health check returning `200 OK`

Italian pages use the same paths under `/it/`, including `/it/me`, `/it/contacts`, `/it/blog` and `/it/blog/feed.xml`. `/it` redirects to `/it/`. The shared IT/EN link switches to the equivalent page, and internal navigation keeps the selected language. Existing English URLs remain unchanged. Language selection uses explicit URLs and works without JavaScript.

Translate page content and the `ui` labels in both YAML files. Certificate names remain official names; each locale keeps the same badge images, credential URLs and contact addresses. Blog translations use matching slugs with separate Markdown sources, such as `lorem.md` and `lorem.it.md`. Keep blog `date` values in the existing canonical English-month or supported ISO format in both files: Italian month names are formatted for display without changing RSS publication dates. Both files load at startup; a missing Italian UI label fails startup rather than rendering an incomplete interface.

Add posts under `blog.posts` in `content.yaml`. Each post can point to a Markdown file in `blog/`:

```yaml
blog:
  posts:
    - slug: first-post
      title: My first post
      date: August 2026
      summary: A short description shown on the blog index.
      tags: [Go, Cloud Run]
      source: first-post.md
```

Then create `blog/first-post.md` using normal Markdown. The post is rendered to safe HTML at startup and included in the deployed container. Inline `body` Markdown is also supported for small posts.

## Where to start customizing

- Edit the homepage name, introduction, and link labels under `homepage` in `content.yaml`; the email address comes from `contact.email`.
- Replace the identity, experience, and projects in `content.yaml`.
- Change shared typography, palettes, header and footer in `static/site.css` and the shared theme behavior in `static/theme.js`. Homepage composition lives in `static/home.css` and its rain interaction in `static/garden.js`. Its transparent illustration is `static/images/gravity-garden.webp`; only the blue cloud emits animated droplets, using `static/images/garden-drop.webp`. The original supplied JPEG is retained as `static/images/gravity-garden-final.jpeg`. The self-hosted Quicksand font and license are in `static/fonts/`.
- Change compact CV, blog, contact, and NFC reading layouts in `static/styles.css`. CV role descriptions use Markdown; native disclosure controls work without JavaScript and expand for printing. All pages use shared header, footer and asset partials in `templates/partials.html`, with one saved theme preference.
- Add images to `static/images/` and reference them as `/static/images/filename.jpg`.
- Browser icons are shared by every page through `site-icons` in `templates/partials.html`. The supplied `static/images/favicon-new.jpeg` is the source for the 16px, 32px, and 48px PNG favicons and the 180px Apple touch icon.

`chi` handles routing and common middleware while remaining compatible with `net/http`. The middleware currently provides request IDs, access logs, panic recovery, client IP handling, and compression.

The server reads both content files once at startup. Restart `go run .` after changing content. The runtime container includes both files and the translated blog sources.

## Deploy to Cloud Run

See [DEPLOYMENT.md](DEPLOYMENT.md) for the Cloud Run and Cloudflare guide. The [justfile](justfile) provides the main commands.

GitHub Actions checks pull requests and automatically deploys successful pushes to `main` to Pinebula through Workload Identity Federation. The [CI/CD setup](DEPLOYMENT.md#6-automated-cicd-with-github-actions) provisions the dedicated identity and GitHub configuration without service account keys.

For manual deployment, set `GCP_PROJECT_ID` if your Google Cloud project is not `pinebula`, then run `just config`, `just enable-apis`, `just create-repository`, and `just deploy`.

## Connect `mrokk.com` through Cloudflare

After deployment, connect `mrokk.com` with:

```sh
just verify-domain
just map-domain
just domain-records
```

If Google asks you to verify ownership, complete the displayed Search Console verification step. Then copy every `resourceRecords` entry from `just domain-records` into Cloudflare DNS. Use the exact record types and values; for the apex domain, Cloud Run commonly returns `@` A/AAAA records.

Set the Cloudflare records to **DNS only** (gray cloud) while the Google-managed HTTPS certificate provisions. Certificate provisioning can take several minutes and occasionally up to 24 hours. Once `https://mrokk.com` works, you can try enabling Cloudflare proxying. If validation or renewal has trouble, leave the records DNS-only and ensure Cloudflare's **Always Use HTTPS** setting is disabled. Cloud Run domain mappings are currently Preview; Google recommends a global external Application Load Balancer for production-grade custom-domain setups. [Cloud Run custom domains](https://docs.cloud.google.com/run/docs/mapping-custom-domains)

The `embed` package includes the templates and CSS in the compiled program, so the app can be run as a single binary later.
