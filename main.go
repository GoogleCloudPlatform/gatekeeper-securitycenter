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

package main

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/googlecloudplatform/gatekeeper-securitycenter/cmd/findings"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/cmd/sources"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/cmd/version"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/signals"

	// blank import for all k8s auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gatekeeper-security",
	Short: "Creates Security Command Center findings from OPA Gatekeeper and Policy Controller violations",
	Long: `gatekeeper-securitycenter is:

- a Kubernetes controller that creates Security Command Center
  findings for violations reported by OPA Gatekeeper/Policy Controller's
  audit controller

- a command-line tool that creates and manages the IAM policies of
  Security Command Center sources`,
}

func init() {
	rootCmd.AddCommand(
		findings.Cmd,
		sources.Cmd,
		version.Cmd,
	)
}

func main() {
	ctx := signals.SetupSignalHandler(context.Background())
	_ = rootCmd.ExecuteContext(ctx)
}
