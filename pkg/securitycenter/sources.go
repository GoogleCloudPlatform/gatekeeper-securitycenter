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
	"strings"

	"google.golang.org/api/iterator"
	securitycenterpb "google.golang.org/genproto/googleapis/cloud/securitycenter/v1"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
)

// GetSource gets a source by its full name in the format `organizations/[organization_id]/sources/[source_id]`
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources/get
func (c *Client) GetSource(ctx context.Context, source string) (*securitycenterpb.Source, error) {
	req := &securitycenterpb.GetSourceRequest{
		Name: source,
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.client.GetSource(ctx, req)
}

// GetSourceNameForDisplayName can be used to check if a source with the same display
// name already exists for the provided organization (case insensitive match).
// Returns the full source name of the existing source with the provided display name.
// If no source exists for the provided display name, this method returns the empty string and nil error.
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources
func (c *Client) GetSourceNameForDisplayName(ctx context.Context, organizationID, displayName string) (string, error) {
	req := &securitycenterpb.ListSourcesRequest{
		Parent: fmt.Sprintf("organizations/%s", organizationID),
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	it := c.client.ListSources(ctx, req)
	for {
		source, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", fmt.Errorf("it.Next error when finding source by display name: %w", err)
		}
		if strings.EqualFold(displayName, source.DisplayName) {
			return source.Name, nil
		}
	}
	return "", nil
}

// ListSources retrieves all sources for the provided numeric organization ID.
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources/list
func (c *Client) ListSources(ctx context.Context, organizationID string) ([]*securitycenterpb.Source, error) {
	req := &securitycenterpb.ListSourcesRequest{
		Parent:   fmt.Sprintf("organizations/%s", organizationID),
		PageSize: c.pageSize,
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	it := c.client.ListSources(ctx, req)
	var sources []*securitycenterpb.Source
	for {
		source, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("it.Next error when listing sources: %w", err)
		}
		sources = append(sources, source)
	}
	return sources, nil
}

// CreateSource creates a source. Returns an error if a source exists for the
// organization with the same displayName.
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources/create
func (c *Client) CreateSource(ctx context.Context, organizationID, displayName, description string) (*securitycenterpb.Source, error) {
	existingSource, err := c.GetSourceNameForDisplayName(ctx, organizationID, displayName)
	if err != nil {
		return nil, err
	}
	if existingSource != "" {
		return nil, fmt.Errorf("source already exists for organizationID=<%v> with displayName=<%v>", organizationID, displayName)
	}
	source := &securitycenterpb.Source{
		DisplayName: displayName,
		Description: description,
	}
	if c.dryRun {
		c.log.Info("(dry-run) creating source", "displayName", source.DisplayName)
		return source, nil
	}
	req := &securitycenterpb.CreateSourceRequest{
		Parent: fmt.Sprintf("organizations/%s", organizationID),
		Source: source,
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.client.CreateSource(ctx, req)
}

// GetIamPolicy for the provided source.
//
// The `source` input argument should be in the format
// `organizations/[organization_id]/sources/[source_id]`
//
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources/getIamPolicy
func (c *Client) GetIamPolicy(ctx context.Context, source string) (*iampb.Policy, error) {
	req := &iampb.GetIamPolicyRequest{
		Options: &iampb.GetPolicyOptions{
			RequestedPolicyVersion: 3,
		},
		Resource: source,
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.client.GetIamPolicy(ctx, req)
}

// SetIamPolicy for the provided source using the provided policy
//
// The `source` input argument should be in the format
// `organizations/[organization_id]/sources/[source_id]`
//
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources/setIamPolicy
func (c *Client) SetIamPolicy(ctx context.Context, source string, policy *iampb.Policy) (*iampb.Policy, error) {
	if c.dryRun {
		c.log.Info("(dry-run) SetIamPolicy", "source", source)
		return policy, nil
	}
	req := &iampb.SetIamPolicyRequest{
		Policy:   policy,
		Resource: source,
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.client.SetIamPolicy(ctx, req)
}
