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

# This script deploys the controller to your Kubernetes cluster.
#
# It requires the following environment variables to be set:
#
# - `SOURCE_NAME`: The full name of the Security Command Center source, in the
#   format `organizations/[ORGANIZATION_ID]/sources/[SOURCE_ID]`;
#
# - `FINDINGS_EDITOR_SA`: A Google service account with permission to create
#   and edit findings for the Security Command Center source, e.g.,
#   `gatekeeper-securitycenter@[PROJECT_ID].iam.gserviceaccount.com`.

set -euf -o pipefail

INVENTORY_NAMESPACE=${INVENTORY_NAMESPACE:-gatekeeper-securitycenter}

if [[ ! -x "kpt" ]]; then
    >&2 echo "Error: kpt not on path: https://googlecontainertools.github.io/kpt/installation/"
    return 1
fi

if [[ -z "$SOURCE_NAME" ]]; then
    >&2 echo "Error: SOURCE_NAME environment variable not set."
    return 1
fi

if [[ -z "$FINDINGS_EDITOR_SA" ]]; then
    >&2 echo "Error: FINDINGS_EDITOR_SA environment variable not set."
    return 1
fi

if [[ -z "$CLUSTER_NAME" ]]; then
    CLUSTER_NAME=$(kubectl config current-context)
fi

kpt cfg set manifests source "$SOURCE_NAME"

kpt cfg set manifests cluster "$CLUSTER_NAME"

kpt cfg annotate manifests \
    --kind ServiceAccount --name gatekeeper-securitycenter-controller \
    --kv "iam.gke.io/gcp-service-account=$FINDINGS_EDITOR_SA"

if [[ ! -f "manifests/inventory-template.yaml" ]]; then
    kpt live init manifests --namespace "$INVENTORY_NAMESPACE"
fi

kpt live apply manifests --reconcile-timeout 2m --output events
