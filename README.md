# gatekeeper-securitycenter

`gatekeeper-securitycenter` allows you to use Security Command Center as a
dashboard for Kubernetes resource policy violations.

`gatekeeper-securitycenter` is:

-   a Kubernetes controller that creates
    [Security Command Center](https://cloud.google.com/security-command-center/docs)
    [findings](https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings)
    for violations reported by the
    [audit controller](https://cloud.google.com/anthos-config-management/docs/how-to/auditing-constraints)
    in
    [Policy Controller](https://cloud.google.com/anthos-config-management/docs/concepts/policy-controller)
    and
    [Open Policy Agent (OPA) Gatekeeper](https://github.com/open-policy-agent/gatekeeper).

-   a command-line tool that creates Security Command Center
    [sources](https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources)
    and manages the IAM policies of the sources.

![Architecture](docs/architecture.svg)

`gatekeeper-securitycenter` requires
[Security Command Center Standard tier](https://cloud.google.com/security-command-center/pricing#standard_tier_pricing).

## Prerequisites

Before installing the `gatekeeper-securitycenter` controller, create all of the
following resources:

-   a Kubernetes cluster, for instance a Google Kubernetes Engine (GKE) cluster
-   Policy Controller or OPA Gatekeeper installed in the Kubernetes cluster
-   a Security Command Center source
-   a Google service account with the
    [Security Center Findings Editor](https://cloud.google.com/security-command-center/docs/access-control)
    role on the Security Command Center source

To create these prerequisite resources, choose one of these options:

1.  Use the [kpt](https://kpt.dev) package in the [`setup`](setup) directory.
    This package creates resources using
    [Config Connector](https://cloud.google.com/config-connector/docs/overview).

2.  Use the shell scripts in the [`scripts`](scripts) directory. These scripts
    create resources using the `gcloud` tool from the
    [Google Cloud SDK](https://cloud.google.com/sdk).

3.  Follow the step-by-step instructions in the accompanying
    [tutorial](https://cloud.google.com/architecture/reporting-policy-controller-audit-violations-security-command-center).

For all options, you must have an appropriate Cloud IAM role for Security
Command Center at the organization level, such as
[Security Center Admin Editor](https://cloud.google.com/security-command-center/docs/access-control).
Your organization administrator can
[grant you this role](https://cloud.google.com/resource-manager/docs/access-control-org).

If your user account is not associated with an
[organization](https://cloud.google.com/resource-manager/docs/creating-managing-organization)
on Google Cloud, you can create an organization resource by signing up for
either [Cloud Identity](https://cloud.google.com/identity) or
[Google Workspace](https://workspace.google.com/) (formerly G Suite) using a
domain you own. Cloud Identity offers a
[free edition](https://gsuite.google.com/signup/gcpidentity/welcome).

## Installing the `gatekeeper-securitycenter` controller

Install the `gatekeeper-securitycenter` controller in your cluster by following
the [documentation in the manifest directory](manifests/README.md).

## Documentation

-   [Tutorial](https://cloud.google.com/architecture/reporting-policy-controller-audit-violations-security-command-center)

-   [Building `gatekeeper-securitycenter`](docs/build.md)

-   [Developing `gatekeeper-securitycenter`](docs/development.md)

-   [Releasing `gatekeeper-securitycenter`](docs/release.md)

## Disclaimer

This is not an officially supported Google product.
