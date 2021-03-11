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

// Source represents a Security Command Center source
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/v1/organizations.sources
type Source struct {
	value string
}

func (s *Source) Add(flags *pflag.FlagSet) {
	flags.StringVar(&s.value, "source", "",
		"full name of the Security Command Center source in the format `organizations/[organization_id]/sources/[source_id]`")
}

func (s *Source) Validate() error {
	sourceNameRegexp, err := regexp.Compile("organizations/[0-9]+/sources/[0-9]+")
	if err != nil {
		return err
	}
	if !sourceNameRegexp.MatchString(s.value) {
		return fmt.Errorf("invalid source name: [%v]", s.value)
	}
	return nil
}

func (s *Source) Value() string {
	return s.value
}
