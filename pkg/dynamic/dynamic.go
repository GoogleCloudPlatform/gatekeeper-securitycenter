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

package dynamic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// Ref: https://pkg.go.dev/k8s.io/client-go/dynamic

const (
	defaultTimeout                  = 60 * time.Second
	gatekeeperConstraintsAPIVersion = "v1beta1"
)

var (
	gatekeeperConstraintTemplateGVR = schema.GroupVersionResource{
		Group:    "templates.gatekeeper.sh",
		Version:  "v1beta1",
		Resource: "constrainttemplates",
	}
)

// Client is a dynamic.Interface wrapper
type Client struct {
	dynamic dynamic.Interface
	log     logr.Logger
	timeout time.Duration
}

// NewClient creates a new dynamic client
func NewClient(log logr.Logger, config *rest.Config) (*Client, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		dynamic: dynamicClient,
		log:     log,
		timeout: defaultTimeout,
	}, nil
}

// GetViolatedConstraints find all constraints that have violations.
// To remove duplicates, it returns a map of constraint UID to constraint.
func (c *Client) GetViolatedConstraints(ctx context.Context, groupResources []schema.GroupResource) (map[types.UID]*unstructured.Unstructured, error) {
	// using a map with constraint UID as the key to avoid duplicated constraints
	violatedConstraints := map[types.UID]*unstructured.Unstructured{}
	for _, groupResource := range groupResources {
		groupVersionResource := groupResource.WithVersion(gatekeeperConstraintsAPIVersion)
		constraints, err := c.listResources(ctx, groupVersionResource)
		if err != nil {
			return nil, err
		}
		for _, constraint := range constraints.Items {
			if _, exists, err := unstructured.NestedSlice(constraint.UnstructuredContent(), "status", "violations"); exists && err == nil {
				violatedConstraints[constraint.GetUID()] = &constraint
			}
		}
	}
	c.log.V(1).Info("Constraint violations", "count", len(violatedConstraints))
	return violatedConstraints, nil
}

// GetResourceByKind returns a resource by trying all the kind-to-GVR mappings
// See explanation in kind.go
func (c *Client) GetResourceByKind(ctx context.Context, kind, name, namespace string, kindToGVR map[string][]schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	possibleGVRs := kindToGVR[kind]
	var r *unstructured.Unstructured
	var err error
	for _, gvr := range possibleGVRs {
		r, err = c.getResource(ctx, gvr, name, namespace)
		if err != nil {
			break // found a match
		}
	}
	if r == nil {
		return nil, fmt.Errorf("could not find resource with kind=[%v] name=[%v] in namespace=[%v] using any of these GroupVersionResource mappings=%+v", kind, name, namespace, possibleGVRs)
	}
	return r, nil
}

// GetConstraintTemplate returns the constraint template for the provided constraint Kind
func (c *Client) GetConstraintTemplate(ctx context.Context, constraintKind string) (*unstructured.Unstructured, error) {
	constraintTemplateName := strings.ToLower(constraintKind)
	c.log.V(2).Info("getting constraint template", "constraintTemplateName", constraintTemplateName)
	return c.dynamic.Resource(gatekeeperConstraintTemplateGVR).Get(ctx, constraintTemplateName, metav1.GetOptions{})
}

// getResource for the provided GVR
func (c *Client) getResource(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string) (*unstructured.Unstructured, error) {
	c.log.V(2).Info("getting resource", "name", name, "namespace", namespace, "apiGroup", gvr.Group, "apiVersion", gvr.Version, "resourceType", gvr.Resource)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.dynamic.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})

}

// listResources lists resources for the provided GVR
func (c *Client) listResources(ctx context.Context, gvr schema.GroupVersionResource) (*unstructured.UnstructuredList, error) {
	c.log.V(2).Info("listing resources", "apiGroup", gvr.Group, "apiVersion", gvr.Version, "resourceType", gvr.Resource)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
}

// SetTimeout for calls to the API server
func (c *Client) SetTimeout(timeout time.Duration) error {
	if timeout.Seconds() <= 0 {
		return fmt.Errorf("invalid timeout: %v", timeout)
	}
	c.timeout = timeout
	return nil
}
