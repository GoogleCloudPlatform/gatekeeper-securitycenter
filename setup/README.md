# kpt package for `gatekeeper-securitycenter` prerequisites

This package creates the following prerequisite resources for the
`gatekeeper-securitycenter` controller using
[Config Connector](https://cloud.google.com/config-connector/docs/overview):

-   a Google Kubernetes Engine (GKE) cluster;
-   Open Policy Agent Gatekeeper installed in the GKE cluster;
-   Google service accounts with Cloud IAM policy bindings; and
-   a Security Command Center source for Gatekeeper audit findings

If you have already set up the prerequisite resources and want to deploy the
`gatekeeper-securitycenter` controller, skip these steps and use the
[`manifests`](https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/tree/main/manifests)
package instead.

## Deploying the controller

Tools required:

-   [Google Cloud SDK](https://cloud.google.com/sdk)
-   [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
-   [kpt](https://kpt.dev/installation/) v1.0.0-beta.1 or later
-   [kustomize](https://kustomize.io/)
-   [jq](https://stedolan.github.io/jq/)

### Fetch the package

```sh
kpt pkg get https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter.git/setup setup
```

### Set environment variables

```sh
source setup/setup.env
```

If you want to use an exisiting GKE cluster and/or existing Google service
accounts, edit the values in [`setup.env`](setup.env) to match the names of
your existing resources before you source the file.

### Create the prerequisite resources

```sh
./setup/setup.sh
```

This script
[initializes](https://kpt.dev/reference/cli/live/init/) and
[applies](https://kpt.dev/reference/cli/live/apply/)
the resource manifests in these directories:

1.  [`config-connector`](config-connector)
2.  [`gatekeeper`](gatekeeper)
3.  [`iam`](iam)
4.  [`securitycenter`](securitycenter)

When the script is done, it prints the values you need to deploy the controller
resources using the kpt package in the
[`manifests`](https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/tree/main/manifests)
directory.

## Troubleshooting

The script is designed to be idempotent. This means that if you encounter
issues, you can run the script again.
