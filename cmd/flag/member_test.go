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

package flag

import (
	"testing"
)

func TestMember_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		// member is mandatory
		{name: "invalid empty value", value: "", wantErr: true},

		// https://cloud.google.com/iam/docs/reference/rest/v1/Policy
		{name: "invalid valid user", value: "foo@example.com", wantErr: true},
		{name: "valid user", value: "user:foo@example.com", wantErr: false},
		{name: "invalid deleted user", value: "deleted:foo@example.com?uid=123456789012345678901", wantErr: true},
		{name: "invalid deleted user without uid", value: "deleted:user:foo@example.com", wantErr: true},
		{name: "valid deleted user", value: "deleted:user:foo@example.com?uid=123456789012345678901", wantErr: false},
		{name: "invalid group", value: "bar@example.com", wantErr: true},
		{name: "valid group", value: "group:bar@example.com", wantErr: false},
		{name: "invalid deleted group", value: "deleted:group:bar@example.com", wantErr: true},
		{name: "valid deleted group", value: "deleted:group:bar@example.com?uid=123456789012345678901", wantErr: false},
		{name: "invalid serviceAccount", value: "123456789012-compute@developer.gserviceaccount.com", wantErr: true},
		{name: "valid serviceAccount", value: "serviceAccount:123456789012-compute@developer.gserviceaccount.com", wantErr: false},
		{name: "invalid deleted serviceAccount", value: "deleted:serviceAccount:123456789012-compute@developer.gserviceaccount.com", wantErr: true},
		{name: "valid deleted serviceAccount", value: "deleted:serviceAccount:123456789012-compute@developer.gserviceaccount.com?uid=123456789012345678901", wantErr: false},
		{name: "invalid domain", value: "example.com", wantErr: true},
		{name: "valid domain", value: "domain:example.com", wantErr: false},

		// https://cloud.google.com/iam/docs/workload-identity-federation
		{name: "invalid principal", value: "iam.googleapis.com/projects/123456789012/locations/global/workloadIdentityPools/pool-id1/subject/subject:name/more", wantErr: true},
		{name: "valid principal", value: "principal://iam.googleapis.com/projects/123456789012/locations/global/workloadIdentityPools/pool-id1/subject/subject:name/more", wantErr: false},
		{name: "invalid deleted principal", value: "deleted:principal://iam.googleapis.com/projects/123456789012/locations/global/workloadIdentityPools/pool-id1/subject/subject:name/more", wantErr: true},
		{name: "valid deleted principal", value: "deleted:principal://iam.googleapis.com/projects/123456789012/locations/global/workloadIdentityPools/pool-id1/subject/subject:name/more?uid=12345", wantErr: false},
		{name: "invalid principalSet", value: "iam.googleapis.com/projects/123456789012/locations/global/workloadIdentityPools/pool-id1/group/group-name", wantErr: true},
		{name: "valid principalSet", value: "principalSet://iam.googleapis.com/projects/123456789012/locations/global/workloadIdentityPools/pool-id1/group/group-name", wantErr: false},
		{name: "invalid deleted principalSet", value: "deleted:principalSet://iam.googleapis.com/projects/123456789012/locations/global/workloadIdentityPools/pool-id1/group/group-name", wantErr: true},
		{name: "valid deketed principalSet", value: "principalSet://iam.googleapis.com/projects/123456789012/locations/global/workloadIdentityPools/pool-id1/group/group-name?uid=12345", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Member{
				value: tt.value,
			}
			if err := m.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
