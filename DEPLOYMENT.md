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

## Updating the site

Change the code, templates, CSS, images, or `content.yaml`, then repeat steps 3 and 4.
