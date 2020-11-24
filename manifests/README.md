# `gatekeeper-securitycenter` `kpt` package

## Description

`kpt` package for the `gatekeeper-securitycenter` controller.

## Usage

### Fetch the package

```bash
kpt pkg get https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/manifests manifests
```

Details: https://googlecontainertools.github.io/kpt/reference/pkg/get/

### View package contents

```bash
kpt cfg tree manifests
```

Details: https://googlecontainertools.github.io/kpt/reference/cfg/tree/

### List setters

```bash
kpt cfg list-setters manifests
```

Details: https://googlecontainertools.github.io/kpt/reference/cfg/list-setters/

### Set values

1.  Set the Security Command Center source name:

    ```bash
    kpt cfg set manifests source $SOURCE_NAME
    ```

    Where `$SOURCE_NAME` is your Security Command Center source in the format
    `organizations/$ORGANIZATION_ID/sources/$SOURCE_ID`.

2.  Set the optional cluster name. You can use any name you like. As an
    example, you can use your current `kubectl` context name:

    ```bash
    kpt cfg set manifests cluster $(kubectl config current-context)
    ```

3.  If you build your own container image, you can change the `image` value:

    ```bash
    kpt cfg set manifests image [YOUR_IMAGE_NAME]
    ```

Details: https://googlecontainertools.github.io/kpt/reference/cfg/set/

### Add annotation

If your Google Kubernetes Engine (GKE) cluster uses
[Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity),
add an annotation for the Google service account `$FINDINGS_EDITOR_SA` to bind
it to the `gatekeeper-securitycenter-controller` Kubernetes service account:

```bash
kpt cfg annotate manifests \
    --kind ServiceAccount --name gatekeeper-securitycenter-controller \
    --kv iam.gke.io/gcp-service-account=$FINDINGS_EDITOR_SA
```

The Google service account must have the
[Security Center Findings Editor](https://cloud.google.com/iam/docs/understanding-roles#security-center-roles)
Cloud IAM role on the source or at the organization level.

Details: https://googlecontainertools.github.io/kpt/reference/cfg/annotate/

### Initialize the package

```bash
kpt live init manifests --namespace gatekeeper-securitycenter
```

Details: https://googlecontainertools.github.io/kpt/reference/live/init/

### Apply the package

```bash
kpt live apply manifests --reconcile-timeout 2m --output table
```

Details: https://googlecontainertools.github.io/kpt/reference/live/apply/
