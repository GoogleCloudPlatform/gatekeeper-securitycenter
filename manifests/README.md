# `gatekeeper-securitycenter` controller kpt package

kpt package for the `gatekeeper-securitycenter` Kubernetes controller.

## Usage

This package assumes that you have already created the
[prerequisite resources](https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter#prerequisites).

### Fetch this package

```bash
VERSION=$(curl -s https://api.github.com/repos/GoogleCloudPlatform/gatekeeper-securitycenter/releases/latest | jq -r '.tag_name')

kpt pkg get https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter.git/manifests@$VERSION manifests
```

### Set source name and cluster name

1.  Set the Security Command Center source name:

    ```bash
    kpt cfg set manifests/ source $SOURCE_NAME
    ```

    Where `$SOURCE_NAME` is your Security Command Center source in the format
    `organizations/$ORGANIZATION_ID/sources/$SOURCE_ID`.

2.  (Optional) Set the cluster name. You can use any name you like, or you can
    leave it blank. If you provide a cluster name, it will be visible in
    Security Command Center. As an example, you can use your current kubectl
    context name:

    ```bash
    kpt cfg set manifests/ cluster $(kubectl config current-context)
    ```

### Add Workload Identity annotation

If your Google Kubernetes Engine (GKE) cluster uses
[Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity),
add an annotation for the Google service account `FINDINGS_EDITOR_SA` to bind
it to the `gatekeeper-securitycenter-controller` Kubernetes service account:

```bash
kpt cfg annotate manifests/ \
    --kind ServiceAccount \
    --name gatekeeper-securitycenter-controller \
    --namespace gatekeeper-securitycenter \
    --kv iam.gke.io/gcp-service-account=$FINDINGS_EDITOR_SA
```

The Google service account must have the
[Security Center Findings Editor](https://cloud.google.com/iam/docs/understanding-roles#security-center-roles)
Cloud IAM role on the source or at the organization level.

If you don't use Workload Identity, see the documentation on
[Authenticating to Google Cloud with service accounts](https://cloud.google.com/kubernetes-engine/docs/tutorials/authenticating-to-cloud-platform)
for alternative instructions on how to provide Google service account
credentials to the `gatekeeper-securitycenter` controller pods.

### Apply the package

```bash
kpt live apply manifests/ --reconcile-timeout=3m --output=table
```
