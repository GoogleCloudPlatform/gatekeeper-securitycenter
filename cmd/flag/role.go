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
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// Role a Cloud IAM role used in an IAM policy binding
// Ref: https://cloud.google.com/iam/docs/understanding-roles
type Role struct {
	value string
}

func (r *Role) Add(flags *pflag.FlagSet) {
	flags.StringVar(&r.value, "role", "",
		"The role of the member, e.g., `roles/securitycenter.findingsEditor`")
}

func (r *Role) Validate() error {
	if !strings.Contains(r.value, "roles/") {
		return fmt.Errorf("invalid role: [%v]", r.value)
	}
	return nil
}

func (r *Role) Value() string {
	return r.value
}
