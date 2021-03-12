# `gatekeeper-securitycenter` controller design

![Architecture](architecture.svg)

The `gatekeeper-securitycenter` controller is based on
[`client-go`](https://github.com/kubernetes/client-go).

It doesn't contain a webhook, and it doesn't introduce any CRDs.

## Control loop

For each iteration of the control loop:

1.  Use a
    [discovery client](https://github.com/kubernetes/client-go/blob/master/discovery)
    to find all
    [(group, resource)](https://github.com/kubernetes/apimachinery/blob/v0.19.4/pkg/runtime/schema/group_version.go#L54)
    types for Gatekeeper constraints by querying using the
    [`constraint` category](https://github.com/open-policy-agent/frameworks/blob/125fdd8ffe0907aa5bf66c95d8939faed5a97f86/constraint/pkg/client/crd_helpers.go#L97).

2.  Add the `v1beta1` version to turn `GroupResource`s into
    [`GroupVersionResource`](https://github.com/kubernetes/apimachinery/blob/v0.19.4/pkg/runtime/schema/group_version.go#L96)s
    (GVR).

3.  Use a
    [dynamic client](https://github.com/kubernetes/client-go/tree/master/dynamic)
    to find all constraint resources for each GVR. These are
    [`Unstructured`](https://github.com/kubernetes/apimachinery/blob/v0.19.4/pkg/apis/meta/v1/unstructured/unstructured.go#L31)
    objects (a wrapper around `map[string]interface{}` with some helpers).

4.  Filter constraints to only keep those that have violations, by checking
    the presence of the
    [`status.violations`](https://github.com/open-policy-agent/gatekeeper#audit)
    field. Use a map to avoid duplicates from the dynamic client, the map key
    is the constraint UID.

5.  For each violated constraint, create a
    [`Constraint`](../pkg/sync/request.go#L45) struct using fields from the
    `Unstructured` constraint object. The struct fields are used when creating
    the finding request. Then:

    -   For each violation, use a dynamic client to get the resource that
        violated the constraint and the constraint template. Use these objects
        to create a [`Resource`](../pkg/sync/request.go#L32) struct. The struct
        fields are used when creating the finding request.

    -   Use the `Constraint` and `Resource` instances to create a
        [`CreateFindingRequest`](https://pkg.go.dev/google.golang.org/genproto/googleapis/cloud/securitycenter/v1#CreateFindingRequest).
        The constraint `Kind` is used as the finding category. The request
        contains a finding ID, see the [section below](#finding-id) for how
        this is determined. For
        [Config Connector resources](https://cloud.google.com/solutions/policy-compliant-resources),
        use the `status.selfLink` value as the resource name. This results in
        Security Command Center rendering a link directly to the Google Cloud
        resource in the finding.

        **Note:** When collecting fields for the `Constraint` and `Resource`
        instances, the controller is tolerant of errors and will default to
        empty string values for fields that aren't required to create a finding
        request.

    -   Collect all finding requests in a map, where the key is the finding
        name in the format
        `organizations/[organization_id]/sources/[source_id]/findings/[finding_id]`.

6.  Iterate over all existing findings for the configured source in Security
    Command Center to sync the
    [finding state](https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings#State)
    with the violations found:

    -   If the existing finding is present in the finding request map, ensure
        that the finding state is `active` (i.e., set it to `active` if it's in
        a different state).

    -   If the existing finding is _not_ present in the finding request map,
        ensure that the finding state is `inactive` (i.e., set it to `inactive`
        if it's in a different state).

7.  Get the subset of the finding requests map that did _not_ already exist for
    the configured source in Security Command Center. For each of these finding
    requests, create new a finding (with `state=active`).

8.  Sleep for the configured interval (default is 2 minutes), then
    rinse-and-repeat.

## Finding ID

The finding ID is a value (1-32 alphanumeric chars) that must be unique for a
source. Because it's client-supplied, the controller ensures it doesn't create
duplicate findings for the same violation on each control loop iteration by
creating the finding ID deterministically.

A change to the constraint, constraint template, or the resource that violated
the constraint results in a new finding. This is because the user could have
[set the state](https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings/setState)
of the previous finding to
[`INACTIVE`](https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings#State),
meaning _"The finding has been fixed, triaged as a non-issue or otherwise
addressed and is no longer active."_ A change to the resource or the policy
means it has to be triaged again.

To detect these changes, the finding ID is determined as follows:

1.  Concatenate the following strings:
    -   constraint UID
    -   constraint spec as JSON string
    -   constraint template UID
    -   constraint template spec as JSON string
    -   the UID of the resource that violated the constraint)
    -   resource spec as JSON string
2.  Calculate the SHA-256 hash of the concatenated string.
3.  Take the first 32 characters of the hash.

## Limitations

-   Policy Controller and Gatekeeper have a
    [default limit of 20 reported violations per constraint](https://github.com/open-policy-agent/gatekeeper#audit).

-   The controller doesn't collect or expose any metrics for monitoring.
