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

# This script creates a Google Kubernetes Engine (GKE) cluster with
# Workload Identity and Gatekeeper.
#
# This is intended as an aid in setting up a development environment and to
# automate steps in the tutorial. This script is _not_ recommended for
# production use.
#
# This script assumes that:
#
# - you have installed the Google Cloud SDK;
#
# - `gcloud` is set to use the project and account you want to use to run this
#   script (verify this with `gcloud config list --format '(core)'`);
#
# - you have installed kubectl (you can install it with
#   `gcloud components install kubectl --quiet`)
#
# To read explanations for the commands used in this script, see the
# instructions in the tutorial:
# https://cloud.google.com/architecture/reporting-policy-controller-audit-violations-security-command-center

set -euf -o pipefail

CLOUDSDK_CORE_ACCOUNT=${CLOUDSDK_CORE_ACCOUNT:-$(gcloud config list --format 'value(core.account)')}
CLOUDSDK_CORE_PROJECT=${CLOUDSDK_CORE_PROJECT:-$(gcloud config list --format 'value(core.project)')}

CLUSTER=${CLUSTER:-gatekeeper-securitycenter}
NUM_NODES=${NUM_NODES:-3}
RELEASE_CHANNEL=${RELEASE_CHANNEL:-regular}
ZONE=${ZONE:-us-central1-f}

GATEKEEPER_VERSION=v3.5.1

gcloud container clusters create "$CLUSTER" \
    --enable-ip-alias \
    --enable-stackdriver-kubernetes \
    --num-nodes "$NUM_NODES" \
    --release-channel "$RELEASE_CHANNEL" \
    --workload-pool "$CLOUDSDK_CORE_PROJECT.svc.id.goog" \
    --zone "$ZONE"

kubectl create clusterrolebinding cluster-admin-binding \
    --clusterrole cluster-admin \
    --user "$CLOUDSDK_CORE_ACCOUNT"

kubectl apply -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/$GATEKEEPER_VERSION/deploy/gatekeeper.yaml

kubectl rollout status deploy gatekeeper-controller-manager -n gatekeeper-system

CLUSTER_NAME=$(kubectl config current-context)
export CLUSTER_NAME

>&2 echo ""
echo CLUSTER_NAME="$CLUSTER_NAME"
