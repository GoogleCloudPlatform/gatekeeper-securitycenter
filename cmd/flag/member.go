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
	"strings"

	"github.com/spf13/pflag"
)

var (
	validPrefixes = []string{
		"user:",
		"group:",
		"serviceAccount:",
		"domain:",
		"principal://",
		"principalSet://",
	}

	deletedSuffixRegexp = regexp.MustCompile(`\?uid=[0-9]+$`)
)

// Member represents a user, group, serviceAccount or domain used in an IAM policy binding
// Ref: https://cloud.google.com/security-command-center/docs/reference/rest/Shared.Types/Policy
type Member struct {
	value string
}

func (m *Member) Add(flags *pflag.FlagSet) {
	flags.StringVar(&m.value, "member", "",
		"The member of the IAM policy binding. Should be of the form `user|group|serviceAccount:email`, `domain:domain`, `principal://...` or `principalSet://...`")
}

func (m *Member) Validate() error {
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(m.value, prefix) {
			return nil
		}
		if strings.HasPrefix(m.value, "deleted:"+prefix) && deletedSuffixRegexp.MatchString(m.value) {
			return nil
		}
	}
	return fmt.Errorf("invalid member: [%v]", m.value)
}

func (m *Member) Value() string {
	return m.value
}
