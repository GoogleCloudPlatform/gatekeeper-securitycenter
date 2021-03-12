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

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	iampb "google.golang.org/genproto/googleapis/iam/v1"

	"github.com/googlecloudplatform/gatekeeper-securitycenter/cmd/flag"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/logging"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/print"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/securitycenter"
)

var (
	addIAMPolicyBindingFlags = flag.New(sourceName, member, role, googleServiceAccount)

	addIAMPolicyBindingCmd = &cobra.Command{
		Use:   "add-iam-policy-binding",
		Short: "Add an IAM policy binding for a Security Command Center source",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return addIAMPolicyBindingFlags.Validate()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return addIAMPolicyBindingRun(cmd.Context())
		},
	}
)

func init() {
	addIAMPolicyBindingFlags.AddToFlagSet(addIAMPolicyBindingCmd.Flags())
}

// addIAMPolicyBindingRun adds an IAM policy binding for the provided SCC source
func addIAMPolicyBindingRun(ctx context.Context) error {
	log := logging.CreateStdLog("add-iam-policy-binding")
	newPolicy, err := addIAMPolicyBinding(ctx, log, sourceName.Value(), member.Value(), role.Value(), googleServiceAccount.Value())
	if err != nil {
		return err
	}
	return print.AsJSON(newPolicy)
}

func addIAMPolicyBinding(ctx context.Context, log logr.Logger, sourceName, member, role, googleServiceAccount string) (*iampb.Policy, error) {
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
	policy.Bindings = append(policy.Bindings, &iampb.Binding{
		Role:    role,
		Members: []string{member},
	})
	return securitycenterClient.SetIamPolicy(ctx, sourceName, policy)
}
