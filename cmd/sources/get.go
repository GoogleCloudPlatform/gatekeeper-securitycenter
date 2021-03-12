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

	"github.com/googlecloudplatform/gatekeeper-securitycenter/cmd/flag"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/logging"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/print"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/securitycenter"
)

var (
	getSourceFlags = flag.New(sourceName, googleServiceAccount)

	getSourceCmd = &cobra.Command{
		Use:   "get",
		Short: "Get a Security Command Center source",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return getSourceFlags.Validate()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return getSourceRun(cmd.Context())
		},
	}
)

func init() {
	getSourceFlags.AddToFlagSet(getSourceCmd.Flags())
}

// getSourceRun prints an existing Security Command Center source
func getSourceRun(ctx context.Context) error {
	err := getIAMPolicyFlags.Validate()
	if err != nil {
		return err
	}

	log := logging.CreateStdLog("get")
	return getSource(ctx, log, sourceName.Value(), googleServiceAccount.Value())
}

func getSource(ctx context.Context, log logr.Logger, sourceName, googleServiceAccount string) error {
	dryRun := false
	securitycenterClient, err := securitycenter.NewClient(ctx, log, googleServiceAccount, dryRun)
	if err != nil {
		return err
	}
	defer securitycenterClient.Close()
	source, err := securitycenterClient.GetSource(ctx, sourceName)
	if err != nil {
		return err
	}
	return print.AsJSON(source)
}
