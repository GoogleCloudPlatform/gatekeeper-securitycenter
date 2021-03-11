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

package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/genproto/googleapis/cloud/securitycenter/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

const scannerName = "GATEKEEPER"

// Resource holds the resource-related values used to create a finding request
type Resource struct {
	Name           string
	Namespace      string
	GVK            schema.GroupVersionKind
	SelfLink       string
	UID            types.UID
	ProjectID      string
	StatusSelfLink string
	Message        string
	SpecJSON       string
}

// Constraint holds the constraint-related values used to create a finding request
type Constraint struct {
	Name             string
	SelfLink         string
	UID              types.UID
	Kind             string
	AuditTime        time.Time
	SpecJSON         string
	TemplateUID      types.UID
	TemplateSelfLink string
	TemplateSpecJSON string
}

// createFindingRequest creates a CreateFindingRequest
func (c *Client) createFindingRequest(constraint *Constraint, resource *Resource) *securitycenter.CreateFindingRequest {
	// Add API server host to all Kubernetes object links and limit to max 255 characters
	resourceSelfLink := fmt.Sprintf("%.255s", fmt.Sprintf("%v%v", c.host, resource.SelfLink))
	constraintSelfLink := fmt.Sprintf("%.255s", fmt.Sprintf("%v%v", c.host, constraint.SelfLink))
	constraintTemplateSelfLink := fmt.Sprintf("%.255s", fmt.Sprintf("%v%v", c.host, constraint.TemplateSelfLink))

	// Config Connector resources have a status.selfLink attribute pointing to the
	// actual Google Cloud resource (not the Kubernetes resource). Cap at 255 chars
	resourceStatusSelfLink := fmt.Sprintf("%.255s", resource.StatusSelfLink)

	// Limit message to 255 chars
	message := fmt.Sprintf("%.255s", resource.Message)

	// Use the status.selfLink attribute as the ResourceName if available.
	// This gives users a direct link from SCC to the resource in the web Cloud Console.
	resourceName := resourceStatusSelfLink
	if resourceName == "" {
		resourceName = resourceSelfLink
	}
	auditTime := constraint.AuditTime
	eventTime, err := ptypes.TimestampProto(auditTime)
	if err != nil {
		c.log.Error(err, "could not convert auditTime to Timestamp proto, using current time instead", "auditTime", auditTime)
		eventTime, err = ptypes.TimestampProto(time.Now())
		if err != nil {
			c.log.Error(err, "could not convert time.Now() to Timestamp proto")
		}
	}
	ID := determineFindingID(constraint, resource)
	return &securitycenter.CreateFindingRequest{
		Parent:    c.source,
		FindingId: ID,
		Finding: &securitycenter.Finding{
			State:        securitycenter.Finding_ACTIVE,
			ResourceName: resourceName,
			Category:     constraint.Kind,
			EventTime:    eventTime,
			ExternalUri:  constraintSelfLink,
			SourceProperties: map[string]*structpb.Value{
				// each source property value must be max 255 chars
				"ScannerName":                {Kind: &structpb.Value_StringValue{StringValue: scannerName}},
				"Explanation":                {Kind: &structpb.Value_StringValue{StringValue: message}},
				"Cluster":                    {Kind: &structpb.Value_StringValue{StringValue: c.cluster}},
				"ConstraintName":             {Kind: &structpb.Value_StringValue{StringValue: constraint.Name}},
				"ConstraintSelfLink":         {Kind: &structpb.Value_StringValue{StringValue: constraintSelfLink}},
				"ConstraintUID":              {Kind: &structpb.Value_StringValue{StringValue: string(constraint.UID)}},
				"ConstraintTemplateSelfLink": {Kind: &structpb.Value_StringValue{StringValue: constraintTemplateSelfLink}},
				"ConstraintTemplateUID":      {Kind: &structpb.Value_StringValue{StringValue: string(constraint.TemplateUID)}},
				"ProjectId":                  {Kind: &structpb.Value_StringValue{StringValue: resource.ProjectID}},
				"ResourceName":               {Kind: &structpb.Value_StringValue{StringValue: resource.Name}},
				"ResourceNamespace":          {Kind: &structpb.Value_StringValue{StringValue: resource.Namespace}},
				"ResourceSelfLink":           {Kind: &structpb.Value_StringValue{StringValue: resourceSelfLink}},
				"ResourceStatusSelfLink":     {Kind: &structpb.Value_StringValue{StringValue: resourceStatusSelfLink}},
				"ResourceUID":                {Kind: &structpb.Value_StringValue{StringValue: string(resource.UID)}},
				"ResourceAPIGroup":           {Kind: &structpb.Value_StringValue{StringValue: resource.GVK.Group}},
				"ResourceAPIVersion":         {Kind: &structpb.Value_StringValue{StringValue: resource.GVK.Version}},
				"ResourceKind":               {Kind: &structpb.Value_StringValue{StringValue: resource.GVK.Kind}},
			},
		},
	}
}

// determineFindingID creates a deterministic finding ID
//
// Inputs:
// - constraint UID
// - constraint spec as JSON string
// - constraint template UID
// - constraint template spec as JSON string
// - resource UID
// - resource spec as JSON string
//
// Related: Forseti implementation (later truncated to 32 chars):
// https://github.com/forseti-security/forseti-security/blob/2644311a2d32113f09915061cb68bd8e1d996821/google/cloud/forseti/services/scanner/dao.py#L367
func determineFindingID(c *Constraint, r *Resource) string {
	uidSha := sha256.Sum256([]byte(fmt.Sprintf("%s%s%s%s%s%s", c.UID, c.SpecJSON, c.TemplateUID, c.TemplateSpecJSON, r.UID, r.SpecJSON)))
	return hex.EncodeToString(uidSha[:])[:32]
}
