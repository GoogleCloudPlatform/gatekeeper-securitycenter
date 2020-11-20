# gatekeeper-securitycenter

`gatekeeper-securitycenter` is

-   A Kubernetes controller that creates Security Command Center
    findings for violations reported by the audit controller in
    [Policy Controller](https://cloud.google.com/anthos-config-management/docs/concepts/policy-controller)
    and
    [Open Policy Agent (OPA) Gatekeeper](https://github.com/open-policy-agent/gatekeeper);
    and

-   A command-line tool that creates and manages the IAM policies of
    [Security Command Center](https://cloud.google.com/security-command-center)
    sources.

## Usage

See the accompanying [tutorial](docs/tutorial.md) for step-by-step
instructions on how to create the Security Command Center source, setting up
the Google service account with the required permissions, and installing the
controller resources in a Google Kubernetes Engine (GKE) cluster.

## Building

Build the command-line tool for your platform:

```bash
go build .
```

Build and publish the container image for the controller using `ko`:

```bash
(cd tools ; go get github.com/google/ko/cmd/ko)
export KO_DOCKER_REPO=gcr.io/$(gcloud config get-value core/project)
ko publish . --base-import-paths --tags latest
```

[`ko`](https://github.com/google/ko) is a command-line tool for building
container images from Go source code. It does not use a `Dockerfile` and it
does not require a local Docker daemon.

If you would like to use a different base image, edit the value of
`defaultBaseImage` in the file [`.ko.yaml`](.ko.yaml).

## Installing

To install the `gatekeeper-securitycenter` controller in your cluster, you
must provide the following inputs:

-   the full name of the Security Command Center source where the controller
    should report findings, in the format
    `organizations/[ORGANIZATION_ID]/sources/[SOURCE_ID]`; and

-   if you are using [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
    (recommended), the Google service account to bind to the Kubernetes
    service account of the controller. The Google service account must
    have the [Security Center Findings Editor](https://cloud.google.com/iam/docs/understanding-roles#security-center-roles)
    role or equivalent permissions on the Security Command Center source, or
    at the organization level.

See the accompanying [tutorial](docs/tutorial.md) for step-by-step
instructions.

## Disclaimer

This is not an officially supported Google product.
