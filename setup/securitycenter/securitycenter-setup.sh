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
# - creates a Security Command Center source; and
#
# - grants the Findings Editor role to a Google service account

set -euf -o pipefail

FINDINGS_EDITOR_SA=$FINDINGS_EDITOR_SA_NAME@$MANAGED_PROJECT_ID.iam.gserviceaccount.com
SOURCES_ADMIN_SA=$SOURCES_ADMIN_SA_NAME@$MANAGED_PROJECT_ID.iam.gserviceaccount.com

SOURCE_DISPLAY_NAME=${SOURCE_DISPLAY_NAME:-"Gatekeeper"}
SOURCE_DESCRIPTION=${SOURCE_DESCRIPTION:-"Reports violations from Gatekeeper audits"}

# Download the gatekeeper-securitycenter command-line tool

VERSION=${VERSION:-$(curl -s https://api.github.com/repos/GoogleCloudPlatform/gatekeeper-securitycenter/releases/latest | jq -r '.tag_name')}
if [[ ! -x "gatekeeper-securitycenter-$VERSION" ]]; then
    >&2 echo Downloading "gatekeeper-securitycenter-$VERSION"
    curl -sSLo "gatekeeper-securitycenter-$VERSION" "https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter/releases/download/${VERSION}/gatekeeper-securitycenter_$(uname -s)_$(uname -m)"
    chmod +x "gatekeeper-securitycenter-$VERSION"
fi

# Check if the Security Command Center source already exists

SOURCE_NAME=$("./gatekeeper-securitycenter-$VERSION" sources list \
    --organization "$ORGANIZATION_ID" \
    --impersonate-service-account "$SOURCES_ADMIN_SA" | \
    jq -r ".[] | select (.display_name==\"$SOURCE_DISPLAY_NAME\") | .name")

# Create the Security Command Center source

SOURCE_NAME=${SOURCE_NAME:-$("./gatekeeper-securitycenter-$VERSION" sources create \
    --organization "$ORGANIZATION_ID" \
    --display-name "$SOURCE_DISPLAY_NAME" \
    --description "$SOURCE_DESCRIPTION" \
    --impersonate-service-account "$SOURCES_ADMIN_SA" | jq -r '.name')}

# Add IAM policy bindings for Security Command Center source

"./gatekeeper-securitycenter-$VERSION" sources add-iam-policy-binding \
    --source "$SOURCE_NAME" \
    --member "serviceAccount:$FINDINGS_EDITOR_SA" \
    --role roles/securitycenter.findingsEditor \
    --impersonate-service-account "$SOURCES_ADMIN_SA" > /dev/null


>&2 echo ""
>&2 echo "Use these values to deploy the gatekeeper-securitycenter controller:"
>&2 echo ""
echo "SOURCE_NAME=$SOURCE_NAME"
echo "FINDINGS_EDITOR_SA=$FINDINGS_EDITOR_SA"
