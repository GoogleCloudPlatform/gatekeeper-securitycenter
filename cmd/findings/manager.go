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
	"context"
	"os"
	"strings"

	"github.com/go-logr/zapr"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/googlecloudplatform/gatekeeper-securitycenter/cmd/flag"
)

var (
	managerFlags = flag.New(kubeconfig, interval, dryRun, source, clusterName)

	managerCmd = &cobra.Command{
		Use:   "manager",
		Short: "Start a Kubernetes controller manager",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return managerFlags.Validate()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return managerRun(cmd.Context())
		},
	}
)

func init() {
	managerFlags.AddToFlagSet(managerCmd.Flags())
}

// managerRun starts a controller manager
func managerRun(ctx context.Context) error {
	zLog, err := createZapLogger()
	if err != nil {
		return err
	}
	defer func() {
		_ = zLog.Sync()
	}()
	log := zapr.NewLogger(zLog).WithName("controller")

	return Start(ctx, log, kubeconfig.Value(), dryRun.Value(), source.Value(), clusterName.Value(), interval.Value())
}

func createZapLogger() (*zap.Logger, error) {
	if strings.ToLower(os.Getenv("DEBUG")) == "true" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
