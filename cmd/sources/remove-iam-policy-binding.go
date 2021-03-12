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

package sources

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	iampb "google.golang.org/genproto/googleapis/iam/v1"

	"github.com/googlecloudplatform/gatekeeper-securitycenter/cmd/flag"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/logging"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/print"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/securitycenter"
)

var (
	removeIAMPolicyBindingFlags = flag.New(sourceName, member, role, googleServiceAccount)

	removeIAMPolicyBindingCmd = &cobra.Command{
		Use:   "remove-iam-policy-binding",
		Short: "Remove an IAM policy binding from a Security Command Center source",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return removeIAMPolicyBindingFlags.Validate()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return removeIAMPolicyBindingRun(cmd.Context())
		},
	}
)

func init() {
	removeIAMPolicyBindingFlags.AddToFlagSet(removeIAMPolicyBindingCmd.Flags())
}

// removeIAMPolicyBindingRun removes a binding from the IAM policy of the provided source
func removeIAMPolicyBindingRun(ctx context.Context) error {
	log := logging.CreateStdLog("remove-iam-policy-binding")
	newPolicy, err := removeIAMPolicyBinding(ctx, log, sourceName.Value(), member.Value(), role.Value(), googleServiceAccount.Value())
	if err != nil {
		return err
	}
	return print.AsJSON(newPolicy)
}

func removeIAMPolicyBinding(ctx context.Context, log logr.Logger, sourceName, member, role, googleServiceAccount string) (*iampb.Policy, error) {
	dryRun := false
	securitycenterClient, err := securitycenter.NewClient(ctx, log, googleServiceAccount, dryRun)
	if err != nil {
		return nil, err
	}
	defer securitycenterClient.Close()
	policy, err := securitycenterClient.GetIamPolicy(ctx, sourceName)
	if err != nil {
		return nil, err
	}
	var foundBinding bool
	for _, binding := range policy.Bindings {
		if binding.Role == role {
			var updatedMembers []string
			for _, existingMember := range binding.Members {
				if existingMember == member {
					foundBinding = true
				} else {
					updatedMembers = append(updatedMembers, existingMember)
				}
			}
			binding.Members = updatedMembers
		}
	}
	if !foundBinding {
		return nil, fmt.Errorf("could not find IAM policy binding for source=<%v> role=<%v> member=<%v>", sourceName, role, member)
	}
	return securitycenterClient.SetIamPolicy(ctx, sourceName, policy)
}
