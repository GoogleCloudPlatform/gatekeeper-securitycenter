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
the cluster, and you can query them using Kubernetes tools such as `kubectl`.

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

In this tutorial you create a source in Security Command Center using a
command-line tool, and you deploy a controller to a Google Kubernetes Engine
(GKE) cluster to synchronize Policy Controller and Gatekeeper constraint
violations to findings in Security Command Center.

<walkthrough-alt>

This diagram shows the components involved in this tutorial:

![Architecture](architecture.svg)

</walkthrough-alt>

If you want to see how this can also apply to policy violations for Google
Cloud resources, check out the
[tutorial on how to create policy-compliant Google Cloud resources using Config Connector with Policy Controller or Gatekeeper](https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/blob/main/docs/config-connector-gatekeeper-tutorial.md).

## Objectives

-   Create a Google Kubernetes Engine (GKE) cluster
-   Install Policy Controller or Gatekeeper
-   Create a Gatekeeper policy and a resource that violates the policy
-   Create a Google service account that can administer Security Command Center
    sources
-   Create a source in Security Command Center
-   Create a Google service account that can create and edit findings for the
    Security Command Center source
-   Create a finding in Security Command Center from a Gatekeeper policy
    violation using the `gatekeeper-securitycenter` command-line tool
-   Deploy the `gatekeeper-securitycenter` controller to the GKE cluster to
    periodically synchronize findings in Security Command Center from
    Policy Controller or Gatekeeper policy violations
-   View findings in your terminal and in the Cloud Console

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

4.  Enable the Google Kubernetes Engine and Security Command Center APIs:

    ```bash
    gcloud services enable \
        container.googleapis.com \
        securitycenter.googleapis.com
    ```

## Creating a GKE cluster

These steps use
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

    You need this role later to deploy the Kubernetes controller. You also need
    it if you install the open source Gatekeeper distribution.

## Installing Policy Controller

