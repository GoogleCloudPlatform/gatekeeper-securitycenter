# Creating policy-compliant Google Cloud resources using Config Connector with Policy Controller and Gatekeeper

This tutorial shows how platform administrators can use
[Policy Controller](https://cloud.google.com/anthos-config-management/docs/concepts/policy-controller)
and
[Open Policy Agent Gatekeeper](https://github.com/open-policy-agent/gatekeeper)
to enforce policies governing the creation of Google Cloud resources with
[Config Connector](https://cloud.google.com/config-connector/docs/overview).

As an example, the tutorial defines a policy that restricts permitted locations
for [Cloud Storage](https://cloud.google.com/storage) buckets.

You can use either Policy Controller or Gatekeeper in this tutorial.

## Introduction

[Policy Controller](https://cloud.google.com/anthos-config-management/docs/concepts/policy-controller)
checks, audits, and enforces your cluster resources' compliance with policies
related to security, regulations, or business rules. Policy Controller is built
from the
[Gatekeeper open source project](https://github.com/open-policy-agent/gatekeeper).

[Config Connector](https://cloud.google.com/config-connector/docs/overview) is
a Kubernetes addon that allows you to create and manage the lifecycle of Google
Cloud resources by describing them using Kubernetes
[custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/).
To create a Google Cloud resource, you create a Kubernetes resource in a
namespace that is managed by Config Connector. Below is an example of how you
describe a [Cloud Storage](https://cloud.google.com/storage) bucket using
Config Connector:

```yaml
apiVersion: storage.cnrm.cloud.google.com/v1beta1
kind: StorageBucket
metadata:
  name: my-bucket
spec:
  location: us-east1
```

By managing your Google Cloud resources using Config Connector, you can apply
Policy Controller and Gatekeeper policies to these resources as you create them
in your Kubernetes cluster. This allows you to prevent or report on creation of
and modifications to resources that violate your policies. For instance, you
can enforce a policy that restricts permitted locations for Cloud Storage
buckets.

This approach based on the
[Kubernetes resource model (KRM)](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/resource-management.md#the-kubernetes-resource-model-krm)
allows you to use a consistent set of tools and workflows to manage both
Kubernetes and Google Cloud resources. This tutorial demonstrates how you can:

-   define policies governing your Google Cloud resources;

-   implement controls that prevent developers and administrators from creating
    Google Cloud resources that violate your policies;

-   implement controls that audit your existing Google Cloud resources against
    your policies; and

-   provide fast feedback to developers and administrators during development,
    by allowing them to validate Google Cloud resource definitions against your
    policies before attempting to apply the definitions to a Kubernetes
    cluster.

## Objectives

-   Create a Google Kubernetes Engine (GKE) cluster
-   Set up Config Connector
-   Install Policy Controller or Gatekeeper
-   Create a Cloud Storage bucket using Config Connector
-   Create a Policy Controller or Gatekeeper policy to restrict permitted
    Cloud Storage bucket locations
-   Verify the policy
-   View policy violations for resources created prior to the policy
-   Validate resources during development
-   Validate resources created outside Config Connector

## Costs

This tutorial uses the following billable components of Google Cloud:

-   [Cloud Asset Inventory](https://cloud.google.com/asset-inventory/pricing)
-   [Cloud Storage](https://cloud.google.com/storage/pricing)
-   [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine/pricing)

To generate a cost estimate based on your projected usage, use the
[pricing calculator](https://cloud.google.com/products/calculator).
New Google Cloud users might be eligible for a free trial.

When you finish, you can avoid continued billing by deleting the resources you
created.
<walkthrough-alt>For more information, see [Cleaning up](#cleaning-up).</walkthrough-alt>

## Before you begin

<!-- {% setvar project_id "YOUR_PROJECT_ID" %} -->

1.  <walkthrough-project-billing-setup></walkthrough-project-billing-setup>

    <walkthrough-alt>

    Open a Linux or macOS terminal and install the following command-line tools:

    - [Google Cloud SDK](https://cloud.google.com/sdk)
    - [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
    - [Go distribution](https://golang.org/dl/) version 1.15 or later
    - [`jq`](https://stedolan.github.io/jq/)

    If you like, you can use [Cloud Shell](https://cloud.google.com/shell):

    [![Open in Cloud Shell](https://gstatic.com/cloudssh/images/open-btn.svg)](https://ssh.cloud.google.com/cloudshell/editor?cloudshell_git_repo=https%3A%2F%2Fgithub.com%2FGoogleCloudPlatform%2Fgatekeeper-securitycenter&cloudshell_tutorial=docs%2Fconfig-connector-gatekeeper-tutorial.md)

    Cloud Shell is a Linux shell environment with the Cloud SDK and the other
    required tools already installed.

    </walkthrough-alt>

2.  Set the Google Cloud project you want to use for this tutorial:

    ```bash
    gcloud config set core/project {{project-id}}
    ```

    where `{{project-id}}` is your
    [Google Cloud project ID](https://cloud.google.com/resource-manager/docs/creating-managing-projects).

    **Note:** Make sure billing is enabled for your Google Cloud project.
    [Learn how to confirm billing is enabled for your project.](https://cloud.google.com/billing/docs/how-to/modify-project)

3.  Define an exported environment variable containing your
    [Google Cloud project ID](https://cloud.google.com/resource-manager/docs/creating-managing-projects):

    ```bash
    export GOOGLE_CLOUD_PROJECT=$(gcloud config get-value core/project)
    ```

4.  Enable the Google Kubernetes Engine API:

    ```bash
    gcloud services enable container.googleapis.com
    ```

5.  Create and navigate to a directory that you use to store files created for
    this tutorial:

    ```bash
    mkdir -p ~/cnrm-gatekeeper-tutorial

    cd ~/cnrm-gatekeeper-tutorial
    ```

## Creating a GKE cluster

1.  Create a GKE cluster with
    [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
    enabled:

    ```bash
    gcloud container clusters create cnrm-gatekeeper-tutorial \
        --enable-ip-alias \
        --enable-stackdriver-kubernetes \
        --release-channel regular \
        --workload-pool $GOOGLE_CLOUD_PROJECT.svc.id.goog
        --zone asia-southeast1-c
    ```

    This creates a cluster in the `asia-southeast1-c` zone. You can use a
    [different zone or region](https://cloud.google.com/compute/docs/regions-zones)
    if you like.

## Setting up Config Connector

**Note:** The Google Cloud Project where you install Config Connector is known
as the _host project_. The projects where you manage resources are known as the
_managed projects_. In this tutorial, you use Config Connector to create Google
Cloud resources in the same project as your GKE cluster, so the host project
and the managed project are the same project.

1.  Download the latest Config Connector release bundle and extract the
    operator manifest:

    ```bash
    gsutil cp gs://configconnector-operator/latest/release-bundle.tar.gz - \
        | tar xz ./operator-system/configconnector-operator.yaml
    ```

2.  Install the Config Connector operator on your cluster:

    ```bash
    kubectl apply -f operator-system/configconnector-operator.yaml
    ```

    The Config Connector operator installs custom resource definitions (CRDs)
    for Google Cloud resources in your GKE cluster.

3.  Create a Kubernetes namespace for the Config Connector resources you will
    create in this tutorial:

    ```bash
    kubectl create namespace tutorial
    ```

4.  Create a Google service account for Config Connector:

    ```bash
    gcloud iam service-accounts create cnrm-gatekeeper-tutorial \
        --display-name "Config Connector Gatekeeper tutorial"
    ```

5.  Grant the
    [Storage Admin role](https://cloud.google.com/storage/docs/access-control/iam-roles)
    to the Google service account:

    ```bash
    gcloud projects add-iam-policy-binding $GOOGLE_CLOUD_PROJECT \
        --member "serviceAccount:cnrm-gatekeeper-tutorial@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com" \
        --role roles/storage.admin
    ```

    In this tutorial you use the Storage Admin role because you use Config
    Connector to create Cloud Storage buckets. In your own environment, you
    should grant the roles required to manage Google Cloud resources you want
    to create with Config Connector. See the Cloud Identity and Access
    Management (IAM) documentation for an
    [overview of predefined roles](https://cloud.google.com/iam/docs/understanding-roles).

6.  Wait for the Config Connector operator to be ready:

    ```bash
    kubectl wait pod/configconnector-operator-0 \
        -n configconnector-operator-system --for=condition=Initialized
    ```

7.  Create a `ConfigConnectorContext` resource that enables Config Connector
    for the Kubernetes namespace and associates it with the Google service
    account:

    ```bash
    cat << EOF > config-connector-context-tutorial.yaml
    apiVersion: core.cnrm.cloud.google.com/v1beta1
    kind: ConfigConnectorContext
    metadata:
      name: configconnectorcontext.core.cnrm.cloud.google.com
      namespace: tutorial
    spec:
      googleServiceAccount: cnrm-gatekeeper-tutorial@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com
    EOF
    ```

    ```bash
    kubectl apply -f config-connector-context-tutorial.yaml
    ```

8.  Wait for the Config Connector controller pod for your namespace to be
    ready:

    ```bash
    kubectl wait -n cnrm-system --for=condition=Ready pod \
        -l cnrm.cloud.google.com/component=cnrm-controller-manager,cnrm.cloud.google.com/scoped-namespace=tutorial
    ```

9.  Bind your Config Connector Kubernetes service account to your Google
    service account by creating a Cloud IAM policy binding:

    ```bash
    gcloud iam service-accounts add-iam-policy-binding \
        cnrm-gatekeeper-tutorial@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com \
        --member "serviceAccount:$GOOGLE_CLOUD_PROJECT.svc.id.goog[cnrm-system/cnrm-controller-manager-tutorial]" \
        --role roles/iam.workloadIdentityUser
    ```

    This binding allows the `cnrm-controller-manager-tutorial` Kubernetes
    service account in the `cnrm-system` namespace to act as the Google service
    account you created.

10. Annotate the namespace to specify which project Config Connector should use
    to create Google Cloud resources (the managed project):

    ```bash
    kubectl annotate namespace tutorial \
        cnrm.cloud.google.com/project-id=$GOOGLE_CLOUD_PROJECT
    ```

11. Wait for the Config Connector pods to be ready:

    ```bash
    kubectl wait -n cnrm-system --for=condition=Ready pod --all
    ```

## Installing Policy Controller

If you have a
[managed Anthos cluster](https://cloud.google.com/anthos/docs/setup/overview#requirements),
follow the steps in this section to install Policy Controller. If not, skip to
the next section to install the open source Gatekeeper distribution instead.

1.  Download the Config Management operator
    [custom resource definition (CRD)](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
    manifest and apply it to your cluster:

    ```bash
    gsutil cp gs://config-management-release/released/latest/config-management-operator.yaml config-management-operator.yaml

    kubectl apply -f config-management-operator.yaml
    ```

2.  Create and apply a `ConfigManagement` manifest based on the operator CRD.
    This manifest instructs the Config Management operator to install the
    Policy Controller components:

    ```bash
    cat << EOF > config-management.yaml
    apiVersion: configmanagement.gke.io/v1
    kind: ConfigManagement
    metadata:
      name: config-management
    spec:
      policyController:
        enabled: true
    EOF
    ```

    ```bash
    kubectl apply -f config-management.yaml
    ```

3.  Wait for Policy Controller to be ready, this could take a few minutes:

    ```bash
    kubectl rollout status deploy gatekeeper-controller-manager \
        -n gatekeeper-system
    ```

    **Note:** It can take some time for the Config Management operator to
    create the Policy Controller namespace and deployments. While this is
    happening, the `kubectl rollout status` command may return an error. If
    this happens, wait a minute and try again.

## Installing Gatekeeper

If you don't have an Anthos entitlement, you can install the open source
Gatekeeper distribution instead of Policy Controller. If you installed Policy
Controller in the previous section, skip to the next section.

1.  Define the Gatekeeper version to install:

    ```bash
    GATEKEEPER_VERSION=v3.1.3
    ```

2.  Install Gatekeeper:

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/$GATEKEEPER_VERSION/deploy/gatekeeper.yaml
    ```

3.  Wait for the Gatekeeper controller manager to be ready:

    ```bash
    kubectl rollout status deploy gatekeeper-controller-manager \
        -n gatekeeper-system
    ```

## Creating a Google Cloud resource using Config Connector

1.  Create a Cloud Storage bucket in the `us-central1` region using Config
    Connector:

    ```bash
    cat << EOF > tutorial-storagebucket-us-central1.yaml
    apiVersion: storage.cnrm.cloud.google.com/v1beta1
    kind: StorageBucket
    metadata:
      name: tutorial-us-central1-$GOOGLE_CLOUD_PROJECT
      namespace: tutorial
    spec:
      location: us-central1
    EOF
    ```

    ```bash
    kubectl apply -f tutorial-storagebucket-us-central1.yaml
    ```

2.  Wait for Config Connector to create the Cloud Storage bucket:

    ```bash
    gsutil ls | grep tutorial
    ```

    You see this output when the Cloud Storage bucket has been created:

    ```terminal
    gs://tutorial-us-central1-{{project-id}}/
    ```

    where `{{project-id}}` is your Google Cloud project ID. If you
    don't see this output, wait a minute and try again.

## Creating a policy

A policy in Policy Controller and Gatekeeper consists of a
[constraint template](https://github.com/open-policy-agent/frameworks/tree/master/constraint#what-is-a-constraint-template)
and a
[constraint](https://github.com/open-policy-agent/frameworks/tree/master/constraint#what-is-a-constraint).
The constraint template contains the policy logic, while the constraint
specifies where the policy applies and input parameters to the policy logic.

1.  Create a constraint template that restricts Cloud Storage bucket locations.
    The constraint template also allows you to specify a list of names for
    exempted buckets:

    ```bash
    cat << EOF > tutorial-storagebucket-location-template.yaml
    apiVersion: templates.gatekeeper.sh/v1beta1
    kind: ConstraintTemplate
    metadata:
      name: gcpstoragelocationconstraintv1
    spec:
      crd:
        spec:
          names:
            kind: GCPStorageLocationConstraintV1
          validation:
            openAPIV3Schema:
              properties:
                locations:
                  type: array
                  items:
                    type: string
                exemptions:
                  type: array
                  items:
                    type: string
      targets:
      - target: admission.k8s.gatekeeper.sh
        rego: |
          package gcpstoragelocationconstraintv1
          allowedLocation(reviewLocation) {
              locations := input.parameters.locations
              satisfied := [ good | location = locations[_]
                                    good = lower(location) == lower(reviewLocation)]
              any(satisfied)
          }
          exempt(reviewName) {
              input.parameters.exemptions[_] == reviewName
          }
          violation[{"msg": msg}] {
              bucketName := input.review.object.metadata.name
              bucketLocation := input.review.object.spec.location
              not allowedLocation(bucketLocation)
              not exempt(bucketName)
              msg := sprintf("Cloud Storage bucket <%v> uses a disallowed location <%v>, allowed cations are %v", [bucketName, bucketLocation, input.parameters.locations])
          }
          violation[{"msg": msg}] {
              not input.parameters.locations
              bucketName := input.review.object.metadata.name
              msg := sprintf("No permitted locations provided in constraint for Cloud Storage cket <%v>", [bucketName])
          }
    EOF
    ```

    ```bash
    kubectl apply -f tutorial-storagebucket-location-template.yaml
    ```

2.  Create a constraint that only allows buckets in the Singapore and Jakarta
    regions (`asia-southeast1` and `asia-southeast2`). The constraint applies
    to the `tutorial` namespace, and it exempts the default Cloud Storage
    bucket for Cloud Build:

    ```bash
    cat << EOF > tutorial-storagebucket-location-constraint.yaml
    apiVersion: constraints.gatekeeper.sh/v1beta1
    kind: GCPStorageLocationConstraintV1
    metadata:
      name: singapore-and-jakarta-only
    spec:
      enforcementAction: deny
      match:
        kinds:
        - apiGroups:
          - storage.cnrm.cloud.google.com
          kinds:
          - StorageBucket
        namespaces:
        - tutorial
      parameters:
        locations:
        - asia-southeast1
        - asia-southeast2
        exemptions:
        - ${GOOGLE_CLOUD_PROJECT}_cloudbuild
    EOF
    ```

    ```bash
    kubectl apply -f tutorial-storagebucket-location-constraint.yaml
    ```

    **Note:** You can delete the namespaces attribute if you want to apply the
    constraint to all namespaces in the GKE cluster. You can also
    [exclude specific namespaces](https://github.com/open-policy-agent/gatekeeper#exempting-namespaces-from-gatekeeper).

## Verifying the policy

1.  Try to create a Cloud Storage bucket in a location that isn't allowed (us-west1):

    ```bash
    cat << EOF > tutorial-storagebucket-us-west1.yaml
    apiVersion: storage.cnrm.cloud.google.com/v1beta1
    kind: StorageBucket
    metadata:
      name: tutorial-us-west1-$GOOGLE_CLOUD_PROJECT
      namespace: tutorial
    spec:
      location: us-west1
    EOF
    ```

    ```bash
    kubectl apply -f tutorial-storagebucket-us-west1.yaml
    ```

    This returns an error that says the bucket location is disallowed.

    <walkthrough-alt>

    The output looks like this:

    ```terminal
    Error from server ([denied by singapore-and-jakarta-only] Cloud Storage bucket <tutorial-us-west1-{{project-id}}> uses a disallowed location <us-west1>, allowed locations are ["asia-southeast1", "asia-southeast2"]): error when creating "tutorial-storagebucket-us-west1.yaml": admission webhook "validation.gatekeeper.sh" denied the request: [denied by singapore-and-jakarta-only] Cloud Storage bucket <tutorial-us-west1-{{project-id}}> uses a disallowed location <us-west1>, allowed locations are ["asia-southeast1", "asia-southeast2"]
    ```

    where `{{project-id}}` is your Google Cloud
    [project ID](https://cloud.google.com/resource-manager/docs/creating-managing-projects).

    </walkthrough-alt>

2.  Create a Cloud Storage bucket in a permitted location (`asia-southeast1`):

    ```bash
    cat << EOF > tutorial-storagebucket-asia-southeast1.yaml
    apiVersion: storage.cnrm.cloud.google.com/v1beta1
    kind: StorageBucket
    metadata:
      name: tutorial-asia-southeast1-$GOOGLE_CLOUD_PROJECT
      namespace: tutorial
    spec:
      location: asia-southeast1
    EOF
    ```

    ```bash
    kubectl apply -f tutorial-storagebucket-asia-southeast1.yaml
    ```

    The output is this:

    ```terminal
    storagebucket.storage.cnrm.cloud.google.com/tutorial-asia-southeast1-{{project-id}} created
    ```

    where `{{project-id}}` is your Google Cloud project ID.

3.  Wait for the bucket to be created:

    ```bash
    gsutil ls | grep tutorial
    ```

    You see this output when the Cloud Storage bucket has been created:

    ```terminal
    gs://tutorial-asia-southeast1-{{project-id}}/
    gs://tutorial-us-central1-{{project-id}}/
    ```

    where {{project-id}} is your Google Cloud project ID. If you don't see this
    output, wait a minute and try again.

## Auditing constraints

The [audit controller](https://github.com/open-policy-agent/gatekeeper#audit)
in Policy Controller and Gatekeeper periodically evaluates resources against
constraints. This allows you to detect policy violations for resources created
prior to the constraint being put in place, and for resources created outside
Config Connector.

1.  View violations for all constraints that use the
    `GCPStorageLocationConstraintV1` constraint template:

    ```bash
    kubectl get gcpstoragelocationconstraintv1 -o json \
        | jq '.items[].status.violations'
    ```

    The output looks like this. You see the bucket you created in `us-central1`
    before creating the constraint:

    ```json
    [
      {
        "enforcementAction": "deny",
        "kind": "StorageBucket",
        "message": "Cloud Storage bucket <tutorial-us-central1-{{project-id}}> uses a disallowed location <us-central1>, allowed locations are [\"asia-southeast1\", \"asia-southeast2\"]",
        "name": "tutorial-us-central1-{{project-id}}",
        "namespace": "tutorial"
      }
    ]
    ```

    where `{{project-id}}` is your Google Cloud project ID.

    **Note:** Policy Controller and Gatekeeper have a
    [default limit on the number of reported violations per constraint](https://github.com/open-policy-agent/gatekeeper#audit).

## Validating resources during development

During development and continuous integration builds, it's helpful to validate
resources against constraints before applying them to a GKE cluster. This
provides fast feedback and allows you to discover issues with resources and
constraints early.

1.  Install [`kpt`](https://googlecontainertools.github.io/kpt/):

    ```bash
    sudo apt-get install -y google-cloud-sdk-kpt
    ```

    `kpt` is a command-line tool that allows you to manage, manipulate,
    customize, and apply Kubernetes resources. You use `kpt` in this tutorial
    to run Gatekeeper as a function without connecting to your GKE cluster.

    **Note:** See the `kpt` website for alternative
    [installation instructions](https://googlecontainertools.github.io/kpt/installation/),
    such as using `brew` for macOS.

2.  Compose and execute a
    [`kpt` pipeline](https://googlecontainertools.github.io/kpt/concepts/functions/#pipeline):

    ```bash
    kpt fn source tutorial-*.yaml \
        | kpt fn run --image gcr.io/kpt-functions/gatekeeper-validate
    ```

    This pipeline uses a
    [`kpt` source function](https://googlecontainertools.github.io/kpt/reference/fn/source/)
    to create a Kubernetes
    [resource list](https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#ResourceList)
    containing the constraint template, the constraint, and the Config
    Connector Cloud Storage bucket resources. Next, the pipeline validates the
    Config Connector Cloud Storage bucket resources against the constraint
    using a
    [`kpt` config function](https://googlecontainertools.github.io/kpt/concepts/functions/)
    called `gatekeeper-validate`. This function is packaged as a container
    image and is available on
    [Container Registry](https://cloud.google.com/container-registry).

    The function reports that the manifest files for Cloud Storage buckets in
    the `us-central1` and `us-west1` regions violate the constraint.

    <walkthrough-alt>

    The output looks like this:

    ```terminal
    Error: Found 2 violations:

    [1] Cloud Storage bucket <tutorial-us-central1-{{project-id}}> uses a disallowed location <us-central1>, allowed locations are ["asia-southeast1", "asia-southeast2"]

    name: "tutorial-us-central1-{{project-id}}"
    path: tutorial-storagebucket-us-central1.yaml

    [2] Cloud Storage bucket <tutorial-us-west1-{{project-id}}> uses a disallowed location <us-west1>, allowed locations are ["asia-southeast1", "asia-southeast2"]

    name: "tutorial-us-west1-{{project-id}}"
    path: tutorial-storagebucket-us-west1.yaml


    error: exit status 1
    ```

    where `{{project-id}}` is your Google Cloud project ID.

    </walkthrough-alt>

## Validating resources created outside Config Connector

We recommend that you use a deployment pipeline to create Google Cloud
resources using Config Connector. However, you may have Google Cloud resources
that were created outside Config Connector, for instance as part of an
emergency change by an administrator in your organization.

You can evaluate your policies against these resources by exporting them, and
either:

-   validating the resources in a `kpt` pipeline as per the previous section;
    or

-   importing the resources into Config Connector.

To export the resources, you can use Cloud Asset Inventory.

1.  Enable the Cloud Asset API:

    ```bash
    gcloud services enable cloudasset.googleapis.com
    ```

2.  Export all Cloud Storage resources in your current project, and store the
    output in the bucket `tutorial-asia-southeast1-{{project-id}}`,
    where `{{project-id}}` is your Google Cloud project ID:

    ```bash
    gcloud asset export \
        --asset-types "storage.googleapis.com/Bucket" \
        --content-type resource \
        --project $GOOGLE_CLOUD_PROJECT \
        --output-path gs://tutorial-asia-southeast1-$GOOGLE_CLOUD_PROJECT/export.ndjson
    ```

    This command starts a background process to export the resources.
    The output looks similar to this:

    ```terminal
    Export in progress for root asset [projects/{{project-id}}].
    Use [gcloud asset operations describe projects/[PROJECT_NUMBER]/operations/ExportAssets/RESOURCE/[UNIQUE_ID]] to check the status of the operation.
    ```

    where `{{project-id}}` is your Google Cloud project ID,
    `[PROJECT_NUMBER]` is your Google Cloud project number, and `[UNIQUE_ID]`
    is the export operation ID.

3.  Use the `gcloud asset operations describe` command from the output of the
    previous command to check if the export is finished. Add the `--format`
    flag to only show the done status:

    ```bash
    gcloud asset operations describe --format 'value(done)' \
        projects/[PROJECT_NUMBER]/operations/ExportAssets/RESOURCE/[UNIQUE_ID]
    ```

    When the export is finished, the output of the command is this:

    ```terminal
    True
    ```

    You can use this command if you want to block and check if the operation is
    done every three seconds:

    ```bash
    until gcloud asset operations describe --format 'value(done)' projects/[PROJECT_NUMBER]/operations/ExportAssets/RESOURCE/[UNIQUE_ID] | grep True ; do sleep 3 ; done
    ```

4.  Copy the file containing the exported resources to your current directory:

    ```bash
    gsutil cp "gs://tutorial-asia-southeast1-$GOOGLE_CLOUD_PROJECT/export.ndjson" .
    ```

    The file is a [newline-delimited JSON file (NDJSON)](http://ndjson.org/),
    containing one resource per line.

5.  Download the
    [`config-connector` command-line tool](https://cloud.google.com/config-connector/docs/how-to/importing-existing-resources)
    to your current directory:

    ```bash
    gsutil cp gs://cnrm/latest/cli.tar.gz - \
        | tar xz --strip-components 3 ./linux/amd64/config-connector
    ```

    If you use macOS, replace `linux` with `darwin` in the command above to get
    a compatible binary.

6.  Use the `config-connector` tool to convert the NDJSON file to a YAML file
    containing Config Connector resource manifests:

    ```bash
    ./config-connector -iam-format=none < export.ndjson > export.yaml
    ```

    The `-iam-format=none` flag skips Cloud IAM policies in the output file. In
    your own environment, you can remove this flag if you want to validate
    Policy Controller or Gatekeeper constraints for Cloud IAM policies.

7.  Use a `kpt` pipeline to validate the resources against the Storage Bucket
    location policy. This pipeline uses a
    [`kpt` config function](https://googlecontainertools.github.io/kpt/concepts/functions/)
    called
    [`set-namespace`](https://googlecontainertools.github.io/kpt/guides/consumer/function/catalog/#transformers)
    to set the namespace metadata attribute value of all the resources. This is
    necessary since the constraint only applies to resources in the `tutorial`
    namespace:

    ```bash
    kpt fn source tutorial-storagebucket-location-*.yaml export.yaml \
        | kpt fn run --image gcr.io/kpt-functions/set-namespace -- namespace=tutorial \
        | kpt fn run --image gcr.io/kpt-functions/gatekeeper-validate
    ```

    The output shows violations for the existing resources you exported.

## Troubleshooting

1.  If Config Connector doesn't create the expected Google Cloud resources, you
    can view the logs of the Config Connector controller manager using this
    command:

    ```bash
    kubectl logs --namespace cnrm-system --container manager \
        --selector cnrm.cloud.google.com/component=cnrm-controller-manager,cnrm.cloud.google.com/scoped-namespace=tutorial
    ```

2.  If Policy Controller or Gatekeeper don't enforce policies correctly, you
    can view logs of the controller manager using this command:

    ```bash
    kubectl logs deployment/gatekeeper-controller-manager \
        --namespace gatekeeper-system
    ```

3.  If Policy Controller or Gatekeeper don't report violations in the status
    field of the constraint objects, you can view logs of the audit controller
    using this command:

    ```bash
    kubectl logs deployment/gatekeeper-audit -n gatekeeper-system
    ```

4.  If you run into other problems with this tutorial, we recommend that you
    review these documents:

    -   [Config Connector troubleshooting](https://cloud.google.com/config-connector/docs/how-to/install-upgrade-uninstall#troubleshooting)
    -   [Config Connector resource reference](https://cloud.google.com/config-connector/docs/reference/overview)
    -   [Gatekeeper debugging](https://github.com/open-policy-agent/gatekeeper#debugging)
    -   [GKE troubleshooting](https://cloud.google.com/kubernetes-engine/docs/troubleshooting)
    -   [Troubleshooting Kubernetes clusters](https://kubernetes.io/docs/tasks/debug-application-cluster/debug-cluster/)

## Cleaning up

To avoid incurring charges to your Google Cloud Platform account for the
resources used in this tutorial, you can delete the project or delete the
individual resources.

### Delete the project

**Caution:** Deleting a project has the following effects:

-   **Everything in the project is deleted.** If you used an existing project
    for this tutorial, when you delete it, you also delete any other work
    you've done in the project.

-   **Custom project IDs are lost.** When you created this project, you might
    have created a custom project ID that you want to use in the future. To
    preserve the URLs that use the project ID, such as an `appspot.com` URL,
    delete selected resources inside the project instead of deleting the whole
    project.

In Cloud Shell, run these commands to delete the current project:

```bash
echo $GOOGLE_CLOUD_PROJECT

gcloud projects delete $GOOGLE_CLOUD_PROJECT
```

### Delete the resources

If you want to keep the Google Cloud project you used in this tutorial, delete the individual resources.

1.  Delete the Cloud Storage bucket location constraint:

    ```bash
    kubectl delete -f tutorial-storagebucket-location-constraint.yaml
    ```

2.  Add the
    [`cnrm.cloud.google.com/force-destroy` annotation](https://cloud.google.com/config-connector/docs/reference/resource-docs/storage/storagebucket#annotations)
    with a string value of `"true"` to all `storagebucket` resources in the
    `tutorial` namespace:

    ```bash
    kubectl annotate storagebucket --all --namespace tutorial \
        cnrm.cloud.google.com/force-destroy=true
    ```

    This annotation is a
    [directive](https://cloud.google.com/config-connector/docs/concepts/resources#object_metadata)
    that allows Config Connector to delete a Cloud Storage bucket when you
    delete the corresponding storagebucket resource in the GKE cluster,
    [even if the bucket contains objects](https://cloud.google.com/config-connector/docs/reference/resource-docs/storage/storagebucket).

3.  Delete the Config Connector resources representing the Cloud Storage
    buckets:

    ```bash
    kubectl delete storagebucket --all --namespace tutorial
    ```

4.  Delete the GKE cluster:

    ```bash
    gcloud container clusters delete cnrm-gatekeeper-tutorial \
        --zone asia-southeast1-c --async --quiet
    ```

5.  Delete the Workload Identity policy binding in Cloud IAM:

    ```bash
    gcloud iam service-accounts remove-iam-policy-binding \
        cnrm-gatekeeper-tutorial@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com \
        --member "serviceAccount:$GOOGLE_CLOUD_PROJECT.svc.id.goog[cnrm-system/cnrm-controller-manager-tutorial]" \
        --role roles/iam.workloadIdentityUser
    ```

6.  Delete the Cloud Storage Admin role binding for the Google service account:

    ```bash
    gcloud projects remove-iam-policy-binding $GOOGLE_CLOUD_PROJECT \
        --member "serviceAccount:cnrm-gatekeeper-tutorial@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com" \
        --role roles/storage.admin
    ```

7.  Delete the Google service account you created for Config Connector:

    ```bash
    gcloud iam service-accounts delete --quiet \
        cnrm-gatekeeper-tutorial@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com
    ```

## What's next

-   Learn how to
    [create findings in Security Command Center for Policy Controller and Gatekeeper constraint violations](https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/blob/main/docs/tutorial.md).

-   Discover how to
    [run Policy Controller or Gatekeeper validation as part of a continuous integration pipeline in Cloud Build](https://cloud.google.com/anthos-config-management/docs/how-to/app-policy-validation-ci-pipeline).

-   Learn more about
    [enforcing policies with Anthos](https://cloud.google.com/architecture/blueprints/anthos-enforcing-policies-blueprint).

-   Learn more about
    [auditing and monitoring for deviation from policy with Anthos](https://cloud.google.com/architecture/blueprints/anthos-auditing-and-monitoring-for-deviation-from-policy-blueprint).

-   Try out other Google Cloud features for yourself. Have a look at our
    [tutorials](https://cloud.google.com/docs/tutorials).
