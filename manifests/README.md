# gatekeeper-securitycenter

## Description

`kpt` package for the `gatekeeper-securitycenter` controller.

## Usage

### Fetch the package

```bash
kpt pkg get https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/manifests gatekeeper-securitycenter
```

Details: https://googlecontainertools.github.io/kpt/reference/pkg/get/

### View package contents

```bash
kpt cfg tree gatekeeper-securitycenter
```

Details: https://googlecontainertools.github.io/kpt/reference/cfg/tree/

### List setters

```bash
kpt cfg list-setters gatekeeper-securitycenter
```

Details: https://googlecontainertools.github.io/kpt/reference/cfg/list-setters/

### Set values

1.  Set the Security Command Center source name:

    ```bash
    kpt cfg set gatekeeper-securitycenter \
        source organizations/$ORGANIZATION_ID/sources/$SOURCE_ID
    ```

    Where `$ORGANIZATION_ID` is your Google Cloud organization ID, and
    `$SOURCE_ID` is your Security Command Center source ID.

2.  Set the optional cluster name. You can use any name you like. As an
    example, you can use your current `kubectl` context name:

    ```bash
    kpt cfg set gatekeeper-securitycenter cluster $(kubectl config current-context)
    ```

Details: https://googlecontainertools.github.io/kpt/reference/cfg/set/

### Add annotation

If your Google Kubernetes Engine (GKE) cluster uses
[Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity),
add an annotation for the Google service account `$GSA_EMAIL` to bind it to the
`gatekeeper-securitycenter-controller` Kubernetes service account:

```bash
kpt cfg annotate gatekeeper-securitycenter \
    --kind ServiceAccount --name gatekeeper-securitycenter-controller \
    --kv iam.gke.io/gcp-service-account=$GSA_EMAIL
```

Details: https://googlecontainertools.github.io/kpt/reference/cfg/annotate/

### Apply the package

```
kpt live init gatekeeper-securitycenter --namespace gatekeeper-securitycenter

kpt live apply gatekeeper-securitycenter --reconcile-timeout 2m --output table
```

Details: https://googlecontainertools.github.io/kpt/reference/live/
