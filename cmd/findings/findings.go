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

package findings

import (
	"github.com/spf13/cobra"

	"github.com/googlecloudplatform/gatekeeper-securitycenter/cmd/flag"
)

var (
	// Cmd is the findings sub-command
	Cmd = &cobra.Command{
		Use:   "findings",
		Short: "Synchronize Gatekeeper audit violations to Security Command Center findings",
	}

	// command-line flags for findings sub-commands
	clusterName          = &flag.Cluster{}                   // cluster identifier, optional
	dryRun               = &flag.DryRun{}                    // skip state-changing operations
	googleServiceAccount = &flag.ImpersonateServiceAccount{} // Google service account to impersonate
	interval             = &flag.Interval{}                  // time in seconds between interations of the control loop
	kubeconfig           = &flag.Kubeconfig{}                // path to kubeconfig, or empty to use in-cluster config
	source               = &flag.Source{}                    // Security Command Center source name
)

func init() {
	Cmd.AddCommand(
		managerCmd,
		syncCmd,
	)
}
