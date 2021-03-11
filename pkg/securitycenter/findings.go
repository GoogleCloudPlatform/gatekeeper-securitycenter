// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package securitycenter

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	securitycenterpb "google.golang.org/genproto/googleapis/cloud/securitycenter/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	errorutils "k8s.io/apimachinery/pkg/util/errors"
)

var (
	errIterator = errors.New("iterator error")
)

// SyncFindings synchronizes the findings already in Security Command Center (SCC) with the
// provided finding requests
func (c *Client) SyncFindings(ctx context.Context, source string, findingRequests map[string]*securitycenterpb.CreateFindingRequest) error {
	c.log.Info("syncing findings", "source", source, "numActiveFindings", len(findingRequests))
	newFindingRequests, err := c.syncFindingsState(ctx, source, findingRequests)
	if err != nil && newFindingRequests == nil {
		return err
	}
	if err != nil {
		c.log.Error(err, "findings state sync errors")
	}
	var createFindingErrs []error
	for _, req := range newFindingRequests {
		if err := c.CreateFinding(ctx, req); err != nil {
			createFindingErrs = append(createFindingErrs, err)
		}
	}
	return errorutils.NewAggregate(createFindingErrs) // returns nil if errs is empty
}

// syncFindingsState updates the state of existing findings for the provided source in Security
// Command Center based on their presence in the findingRequests input parameter.
//
// Existing findings that are present in the findingRequests input have their state set to ACTIVE.
// Existing findings that are _not_ present in the findingRequests input have their state set to INACTIVE.
//
// The `source` input parameter should be of the format `organizations/[organization_id]/sources/[source_id]`
// To sync across all sources provide a "-" as the source_id.
//
// The key in the findingRequests map is the full finding name of the format
// `organizations/[organization_id]/sources/[source_id]/findings/[finding_id]`
//
// Returns the subset of findingRequests from the input that were _not_ already present in SCC.
// These request objects can then be used to create new findings.
//
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings/setState
func (c *Client) syncFindingsState(ctx context.Context, source string, findingRequests map[string]*securitycenterpb.CreateFindingRequest) ([]*securitycenterpb.CreateFindingRequest, error) {
	if c.log.V(2).Enabled() {
		for findingName := range findingRequests {
			c.log.V(2).Info("findingRequest", "findingIDToName", findingName)
		}
	}
	syncedFindingNames := map[string]bool{}
	var pageToken string
	ensureStateFn := func(ctx context.Context, finding *securitycenterpb.Finding) (*securitycenterpb.Finding, error) {
		c.log.V(2).Info("ensure state", "finding", finding.Name)
		_, exists := findingRequests[finding.Name]
		return c.ensureFindingState(ctx, finding, exists)
	}
	var ensureStateFnErrors []error
	for {
		syncedFindingsFromPage, pageToken, err := c.mapFindingsPage(ctx, source, pageToken, ensureStateFn)
		if err != nil && errors.Is(err, errIterator) {
			return nil, err // iterator error, stop
		}
		switch mapErr := err.(type) {
		case errorutils.Aggregate:
			// err is collection of errors from applying ensureStateFn. Collect them.
			ensureStateFnErrors = append(ensureStateFnErrors, mapErr.Errors()...)
		default:
			if err != nil {
				return nil, err // some other error, stop
			}
		}
		for _, syncedFinding := range syncedFindingsFromPage {
			c.log.V(2).Info("synced finding state", "name", syncedFinding.Name)
			syncedFindingNames[syncedFinding.Name] = true // add findings even if there was an error applying the ensureStateFn
		}
		if pageToken == "" {
			break
		}
	}
	unsyncedFindingRequests := c.filterUnsyncedFindingRequests(findingRequests, syncedFindingNames)
	if len(ensureStateFnErrors) > 0 {
		return unsyncedFindingRequests, errors.Wrap(errorutils.NewAggregate(ensureStateFnErrors), "findings state sync errors")
	}
	return unsyncedFindingRequests, nil
}

// filterUnsyncedFindingRequests returns the subset of findingRequests that do not have corresponding
// finding IDs in syncedFindingNames. This subset represents findings that don't yet exist.
func (c *Client) filterUnsyncedFindingRequests(findingRequests map[string]*securitycenterpb.CreateFindingRequest, syncedFindingNames map[string]bool) []*securitycenterpb.CreateFindingRequest {
	var unsyncedFindingRequests []*securitycenterpb.CreateFindingRequest
	for _, req := range findingRequests {
		reqFindingName := fmt.Sprintf("%s/findings/%s", req.Parent, req.FindingId)
		if !syncedFindingNames[reqFindingName] {
			c.log.V(2).Info("unsyncedFindingRequest", "findingId", req.FindingId)
			unsyncedFindingRequests = append(unsyncedFindingRequests, req)
		}
	}
	return unsyncedFindingRequests
}

