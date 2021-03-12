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

package securitycenter

import (
	"context"
	"fmt"
	"time"

	securitycenterv1 "cloud.google.com/go/securitycenter/apiv1"
	"github.com/go-logr/logr"
	"google.golang.org/api/option"

	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/logging"
)

const (
	defaultTimeout  = 60 * time.Second
	defaultPageSize = 1000
)

// Client for the Security Command Center API v1. Wraps the googleapis client.
type Client struct {
	client   *securitycenterv1.Client
	timeout  time.Duration
	pageSize int32
	log      logr.Logger
	dryRun   bool
}

// Close cleans up
func (c *Client) Close() error {
	return c.client.Close()
}

// NewClient creates a Client for the Security Command Center API v1.
// Remember to `defer Close()` to clean up.
//
// Note: All methods creates child contextx (with timeouts) from the provided context.
//
// Optional arguments:
// - a logr.Logger. Default to stdr, which is a wrapper of Go's system log package.
// - a Google Service Account to impersonate. Defaults to no impersonation for empty string.
// - ClientOptions from the google.golang.org/api/option package
func NewClient(ctx context.Context, log logr.Logger, googleServiceAccount string, dryRun bool, opts ...option.ClientOption) (*Client, error) {
	if log == nil {
		log = logging.CreateStdLog("securitycenter")
	}
	if dryRun {
		log.Info("enabling dry-run mode")
	}
	if googleServiceAccount != "" {
		log.V(1).Info("impersonating Google service account", "serviceAccount", googleServiceAccount)
		opts = append(opts,
			option.ImpersonateCredentials(googleServiceAccount),
			option.WithScopes(securitycenterv1.DefaultAuthScopes()...))
	}
	securitycenterClient, err := securitycenterv1.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("could not create securitycenter client: %w", err)
	}
	return &Client{
		client:   securitycenterClient,
		timeout:  defaultTimeout,
		log:      log,
		pageSize: defaultPageSize,
		dryRun:   dryRun,
	}, nil
}

// SetTimeout for calls to the Security Center API
func (c *Client) SetTimeout(timeout time.Duration) error {
	if timeout.Seconds() <= 0 {
		return fmt.Errorf("Invalid timeout: %v", timeout)
	}
	c.timeout = timeout
	return nil
}

// SetPageSize for list calls to the Security Center API
func (c *Client) SetPageSize(pageSize int) error {
	if pageSize < 1 || pageSize > 1000 {
		return fmt.Errorf("Invalid pageSize: %v", pageSize)
	}
	c.pageSize = int32(pageSize)
	return nil
}
