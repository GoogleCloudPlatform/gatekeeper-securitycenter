# Reporting Policy Controller and Gatekeeper audit violations in Security Command Center

This tutorial shows platform security administrators how to report policy
violations from
[Policy Controller](https://cloud.google.com/anthos-config-management/docs/concepts/policy-controller)
and
[Open Policy Agent (OPA) Gatekeeper](https://github.com/open-policy-agent/gatekeeper)
as findings in
[Security Command Center](https://cloud.google.com/security-command-center).

This allows you to view and manage policy violations for your Kubernetes
resources alongside
[other vulnerability and security findings](https://cloud.google.com/security-command-center/docs/concepts-vulnerabilities-findings).

To complete this tutorial, you must have an appropriate editor role for
Security Command Center at the organization level, such as
[Security Center Admin Editor](https://cloud.google.com/security-command-center/docs/access-control).

You can use either Policy Controller or Gatekeeper in this tutorial.

## Introduction

[Policy Controller](https://cloud.google.com/anthos-config-management/docs/concepts/policy-controller)
checks, audits, and enforces your Kubernetes cluster resources' compliance with
policies related to security, regulations, or business rules. Policy Controller
is built from the
[Gatekeeper open source project](https://github.com/open-policy-agent/gatekeeper).

The [audit functionality](https://github.com/open-policy-agent/gatekeeper#audit)
in Policy Controller and Gatekeeper allow you to implement detective controls.
It periodically evaluates resources against policies and creates violations for
resources that don't conform to the policies. These violations are stored in
the cluster, and you can query them using Kubernetes tools such as kubectl.

To make these violations visible, and to help you take actions such as alerting
and remediation, you can use
[Security Command Center](https://cloud.google.com/security-command-center).
Security Command Center provides a dashboard and APIs for surfacing,
understanding, and remediating security and data risks across an organization,
for Google Cloud resources, Kubernetes resources, and hybrid or multi-cloud
resources.

Security Command Center displays possible security risks and policy violations,
called
[findings](https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings).
Findings come from
[sources](https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources),
which are mechanisms that can detect and report risks and violations. Security
Command Center includes
[built-in services](https://cloud.google.com/security-command-center/docs/concepts-security-command-center-overview),
and you can add third-party sources and your own sources.

This tutorial and associated source code shows you how to create a source and
findings in Security Command Center for Policy Controller and Gatekeeper policy
violations.

<walkthrough-alt>

![Architecture](architecture.svg)

</walkthrough-alt>

## Costs

This tutorial uses the following billable components of Google Cloud:

-   [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine/pricing)
-   [Security Command Center Standard tier](https://cloud.google.com/security-command-center/pricing#standard_tier_pricing)

To generate a cost estimate based on your projected usage, use the
[pricing calculator](https://cloud.google.com/products/calculator).
New Google Cloud users might be eligible for a free trial.

When you finish, you can avoid continued billing by deleting the resources you
created. For more information, see [Cleaning up](#cleaning-up).

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

    [![Open in Cloud Shell](https://gstatic.com/cloudssh/images/open-btn.svg)](https://ssh.cloud.google.com/cloudshell/editor?cloudshell_git_repo=https%3A%2F%2Fgithub.com%2FGoogleCloudPlatform%2Fgatekeeper-securitycenter&cloudshell_tutorial=docs%2Ftutorial.md)

    Cloud Shell is a Linux shell environment with the Cloud SDK and other tools
    already installed.

    </walkthrough-alt>

2.  Set the Google Cloud project you want to use for this tutorial:

    ```bash
    gcloud config set core/project {{project-id}}
    ```

    where `{{project-id}}` is your
    [Google Cloud project ID](https://cloud.google.com/resource-manager/docs/creating-managing-projects).

    **Note:** Make sure that billing is enabled for your Google Cloud project.
    [Learn how to confirm billing is enabled for your project.](https://cloud.google.com/billing/docs/how-to/modify-project)

3.  Define an exported environment variable containing your
    [Google Cloud project ID](https://cloud.google.com/resource-manager/docs/creating-managing-projects):

    ```bash
    export GOOGLE_CLOUD_PROJECT=$(gcloud config get-value core/project)
    ```

4.  Define an environment variable containing your
    [Google Cloud organization ID](https://cloud.google.com/resource-manager/docs/creating-managing-organization#retrieving_your_organization_id):

    ```bash
    ORGANIZATION_ID=$(gcloud projects get-ancestors \
        $GOOGLE_CLOUD_PROJECT --format json \
        | jq -r '.[] | select (.type=="organization") | .id')
    ```

5.  Enable the Google Kubernetes Engine and Security Command Center APIs:

    ```bash
    gcloud services enable \
        container.googleapis.com \
        securitycenter.googleapis.com
    ```

## Creating a Security Command Center source

Security Command Center records
[findings](https://cloud.google.com/security-command-center/docs/how-to-api-list-findings)
against
[sources](https://cloud.google.com/security-command-center/docs/how-to-configure-security-command-center).
Follow these steps to create a source for findings from Policy Controller and
Gatekeeper.

1.  Create a
    [Google service account](https://cloud.google.com/iam/docs/service-accounts)
    and store the service account name in an environment variable:

    ```bash
    SOURCES_ADMIN_SA=$(gcloud iam service-accounts create \
        securitycenter-sources-admin \
        --display-name "Security Command Center sources admin" \
        --format 'value(email)')
    ```

    You use this Google service account to administer Security Command Center
    sources.

2.  Grant the
    [Security Center Sources Admin](https://cloud.google.com/iam/docs/understanding-roles#security-center-roles)
    Cloud IAM role to the Google service account at the organization level:

    ```
    gcloud organizations add-iam-policy-binding $ORGANIZATION_ID \
        --member "serviceAccount:$SOURCES_ADMIN_SA" \
        --role roles/securitycenter.sourcesAdmin
    ```

3.  Grant the
    [Service Usage Consumer](https://cloud.google.com/iam/docs/understanding-roles#service-usage-roles)
    role to the Google service account at the organization level:

    ```bash
    gcloud organizations add-iam-policy-binding $ORGANIZATION_ID \
        --member "serviceAccount:$SOURCES_ADMIN_SA" \
        --role roles/serviceusage.serviceUsageConsumer
    ```

4.  Grant your user identity the
    [Service Account Token Creator](https://cloud.google.com/iam/docs/understanding-roles#service-accounts-roles)
    role for the Google service account:

    ```bah
    gcloud iam service-accounts add-iam-policy-binding \
        $SOURCES_ADMIN_SA \
        --member "user:$(gcloud config get-value account)" \
        --role roles/iam.serviceAccountTokenCreator
    ```

    This allows your user identity to
    [impersonate](https://cloud.google.com/iam/docs/impersonating-service-accounts)
    the Google service account.

5.  Download the latest `gatekeeper-securitycenter` binary for your platform:

    ```bash
    VERSION=$(curl -s https://api.github.com/repos/GoogleCloudPlatform/gatekeeper-securitycenter/latest | jq -r '.tag_name')
    OS=$(go env GOOS)
    ARCH=$(go env GOARCH)

    curl -sLo gatekeeper-securitycenter "https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/releases/download/${VERSION}/gatekeeper-securitycenter_${OS}_${ARCH}"

    chmod +x gatekeeper-securitycenter
    ```

    **Note:** If you prefer to build your own binary, follow the instructions
    in the [`README.md`](../README.md).

6.  Create a Security Command Center source for your organization. Capture the
    full source name in an exported environment variable:

    ```bash
    export SOURCE_NAME=$(./gatekeeper-securitycenter sources create \
        --organization $ORGANIZATION_ID \
        --display-name "Gatekeeper" \
        --description "Reports violations from Gatekeeper audits" \
        --impersonate-service-account $SOURCES_ADMIN_SA | jq -r '.name')
    ```

## Creating Security Command Center findings using a Kubernetes controller

You can deploy `gatekeeper-securitycenter` as a controller in your Kubernetes
cluster. This controller periodically checks for constraint violations and
creates a finding in Security Command Center for each violation.

These steps assume that you use
[Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine)
with
[Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity).

1.  Create a GKE cluster with Workload Identity:

    ```bash
    gcloud container clusters create gatekeeper-securitycenter-tutorial \
        --enable-ip-alias \
        --enable-stackdriver-kubernetes \
        --release-channel regular \
        --workload-pool $GOOGLE_CLOUD_PROJECT.svc.id.goog \
        --zone us-central1-f
    ```

    This command creates a cluster in the `us-central1-f` zone. You can use a
    [different zone or region](https://cloud.google.com/compute/docs/regions-zones)
    if you like.

2.  Grant the cluster-admin cluster role to your Google identity:

    ```bash
    kubectl create clusterrolebinding cluster-admin-binding \
        --clusterrole cluster-admin \
        --user $(gcloud config get-value core/account)
    ```

3.  Install Gatekeeper:

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/master/deploy/gatekeeper.yaml
    ```

    **Note:** If you have an
    [Anthos entitlement](https://cloud.google.com/anthos/pricing),
    you can
    [install Policy Controller](https://cloud.google.com/anthos-config-management/docs/how-to/installing-policy-controller)
    instead of Gatekeeper.

4.  Create a Google service account and store the service account name in an
    environment variable:

    ```bash
    GATEKEEPER_FINDINGS_EDITOR_SA=$(gcloud iam service-accounts create \
        gatekeeper-securitycenter \
        --display-name "Security Command Center Gatekeeper findings editor" \
        --format 'value(email)')
    ```

    You use this Google service account to create findings for the Security
    Command Center source.

5.  Grant the
    [Security Center Findings Editor](https://cloud.google.com/iam/docs/understanding-roles#security-center-roles)
    role to the Google service account for the source:

    ```bash
    ./gatekeeper-securitycenter sources add-iam-policy-binding \
        --source $SOURCE_NAME \
        --member "serviceAccount:$GATEKEEPER_FINDINGS_EDITOR_SA" \
        --role roles/securitycenter.findingsEditor \
        --impersonate-service-account $SOURCES_ADMIN_SA
    ```

    **Note:** You can grant the role at the organization level instead of on
    the source. By granting at the organization level, you grant the Google
    service account permission to edit findings for all sources.

6.  Create a
    [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
    Cloud IAM policy binding to allow the Kubernetes service account
    `gatekeeper-securitycenter-controller` in the namespace
    `gatekeeper-securitycenter` to impersonate the findings editor Google
    service account:

    ```bash
    gcloud iam service-accounts add-iam-policy-binding \
        $GATEKEEPER_FINDINGS_EDITOR_SA \
        --member "serviceAccount:$GOOGLE_CLOUD_PROJECT.svc.id.goog[gatekeeper-securitycenter/gatekeeper-securitycenter-controller]" \
        --role roles/iam.workloadIdentityUser
    ```

    You create the Kubernetes service account and namespace when you deploy the
    controller.

7.  Install `kpt`:

    ```bash
    gcloud components install kpt --quiet
    ```

    `kpt` is a command-line tool that allows you to manage, manipulate,
    customize, and apply Kubernetes resources. You use `kpt` in this tutorial
    to customize the resource manifests for your environment.

    **Note:** See the `kpt` website for alternative
    [installation instructions](https://googlecontainertools.github.io/kpt/installation/)
    such as using `brew` for macOS.

8.  Fetch the `kpt` package for `gatekeeper-securitycenter`:

    ```bash
    kpt pkg get https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/manifests gatekeeper-securitycenter
    ```

    This creates a directory called `gatekeeper-securitycenter` containing the
    resource manifests files for the controller.

9.  Set the Security Command Center source name:

    ```bash
    kpt cfg set gatekeeper-securitycenter source-name $SOURCE_NAME
    ```

10. Set the optional cluster name. You can use any name you like. For this
    tutorial, use your current `kubectl` context name:

    ```bash
    kpt cfg set gatekeeper-securitycenter cluster $(kubectl config current-context)
    ```

11. Add the Workload Identity annotation to the Kubernetes service account used
    by the controller, to bind it to the findings editor Google service
    account:

    ```bash
    kpt cfg annotate gatekeeper-securitycenter \
        --kind ServiceAccount --name gatekeeper-securitycenter-controller \
        --kv iam.gke.io/gcp-service-account=$GATEKEEPER_FINDINGS_EDITOR_SA
    ```

12. Apply the controller resources to your cluster:

    ```bash
    kpt live init gatekeeper-securitycenter

    kpt live apply gatekeeper-securitycenter --reconcile-timeout=2m --output=table
    ```

    This command creates the following resources in your cluster:

    -   a namespace called `gatekeeper-securitycenter`;
    -   a service account called `gatekeeper-securitycenter-controller`;
    -   a cluster role that provides
        [`get` and `list` access](https://kubernetes.io/docs/reference/access-authn-authz/authorization/#determine-the-request-verb)
        to all resources in all
        [API groups](https://kubernetes.io/docs/reference/using-api/api-overview/#api-groups)
        (this is required because the controller retrieves the resources that
        caused policy violations);
    -   a cluster role binding that grants the cluster role to the service
        account; and
    -   a deployment called `gatekeeper-securitycenter-controller-manager`; and
    -   a config map called `gatekeeper-securitycenter-config` that contains
        configuration values for the deployment.

    The command also waits for the resources to be ready.

    **Note:** If you prefer, you can apply the resources using
    `kubectl apply -f gatekeeper-securitycenter` instead of using `kpt live`.

13. Verify that the controller can read constraint violations and communicate
    with the Security Command Center API by following the controller log:

    ```bash
    kubectl logs deployment/gatekeeper-securitycenter-controller-manager \
        --namespace gatekeeper-securitycenter --follow
    ```

**Note:** Policy Controller and Gatekeeper have a
[default limit on the number of reported violations per constraint](https://github.com/open-policy-agent/gatekeeper#audit).

## Viewing findings

You can view Security Command Center findings on the terminal and in the Cloud
Console.

**Note:** It may take a few minutes for new findings and changes to existing
findings to appear in Security Command Center. If you don't see the expected
findings, wait a few minutes and try again.

1.  Use the `gcloud` tool to list findings for your organization and source:

    ```bash
    gcloud scc findings list $ORGANIZATION_ID \
        --source $(basename $SOURCE_NAME) \
        --format json
    ```

    This command uses `basename` to get the numeric source ID from the full
    source name.

    To understand what the finding attributes mean, see
    [the Finding resource in the Security Command Center API](https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings).

2.  View the findings in the Findings tab of the Security Command Center
    dashboard in the Cloud Console:

    [Go to Security Command Center Findings](https://console.cloud.google.com/security/command-center/findings)

    Select your organization and click **Select**, then click
    **View by Source type**. In the **Source type** list, click **Gatekeeper**.
    On the right, you see the list of findings. If you click on a finding, you
    can see the finding attributes and source properties.

    If you don't see **Gatekeeper** in the **Source type** list, clear any
    filters in the list of findings on the right.

    **Note:** Some of the source properties will be missing (empty) for some
    findings. This is expected, and it is because some resources contain more
    attributes than others. For instance, findings for Pod resources will not
    have a value for the **ResourceAPIGroup** source property because Pod
    doesn't belong to a Kubernetes API group. Only
    [Config Connector](https://cloud.google.com/config-connector/docs/overview)
    resources will have a value for the **ProjectId** source property.

3.  If you fix a resource so it no longer causes a violation, the finding state
    will be set to `inactive`. It can take a few minutes for this change to be
    visible in Security Command Center.

## Troubleshooting

1.  If the `gatekeeper-securitycenter` controller doesn't create findings in
    Security Command Center, you can view logs of the controller manager using
    this command:

    ```bash
    kubectl logs deployment/gatekeeper-securitycenter-controller-manager \
        --namespace gatekeeper-securitycenter
    ```

2.  If the `gatekeeper-securitycenter` controller or command-line tool report
    errors, you can increase the verbosity of the log output by setting the
    `DEBUG` environment variable to `true`:

    ```bash
    DEBUG=true ./gatekeeper-securitycenter [subcommand] [flags]
    ```

3.  If you run into other problems, review these documents:

    -   [Gatekeeper debugging](https://github.com/open-policy-agent/gatekeeper#debugging)
    -   [GKE troubleshooting](https://cloud.google.com/kubernetes-engine/docs/troubleshooting)
    -   [Troubleshooting Kubernetes clusters](https://kubernetes.io/docs/tasks/debug-application-cluster/debug-cluster/)

## Cleaning up

1.  Delete the GKE cluster:

    ```bash
    gcloud container clusters delete gatekeeper-securitycenter-tutorial \
        --zone us-central1-f --async --quiet
    ```

2.  Delete the Cloud IAM policy bindings:

    ```bash
    GOOGLE_CLOUD_PROJECT=$(gcloud config get-value core/project)

    ORGANIZATION_ID=$(gcloud projects get-ancestors $GOOGLE_CLOUD_PROJECT \
        --format json | jq -r '.[] | select (.type=="organization") | .id')

    ./gatekeeper-securitycenter sources remove-iam-policy-binding \
        --source $SOURCE_NAME \
        --member "serviceAccount:gatekeeper-securitycenter@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com" \
        --role roles/securitycenter.findingsEditor \
        --impersonate-service-account securitycenter-sources-admin@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com

    gcloud iam service-accounts remove-iam-policy-binding \
        gatekeeper-securitycenter@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com \
        --member "serviceAccount:$GOOGLE_CLOUD_PROJECT.svc.id.goog[gatekeeper-securitycenter/gatekeeper-securitycenter-controller]" \
        --role roles/iam.workloadIdentityUser

    gcloud iam service-accounts remove-iam-policy-binding \
        securitycenter-sources-admin@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com \
        --member "user:$(gcloud config get-value account)" \
        --role roles/iam.serviceAccountTokenCreator

    gcloud organizations remove-iam-policy-binding $ORGANIZATION_ID \
        --member "serviceAccount:securitycenter-sources-admin@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com" \
        --role roles/serviceusage.serviceUsageConsumer

    gcloud organizations remove-iam-policy-binding $ORGANIZATION_ID \
        --member "serviceAccount:securitycenter-sources-admin@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com" \
        --role roles/securitycenter.sourcesAdmin
    ```

3.  Delete the Google service accounts:

    ```bash
    gcloud iam service-accounts delete --quiet \
        gatekeeper-securitycenter@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com

    gcloud iam service-accounts delete --quiet \
        securitycenter-sources-admin@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com
    ```

## What's next

-   Discover how to
    [run Policy Controller validation as part of a continuous integration pipeline in Cloud Build](https://cloud.google.com/anthos-config-management/docs/how-to/app-policy-validation-ci-pipeline).

-   Learn how to
    [set up notifications for Security Command Center findings](https://cloud.google.com/security-command-center/docs/how-to-notifications).

-   Learn more about
    [enforcing policies with Anthos](https://cloud.google.com/architecture/blueprints/anthos-enforcing-policies-blueprint).

-   Learn more about
    [auditing and monitoring for deviation from policy with Anthos](https://cloud.google.com/architecture/blueprints/anthos-auditing-and-monitoring-for-deviation-from-policy-blueprint).

-   Try out other Google Cloud features for yourself. Have a look at our
    [tutorials](https://cloud.google.com/docs/tutorials).
