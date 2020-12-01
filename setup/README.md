# `gatekeeper-securitycenter-setup` `kpt` package

This package creates prerequisite resources for the `gatekeeper-securitycenter`
controller:

-   a Google Kubernetes Engine (GKE) cluster;
-   Open Policy Agent Gatekeeper installed in the GKE cluster;
-   Google service accounts with Cloud IAM policy bindings; and
-   a Security Command Center source for Gatekeeper audit findings

## Description

This package creates the Google service accounts and Cloud IAM policy bindings
using
[Config Connector](https://cloud.google.com/config-connector/docs/overview).

If you have already set up the prerequisite resources listed above and just
want to deploy the `gatekeeper-securitycenter` controller, use the
[`manifests`](../manifests) package instead.

## Usage

Tools required:

-   [Google Cloud SDK](https://cloud.google.com/sdk)
-   [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
-   [`kpt`](https://googlecontainertools.github.io/kpt/)
-   [Go distribution](https://golang.org/dl/)
-   [`jq`](https://stedolan.github.io/jq/)

### Fetch the package

```bash
DIR=gatekeeper-securitycenter-setup

kpt pkg get https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter.git/setup "$DIR"
```

Details: https://googlecontainertools.github.io/kpt/reference/pkg/get/

### View package content

```bash
kpt cfg tree "$DIR"
```

Details: https://googlecontainertools.github.io/kpt/reference/cfg/tree/

### List setters

```bash
kpt cfg list-setters "$DIR"
```

Details: https://googlecontainertools.github.io/kpt/reference/cfg/list-setters/

### Set environment variables

```bash
source "$DIR/setup.env"
```

If you want to use an exisiting GKE cluster and/or existing Google service
accounts, edit the values in [`setup.env`](setup.env) to match the names of
your existing resources before you source the file.

### Create the prerequisite resources

```bash
"$DIR/setup.sh"
```

This script
[initializes](https://googlecontainertools.github.io/kpt/reference/live/init/)
and
[applies](https://googlecontainertools.github.io/kpt/reference/live/apply/)
the resource manifests in these directories:

1.  [`config-connector`](config-connector)
2.  [`gatekeeper`](gatekeeper)
3.  [`iam`](iam)
4.  [`securitycenter`](securitycenter)

When the script is done, it prints the values you need to deploy the controller
resources using the `kpt` package in the [`manifests`](../manifests) directory.

##

The script is designed to be idempotent. This means that if you encounter
issues, you can run the script again
