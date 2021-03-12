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
	"time"

	"github.com/go-logr/logr"

	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/sync"
)

// Start the control loop
// The control loop retrieves Gatekeeper audit constraint violations and
// creates a finding in Security Command Center for each violation.
func Start(ctx context.Context, log logr.Logger, kubeconfig string, dryRun bool, sourceName, clusterName string, intervalSeconds int) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	client, err := sync.NewClient(ctx, log, kubeconfig, dryRun, sourceName, clusterName, "")
	if err != nil {
		return err
	}
	defer client.Close()
	log.Info("Starting control loop")
	for {
		if err := client.Sync(ctx); err != nil {
			log.Error(err, "sync failed")
		}
		select {
		case <-ctx.Done():
			log.Info("Stopping control loop")
			return nil
		case <-time.After(time.Duration(intervalSeconds) * time.Second):
			continue
		}
	}
}
