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
	listSourcesFlags = flag.New(organizationID, sourceLimit, googleServiceAccount)

	listSourcesCmd = &cobra.Command{
		Use:   "list",
		Short: "List existing Security Command Center sources",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return listSourcesFlags.Validate()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return listSourcesRun(cmd.Context())
		},
	}
)

func init() {
	listSourcesFlags.AddToFlagSet(listSourcesCmd.Flags())
}

// listSourcesRun lists existing Security Command Center security sources
func listSourcesRun(ctx context.Context) error {
	log := logging.CreateStdLog("list")
	return listSources(ctx, log, organizationID.Value(), sourceLimit.Value(), googleServiceAccount.Value())
}

func listSources(ctx context.Context, log logr.Logger, organizationID string, sourceLimit int, googleServiceAccount string) error {
	dryRun := false
	securitycenterClient, err := securitycenter.NewClient(ctx, log, googleServiceAccount, dryRun)
	if err != nil {
		return err
	}
	defer securitycenterClient.Close()
	err = securitycenterClient.SetPageSize(sourceLimit)
	if err != nil {
		return err
	}
	sources, err := securitycenterClient.ListSources(ctx, organizationID)
	if err != nil {
		return err
	}
	return print.AsJSON(sources)
}
