#!/usr/bin/env python3
"""Provision this repository's keyless GitHub Actions deployment identity.

Run with an authenticated gcloud administrator and gh repository administrator.
Safe to rerun: existing resources are reused and IAM bindings are additive.
"""

import json
import shutil
import subprocess
import sys
import time

PROJECT = "pinebula"
REGION = "europe-west1"
REPOSITORY = "rokkian/mrokk-website"
REGISTRY = "personal-site"
SERVICE = "mrokk-web"
POOL = "mrokk-website"
PROVIDER = "github"
DEPLOY_ACCOUNT = "github-mrokk-deploy"
ENVIRONMENT = "production"
WORKFLOW = ".github/workflows/ci.yml"


def command(*args, payload=None):
    result = subprocess.run(
        args,
        input=json.dumps(payload) if payload is not None else None,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    result.check_returncode()
    if result.stderr:
        print(result.stderr, end="", file=sys.stderr)
    return result.stdout.strip()


def cloud(*args):
    output = command(
        "gcloud", *args, f"--project={PROJECT}", "--quiet", "--format=json"
    )
    return json.loads(output) if output else None


def github(endpoint, method="GET", payload=None):
    args = ["gh", "api", f"repos/{REPOSITORY}/{endpoint}".rstrip("/"), "--method", method]
    if payload is not None:
        args.extend(["--input", "-"])
    output = command(*args, payload=payload)
    return json.loads(output) if output else None


def bind(*args):
    # Newly created service accounts may take a short time to reach IAM.
    for attempt in range(6):
        try:
            return cloud(*args)
        except subprocess.CalledProcessError as error:
            if "does not exist" not in error.stderr or attempt == 5:
                raise
            print("Waiting for the new service account to propagate...", flush=True)
            time.sleep(min(2 ** (attempt + 1), 16))


def main():
    for executable in ("gcloud", "gh"):
        if not shutil.which(executable):
            sys.exit(f"Install and authenticate {executable} before running this script.")

    print("Checking the existing repository, Cloud Run service and registry...", flush=True)
    repo = github("")
    if repo["default_branch"] != "main":
        sys.exit("This workflow expects the repository's default branch to be main.")
    service = cloud("run", "services", "describe", SERVICE, f"--region={REGION}")
    runtime_account = service["spec"]["template"]["spec"]["serviceAccountName"]
    cloud("artifacts", "repositories", "describe", REGISTRY, f"--location={REGION}")
    project_number = str(cloud("projects", "describe", PROJECT)["projectNumber"])
    deploy_account = f"{DEPLOY_ACCOUNT}@{PROJECT}.iam.gserviceaccount.com"
    pool_name = f"projects/{project_number}/locations/global/workloadIdentityPools/{POOL}"
    provider_name = f"{pool_name}/providers/{PROVIDER}"

    print("Enabling the required APIs...", flush=True)
    cloud(
        "services", "enable", "iam.googleapis.com", "iamcredentials.googleapis.com",
        "sts.googleapis.com", "cloudresourcemanager.googleapis.com",
        "run.googleapis.com", "artifactregistry.googleapis.com",
    )

    print("Configuring the dedicated service account and WIF provider...", flush=True)
    accounts = cloud("iam", "service-accounts", "list")
    if not any(account["email"] == deploy_account for account in accounts):
        cloud(
            "iam", "service-accounts", "create", DEPLOY_ACCOUNT,
            "--display-name=GitHub Actions deploy for mrokk-website",
        )
    pools = cloud("iam", "workload-identity-pools", "list", "--location=global")
    if not any(pool["name"] == pool_name for pool in pools):
        cloud(
            "iam", "workload-identity-pools", "create", POOL, "--location=global",
            "--display-name=mrokk-website GitHub Actions",
        )

    # Numeric IDs cannot be reclaimed after a repository or owner is deleted.
    # The ref, workflow and environment also exclude PR and other-branch tokens.
    # Check the environment claim directly: GitHub's immutable subject format
    # includes owner/repository IDs, so matching a legacy sub rejects valid jobs.
    condition = " && ".join([
        f"assertion.repository_id == '{repo['id']}'",
        f"assertion.repository_owner_id == '{repo['owner']['id']}'",
        "assertion.ref == 'refs/heads/main'",
        f"assertion.workflow_ref == '{REPOSITORY}/{WORKFLOW}@refs/heads/main'",
        f"assertion.environment == '{ENVIRONMENT}'",
        "assertion.event_name in ['push', 'workflow_dispatch']",
    ])
    providers = cloud(
        "iam", "workload-identity-pools", "providers", "list",
        "--location=global", f"--workload-identity-pool={POOL}",
    )
    operation = "update-oidc" if any(
        provider["name"] == provider_name for provider in providers
    ) else "create-oidc"
    cloud(
        "iam", "workload-identity-pools", "providers", operation, PROVIDER,
        "--location=global", f"--workload-identity-pool={POOL}",
        "--issuer-uri=https://token.actions.githubusercontent.com",
        "--attribute-mapping=google.subject=assertion.sub,"
        "attribute.repository_id=assertion.repository_id,"
        "attribute.repository_owner_id=assertion.repository_owner_id",
        f"--attribute-condition={condition}",
    )

    print("Granting resource-scoped deployment permissions...", flush=True)
    member = f"--member=serviceAccount:{deploy_account}"
    bind(
        "artifacts", "repositories", "add-iam-policy-binding", REGISTRY,
        f"--location={REGION}", member, "--role=roles/artifactregistry.writer",
    )
    bind(
        "run", "services", "add-iam-policy-binding", SERVICE,
        f"--region={REGION}", member, "--role=roles/run.developer",
    )
    bind(
        "iam", "service-accounts", "add-iam-policy-binding", runtime_account,
        member, "--role=roles/iam.serviceAccountUser", "--condition=None",
    )
    bind(
        "iam", "service-accounts", "add-iam-policy-binding", deploy_account,
        f"--member=principalSet://iam.googleapis.com/{pool_name}"
        f"/attribute.repository_id/{repo['id']}",
        "--role=roles/iam.workloadIdentityUser", "--condition=None",
    )

    print("Configuring the production environment and GitHub variables...", flush=True)
    github(f"environments/{ENVIRONMENT}", "PUT", {
        "deployment_branch_policy": {
            "protected_branches": False,
            "custom_branch_policies": True,
        },
    })
    policies = github(f"environments/{ENVIRONMENT}/deployment-branch-policies")
    if not any(
        policy["name"] == "main" and policy["type"] == "branch"
        for policy in policies["branch_policies"]
    ):
        github(f"environments/{ENVIRONMENT}/deployment-branch-policies", "POST", {
            "name": "main", "type": "branch",
        })
    for name, value in {
        "GCP_WIF_PROVIDER": provider_name,
        "GCP_SERVICE_ACCOUNT": deploy_account,
        "GCP_RUNTIME_SERVICE_ACCOUNT": runtime_account,
    }.items():
        command("gh", "variable", "set", name, "--repo", REPOSITORY, "--body", value)

    print(f"Configured {provider_name}")
    print(f"Deployment account: {deploy_account}")
    print(f"Runtime account: {runtime_account}")
    print("IAM propagation can take a few minutes. Push the workflow to main to deploy.")


if __name__ == "__main__":
    try:
        main()
    except subprocess.CalledProcessError as error:
        if error.stderr:
            print(error.stderr, end="", file=sys.stderr)
        sys.exit(f"Setup stopped (exit {error.returncode}). Fix the error above and rerun.")
