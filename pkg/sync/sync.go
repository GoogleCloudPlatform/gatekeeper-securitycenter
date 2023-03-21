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

package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	securitycenterpb "google.golang.org/genproto/googleapis/cloud/securitycenter/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/discovery"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/dynamic"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/print"
	"github.com/googlecloudplatform/gatekeeper-securitycenter/pkg/securitycenter"
)

const (
	cnrmAnnotationProjectID = "cnrm.cloud.google.com/project-id"
	esAnnotationSeverity    = "gatekeeper.epidemicsound.com/severity"
)

// Client to sync audit violations to Security Command Center
type Client struct {
	log                  logr.Logger
	dryRun               bool
	securitycenterClient *securitycenter.Client
	discoveryClient      *discovery.Client
	dynamicClient        *dynamic.Client
	host                 string
	source               string
	cluster              string
}

// Close cleans up resources, use with defer
func (c *Client) Close() error {
	return c.securitycenterClient.Close()
}

// NewClient creates a Client that reads audit violations and creates findings.
// Use defer Client.Close() to clean up.
func NewClient(ctx context.Context, log logr.Logger, kubeconfig string, dryRun bool, source, clusterName, googleServiceAccount string) (*Client, error) {
	var config *rest.Config
	var err error
	if kubeconfig == "" {
		log.V(2).Info("using in-cluster config")
		config, err = rest.InClusterConfig()
	} else {
		log.V(2).Info("using kubeconfig file", "kubeconfig", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, err
	}
	discoveryClient, err := discovery.NewClient(log, config)
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewClient(log, config)
	if err != nil {
		return nil, err
	}
	securitycenterClient, err := securitycenter.NewClient(ctx, log, googleServiceAccount, dryRun)
	if err != nil {
		return nil, err
	}
	return &Client{
		log:                  log,
		dryRun:               dryRun,
		securitycenterClient: securitycenterClient,
		discoveryClient:      discoveryClient,
		dynamicClient:        dynamicClient,
		host:                 config.Host,
		source:               source,
		cluster:              clusterName,
	}, nil
}

// Sync retrieves Gatekeeper audit constraint violations and creates a
// finding in Security Command Center for each violation.
func (c *Client) Sync(ctx context.Context) error {
	groupResources, err := c.discoveryClient.GetConstraintGroupResources()
	if err != nil {
		return err
	}
	violatedConstraints, err := c.dynamicClient.GetViolatedConstraints(ctx, groupResources)
	if err != nil {
		return err
	}
	kindToGVR, err := c.discoveryClient.CreateKindToGVRMap()
	if err != nil {
		return err
	}

	// For each constraint that contains audit violations,
	// for each audit violation,
	// get the resource that caused the violation
	// and use attributes of the constraint, the violation, and the resource to create a finding request.
	findingRequests := map[string]*securitycenterpb.CreateFindingRequest{} // key is full finding name
	for _, unstructuredConstraint := range violatedConstraints {
		constraint := c.getConstraint(ctx, unstructuredConstraint)
		resources := c.getViolatingResourcesForConstraint(ctx, unstructuredConstraint, kindToGVR)
		for _, resource := range resources {
			req := c.createFindingRequest(constraint, resource)
			findingName := fmt.Sprintf("%s/findings/%s", req.Parent, req.FindingId)
			findingRequests[findingName] = req
		}
	}

	if c.dryRun {
		return printFindingRequests(findingRequests)
	}
	if err := c.securitycenterClient.SyncFindings(ctx, c.source, findingRequests); err != nil {
		return fmt.Errorf("could not sync findings: %w", err)
	}
	return nil
}

// getConstraint creates a Constraint struct from an unstructured constraint.
// It's intentionally forgiving of errors and defaults to empty string values
// for fields that aren't required to create a finding.
func (c *Client) getConstraint(ctx context.Context, constraint *unstructured.Unstructured) *Constraint {
	name := constraint.GetName()
	selfLink := constraint.GetSelfLink()
	uid := constraint.GetUID()
	constraintKind := constraint.GetKind()
	specJSON, err := getSpecAsJSON(constraint)
	if err != nil {
		c.log.Error(err, "could not get constraint spec as JSON string")
	}
	var templateUID types.UID
	var templateSelfLink string
	var templateSpecJSON string
	template, err := c.dynamicClient.GetConstraintTemplate(ctx, constraintKind)
	if err != nil {
		c.log.Error(err, "could not get constraint template", "constraintKind", constraintKind)
	} else {
		templateUID = template.GetUID()
		templateSelfLink = template.GetSelfLink()
		templateSpecJSON, err = getSpecAsJSON(template)
		if err != nil {
			c.log.Error(err, "could not get constraint template spec as JSON string")
		}
	}
	auditTimestamp, _, _ := unstructured.NestedString(constraint.UnstructuredContent(), "status", "auditTimestamp")
	auditTime, err := time.Parse(time.RFC3339, auditTimestamp)
	if err != nil {
		c.log.Error(err, "could not parse auditTimestamp, using current time instead", "auditTimestamp", auditTimestamp)
		auditTime = time.Now()
	}
	return &Constraint{
		Name:             name,
		SelfLink:         selfLink,
		UID:              uid,
		Kind:             constraintKind,
		AuditTime:        auditTime,
		SpecJSON:         specJSON,
		TemplateUID:      templateUID,
		TemplateSelfLink: templateSelfLink,
		TemplateSpecJSON: templateSpecJSON,
	}
}

func (c *Client) getViolatingResourcesForConstraint(ctx context.Context, constraint *unstructured.Unstructured, kindToGVR map[string][]schema.GroupVersionResource) []*Resource {
	violations := getViolationsForConstraint(c.log, constraint)
	var resources []*Resource
	for _, violation := range violations {
		resource, err := c.getResource(ctx, violation, kindToGVR)
		if err != nil {
			c.log.Error(err, "skipping violation")
		} else {
			resources = append(resources, resource)
		}
	}
	return resources
}

func getViolationsForConstraint(log logr.Logger, constraint *unstructured.Unstructured) []map[string]interface{} {
	var violations []map[string]interface{}
	rawViolations, exists, err := unstructured.NestedSlice(constraint.UnstructuredContent(), "status", "violations")
	if !exists || err != nil {
		return violations
	}
	for _, rawViolation := range rawViolations {
		violation, ok := rawViolation.(map[string]interface{})
		if !ok {
			log.Error(fmt.Errorf("could not cast violation to map[string]interface{}: %+v", rawViolation), "skipping violation")
		} else {
			violations = append(violations, violation)
		}
	}
	return violations
}

// getResource collects resource information for a violation
func (c *Client) getResource(ctx context.Context, violation map[string]interface{}, kindToGVR map[string][]schema.GroupVersionResource) (*Resource, error) {
	name, _, _ := unstructured.NestedString(violation, "name")
	namespace, _, _ := unstructured.NestedString(violation, "namespace")
	kind, _, _ := unstructured.NestedString(violation, "kind")
	resource, err := c.dynamicClient.GetResourceByKind(ctx, kind, name, namespace, kindToGVR)
	if err != nil {
		return nil, err
	}
	// Config Connector resources have a status.selfLink attribute pointing to the
	// actual Google Cloud resource (not the Kubernetes resource).
	statusSelfLink, _, _ := unstructured.NestedString(resource.UnstructuredContent(), "status", "selfLink")
	// Config Connector resources have an annotation pointing to the
	// project ID of the Google Cloud resource. Get it if available.
	projectID := resource.GetAnnotations()[cnrmAnnotationProjectID]
	// In Epidemic, we set a custom annotation to indicate the Constraint's
	// Severity when triggered. Get it if available.
	severity := resource.GetAnnotations()[esAnnotationSeverity]
	message, _, _ := unstructured.NestedString(violation, "message")
	specJSON, err := getSpecAsJSON(resource)
	if err != nil {
		c.log.Error(err, "could not get resource spec as JSON string")
	}
	return &Resource{
		Name:           name,
		Namespace:      namespace,
		GVK:            resource.GroupVersionKind(),
		SelfLink:       resource.GetSelfLink(),
		UID:            resource.GetUID(),
		ProjectID:      projectID,
		Severity:       severity,
		StatusSelfLink: statusSelfLink,
		Message:        message,
		SpecJSON:       specJSON,
	}, nil
}

// getSpecAsJSON returns a string containing the spec field of the provided object
func getSpecAsJSON(obj *unstructured.Unstructured) (string, error) {
	var specJSONBytes []byte
	specMap, exists, err := unstructured.NestedMap(obj.UnstructuredContent(), "spec")
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("no spec on object: %+v", obj)
	}
	specJSONBytes, err = json.Marshal(specMap)
	if err != nil {
		return "", err
	}
	return string(specJSONBytes), err
}

func printFindingRequests(requestMap map[string]*securitycenterpb.CreateFindingRequest) error {
	var requests []*securitycenterpb.CreateFindingRequest
	for _, req := range requestMap {
		requests = append(requests, req)
	}
	return print.AsJSON(requests)
}
