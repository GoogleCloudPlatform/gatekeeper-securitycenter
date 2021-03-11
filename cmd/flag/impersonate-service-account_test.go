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

func TestImpersonateServiceAccount_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "empty value accepted since flag is optional",
			value:   "",
			wantErr: false,
		},
		{
			name:    "service account accepted",
			value:   "123456789012-compute@developer.gserviceaccount.com",
			wantErr: false,
		},
		{
			name:    "error on gmail account",
			value:   "test@gmail.com",
			wantErr: true,
		},
		{
			name:    "unique id accepted",
			value:   "123456789012345678901",
			wantErr: false,
		},
		{
			name:    "error on arbitrary alphanumeric string",
			value:   "abcde12345",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ImpersonateServiceAccount{
				value: tt.value,
			}
			if err := a.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
