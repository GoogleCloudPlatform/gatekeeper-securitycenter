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
	"github.com/spf13/cobra"

	"github.com/googlecloudplatform/gatekeeper-securitycenter/cmd/flag"
)

var (
	// Cmd is the sources sub-command
	Cmd = &cobra.Command{
		Use:   "sources",
		Short: "Manage Security Command Center sources and their IAM policies",
	}

	// command-line flags for sources sub-commands
	member               = &flag.Member{}                    // member for Cloud IAM policy binding
	description          = &flag.Description{}               // SCC source description
	displayName          = &flag.DisplayName{}               // SCC source display name
	googleServiceAccount = &flag.ImpersonateServiceAccount{} // Google service account to impersonate
	organizationID       = &flag.OrganizationID{}            // Google Cloud organization ID
	role                 = &flag.Role{}                      // role for Cloud IAM policy finding
	sourceLimit          = &flag.SourceLimit{}               // limit on number of SCC sources to list
	sourceName           = &flag.Source{}                    // Security Command Center source name
)

func init() {
	Cmd.AddCommand(
		addIAMPolicyBindingCmd,
		createSourceCmd,
		getIAMPolicyCmd,
		getSourceCmd,
		listSourcesCmd,
		removeIAMPolicyBindingCmd,
	)
}