// mapFindingsPage applies the mapFn to each finding returned by the call to the ListFindings API method.
// Returns all findings from where the mapFn returned non-nil, and the pageToken to allow the
// caller to repeat this for the next page.
//
// The returned error is either:
//
// - `iteratorError` if there is an error iterating over the findings in the page. The caller
//   should check for this using errors.Is(err, itegatorError) and stop fetching pages if true; or
//
// - an aggregate of the errors when applying the mapFn to the findings in the page. In this case,
//   the caller can choose to stop fetching pages, or continue and collect all errors to report at
//   the end.
//
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings/list
func (c *Client) mapFindingsPage(ctx context.Context, source string, pageToken string, mapFn func(ctx context.Context, finding *securitycenterpb.Finding) (*securitycenterpb.Finding, error)) ([]*securitycenterpb.Finding, string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	req := &securitycenterpb.ListFindingsRequest{
		Parent:    source,
		PageSize:  c.pageSize,
		PageToken: pageToken,
	}
	c.log.V(2).Info("listing findings", "source", req.Parent, "pageSize", req.PageSize, "pageToken", req.PageToken)
	it := c.client.ListFindings(ctx, req)
	pageToken = it.PageInfo().Token
	var mappedFindings []*securitycenterpb.Finding
	var mapFnErrs []error
	for {
		finding, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return mappedFindings, pageToken, errorutils.NewAggregate([]error{errIterator, err})
		}
		mappedFinding, err := mapFn(ctx, finding.Finding)
		if err != nil {
			mapFnErrs = append(mapFnErrs, err)
		}
		if mappedFinding != nil {
			mappedFindings = append(mappedFindings, mappedFinding)
		}
	}
	return mappedFindings, pageToken, errorutils.NewAggregate(mapFnErrs)
}

// CreateFinding using the provided CreateFindingRequest
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings/create
func (c *Client) CreateFinding(ctx context.Context, req *securitycenterpb.CreateFindingRequest) error {
	if c.dryRun {
		c.log.Info("(dry-run) skip create finding", "findingIDToName", fmt.Sprintf("%v/findings/%v", req.Parent, req.FindingId), "constraintTemplate", req.Finding.Category, "resourceName", req.Finding.ResourceName, "constraintUri", req.Finding.ExternalUri)
		return nil
	}
	c.log.Info("create finding", "findingName", fmt.Sprintf("%v/findings/%v", req.Parent, req.FindingId), "constraintTemplate", req.Finding.Category, "resourceName", req.Finding.ResourceName, "constraintUri", req.Finding.ExternalUri)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.CreateFinding(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.AlreadyExists {
			return err
		}
		c.log.Error(err, "finding already exists", "findingIDToName", fmt.Sprintf("%v/findings/%v", req.Parent, req.FindingId), "constraintTemplate", req.Finding.Category, "resourceName", req.Finding.ResourceName, "constraintUri", req.Finding.ExternalUri)
		return nil
	}
	return nil
}

// ensureFindingState ensures the finding state matches the provided desired state.
// If the finding is already in the correct state, this is a noop.
//
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings/setState
func (c *Client) ensureFindingState(ctx context.Context, finding *securitycenterpb.Finding, active bool) (*securitycenterpb.Finding, error) {
	var newState securitycenterpb.Finding_State
	if active {
		newState = securitycenterpb.Finding_ACTIVE
	} else {
		newState = securitycenterpb.Finding_INACTIVE
	}
	if finding.State == newState {
		c.log.V(2).Info("finding already in desired state", "findingIDToName", finding.Name, "state", newState.String())
		return finding, nil
	}
	return c.setFindingState(ctx, finding, newState)
}

// setFindingState sets the finding state according to the input.
//
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources.findings/setState
func (c *Client) setFindingState(ctx context.Context, finding *securitycenterpb.Finding, newState securitycenterpb.Finding_State) (*securitycenterpb.Finding, error) {
	if c.dryRun {
		c.log.Info("(dry-run) skip set finding state", "findingIDToName", finding.Name, "state", newState.String())
		return finding, nil
	}
	now, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return nil, err
	}
	c.log.Info("set finding state", "findingIDToName", finding.Name, "state", newState.String())
	req := &securitycenterpb.SetFindingStateRequest{
		Name:      finding.Name,
		State:     newState,
		StartTime: now,
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	c.log.Info("updating finding state", "findingIDToName", finding.Name, "state", newState.String())
	return c.client.SetFindingState(ctx, req)
}