If you have an [Anthos entitlement](https://cloud.google.com/anthos/pricing),
follow the steps in this section to install Policy Controller. If not, skip to
the next section to install the open source Gatekeeper distribution instead.

1.  Enable the Anthos API:

    ```bash
    gcloud services enable anthos.googleapis.com
    ```

2.  Download the Config Management operator
    [custom resource definition (CRD)](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
    manifest and apply it to your cluster:

    ```bash
    gsutil cp gs://config-management-release/released/latest/config-management-operator.yaml config-management-operator.yaml

    kubectl apply -f config-management-operator.yaml
    ```

3.  Create and apply a `ConfigManagement` manifest based on the operator CRD.
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

    kubectl apply -f config-management.yaml
    ```

4.  Wait for Policy Controller to be ready, this could take a few minutes:

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
Controller in the previous section, skip to the
[next section](#creating-a-policy).

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

## Creating a policy

A policy in Policy Controller and Gatekeeper consists of a
[constraint template](https://github.com/open-policy-agent/frameworks/tree/master/constraint#what-is-a-constraint-template)
and a
[constraint](https://github.com/open-policy-agent/frameworks/tree/master/constraint#what-is-a-constraint).
The constraint template contains the policy logic, and the constraint specifies
where the policy applies and input parameters to the policy logic.

In this section you create a policy for Kubernetes pods and a pod that violates
the policy. The policy requires that pod specifications use container images
from approved repositories.

1.  Clone the Gatekeeper library repository, navigate to the repository
    directory, and check out a known commit:

    ```bash
    git clone https://github.com/open-policy-agent/gatekeeper-library.git ~/gatekeeper-library

    cd ~/gatekeeper-library

    git checkout ce24dd6802b8c845f80a27731b9095cc0864726f
    ```

2.  Create a pod called `nginx-disallowed` in the `default` namespace:

    ```bash
    kubectl apply --namespace default -f \
        library/general/allowedrepos/samples/repo-must-be-openpolicyagent/example_disallowed.yaml
    ```

    This pod uses a container image from a repository that is not approved by
    the policy.

3.  Create a constraint template called `k8sallowedrepos`:

    ```bash
    kubectl apply -f library/general/allowedrepos/template.yaml
    ```

14. Create a constraint called `repo-is-openpolicyagent`:

    ```bash
    kubectl apply -f \
        library/general/allowedrepos/samples/repo-must-be-openpolicyagent/constraint.yaml
    ```

    This constraint applies to all pods in the `default` namespace.

## Auditing constraints

The [audit controller](https://github.com/open-policy-agent/gatekeeper#audit)
in Policy Controller and Gatekeeper periodically evaluates resources against
constraints. This allows you to detect policy-violating resources created prior
to the constraint.

1.  View violations for all constraints by querying using the `constraint`
    category:

    ```bash
    kubectl get constraint -o json | jq '.items[].status.violations'
    ```

    The output looks like this. You see a violation for the pod you created
    before creating the constraint:

    ```terminal
    [
      {
        "enforcementAction": "deny",
        "kind": "Pod",
        "message": "container <nginx> has an invalid image repo <nginx>, allowed repos are [\"openpolicyagent\"]",
        "name": "nginx-disallowed",
        "namespace": "default"
      }
    ]
    ```

    If you see `null` instead of the output above, wait a minute and try again.
    This can happen because the Gatekeeper audit hasn't run since you created
    the constraint. The audit runs every minute by default.

**Note:** Policy Controller and Gatekeeper have a
[default limit on the number of reported violations per constraint](https://github.com/open-policy-agent/gatekeeper#audit).

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

2.  Define an environment variable containing your
    [Google Cloud organization ID](https://cloud.google.com/resource-manager/docs/creating-managing-organization#retrieving_your_organization_id):

    ```bash
    ORGANIZATION_ID=$(gcloud projects get-ancestors $GOOGLE_CLOUD_PROJECT \
        --format json | jq -r '.[] | select (.type=="organization") | .id')
    ```

3.  Grant the
    [Security Center Sources Admin](https://cloud.google.com/iam/docs/understanding-roles#security-center-roles)
    Cloud Identity and Access Management (IAM) role to the sources admin
    Google service account at the organization level:

    ```bash
    gcloud organizations add-iam-policy-binding \
        $ORGANIZATION_ID \
        --member "serviceAccount:$SOURCES_ADMIN_SA" \
        --role roles/securitycenter.sourcesAdmin
    ```

    This role provides the
    [`securitycenter.sources.*`](https://cloud.google.com/iam/docs/permissions-reference)
    permissions required to administer sources.

4.  Grant the
    [Service Usage Consumer](https://cloud.google.com/iam/docs/understanding-roles#service-usage-roles)
    role to the sources Google service account at the organization level:

    ```bash
    gcloud organizations add-iam-policy-binding \
        $ORGANIZATION_ID \
        --member "serviceAccount:$SOURCES_ADMIN_SA" \
        --role roles/serviceusage.serviceUsageConsumer
    ```

    This role provides the
    [`serviceusage.services.use`](https://cloud.google.com/iam/docs/permissions-reference)
    permission which is required to allow other identities to
    [impersonate](https://cloud.google.com/iam/docs/impersonating-service-accounts)
    the Google service account when using Google Cloud client libraries, as
    long as the other identities have the necessary permissions.

5.  Grant your user identity the
    [Service Account Token Creator](https://cloud.google.com/iam/docs/understanding-roles#service-accounts-roles)
    role for the sources admin Google service account:

    ```bash
    gcloud iam service-accounts add-iam-policy-binding \
        $SOURCES_ADMIN_SA \
        --member "user:$(gcloud config get-value account)" \
        --role roles/iam.serviceAccountTokenCreator
    ```

    This allows your user identity to
    [impersonate](https://cloud.google.com/iam/docs/impersonating-service-accounts)
    the Google service account.

6.  Download the latest version of the `gatekeeper-securitycenter` command-line
    tool for your platform:

    ```bash
    VERSION=$(curl -s https://api.github.com/repos/GoogleCloudPlatform/gatekeeper-securitycenter/releases/latest | jq -r '.tag_name')
    OS=$(go env GOOS)
    ARCH=$(go env GOARCH)

    curl -sLo gatekeeper-securitycenter "https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/releases/download/${VERSION}/gatekeeper-securitycenter_${OS}_${ARCH}"

    chmod +x gatekeeper-securitycenter
    ```

    **Note:** If you prefer to build your own binary, follow the instructions
    in the [`README.md`](https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/blob/main/README.md#build).

7.  Use the `gatekeeper-securitycenter` tool to create a Security Command
    Center source for your organization. Capture the full source name in an
    exported environment variable:

    ```bash
    export SOURCE_NAME=$(./gatekeeper-securitycenter sources create \
        --organization $ORGANIZATION_ID \
        --display-name "Gatekeeper" \
        --description "Reports violations from Gatekeeper audits" \
        --impersonate-service-account $SOURCES_ADMIN_SA | jq -r '.name')
    ```

    This command creates a source with the display name **Gatekeeper**. This
    display name is visible in the Security Command Center console. You can use
    a different display name and description if you like.

    If you get a response with the error message
    `The caller does not have permission`, wait a minute and try again. This
    can happen if the Cloud IAM bindings haven't taken effect yet.

## Creating findings using the command line

You can create Security Command Center findings from Policy Controller and
Gatekeeper constraint violations using the `gatekeeper-securitycenter` tool as
part of a build pipeline or scheduled task.

1.  Create a Google service account and store the service account name in an
    environment variable:

    ```bash
    FINDINGS_EDITOR_SA=$(gcloud iam service-accounts create \
        gatekeeper-securitycenter \
        --display-name "Security Command Center Gatekeeper findings editor" \
        --format 'value(email)')
    ```

    You use this Google service account to create findings for your Security
    Command Center source.

2.  Use the `gatekeeper-securitycenter` tool to grant the
    [Security Center Findings Editor](https://cloud.google.com/iam/docs/understanding-roles#security-center-roles)
    role to the findings editor Google service account for your Security
    Command Center source:

    ```bash
    ./gatekeeper-securitycenter sources add-iam-policy-binding \
        --source $SOURCE_NAME \
        --member "serviceAccount:$FINDINGS_EDITOR_SA" \
        --role roles/securitycenter.findingsEditor \
        --impersonate-service-account $SOURCES_ADMIN_SA
    ```

    This role provides the
    [`securitycenter.findings.*`](https://cloud.google.com/iam/docs/permissions-reference)
    permissions required to create and edit findings.

    When you execute this command, you impersonate the sources admin Google
    service account.

    **Note:** You can grant the role at the organization level instead of on
    the source. By granting at the organization level, you grant the Google
    service account permission to edit findings for all sources.

3.  Grant the
    [Service Usage Consumer](https://cloud.google.com/iam/docs/understanding-roles#service-usage-roles)
    role to the findings editor Google service account at the organization
    level:

    ```bash
    gcloud organizations add-iam-policy-binding $ORGANIZATION_ID \
        --member "serviceAccount:$FINDINGS_EDITOR_SA" \
        --role roles/serviceusage.serviceUsageConsumer
    ```

4.  Grant your user identity the
    [Service Account Token Creator](https://cloud.google.com/iam/docs/understanding-roles#service-accounts-roles)
    role for the findings editor Google service account:

    ```bash
    gcloud iam service-accounts add-iam-policy-binding \
        $FINDINGS_EDITOR_SA \
        --member "user:$(gcloud config get-value account)" \
        --role roles/iam.serviceAccountTokenCreator
    ```

5.  Run `gatekeeper-securitycenter findings sync` in dry-run mode. This
    command retrieves audit violations from the cluster, but it prints
    findings to the console instead of creating them in Security Command
    Center:

    ```bash
    ./gatekeeper-securitycenter findings sync --dry-run=true
    ```

    This command uses your current
    [kubeconfig context](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)
    by default. If you want to use a different kubeconfig file, use the
    `-kubeconfig` flag.

    <walkthrough-alt>

    The output looks similar to this:

    ```terminal
    [
      {
        "finding_id": "0be44bcf181ef03162eed40126a500a0",
        "finding": {
          "resource_name": "https://[apiserver]/api/v1/namespaces/default/pods/    nginx-disallowed",
          "state": 1,
          "category": "K8sAllowedRepos",
          "external_uri": "https://[apiserver]/apis/constraints.gatekeeper.sh/v1beta1/k8sallowedrepos/repo-is-openpolicyagent",
          "source_properties": {
            "Cluster": "",
            "ConstraintName": "repo-is-openpolicyagent",
            "ConstraintSelfLink": "https://[apiserver]/apis/constraints.gatekeeper.sh/v1beta1/k8sallowedrepos/repo-is-openpolicyagent",
            "ConstraintTemplateSelfLink": "https://[apiserver]/apis/templates.    gatekeeper.sh/v1beta1/constrainttemplates/k8sallowedrepos",
            "ConstraintTemplateUID": "e35b1c39-15f7-4a7a-afae-1637b44e81b2",
            "ConstraintUID": "b904dddb-0a23-4f4f-81bb-0103de838d3e",
            "Explanation": "container \u003cnginx\u003e has an invalid image repo \u003cnginx\u003e, allowed repos are [\"openpolicyagent\"]",
            "ProjectId": "",
            "ResourceAPIGroup": "",
            "ResourceAPIVersion": "v1",
            "ResourceKind": "Pod",
            "ResourceName": "nginx-disallowed",
            "ResourceNamespace": "default",
            "ResourceSelfLink": "https://[apiserver]/api/v1/namespaces/default/pods/nginx-disallowed",
            "ResourceStatusSelfLink": "",
            "ResourceUID": "8ddd752f-e620-43ea-b966-4ae2ae507c67",
            "ScannerName": "GATEKEEPER"
          },
          "event_time": {
            "seconds": 1606287680
          }
        }
      }
    ]
    ```

    where `[apiserver]` is the IP address or hostname of your GKE cluster
    [API server](https://kubernetes.io/docs/concepts/overview/components/#kube-apiserver).

    </walkthrough-alt>

    To learn what the attributes mean, see the
    [Finding resource](https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings)
    in the Security Command Center API.

6.  Run `gatekeeper-securitycenter findings sync` to create findings in
    Security Command Center. Specify the source you created in the previous
    section:

    ```bash
    ./gatekeeper-securitycenter findings sync \
        --source $SOURCE_NAME \
        --impersonate-service-account $FINDINGS_EDITOR_SA
    ```

    When you execute this command, you impersonate the findings editor Google
    service account.

    The output of the command includes a log entry with the message
    `create finding`. This means that the `gatekeeper-securitycenter`
    command-line tool created a finding.

    The `findingID` attribute of that log entry contains the full name of the
    finding in the format
    `organizations/[organization_id]/sources/[source_id]/findings/[finding_id]`,
    where `[organization_id]` is your Google Cloud organization ID,
    `[source_id]` is your Security Command Center source ID, and `[finding_id]`
    is the finding ID.

    <walkthrough-alt>

    To view the finding, see the section
    [Viewing findings](#viewing-findings).

    </walkthrough-alt>

## Creating findings using a Kubernetes controller

You can deploy `gatekeeper-securitycenter` as a
[controller](https://kubernetes.io/docs/concepts/architecture/controller/)
in your Kubernetes cluster. This controller periodically checks for constraint
violations and creates a finding in Security Command Center for each violation.

If the resource becomes compliant, the controller sets the state of the
existing finding to _inactive_.

1.  Create a
    [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
    Cloud IAM policy binding to allow the Kubernetes service account
    `gatekeeper-securitycenter-controller` in the namespace
    `gatekeeper-securitycenter` to impersonate the findings editor Google
    service account:

    ```bash
    gcloud iam service-accounts add-iam-policy-binding \
        $FINDINGS_EDITOR_SA \
        --member "serviceAccount:$GOOGLE_CLOUD_PROJECT.svc.id.goog[gatekeeper-securitycenter/gatekeeper-securitycenter-controller]" \
        --role roles/iam.workloadIdentityUser
    ```

    You create the Kubernetes service account and namespace when you deploy the
    controller.

2.  Install [`kpt`](https://googlecontainertools.github.io/kpt/):

    ```bash
    sudo apt-get install -y google-cloud-sdk-kpt
    ```

    <walkthrough-alt>

    If you _don't_ use Cloud Shell, install `kpt` using this command instead:

    ```bash
    gcloud components install kpt --quiet
    ```

    </walkthrough-alt>

    `kpt` is a command-line tool that allows you to manage, manipulate,
    customize, and apply Kubernetes resources. You use `kpt` in this tutorial
    to customize the resource manifests for your environment.

    **Note:** See the `kpt` website for alternative
    [installation instructions](https://googlecontainertools.github.io/kpt/installation/),
    such as using `brew` for macOS.

3.  Fetch the latest version of the `kpt` package for
    `gatekeeper-securitycenter`:

    ```bash
    VERSION=$(curl -s https://api.github.com/repos/GoogleCloudPlatform/gatekeeper-securitycenter/releases/latest | jq -r '.tag_name')

    kpt pkg get https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter.git/manifests@$VERSION manifests
    ```

    This creates a directory called `manifests` containing the
    resource manifests files for the controller.

4.  Set the Security Command Center source name:

    ```bash
    kpt cfg set manifests source $SOURCE_NAME
    ```

5.  Set the cluster name. This step is optional, and you can use any name you
    like. For this tutorial, use your current `kubectl` context name:

    ```bash
    kpt cfg set manifests cluster $(kubectl config current-context)
    ```

    **Note:** Other
    [setters](https://googlecontainertools.github.io/kpt/guides/consumer/set/)
    are available to customize the package. Run the command
    `kpt cfg list-setters manifests` to list the available
    setters and their values.

6.  Add the Workload Identity annotation to the Kubernetes service account used
    by the controller, to bind it to the findings editor Google service
    account:

    ```bash
    kpt cfg annotate manifests \
        --kind ServiceAccount \
        --name gatekeeper-securitycenter-controller \
        --namespace gatekeeper-securitycenter \
        --kv iam.gke.io/gcp-service-account=$FINDINGS_EDITOR_SA
    ```

7.  Initialize the package to allow `kpt` to track the controller resources in
    your cluster:

    ```bash
    kpt live init manifests --namespace gatekeeper-securitycenter
    ```

8.  Apply the controller resources to your cluster:

    ```bash
    kpt live apply manifests --reconcile-timeout 2m --output events
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

9.  Verify that the controller can read constraint violations and communicate
    with the Security Command Center API by following the controller log:

    ```bash
    kubectl logs deployment/gatekeeper-securitycenter-controller-manager \
        --namespace gatekeeper-securitycenter --follow
    ```

    You see log entries with the message `syncing findings`.

    Press **Ctrl+C** to stop following the log.

10. To verify that the controller can create new findings, create a policy and
    a resource that violates the policy. The policy requires that pod
    specifications refer to container images using digests.

    Navigate to the Gatekeeper library repository directory:

    ```bash
    cd ~/gatekeeper-library
    ```

11. Create a pod called `opa-disallowed` in the `default` namespace:

    ```bash
    kubectl apply --namespace default -f \
        library/general/imagedigests/samples/container-image-must-have-digest/example_disallowed.yaml
    ```

    This pod specification refers to a container image by tag instead of by
    digest.

12. Create a constraint template called `k8simagedigests`:

    ```bash
    kubectl apply -f library/general/imagedigests/template.yaml
    ```

13. Create a constraint called `container-image-must-have-digest`:

    ```bash
    kubectl apply -f \
        library/general/imagedigests/samples/container-image-must-have-digest/constraint.yaml
    ```

    This constraint only applies to the `default` namespace.

14. Following the controller log:

    ```bash
    kubectl logs deployment/gatekeeper-securitycenter-controller-manager \
        --namespace gatekeeper-securitycenter --follow
    ```

    After a few minutes, you see a log entry with the message
    `create finding`. This means that the `gatekeeper-securitycenter`
    controller created a finding.

    Press **Ctrl+C** to stop following the log.

15. To verify that the controller can set the finding state to
    [`INACTIVE`](https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings#State)
    when a violation is no longer reported by Policy Controller or Gatekeeper,
    delete the pod called `opa-disallowed` in the `default` namespace:

    ```bash
    kubectl delete pod opa-disallowed --namespace default
    ```

16. Follow the controller log:

    ```bash
    kubectl logs deployment/gatekeeper-securitycenter-controller-manager \
        --namespace gatekeeper-securitycenter --follow
    ```

    After a few minutes, you see a log entry with the message
    `updating finding state` and the attribute `"state":"INACTIVE"`.
    This means the controller set the finding state to inactive.

    Press **Ctrl+C** to stop following the log.

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

    You use the `basename` command to get the numeric source ID from the full
    source name.

    <walkthrough-alt>

    The output looks similar to this:

    ```terminal
    [
      {
        "finding": {
          "category": "K8sAllowedRepos",
          "createTime": "2020-11-25T06:58:47.213Z",
          "eventTime": "2020-11-25T06:58:20Z",
          "externalUri": "https://[apiserver]/apis/constraints.gatekeeper.sh/v1beta1/k8sallowedrepos/repo-is-openpolicyagent",
          "name": "organizations/[organization_id]/sources/[source_id]/findings/    [finding_id]",
          "parent": "organizations/[organization_id]/sources/[source_id]",
          "resourceName": "https://[apiserver]/api/v1/namespaces/default/pods/    nginx-disallowed",
          "securityMarks": {
            "name": "organizations/[organization_id]/sources/[source_id]/findings/[finding_id]/securityMarks"
          },
          "sourceProperties": {
            "Cluster": "[cluster-name]",
            "ConstraintName": "repo-is-openpolicyagent",
            "ConstraintSelfLink": "https://[apiserver]/apis/constraints.gatekeeper.sh/v1beta1/k8sallowedrepos/repo-is-openpolicyagent",
            "ConstraintTemplateSelfLink": "https://[apiserver]/apis/templates.    gatekeeper.sh/v1beta1/constrainttemplates/k8sallowedrepos",
            "ConstraintTemplateUID": "e35b1c39-15f7-4a7a-afae-1637b44e81b2",
            "ConstraintUID": "b904dddb-0a23-4f4f-81bb-0103de838d3e",
            "Explanation": "container <nginx> has an invalid image repo <nginx>, allowed repos are [\"openpolicyagent\"]",
            "ProjectId": "",
            "ResourceAPIGroup": "",
            "ResourceAPIVersion": "v1",
            "ResourceKind": "Pod",
            "ResourceName": "nginx-disallowed",
            "ResourceNamespace": "default",
            "ResourceSelfLink": "https://[apiserver]/api/v1/namespaces/default/pods/nginx-disallowed",
            "ResourceStatusSelfLink": "",
            "ResourceUID": "8ddd752f-e620-43ea-b966-4ae2ae507c67",
            "ScannerName": "GATEKEEPER"
          },
          "state": "ACTIVE"
        },
        "resource": {
          "name": "https://[apiserver]/api/v1/namespaces/default/pods/nginx-disallowed"
        }
      },
      {
        "finding": {
          "category": "K8sImageDigests",
          [...]
      }
    ]
    ```

    where `[apiserver]` is the IP address or hostname of your GKE cluster
    [API server](https://kubernetes.io/docs/concepts/overview/components/#kube-apiserver),
    `[organization_id]` is your Google Cloud organization ID, `[source_id]` is
    your Security Command Center source ID, and `[finding_id]` is the
    finding ID.

    </walkthrough-alt>

    To learn what the finding attributes mean, see
    [the Finding resource in the Security Command Center API](https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings).

    **Note:** Some of the source properties will be missing (empty) for some
    findings. This is expected, and it is because some resources contain more
    attributes than others. For instance, findings for Pod resources will not
    have a value for the **ResourceAPIGroup** source property because Pod
    doesn't belong to a Kubernetes API group. Only
    [Config Connector](https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/blob/main/docs/config-connector-gatekeeper-tutorial.md)
    resources will have a value for the **ProjectId** source property.

2.  View the findings in the Findings tab of the Security Command Center
    dashboard in the Cloud Console:

    [Go to Security Command Center Findings](https://console.cloud.google.com/security/command-center/findings)

    Select your organization and click **Select**, then click
    **View by Source type**. In the **Source type** list, click **Gatekeeper**.
    On the right, you see the list of findings. If you click on a finding, you
    can see the finding attributes and source properties.

    If you don't see **Gatekeeper** in the **Source type** list, clear any
    filters in the list of findings on the right.

    If a resource no longer causes a violation because of a change to the
    resource or the policy, the controller sets the finding state to
    _inactive_. It can take a few minutes for this change to be visible in
    Security Command Center.

    By default, the Security Command Center dashboard shows active findings.
    You can see inactive findings by turning off the toggle
    **Show Only Active Findings**.

## Troubleshooting

1.  If Policy Controller or Gatekeeper don't report violations in the `status`
    field of the constraint objects, you can view logs of the audit controller
    using this command:

    ```bash
    kubectl logs deployment/gatekeeper-audit -n gatekeeper-system
    ```

2.  If the `gatekeeper-securitycenter` controller doesn't create findings in
    Security Command Center, you can view logs of the controller manager using
    this command:

    ```bash
    kubectl logs deployment/gatekeeper-securitycenter-controller-manager \
        --namespace gatekeeper-securitycenter
    ```

3.  If the `gatekeeper-securitycenter` controller or command-line tool report
    errors, you can increase the verbosity of the log output by setting the
    `DEBUG` environment variable to `true`:

    ```
    DEBUG=true ./gatekeeper-securitycenter [subcommand] [flags]
    ```

3.  If you run into other problems with this tutorial, we recommend that you
    review these documents:

    -   [Gatekeeper debugging](https://github.com/open-policy-agent/gatekeeper#debugging)
    -   [GKE troubleshooting](https://cloud.google.com/kubernetes-engine/docs/troubleshooting)
    -   [Troubleshooting Kubernetes clusters](https://kubernetes.io/docs/tasks/debug-application-cluster/debug-cluster/)

## Automating the setup

For future deployments, you can automate some of the steps in this tutorial by
downloading and running a setup script that performs the following tasks:

-   Creates the sources admin and findings editor Google service accounts.
-   Grants the Cloud IAM role bindings to the Google service accounts.
-   Grants permissions to allow your Google identity to impersonate the Google
    service accounts.
-   Creates the Security Command Center source.

1.  Download the setup script:

    ```
    curl -sLO https://raw.githubusercontent.com/GoogleCloudPlatform/gatekeeper-securitycenter/main/scripts/setup.sh
    ```

2.  Execute the script:

    ```
    bash setup.sh
    ```

    The script prints out the names of the Security Command Center source
    (`SOURCE_NAME`) and the findings editor Google service account
    (`FINDINGS_EDITOR_SA`). You need these values to deploy the controller to
    your GKE cluster.

## Cleaning up

To avoid incurring further charges to your Google Cloud Platform account for
the resources used in this tutorial, delete the individual resources.

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
        gatekeeper-securitycenter@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com \
        --member "user:$(gcloud config get-value account)" \
        --role roles/iam.serviceAccountTokenCreator

    gcloud organizations remove-iam-policy-binding $ORGANIZATION_ID \
        --member "serviceAccount:gatekeeper-securitycenter@$GOOGLE_CLOUD_PROJECT.iam.gserviceaccount.com" \
        --role roles/serviceusage.serviceUsageConsumer

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

-   Learn how to
    [create policy-compliant Google Cloud resources using Config Connector and Policy Controller or Gatekeeper](https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/blob/main/docs/config-connector-gatekeeper-tutorial.md).

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
