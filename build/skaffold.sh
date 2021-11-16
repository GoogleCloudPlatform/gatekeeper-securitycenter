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

# Skaffold uses this script to build the container image.

set -euf -o pipefail

if ! [ -x "$(command -v ko)" ]; then
    pushd $(mktemp -d)
    curl -L https://github.com/google/ko/archive/v0.9.3.tar.gz | tar --strip-components 1 -zx
    go build -o $(go env GOPATH)/bin/ko .
    popd
fi

if ! [ -x "$(command -v crane)" ]; then
    pushd $(mktemp -d)
    curl -L https://github.com/google/go-containerregistry/archive/v0.7.0.tar.gz | tar --strip-components 1 -zx
    go build -o $(go env GOPATH)/bin/crane .
    popd
fi

image_tar=$(mktemp)

KO_DOCKER_REPO=${KO_DOCKER_REPO:-$SKAFFOLD_DEFAULT_REPO}
export KO_DOCKER_REPO

ko publish --tarball $image_tar --push=false .

crane push $image_tar $IMAGE
