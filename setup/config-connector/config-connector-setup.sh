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

# This script creates a GKE cluster and installs Config Connector.

set -ef -o pipefail

GCLOUD_PROJECT_ID=${CLOUDSDK_CORE_PROJECT:-$(gcloud config list --format 'value(core.project)')}
HOST_PROJECT_ID=${HOST_PROJECT_ID:-$GCLOUD_PROJECT_ID}
MANAGED_PROJECT_ID=${MANAGED_PROJECT_ID:-$HOST_PROJECT_ID}

echo "Host project: $HOST_PROJECT_ID (override with \$HOST_PROJECT_ID)"
echo "Managed project: $MANAGED_PROJECT_ID (override with \$MANAGED_PROJECT_ID)"

GCLOUD_CLUSTER=${CLOUDSDK_CONTAINER_CLUSTER:-$(gcloud config list --format 'value(container.cluster)')}
DEFAULT_CLUSTER=${GCLOUD_CLUSTER:-config-connector}
CLUSTER=${CLUSTER:-$DEFAULT_CLUSTER}

GCLOUD_ZONE=${CLOUDSDK_COMPUTE_ZONE:-$(gcloud config list --format 'value(compute.zone)')}
DEFAULT_ZONE=${GCLOUD_ZONE:-us-central1-f}
ZONE=${ZONE:-$DEFAULT_ZONE}

GOOGLE_CLOUD_APIS=${GOOGLE_CLOUD_APIS:-""}
gcloud services enable container.googleapis.com $GOOGLE_CLOUD_APIS --project "$HOST_PROJECT_ID"
if [ ! -z "$GOOGLE_CLOUD_APIS" ]; then
    gcloud services enable $GOOGLE_CLOUD_APIS --project "$MANAGED_PROJECT_ID"
fi

if gcloud container clusters describe "$CLUSTER" --project "$HOST_PROJECT_ID" --zone "$ZONE" --no-user-output-enabled 2> /dev/null ; then
    echo "Using existing GKE cluster $CLUSTER in zone $ZONE"
else
    MAX_CPU=${MAX_CPU:-12}
    MAX_MEMORY=${MAX_MEMORY:-45}
    RELEASE_CHANNEL=${RELEASE_CHANNEL:-regular}
    CLUSTER_NODES_TAGS=${CLUSTER_NODES_TAGS:-"gke-$CLUSTER-default-pool"}

    echo "Creating GKE cluster $CLUSTER for host project in zone $ZONE (override with \$CLUSTER and \$ZONE)"
    gcloud container clusters create "$CLUSTER" \
        --enable-autoprovisioning --max-cpu "$MAX_CPU" --max-memory "$MAX_MEMORY" \
        --enable-ip-alias \
        --enable-stackdriver-kubernetes \
        --project "$HOST_PROJECT_ID" \
        --release-channel "$RELEASE_CHANNEL" \
        --tags "$CLUSTER_NODES_TAGS" \
        --verbosity error \
        --workload-pool "$HOST_PROJECT_ID.svc.id.goog" \
        --zone "$ZONE"
fi

KCC_SA_NAME=${KCC_SA_NAME:-config-connector}
KCC_SA=$KCC_SA_NAME@$HOST_PROJECT_ID.iam.gserviceaccount.com
KCC_SA_DISPLAY_NAME=${KCC_SA_DISPLAY_NAME:-"Config Connector service account"}
KCC_SA_ORGANIZATION_ROLES=${KCC_SA_ORGANIZATION_ROLES:-""}
KCC_SA_PROJECT_ROLES=${KCC_SA_PROJECT_ROLES:-"roles/editor"}

if gcloud iam service-accounts describe "$KCC_SA" --project="$HOST_PROJECT_ID" --no-user-output-enabled 2> /dev/null ; then
    echo "Using existing Google service account $KCC_SA"
else
    echo "Creating Google service account $KCC_SA (override with \$KCC_SA_NAME)"
    gcloud iam service-accounts create \
        "$KCC_SA_NAME" \
        --display-name "$KCC_SA_DISPLAY_NAME" \
        --project "$HOST_PROJECT_ID"
fi

ORGANIZATION_ID=${ORGANIZATION_ID:-$(gcloud projects get-ancestors "$MANAGED_PROJECT_ID" \
    --format json | jq -r '.[] | select (.type=="organization") | .id')}

for ROLE in $KCC_SA_ORGANIZATION_ROLES ; do
    echo "Granting $ROLE to $KCC_SA for organization $ORGANIZATION_ID"
    gcloud organizations add-iam-policy-binding \
        "$ORGANIZATION_ID" \
        --member "serviceAccount:$KCC_SA" \
        --role "$ROLE" \
        --no-user-output-enabled
done

for ROLE in $KCC_SA_PROJECT_ROLES ; do
    echo "Granting $ROLE to $KCC_SA for project $MANAGED_PROJECT_ID"
    gcloud projects add-iam-policy-binding \
        "$MANAGED_PROJECT_ID" \
        --member "serviceAccount:$KCC_SA" \
        --role "$ROLE" \
        --no-user-output-enabled
done

KCC_NS=${KCC_NS:-config-connector}
echo "Using namespace $KCC_NS for Config Connector resources in project $MANAGED_PROJECT_ID"

echo "Binding $KCC_SA to namespaces/cnrm-system/serviceaccounts/cnrm-controller-manager-$KCC_NS"
gcloud iam service-accounts add-iam-policy-binding \
    "$KCC_SA" \
    --member "serviceAccount:$HOST_PROJECT_ID.svc.id.goog[cnrm-system/cnrm-controller-manager-$KCC_NS]" \
    --role roles/iam.workloadIdentityUser \
    --no-user-output-enabled

SCRIPT_DIR=$(dirname "$0")
KCC_VERSION=${KCC_VERSION:-latest}
if [ ! -f "$SCRIPT_DIR/configconnector-operator.yaml" ]; then
    (cd "$SCRIPT_DIR" ; gsutil cp "gs://configconnector-operator/$KCC_VERSION/release-bundle.tar.gz" - \
        | tar xz --strip-components 2 ./operator-system/configconnector-operator.yaml)
fi
