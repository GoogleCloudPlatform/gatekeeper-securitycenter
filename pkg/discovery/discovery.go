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

package discovery

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

// Ref: https://pkg.go.dev/k8s.io/client-go/discovery

const (
	defaultTimeout               = 60 * time.Second
	gatekeeperConstraintCategory = "constraint"
)

// Client is a wrapper for discovery.DiscoveryClient
type Client struct {
	discovery discovery.DiscoveryInterface
	log       logr.Logger
	timeout   time.Duration
}

// NewClient creates a new discovery client
func NewClient(log logr.Logger, config *rest.Config) (*Client, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		discovery: discoveryClient,
		log:       log,
		timeout:   defaultTimeout,
	}, nil
}

// GetConstraintGroupResources returns constraint types by category
func (c *Client) GetConstraintGroupResources() ([]schema.GroupResource, error) {
	c.log.V(2).Info("discovering resources types for category", "category", gatekeeperConstraintCategory)
	categoryExpander := restmapper.NewDiscoveryCategoryExpander(c.discovery)
	groupResources, exists := categoryExpander.Expand(gatekeeperConstraintCategory)
	if !exists {
		c.log.Info("could not find constraint resource types", "category", gatekeeperConstraintCategory)
	}
	return groupResources, nil
}

// CreateKindToGVRMap builds a mapping from kind to a slice of possible
// GroupVersionResources based on API resources from the discovery.DiscoveryClient.
// It does not map subresources (e.g., pods/status).
// This is a hack to work around the limited data available in constraint
// violations (kind only, no API group or version).
// Remove this as soon as a release containing this PR is available:
// https://github.com/open-policy-agent/gatekeeper/pull/855
func (c *Client) CreateKindToGVRMap() (map[string][]schema.GroupVersionResource, error) {
	c.log.V(1).Info("creating Kind to GroupVersionResource mappings")
	kindToGVR := map[string][]schema.GroupVersionResource{}
	apiGroupResources, err := restmapper.GetAPIGroupResources(c.discovery)
	if err != nil {
		return nil, err
	}
	for _, apiGroupResource := range apiGroupResources {
		group := apiGroupResource.Group.Name
		for version, versionedResource := range apiGroupResource.VersionedResources {
			for _, apiResource := range versionedResource {
				if !strings.Contains(apiResource.Name, "/") {
					// conditional to skip subresources, such as `pods/status`
					gvr := schema.GroupVersionResource{
						Group:    group,
						Version:  version,
						Resource: apiResource.Name,
					}
					kindToGVR[apiResource.Kind] = append(kindToGVR[apiResource.Kind], gvr)
				}
			}
		}
	}
	if c.log.V(2).Enabled() {
		for kind, gvrs := range kindToGVR {
			c.log.V(2).Info("kindToGVRMapping", "kind", kind, "GVRs", gvrs)
		}
	}
	return kindToGVR, nil
}

// SetTimeout for calls to the API server
func (c *Client) SetTimeout(timeout time.Duration) error {
	if timeout.Seconds() < 0 {
		return fmt.Errorf("Invalid timeout: %v", timeout)
	}
	c.timeout = timeout
	return nil
}
