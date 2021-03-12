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
	securitycenterpb "google.golang.org/genproto/googleapis/cloud/securitycenter/v1"

	"github.com/googlecloudplatform/gatekeeper-securitycenter/cmd/flag"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/logging"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/print"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/securitycenter"
)

var (
	createSourceFlags = flag.New(organizationID, displayName, description, googleServiceAccount)

	createSourceCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a Security Command Center source",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return createSourceFlags.Validate()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return createSourceRun(cmd.Context())
		},
	}
)

func init() {
	createSourceFlags.AddToFlagSet(createSourceCmd.Flags())
}

// createSourceRun creates a Security Command Center source
func createSourceRun(ctx context.Context) error {
	log := logging.CreateStdLog("create")
	source, err := createSource(ctx, log, organizationID.Value(), displayName.Value(), description.Value(), googleServiceAccount.Value())
	if err != nil {
		return err
	}
	return print.AsJSON(source)
}

func createSource(ctx context.Context, log logr.Logger, organizationID, displayName, description, googleServiceAccount string) (*securitycenterpb.Source, error) {
	dryRun := false
	securitycenterClient, err := securitycenter.NewClient(ctx, log, googleServiceAccount, dryRun)
	if err != nil {
		return nil, err
	}
	defer securitycenterClient.Close()
	return securitycenterClient.CreateSource(ctx, organizationID, displayName, description)
}
