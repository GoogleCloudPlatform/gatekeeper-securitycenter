# gatekeeper-securitycenter

`gatekeeper-securitycenter` is

-   a Kubernetes controller that creates
    [Security Command Center](https://cloud.google.com/security-command-center)
    [findings](https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings)
    for violations reported by the
    [audit controller](https://cloud.google.com/anthos-config-management/docs/how-to/auditing-constraints)
    in
    [Policy Controller](https://cloud.google.com/anthos-config-management/docs/concepts/policy-controller)
    and
    [Open Policy Agent (OPA) Gatekeeper](https://github.com/open-policy-agent/gatekeeper);
    and

-   a command-line tool that creates and manages the IAM policies of
    Security Command Center
    [sources](https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources).

## Tutorial

See the accompanying [tutorial](docs/tutorial.md) for detailed explanation and
step-by-step instructions on how to create a Security Command Center source,
set up the Google service accounts with the required permissions, and
installing the controller resources in a Google Kubernetes Engine (GKE)
cluster.

## Install

To install the `gatekeeper-securitycenter` controller in your cluster, you
must provide the following inputs:

-   the full name of the Security Command Center source where the controller
    should report findings, in the format
    `organizations/[ORGANIZATION_ID]/sources/[SOURCE_ID]`; and

-   if you use
    [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
    (recommended), the Google service account to bind to the Kubernetes
    service account of the controller. The Google service account must
    have the
    [Security Center Findings Editor](https://cloud.google.com/iam/docs/understanding-roles#security-center-roles)
    role or equivalent permissions on the Security Command Center source, or at
    the organization level.

You can run the provided [setup script](scripts/setup.sh) to create these
resources and set up the necessary permissions. To run this script, you must
have an appropriate role for Security Command Center at the organization level,
such as
[Security Center Admin Editor](https://cloud.google.com/security-command-center/docs/access-control).

After you have created the resources, you can deploy the controller by running
the provided <scripts/deploy.sh> script, or you can follow the steps in the
[`kpt` package](manifests/README.md).

## Build

Build the command-line tool for your platform:

```bash
go build .
```

Build and publish the container image for the controller:

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

## Development

1.  Install [Skaffold](https://skaffold.dev/docs/install/).

2.  Install [`kpt`](https://googlecontainertools.github.io/kpt/installation/).

3.  Create a development GKE cluster with Workload Identity, and install
    Policy Controller or Gatekeeper. If you like, you can use the provided
    `dev-cluster.sh` shell script:

    ```bash
    ./scripts/dev-cluster.sh
    ```

4.  Create your Security Command Center source (`$SOURCE_NAME`) and set up your
    findings editor Google service account (`$FINDINGS_EDITOR_SA`) with the
    required permissions:

    ```bash
    ./scripts/setup.sh
    ```

    The script prints out values for `$SOURCE_NAME` and `$FINDINGS_EDITOR_SA`.
    Set these as environment variables for use in later steps.

5.  Create a copy of the `manifests` directory called `.kpt-skaffold`. This
    directory stores your manifests for development purposes:

    ```bash
    cp -r manifests .kpt-skaffold
    ```

6.  Set the name of your Security Command Center source:

    ```bash
    kpt cfg set .kpt-skaffold source $SOURCE_NAME
    ```

7.  Set the image name to the Go import path, with the prefix `ko://`:

    ```bash
    kpt cfg set .kpt-skaffold image ko://github.com/GoogleCloudPlatform/gatekeeper-securitycenter
    ```

8.  If you use a GKE cluster with Workload Identity, add the Workload Identity
    annotation to the Kubernetes service account used by the controller:

    ```bash
    kpt cfg annotate .kpt-skaffold \
        --kind ServiceAccount \
        --name gatekeeper-securitycenter-controller \
        --namespace gatekeeper-securitycenter \
        --kv iam.gke.io/gcp-service-account=$FINDINGS_EDITOR_SA
    ```

9.  Deploy the resources and start the Skaffold development mode watch loop:

    ```bash
    skaffold dev
    ```

    `kpt` creates a directory called `.kpt-hydrated` to store the hydrated
    manifests and the `inventory-template.yaml` file.

## Disclaimer

This is not an officially supported Google product.
