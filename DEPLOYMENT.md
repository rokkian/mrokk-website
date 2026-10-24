# Deployment guide

This deploys the Go website to Cloud Run and connects `mrokk.com` through Cloudflare.

## 1. Set your project values

Install and authenticate the [Google Cloud CLI](https://cloud.google.com/sdk/docs/install), then run:

```sh
gcloud auth login
gcloud config set project YOUR_PROJECT_ID
gcloud config set run/region europe-west1
```

## 2. Enable services and create a registry

Run once per Google Cloud project:

```sh
gcloud services enable run.googleapis.com artifactregistry.googleapis.com cloudbuild.googleapis.com

gcloud artifacts repositories create personal-site \
  --repository-format=docker \
  --location=europe-west1
```

If the repository already exists, skip that command.

## 3. Build and push the image

Run this after every application or content change:

```sh
gcloud builds submit \
  --tag europe-west1-docker.pkg.dev/YOUR_PROJECT_ID/personal-site/mrokk-web:latest
```

## 4. Deploy the Cloud Run service

Replace `PROJECT_ID` in [deploy/cloudrun.yaml](deploy/cloudrun.yaml) with your actual project ID, then run:

```sh
gcloud run services replace deploy/cloudrun.yaml \
  --region=europe-west1

gcloud run services add-iam-policy-binding mrokk-web \
  --region=europe-west1 \
  --member="allUsers" \
  --role="roles/run.invoker"
```

Check the generated Cloud Run URL:

```sh
gcloud run services describe mrokk-web \
  --region=europe-west1 \
  --format='value(status.url)'
```

## 5. Connect `mrokk.com` in Cloudflare

Create the Cloud Run domain mapping:

```sh
gcloud beta run domain-mappings create \
  --service=mrokk-web \
  --domain=mrokk.com \
  --region=europe-west1
```

Display the required DNS records:

```sh
gcloud beta run domain-mappings describe \
  --domain=mrokk.com \
  --region=europe-west1
```

In Cloudflare, add every record listed under `resourceRecords`. Use **DNS only** while the Google-managed certificate is provisioning. After `https://mrokk.com` works, Cloudflare proxying can be enabled if desired.

## 6. Automated CI/CD with GitHub Actions

[CI and deploy](.github/workflows/ci.yml) replaces the automated Cloud Build path:

1. Pull requests and pushes to `main` run formatting checks, `go vet`, and Go tests with the race detector. Go's version comes from `go.mod`.
2. A separate job builds the Linux/amd64 runtime image with Buildx and GitHub's layer cache, starts it, and checks `/healthz`, `/`, and `/it/`. This also validates the runtime content files and embedded assets on pull requests.
3. Once both jobs pass on `main`, the `production` deployment job authenticates through GitHub OIDC and Google Workload Identity Federation (WIF). It publishes full-commit-SHA and `latest` tags to `europe-west1-docker.pkg.dev/pinebula/personal-site/mrokk-web`, with build provenance and an SBOM.
4. Cloud Run deploys the immutable **image digest**, routes traffic to the new revision, and checks its public `/healthz`. The Actions summary records the commit, digest, and service URL.

The service uses the existing runtime service account, one CPU, 256 MiB RAM, port 8080, and revision autoscaling from zero to one instance, matching `cloudbuild.yaml`. Its existing public access and domain mapping are preserved. Failed checks stop deployment; a failed post-deploy health check marks the run as failed and does not automatically roll back traffic.

### One-time WIF setup

Authenticate `gcloud` as a Pinebula administrator and `gh` as a repository administrator, then run from the repository root:

```sh
python3 deploy/setup-github-wif.py
```

The [setup script](deploy/setup-github-wif.py) checks the existing service and registry, enables the required APIs, creates or reuses the dedicated identity, and configures GitHub. It can be rerun after a partial failure. It creates:

- Workload Identity Pool `mrokk-website` with provider `github` in project `pinebula`.
- Service account `github-mrokk-deploy@pinebula.iam.gserviceaccount.com`.
- GitHub environment `production`, with a deployment branch rule for `main`.
- Repository variables `GCP_WIF_PROVIDER`, `GCP_SERVICE_ACCOUNT`, and `GCP_RUNTIME_SERVICE_ACCOUNT`. These identifiers are configuration, not secrets.

The WIF provider checks the numeric repository and owner IDs, `refs/heads/main`, the exact workflow path `.github/workflows/ci.yml`, the `production` subject, and the `push` or `workflow_dispatch` event. Only the deployment job requests `id-token: write`; PR jobs have read-only repository permissions and no Google Cloud credentials. This follows Google's guidance to [restrict federation using numeric IDs](https://docs.cloud.google.com/iam/docs/workload-identity-federation-with-deployment-pipelines).

The deployment account has only these resource-scoped roles:

| Role | Resource |
| --- | --- |
| `roles/artifactregistry.writer` | `personal-site` Artifact Registry repository |
| `roles/run.developer` | Existing `mrokk-web` Cloud Run service |
| `roles/iam.serviceAccountUser` | Existing Cloud Run runtime service account |
| `roles/iam.workloadIdentityUser` | Deployment account, granted to this repository's federated identity |

No service account key is generated. The Google authentication action's temporary `gha-creds-*.json` files are excluded from Git, Docker contexts, and Cloud Build uploads. Cloud Run authentication is refreshed after image publishing to account for the short OIDC token lifetime. See the [authentication action](https://github.com/google-github-actions/auth) for details.

### Operation and maintenance

Push to `main` to deploy, or select **Actions → CI and deploy → Run workflow** with branch `main`. IAM changes can take a few minutes to propagate after the initial setup. Production runs are serialized; superseded pending runs can be replaced by a newer push, while active production runs are allowed to finish. New PR runs cancel superseded PR checks.

All third-party actions are pinned to commit SHAs. [Dependabot](.github/dependabot.yml) checks monthly for updates to GitHub Actions, Go modules (including indirect dependencies), and Docker base images. Actions updates are grouped; Go minor and patch updates are grouped, while Go major updates and Docker image updates get separate PRs. Dependency PRs run the same Go and container checks; merging to `main` triggers deployment. If the repository, owner, workflow path, branch, or environment changes, update the workflow and setup script together and rerun setup.

`cloudbuild.yaml` remains available for manual use. Keep Cloud Build push triggers disabled when using GitHub Actions to avoid competing deployments.

To roll back to a previously healthy revision:

```sh
gcloud run revisions list --project=pinebula --region=europe-west1 --service=mrokk-web
gcloud run services update-traffic mrokk-web \
  --project=pinebula --region=europe-west1 --to-revisions=REVISION_NAME=100
```

Choose the known healthy revision from the list. The next successful Actions deployment moves traffic to its new revision again.

## Updating the site

- **Automated**: Push your changes to `main` on GitHub; GitHub Actions runs checks, builds and publishes the image, and deploys through WIF.
- **Manual**: Run steps 3 and 4 (or `just deploy`).

