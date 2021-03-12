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
	getIAMPolicyFlags = flag.New(sourceName, googleServiceAccount)

	getIAMPolicyCmd = &cobra.Command{
		Use:   "get-iam-policy",
		Short: "Get the IAM policy for a Security Command Center source",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return getIAMPolicyFlags.Validate()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return getIAMPolicyRun(cmd.Context())
		},
	}
)

func init() {
	getIAMPolicyFlags.AddToFlagSet(getIAMPolicyCmd.Flags())
}

// getIAMPolicyRun prints the IAM policy for the provided source
func getIAMPolicyRun(ctx context.Context) error {
	log := logging.CreateStdLog("get-iam-policy")
	policy, err := getIAMPolicy(ctx, log, sourceName.Value(), googleServiceAccount.Value())
	if err != nil {
		return err
	}
	return print.AsJSON(policy)

}

func getIAMPolicy(ctx context.Context, log logr.Logger, sourceName, googleServiceAccount string) (*iampb.Policy, error) {
	dryRun := false
	securitycenterClient, err := securitycenter.NewClient(ctx, log, googleServiceAccount, dryRun)
	if err != nil {
		return nil, err
	}
	defer securitycenterClient.Close()
	return securitycenterClient.GetIamPolicy(ctx, sourceName)
}
