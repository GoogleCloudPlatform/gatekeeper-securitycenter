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
	"testing"
	"time"

	logrtesting "github.com/go-logr/logr/testing"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	securitycenterpb "google.golang.org/genproto/googleapis/cloud/securitycenter/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	host    = "https://apiserver:443"
	source  = "organizations/123/sources/456"
	cluster = "my-cluster"
	client  = &Client{
		log:     logrtesting.NullLogger{},
		host:    host,
		source:  source,
		cluster: cluster,
	}

	// ignoreUnexported compares two CreateFindingRequest objects on all exported fields
	ignoreUnexported = cmpopts.IgnoreUnexported(
		securitycenterpb.CreateFindingRequest{},
		securitycenterpb.Finding{},
		structpb.Value{},
		timestamppb.Timestamp{},
	)

	// ignoreFindingID ignores the FindingId field when comparing two CreateFindingRequest objects
	ignoreFindingID = cmpopts.IgnoreFields(securitycenterpb.CreateFindingRequest{}, "FindingId")

	// resourceNameOnly compares two CreateFindingRequest objects on the Finding.ResourceName fields only
	resourceNameOnly = cmp.Comparer(func(l, r *securitycenterpb.CreateFindingRequest) bool {
		return l.Finding.ResourceName == r.Finding.ResourceName
	})
)

func TestClient_createFindingRequest(t *testing.T) {
	now := time.Now()
	nowpb, _ := ptypes.TimestampProto(now)

	tests := []struct {
		name       string
		cmpOptions []cmp.Option
		constraint *Constraint
		resource   *Resource
		want       *securitycenterpb.CreateFindingRequest
	}{
		{
			name: "create finding request for non-KCC resource violation",
			cmpOptions: []cmp.Option{
				ignoreUnexported,
				ignoreFindingID,
			},
			constraint: &Constraint{
				Name:             "constraintName",
				SelfLink:         "/constraintSelfLink",
				UID:              "constraintUID",
				Kind:             "constraintKind",
				AuditTime:        now,
				SpecJSON:         "constraintSpecJSON",
				TemplateUID:      "constraintTemplateUID",
				TemplateSelfLink: "/constraintTemplateSelfLink",
				TemplateSpecJSON: "constraintTemplateSpecJSON",
			},
			resource: &Resource{
				Name:      "resourceName",
				Namespace: "resourceNamespace",
				GVK: schema.GroupVersionKind{
					Group:   "resourceGVKGroup",
					Version: "resourceGVKVersion",
					Kind:    "resourceGVKKind",
				},
				SelfLink:       "/resourceSelfLink",
				UID:            "resourceUID",
				ProjectID:      "resourceProjectID",
				StatusSelfLink: "", // empty for non-KCC resources
				Message:        "violationMessage",
				SpecJSON:       "resourceSpecJSON",
			},
			want: &securitycenterpb.CreateFindingRequest{
				Parent: "organizations/123/sources/456",
				// FindingId: "d0cbf936dbd346c0b7a772eac241cbb",
				Finding: &securitycenterpb.Finding{
					ResourceName: "https://apiserver:443/resourceSelfLink",
					State:        securitycenterpb.Finding_ACTIVE,
					Category:     "constraintKind",
					ExternalUri:  "https://apiserver:443/constraintSelfLink",
					EventTime:    nowpb,
					SourceProperties: map[string]*structpb.Value{
						"Cluster":                    structpb.NewStringValue("my-cluster"),
						"ConstraintName":             structpb.NewStringValue("constraintName"),
						"ConstraintSelfLink":         structpb.NewStringValue("https://apiserver:443/constraintSelfLink"),
						"ConstraintTemplateSelfLink": structpb.NewStringValue("https://apiserver:443/constraintTemplateSelfLink"),
						"ConstraintTemplateUID":      structpb.NewStringValue("constraintTemplateUID"),
						"ConstraintUID":              structpb.NewStringValue("constraintUID"),
						"Explanation":                structpb.NewStringValue("violationMessage"),
						"ProjectId":                  structpb.NewStringValue("resourceProjectID"),
						"ResourceAPIGroup":           structpb.NewStringValue("resourceGVKGroup"),
						"ResourceAPIVersion":         structpb.NewStringValue("resourceGVKVersion"),
						"ResourceKind":               structpb.NewStringValue("resourceGVKKind"),
						"ResourceName":               structpb.NewStringValue("resourceName"),
						"ResourceNamespace":          structpb.NewStringValue("resourceNamespace"),
						"ResourceSelfLink":           structpb.NewStringValue("https://apiserver:443/resourceSelfLink"),
						"ResourceStatusSelfLink":     structpb.NewStringValue(""),
						"ResourceUID":                structpb.NewStringValue("resourceUID"),
						"ScannerName":                structpb.NewStringValue("GATEKEEPER"),
					},
				},
			},
		},
		{
			name: "use resourceStatusSelfLink for KCC resource violation",
			cmpOptions: []cmp.Option{
				resourceNameOnly,
			},
			constraint: &Constraint{},
			resource: &Resource{
				SelfLink:       "/doNotUseThis",
				StatusSelfLink: "https://www.googleapis.com/storage/v1/b/bucket-name", // not empty for KCC resources
			},
			want: &securitycenterpb.CreateFindingRequest{
				Finding: &securitycenterpb.Finding{
					ResourceName: "https://www.googleapis.com/storage/v1/b/bucket-name",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.createFindingRequest(tt.constraint, tt.resource)
			if diff := cmp.Diff(tt.want, got, tt.cmpOptions...); diff != "" {
				t.Errorf("createFindingRequest() (%s) mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

// func TestClient_createFindingRequestConfigConnector(t *testing.T) {
// 	constraint := &Constraint{}
// 	resource := &Resource{
// 		SelfLink:       "/doNotUseThis",
// 		StatusSelfLink: "https://www.googleapis.com/storage/v1/b/bucket-name", // not empty for KCC resources
// 	}
// 	request := client.createFindingRequest(constraint, resource)
// 	got := request.Finding.ResourceName
// 	want := resource.StatusSelfLink
// 	if got != want {
// 		t.Errorf("createFindingRequest() (KCC resource) mismatch: got: %s, want: %s", got, want)
// 	}
// }

func Test_determineFindingID(t *testing.T) {
	tests := []struct {
		name       string
		constraint *Constraint
		resource   *Resource
		want       string
	}{
		{
			name: "all inputs available",
			constraint: &Constraint{
				UID:              "constraintUID",
				SpecJSON:         "constraintSpecJSON",
				TemplateUID:      "constraintTemplateUID",
				TemplateSpecJSON: "constraintTemplateSpecJSON",
			},
			resource: &Resource{
				UID:      "resourceUID",
				SpecJSON: "resourceSpecJSON",
			},
			want: "d0cbf936dbd346c0b7a772eac241cbbd",
		},
		{
			name:       "all inputs empty",
			constraint: &Constraint{},
			resource:   &Resource{},
			want:       "e3b0c44298fc1c149afbf4c8996fb924",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineFindingID(tt.constraint, tt.resource)
			wantLength := 32
			if len(got) != wantLength {
				t.Errorf("determineFindingID() (%s) length %v, want %v", tt.name, len(got), wantLength)
			}
			if got != tt.want {
				t.Errorf("determineFindingID() (%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
