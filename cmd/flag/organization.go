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
	"regexp"

	"github.com/spf13/pflag"
)

// OrganizationID represents the numeric Google Cloud organization ID.
// Ref: https://cloud.google.com/resource-manager/docs/creating-managing-organization#retrieving_your_organization_id
type OrganizationID struct {
	value string
}

func (o *OrganizationID) Add(flags *pflag.FlagSet) {
	flags.StringVar(&o.value, "organization", "",
		"The numeric Google Cloud organization ID, see <https://cloud.google.com/resource-manager/docs/creating-managing-organization#retrieving_your_organization_id>")
}

func (o *OrganizationID) Validate() error {
	organizationIDRegexp, err := regexp.Compile("[0-9]+")
	if err != nil {
		return err
	}
	if !organizationIDRegexp.MatchString(o.value) {
		return fmt.Errorf("invalid organization ID: [%v]", o.value)
	}
	return nil
}

func (o *OrganizationID) Value() string {
	return o.value
}
