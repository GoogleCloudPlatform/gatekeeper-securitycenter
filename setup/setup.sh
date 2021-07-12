#!/usr/bin/env bash
#
# Copyright 2021 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This scripts orchestrates the creation of the prerequisite resources for the
# gatekeeper-securitycenter Kubernetes controller.
#
# It uses kpt (https://kpt.dev) to manage resources manifests and
# Config Connector to create the Google Cloud resources.

set -ef -o pipefail

DIR=$(dirname "$0")
SET_BY=$(basename "$0")

if [ -z "$ADMIN_USER" ]; then
    echo Run \"source "$DIR/setup.env"\" to set environment variables before running this script.
    exit 1
fi

# Set values from `setup.env`

kpt fn eval "$DIR" \
    --image "gcr.io/kpt-fn/apply-setters:$KPT_FN_APPLY_SETTERS_VERSION" -- \
    "admin-user=$ADMIN_USER" \
    "findings-editor-gsa-name=$FINDINGS_EDITOR_SA_NAME" \
    "host-project-id=$HOST_PROJECT_ID" \
    "kcc-gsa-name=$KCC_SA_NAME" \
    "kcc-namespace=$KCC_NS" \
    "managed-project-id=$MANAGED_PROJECT_ID" \
    "node-network-tags=gke-$CLUSTER-default-pool" \
    "organization-id=$ORGANIZATION_ID" \
    "sources-admin-gsa-name=$SOURCES_ADMIN_SA_NAME"

# Create a GKE cluster and deploy Config Connector.

(cd "$DIR/config-connector" ; ./config-connector-setup.sh)
if [ ! -f "$DIR/config-connector/inventory-template.yaml" ]; then
    kpt live init "$DIR/config-connector" --namespace "$KCC_NS" --inventory-id config-connector
fi
kpt live apply "$DIR/config-connector" --reconcile-timeout 3m --output events
kubectl wait -n cnrm-system --for=condition=Ready pod -l \
    "cnrm.cloud.google.com/component=cnrm-controller-manager,cnrm.cloud.google.com/scoped-namespace=$KCC_NS"

# Deploy Open Policy Agent (OPA) Gatekeeper.

(cd "$DIR/gatekeeper" ; ./gatekeeper-setup.sh)
if [ ! -f "$DIR/gatekeeper/inventory-template.yaml" ]; then
    kpt live init "$DIR/gatekeeper" --namespace "$KCC_NS" --inventory-id gatekeeper
fi
kpt live apply "$DIR/gatekeeper" --reconcile-timeout 5m --output events

# Create Google service accounts and Cloud IAM policy bindings for
# Security Command Center using Config Connector.

if [ ! -f "$DIR/iam/inventory-template.yaml" ]; then
    kpt live init "$DIR/iam" --namespace "$KCC_NS" --inventory-id iam
fi
kpt live apply "$DIR/iam" --reconcile-timeout 5m --output events

# Create the Security Command Center source and set permissions on the source
# using the `gatekeeper-securitycenter` command-line tool.

(cd "$DIR/securitycenter" ; ./securitycenter-setup.sh)

echo ""
echo "You can now deploy the gatekeeper-securitycenter controller."
echo "Follow the instructions in the manifests package:"
echo "<https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/tree/main/manifests>"
