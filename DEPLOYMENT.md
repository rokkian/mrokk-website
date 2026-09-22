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

## 6. Automated CI/CD with Cloud Build (GitHub)

The repository includes a [`cloudbuild.yaml`](cloudbuild.yaml) pipeline that runs tests, builds and tags the container image (with `$SHORT_SHA` and `latest`), pushes to Artifact Registry, and deploys to Cloud Run.

### Step A: Grant Cloud Build permissions

To deploy to Cloud Run, Cloud Build's service account needs the appropriate IAM roles. Run:

```sh
PROJECT_ID="pinebula"  # or your project ID
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
CB_SA="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"

# Allow Cloud Build to manage Cloud Run
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${CB_SA}" \
  --role="roles/run.admin"

# Allow Cloud Build to push container images to Artifact Registry
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${CB_SA}" \
  --role="roles/artifactregistry.writer"

# Allow Cloud Build to act as the runtime service account
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${CB_SA}" \
  --role="roles/iam.serviceAccountUser"
```

> **Note**: If your project uses the Compute Engine service account for Cloud Build (`${PROJECT_NUMBER}-compute@developer.gserviceaccount.com`), grant the roles to that account instead.

### Step B: Connect your GitHub repository

1. In the Google Cloud Console, navigate to **Cloud Build** > **Triggers** (or **Repositories**).
2. Click **Connect Repository** (or **Manage Repositories**).
3. Select **GitHub (Cloud Build GitHub App)** or **2nd gen** and authenticate.
4. Select your public repository (`rokkian/mrokk-website`).

### Step C: Create the deployment trigger

1. Go to **Cloud Build** > **Triggers** > **Create Trigger**.
2. Set the following options:
   - **Name**: `deploy-mrokk-web-on-push`
   - **Event**: `Push to a branch`
   - **Source**: Select your connected repository (`rokkian/mrokk-website`)
   - **Branch**: `^main$`
   - **Configuration**: `Cloud Build configuration file (yaml or json)`
   - **Location**: `Repository` at `/cloudbuild.yaml`
3. Click **Create**.

> **Security Note for Public Repositories**: Because this is a public repository, only configure deployment triggers on pushes to `main`. Do not configure automated deployment triggers on untrusted Pull Requests.

### Step D: Test the pipeline

Push a commit to `main`, or click **Run** on the trigger in the Cloud Build Console to verify the build, test, and deploy workflow.

## Updating the site

- **Automated**: Push your changes to `main` on GitHub; Cloud Build will automatically run tests, build the image, and deploy.
- **Manual**: Run steps 3 and 4 (or `just deploy`).

