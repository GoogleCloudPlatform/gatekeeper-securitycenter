# `gatekeeper-securitycenter` controller manifests

Manifests for the `gatekeeper-securitycenter` Kubernetes controller.

## Usage

These instructions assume that you have already created the
[prerequisite resources](https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter#prerequisites).

### Tools required

-   [kpt](https://kpt.dev/installation/) v1.0.0-beta.7 or later

### Fetch the manifests

```sh
VERSION=v0.3.0
kpt pkg get https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter.git/manifests@$VERSION manifests
```

### Set source name and cluster name

1.  Set the Security Command Center source name:

    ```sh
    kpt fn eval manifests \
        --image gcr.io/kpt-fn/apply-setters:v0.2 -- \
        "source=$SOURCE_NAME"
    ```

    Where `$SOURCE_NAME` is your Security Command Center source in the format
    `organizations/$ORGANIZATION_ID/sources/$SOURCE_ID`.

2.  (Optional) Set the cluster name. You can use any name you like, or you can
    leave it blank. If you provide a cluster name, it will be visible in
    Security Command Center. As an example, you can use your current kubectl
    context name:

    ```sh
    kpt fn eval manifests \
        --image gcr.io/kpt-fn/apply-setters:v0.2 -- \
        "cluster=$(kubectl config current-context)"
    ```

### Add Workload Identity annotation

If your Google Kubernetes Engine (GKE) cluster uses
[Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity),
add an annotation for the Google service account `FINDINGS_EDITOR_SA` to bind
it to the `gatekeeper-securitycenter-controller` Kubernetes service account:

```sh
kpt fn eval \
    --image gcr.io/kpt-fn/set-annotations:v0.1.4 \
    --match-kind ServiceAccount \
    --match-name gatekeeper-securitycenter-controller \
    --match-namespace gatekeeper-securitycenter -- \
    "iam.gke.io/gcp-service-account=$FINDINGS_EDITOR_SA"
```

The Google service account must have the
[Security Center Findings Editor](https://cloud.google.com/iam/docs/understanding-roles#security-center-roles)
Cloud IAM role on the source or at the organization level.

If you don't use GKE Workload Identity, see the documentation on
[Authenticating to Google Cloud with service accounts](https://cloud.google.com/kubernetes-engine/docs/tutorials/authenticating-to-cloud-platform)
for alternative instructions on how to provide Google service account
credentials to the `gatekeeper-securitycenter` controller pods.

### Setup inventory tracking

```sh
kpt live init manifests
```

### Apply the manifests

```sh
kpt live apply manifests --reconcile-timeout=3m
```
