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
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

// ImpersonateServiceAccount represents a Google service account to
// impersonate when making requests to Google Cloud APIs. Only relevant for
// CLI usage. The controller should use Workload Identity to bind the
// Kubernetes service account to a Google service account instead of using
// this flag.
// Ref: https://cloud.google.com/iam/docs/impersonating-service-accounts
type ImpersonateServiceAccount struct {
	value string
}

func (a *ImpersonateServiceAccount) Add(flags *pflag.FlagSet) {
	flags.StringVar(&a.value, "impersonate-service-account", "",
		"(optional) Google service account to impersonate")
}

func (a *ImpersonateServiceAccount) Validate() error {
	if a.value == "" {
		return nil
	}
	if strings.HasSuffix(a.value, ".gserviceaccount.com") {
		return nil
	}
	if _, err := strconv.ParseFloat(a.value, 64); err == nil {
		return nil
	}
	return fmt.Errorf("invalid Google service account: [%v]", a.value)
}

func (a *ImpersonateServiceAccount) Value() string {
	return a.value
}
