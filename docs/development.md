# Developing `gatekeeper-securitycenter`

1.  Install these command-line tools:

    - [`gcloud`](https://cloud.google.com/sdk/docs/install)
    - [kpt](https://kpt.dev/installation/) v1.0.0-beta.7 or later
    - [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
    - [Skaffold](https://skaffold.dev/) v1.37.2 or later

2.  Create a development GKE cluster with Workload Identity, and install
    OPA Gatekeeper. If you like, you can use the provided `dev-cluster.sh`
    shell script:

    ```sh
    ./scripts/dev-cluster.sh
    ```

3.  Create your Security Command Center source (`SOURCE_NAME`) and set up your
    findings editor Google service account (`FINDINGS_EDITOR_SA`) with the
    required permissions:

    ```sh
    ./scripts/iam-setup.sh
    ```

    The script prints out values for `SOURCE_NAME` and `FINDINGS_EDITOR_SA`.
    Set these as environment variables for use in later steps.

4.  Set the name of your Security Command Center source:

    ```sh
    kpt fn eval manifests \
        --image gcr.io/kpt-fn/apply-setters:v0.2 -- \
        "source=$SOURCE_NAME"
    ```

5.  If you use a GKE cluster with Workload Identity, add the Workload Identity
    annotation to the Kubernetes service account used by the controller:

    ```sh
    kpt fn eval manifests \
        --image gcr.io/kpt-fn/set-annotations:v0.1.4 \
        --match-kind ServiceAccount \
        --match-name gatekeeper-securitycenter-controller \
        --match-namespace gatekeeper-securitycenter -- \
        "iam.gke.io/gcp-service-account=$FINDINGS_EDITOR_SA"
    ```

6.  Define the base image registry path for Skaffold:

    ```sh
    export SKAFFOLD_DEFAULT_REPO=gcr.io/$(gcloud config get project)
    ```

7.  Deploy the resources and start the Skaffold development mode watch loop:

    ```sh
    skaffold dev
    ```
