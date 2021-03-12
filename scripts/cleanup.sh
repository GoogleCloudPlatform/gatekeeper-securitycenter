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

# This script deletes:
#
# - the Cloud IAM role bindings; and
#
# - the Google service accounts
#
# created by the setup script.
#
# It does not delete the GKE cluster or any resources in the cluster.
#
# This script assumes:
#
# - you have installed the Google Cloud SDK, and `gcloud` is set to use the
#   project and account you want to use to run this script; and
#
# - you have installed the Go distribution, or you can set values for `OS`
#   (e.g., `linux`, `darwin`) and `ARCH` (e.g., `amd64`, `arm64`) before
#   running this script.

set -euf -o pipefail

CLOUDSDK_CORE_ACCOUNT=${CLOUDSDK_CORE_ACCOUNT:-$(gcloud config list --format 'value(core.account)')}
CLOUDSDK_CORE_PROJECT=${CLOUDSDK_CORE_PROJECT:-$(gcloud config list --format 'value(core.project)')}
ORGANIZATION_ID=$(gcloud projects get-ancestors "$CLOUDSDK_CORE_PROJECT" \
    --format json | jq -r '.[] | select (.type=="organization") | .id')

SOURCES_ADMIN_SA_NAME=${SOURCES_ADMIN_SA_NAME:-securitycenter-sources-admin}
SOURCES_ADMIN_SA=$SOURCES_ADMIN_SA_NAME@$CLOUDSDK_CORE_PROJECT.iam.gserviceaccount.com

FINDINGS_EDITOR_SA_NAME=${FINDINGS_EDITOR_SA_NAME:-gatekeeper-securitycenter}
FINDINGS_EDITOR_SA=$FINDINGS_EDITOR_SA_NAME@$CLOUDSDK_CORE_PROJECT.iam.gserviceaccount.com

K8S_NAMESPACE=${K8S_NAMESPACE:-gatekeeper-securitycenter}
K8S_SA=${K8S_SA:-gatekeeper-securitycenter-controller}

VERSION=${VERSION:-$(curl -s https://api.github.com/repos/GoogleCloudPlatform/gatekeeper-securitycenter/releases/latest | jq -r '.tag_name')}

if [[ ! -x "gatekeeper-securitycenter-$VERSION" ]]; then
    curl -sSLo "gatekeeper-securitycenter-$VERSION" "https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/releases/download/${VERSION}/gatekeeper-securitycenter_$(uname -s)_$(uname -m)"
    chmod +x "gatekeeper-securitycenter-$VERSION"
fi

"./gatekeeper-securitycenter-$VERSION" sources remove-iam-policy-binding \
    --source "$SOURCE_NAME" \
    --member "serviceAccount:$FINDINGS_EDITOR_SA" \
    --role roles/securitycenter.findingsEditor \
    --impersonate-service-account "$SOURCES_ADMIN_SA" > /dev/null

gcloud iam service-accounts remove-iam-policy-binding \
    "$FINDINGS_EDITOR_SA" \
    --member "serviceAccount:$CLOUDSDK_CORE_PROJECT.svc.id.goog[$K8S_NAMESPACE/$K8S_SA]" \
    --role roles/iam.workloadIdentityUser > /dev/null

gcloud iam service-accounts remove-iam-policy-binding \
    "$FINDINGS_EDITOR_SA" \
    --member "user:$CLOUDSDK_CORE_ACCOUNT" \
    --role roles/iam.serviceAccountTokenCreator > /dev/null

gcloud organizations remove-iam-policy-binding \
    "$ORGANIZATION_ID" \
    --member "serviceAccount:$FINDINGS_EDITOR_SA" \
    --role roles/serviceusage.serviceUsageConsumer > /dev/null

gcloud iam service-accounts remove-iam-policy-binding \
    "$SOURCES_ADMIN_SA" \
    --member "user:$CLOUDSDK_CORE_ACCOUNT" \
    --role roles/iam.serviceAccountTokenCreator > /dev/null

gcloud organizations remove-iam-policy-binding \
    "$ORGANIZATION_ID" \
    --member "serviceAccount:$SOURCES_ADMIN_SA" \
    --role roles/serviceusage.serviceUsageConsumer > /dev/null

gcloud organizations remove-iam-policy-binding \
    "$ORGANIZATION_ID" \
    --member "serviceAccount:$SOURCES_ADMIN_SA" \
    --role roles/securitycenter.sourcesAdmin > /dev/null

gcloud iam service-accounts delete "$FINDINGS_EDITOR_SA"

gcloud iam service-accounts delete "$SOURCES_ADMIN_SA"
