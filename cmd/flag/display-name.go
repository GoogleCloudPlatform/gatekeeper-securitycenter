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

	"github.com/spf13/pflag"
)

const defaultDisplayName = "Gatekeeper"

// DisplayName of a Security Command Center source.
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources
type DisplayName struct {
	value string
}

func (d *DisplayName) Add(flags *pflag.FlagSet) {
	flags.StringVar(&d.value, "display-name", defaultDisplayName,
		"(optional) display name of the Security Command Center source")
}

func (d *DisplayName) Validate() error {
	if len(d.value) < 1 || len(d.value) > 64 {
		return fmt.Errorf("invalid display name: [%v], must be between 1 and 64 characters", d.value)
	}
	return nil
}

func (d *DisplayName) Value() string {
	return d.value
}
