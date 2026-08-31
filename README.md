# Personal website in Go

Small server-rendered personal website using Go, the `chi` router, and Go's standard library. Site content lives in `content.yaml`.

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
- `/contacts` — public contact links
- `/mattrolled` — event NFC joke page with configurable video and contact links
- `/healthz` — lightweight health check returning `200 OK`

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

- Edit the copy and contact address in `templates/index.html`.
- Replace the identity, experience, and projects in `content.yaml`.
- Change the visual style in `static/styles.css`.
- Add images to `static/images/` and reference them as `/static/images/filename.jpg`.

`chi` handles routing and common middleware while remaining compatible with `net/http`. The middleware currently provides request IDs, access logs, panic recovery, client IP handling, and compression.

The server reads `content.yaml` once at startup. Restart `go run .` after changing content.

## Deploy to Cloud Run

See [DEPLOYMENT.md](DEPLOYMENT.md) for the Cloud Run and Cloudflare guide. The [justfile](justfile) provides the main commands.

Set `GCP_PROJECT_ID` if your Google Cloud project is not `pinebula`, then run `just config`, `just enable-apis`, `just create-repository`, and `just deploy`.

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
