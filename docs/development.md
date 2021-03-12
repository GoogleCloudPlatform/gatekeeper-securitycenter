# Developing `gatekeeper-securitycenter`

1.  Install [kpt](https://googlecontainertools.github.io/kpt),
    [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/),
    [kustomize](https://kustomize.io/), and
    [Skaffold](https://skaffold.dev/):

    ```bash
    gcloud components install kpt kustomize skaffold --quiet
    ```

2.  Create a development GKE cluster with Workload Identity, and install
    Policy Controller or Gatekeeper. If you like, you can use the provided
    `dev-cluster.sh` shell script:

    ```bash
    ./scripts/dev-cluster.sh
    ```

3.  Create your Security Command Center source (`SOURCE_NAME`) and set up your
    findings editor Google service account (`FINDINGS_EDITOR_SA`) with the
    required permissions:

    ```bash
    ./scripts/iam-setup.sh
    ```

    The script prints out values for `SOURCE_NAME` and `FINDINGS_EDITOR_SA`.
    Set these as environment variables for use in later steps.

4.  Set the name of your Security Command Center source:

    ```bash
    kpt cfg set manifests/ source $SOURCE_NAME
    ```

5.  If you use a GKE cluster with Workload Identity, add the Workload Identity
    annotation to the Kubernetes service account used by the controller:

    ```bash
    kpt cfg annotate manifests/ \
        --kind ServiceAccount \
        --name gatekeeper-securitycenter-controller \
        --namespace gatekeeper-securitycenter \
        --kv iam.gke.io/gcp-service-account=$FINDINGS_EDITOR_SA
    ```

6.  Define the base image registry path for Skaffold:

    ```bash
    export SKAFFOLD_DEFAULT_REPO=gcr.io/$(gcloud config get-value core/project)
    ```

7.  Deploy the resources and start the Skaffold development mode watch loop:

    ```bash
    skaffold dev
    ```

    Skaffold creates a directory called `.kpt-hydrated` to store the hydrated
    manifests and the inventory template.
