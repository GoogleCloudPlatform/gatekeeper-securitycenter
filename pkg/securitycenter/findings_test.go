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
	"os"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	logrtesting "github.com/go-logr/logr/testing"
	securitycenterpb "google.golang.org/genproto/googleapis/cloud/securitycenter/v1"

	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/logging"
)

const source = "organizations/123/sources/456"

var logger logr.Logger = logrtesting.NullLogger{}

func init() {
	if strings.ToLower(os.Getenv("DEBUG")) == "true" {
		logger = logging.CreateStdLog("test")
	}
}

func findingIDToName(id string) string {
	return source + "/findings/" + id
}

func Test_SyncFindings(t *testing.T) {
	ctx := context.Background()
	client, err := NewClient(ctx, logger, "", false, clientOptionsForMockServer)
	if err != nil {
		t.Fatal(err)
	}
	findingRequests := map[string]*securitycenterpb.CreateFindingRequest{
		findingIDToName("1"): {FindingId: "1", Parent: source, Finding: &securitycenterpb.Finding{}}, // should not change
		// finding 2 intentionally skipped, implementation should set state of existing finding to inactive
		findingIDToName("3"): {FindingId: "3", Parent: source, Finding: &securitycenterpb.Finding{}}, // should become active
		findingIDToName("4"): {FindingId: "4", Parent: source, Finding: &securitycenterpb.Finding{ // new finding
			State: securitycenterpb.Finding_ACTIVE},
		},
	}

	response0ListFindings := &securitycenterpb.ListFindingsResponse{
		ListFindingsResults: []*securitycenterpb.ListFindingsResponse_ListFindingsResult{
			{Finding: &securitycenterpb.Finding{
				Name:   findingIDToName("1"),
				Parent: source,
				State:  securitycenterpb.Finding_ACTIVE,
			}},
			{Finding: &securitycenterpb.Finding{
				Name:   findingIDToName("2"),
				Parent: source,
				State:  securitycenterpb.Finding_ACTIVE,
			}},
		},
		NextPageToken: "page1",
	}
	mockSecurityCenter.resps = append(mockSecurityCenter.resps, response0ListFindings)

	response1SetFindingState := &securitycenterpb.Finding{
		Name:   findingIDToName("2"),
		Parent: source,
		State:  securitycenterpb.Finding_INACTIVE,
	}
	mockSecurityCenter.resps = append(mockSecurityCenter.resps, response1SetFindingState)

	response2ListFindings := &securitycenterpb.ListFindingsResponse{
		ListFindingsResults: []*securitycenterpb.ListFindingsResponse_ListFindingsResult{
			{Finding: &securitycenterpb.Finding{
				Name:   findingIDToName("3"),
				Parent: source,
				State:  securitycenterpb.Finding_INACTIVE,
			}},
		},
		NextPageToken: "",
	}
	mockSecurityCenter.resps = append(mockSecurityCenter.resps, response2ListFindings)

	response3SetFindingState := &securitycenterpb.Finding{
		Name:   findingIDToName("3"),
		Parent: source,
		State:  securitycenterpb.Finding_ACTIVE,
	}
	mockSecurityCenter.resps = append(mockSecurityCenter.resps, response3SetFindingState)

	response4CreateFinding := &securitycenterpb.Finding{
		Name:   findingIDToName("4"),
		Parent: source,
		State:  securitycenterpb.Finding_ACTIVE,
	}
	mockSecurityCenter.resps = append(mockSecurityCenter.resps, response4CreateFinding)

	if err = client.SyncFindings(ctx, "source", findingRequests); err != nil {
		t.Fatal(err)
	}

	if len(mockSecurityCenter.resps) > 0 {
		t.Errorf("unused responses: %+v", mockSecurityCenter.resps)
	}

	request1SetFindingState, ok := mockSecurityCenter.reqs[1].(*securitycenterpb.SetFindingStateRequest)
	if !ok {
		t.Errorf("expected type securitycenterpb.SetFindingStateRequest, got %T", mockSecurityCenter.reqs[1])
	}
	if request1SetFindingState.Name != findingIDToName("2") {
		t.Errorf("expected %s, got %s", findingIDToName("2"), request1SetFindingState.Name)
	}
	if request1SetFindingState.State != securitycenterpb.Finding_INACTIVE {
		t.Errorf("expected state %s, got %s", securitycenterpb.Finding_INACTIVE, request1SetFindingState.State)
	}

	request3SetFindingState, ok := mockSecurityCenter.reqs[3].(*securitycenterpb.SetFindingStateRequest)
	if !ok {
		t.Errorf("expected type securitycenterpb.SetFindingStateRequest, got %T", mockSecurityCenter.reqs[3])
	}
	if request3SetFindingState.Name != findingIDToName("3") {
		t.Errorf("expected %s, got %s", findingIDToName("3"), request3SetFindingState.Name)
	}
	if request3SetFindingState.State != securitycenterpb.Finding_ACTIVE {
		t.Errorf("expected state %s, got %s", securitycenterpb.Finding_ACTIVE, request3SetFindingState.State)
	}

	request4CreateFinding, ok := mockSecurityCenter.reqs[4].(*securitycenterpb.CreateFindingRequest)
	if !ok {
		t.Errorf("expected type securitycenterpb.CreateFindingRequest, got %T", mockSecurityCenter.reqs[4])
	}
	if request4CreateFinding.FindingId != "4" {
		t.Errorf("expected findingID 4, got %s", request4CreateFinding.FindingId)
	}
	if request4CreateFinding.Parent != source {
		t.Errorf("expected Parent %s, got %s", source, request4CreateFinding.Parent)
	}
	if request4CreateFinding.Finding.State != securitycenterpb.Finding_ACTIVE {
		t.Errorf("expected state %s, got %s", securitycenterpb.Finding_ACTIVE, request4CreateFinding.Finding.State)
	}
}
