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
- `/rickrolled` — event NFC joke page with configurable video and contact links

Add posts under `blog.posts` in `content.yaml`. Content remains centralized in that file.

## Where to start customizing

- Edit the copy and contact address in `templates/index.html`.
- Replace the identity, experience, and projects in `content.yaml`.
- Change the visual style in `static/styles.css`.
- Add images to `static/images/` and reference them as `/static/images/filename.jpg`.

`chi` handles routing and common middleware while remaining compatible with `net/http`. The middleware currently provides request IDs, access logs, panic recovery, client IP handling, and compression.

The server reads `content.yaml` once at startup. Restart `go run .` after changing content.

## Deploy to Cloud Run

See [DEPLOYMENT.md](DEPLOYMENT.md) for the concise deployment and Cloudflare operations guide.

The deployment files are in `Dockerfile` and `deploy/cloudrun.yaml`. Replace `PROJECT_ID` in the manifest with your Google Cloud project ID.

```sh
gcloud config set project PROJECT_ID
gcloud services enable run.googleapis.com artifactregistry.googleapis.com
gcloud artifacts repositories create personal-site \
  --repository-format=docker \
  --location=europe-west1

gcloud builds submit \
  --tag europe-west1-docker.pkg.dev/PROJECT_ID/personal-site/mrokk-web:latest

sed 's/PROJECT_ID/your-real-project-id/g' deploy/cloudrun.yaml > /tmp/mrokk-cloudrun.yaml
gcloud run services replace /tmp/mrokk-cloudrun.yaml --region=europe-west1
gcloud run services add-iam-policy-binding mrokk-web \
  --region=europe-west1 \
  --member="allUsers" \
  --role="roles/run.invoker"
```

Cloud Run will print a service URL after deployment. The manifest is the service configuration; the container image must be rebuilt whenever Go code, templates, CSS, JavaScript, images, or `content.yaml` changes.

## Connect `mrokk.com` through Cloudflare

After the service is deployed, map `mrokk.com` to `mrokk-web` using Cloud Run's custom-domain flow. Google will provide the DNS records required for the mapping. Add those records in Cloudflare DNS. If Cloudflare proxying causes certificate validation trouble, temporarily set the record to DNS-only while the Google-managed certificate provisions.

The `embed` package includes the templates and CSS in the compiled program, so the app can be run as a single binary later.
