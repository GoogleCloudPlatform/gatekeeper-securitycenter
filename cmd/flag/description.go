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

const defaultDescription = "Reports violations from Gatekeeper audits"

// Description of a Security Command Center source.
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources
type Description struct {
	value string
}

func (d *Description) Add(flags *pflag.FlagSet) {
	flags.StringVar(&d.value, "description", defaultDescription,
		"(optional) description of the Security Command Center source")
}

func (d *Description) Validate() error {
	if len(d.value) > 1024 {
		return fmt.Errorf("invalid description: [%v], must be between max 1024 characters", d.value)
	}
	return nil
}

func (d *Description) Value() string {
	return d.value
}
