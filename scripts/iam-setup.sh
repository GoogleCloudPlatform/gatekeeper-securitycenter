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

# This script:
#
# - creates a Security Command Center source;
#
# - creates the Google service accounts; and
#
# - grants the Cloud IAM roles
#
# required to use the `gatekeeper-securitycenter` CLI and controller.
#
# Source this script to set up environment variables to use with `kpt cfg set`.
#
# If you have an existing Security Command Center source that you want to use,
# set the SOURCE_NAME environment variable before running this script.
#
# This script assumes:
#
# - you have installed the Google Cloud SDK, and `gcloud` is set to use the
#   project and account you want to use to run this script;
#
# - you already have a GKE cluster;
#
# - Workload Identity is enabled in the GKE cluster;
#
# - you have installed jq; and.
#
# To read explanations for the commands used in this script, see the
# instructions in the tutorial:
# https://cloud.google.com/architecture/reporting-policy-controller-audit-violations-security-command-center

set -euf -o pipefail

CLOUDSDK_CORE_ACCOUNT=${CLOUDSDK_CORE_ACCOUNT:-$(gcloud config list --format 'value(core.account)')}
CLOUDSDK_CORE_PROJECT=${CLOUDSDK_CORE_PROJECT:-$(gcloud config list --format 'value(core.project)')}
ORGANIZATION_ID=$(gcloud projects get-ancestors "$CLOUDSDK_CORE_PROJECT" \
    --format json | jq -r '.[] | select (.type=="organization") | .id')

SOURCES_ADMIN_SA_NAME=${SOURCES_ADMIN_SA_NAME:-securitycenter-sources-admin}
SOURCES_ADMIN_SA=$SOURCES_ADMIN_SA_NAME@$CLOUDSDK_CORE_PROJECT.iam.gserviceaccount.com

FINDINGS_EDITOR_SA_NAME=${FINDINGS_EDITOR_SA_NAME:-gatekeeper-securitycenter}
FINDINGS_EDITOR_SA=$FINDINGS_EDITOR_SA_NAME@$CLOUDSDK_CORE_PROJECT.iam.gserviceaccount.com

SOURCE_DISPLAY_NAME=${SOURCE_DISPLAY_NAME:-"Gatekeeper"}
SOURCE_DESCRIPTION=${SOURCE_DESCRIPTION:-"Reports violations from Gatekeeper audits"}

K8S_NAMESPACE=${K8S_NAMESPACE:-gatekeeper-securitycenter}
K8S_SA=${K8S_SA:-gatekeeper-securitycenter-controller}

VERSION=${VERSION:-$(curl -s https://api.github.com/repos/GoogleCloudPlatform/gatekeeper-securitycenter/releases/latest | jq -r '.tag_name')}

if [[ ! -x "gatekeeper-securitycenter-$VERSION" ]]; then
    >&2 echo Downloading "gatekeeper-securitycenter-$VERSION"
    curl -sSLo "gatekeeper-securitycenter-$VERSION" "https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/releases/download/${VERSION}/gatekeeper-securitycenter_$(uname -s)_$(uname -m)"
    chmod +x "gatekeeper-securitycenter-$VERSION"
fi

# Create the sources admin Google service account

if ! gcloud iam service-accounts describe "$SOURCES_ADMIN_SA" > /dev/null 2>&1 ; then
    >&2 echo Creating Google service account "$SOURCES_ADMIN_SA"
    gcloud iam service-accounts create "$SOURCES_ADMIN_SA_NAME" \
        --display-name "Security Command Center sources admin"
fi

gcloud organizations add-iam-policy-binding \
    "$ORGANIZATION_ID" \
    --member "serviceAccount:$SOURCES_ADMIN_SA" \
    --role roles/securitycenter.sourcesAdmin > /dev/null

gcloud organizations add-iam-policy-binding \
    "$ORGANIZATION_ID" \
    --member "serviceAccount:$SOURCES_ADMIN_SA" \
    --role roles/serviceusage.serviceUsageConsumer > /dev/null

gcloud iam service-accounts add-iam-policy-binding \
    "$SOURCES_ADMIN_SA" \
    --member "user:$CLOUDSDK_CORE_ACCOUNT" \
    --role roles/iam.serviceAccountTokenCreator > /dev/null

# Create the Security Command Center source if it doesn't exist

SOURCE_NAME=${SOURCE_NAME:-$("./gatekeeper-securitycenter-$VERSION" sources list \
    --organization "$ORGANIZATION_ID" \
    --impersonate-service-account "$SOURCES_ADMIN_SA" \
    | jq -r ".[] | select(.display_name==\"$SOURCE_DISPLAY_NAME\") | .name")}
SOURCE_NAME=${SOURCE_NAME:-$("./gatekeeper-securitycenter-$VERSION" sources create \
    --organization "$ORGANIZATION_ID" \
    --display-name "$SOURCE_DISPLAY_NAME" \
    --description "$SOURCE_DESCRIPTION" \
    --impersonate-service-account "$SOURCES_ADMIN_SA" | jq -r '.name')}

# Create the findings editor Google service account

if ! gcloud iam service-accounts describe "$FINDINGS_EDITOR_SA" > /dev/null 2>&1 ; then
    >&2 echo Creating Google service account "$FINDINGS_EDITOR_SA"
    gcloud iam service-accounts create "$FINDINGS_EDITOR_SA_NAME" \
        --display-name "Security Command Center Gatekeeper findings editor"
fi

"./gatekeeper-securitycenter-$VERSION" sources add-iam-policy-binding \
    --source "$SOURCE_NAME" \
    --member "serviceAccount:$FINDINGS_EDITOR_SA" \
    --role roles/securitycenter.findingsEditor \
    --impersonate-service-account "$SOURCES_ADMIN_SA" > /dev/null

# Comment this out if you aren't using Workload Identity in your GKE cluster
gcloud iam service-accounts add-iam-policy-binding \
    "$FINDINGS_EDITOR_SA" \
    --member "serviceAccount:$CLOUDSDK_CORE_PROJECT.svc.id.goog[$K8S_NAMESPACE/$K8S_SA]" \
    --role roles/iam.workloadIdentityUser > /dev/null

# Comment this out if you don't plan to use the `findings sync` subcommand
gcloud organizations add-iam-policy-binding \
    "$ORGANIZATION_ID" \
    --member "serviceAccount:$FINDINGS_EDITOR_SA" \
    --role roles/serviceusage.serviceUsageConsumer > /dev/null

# Comment this out if you don't plan to use the `findings sync` subcommand
gcloud iam service-accounts add-iam-policy-binding \
    "$FINDINGS_EDITOR_SA" \
    --member "user:$CLOUDSDK_CORE_ACCOUNT" \
    --role roles/iam.serviceAccountTokenCreator > /dev/null

>&2 echo ""
echo SOURCES_ADMIN_SA="$SOURCES_ADMIN_SA"
echo FINDINGS_EDITOR_SA="$FINDINGS_EDITOR_SA"
echo SOURCE_NAME="$SOURCE_NAME"
