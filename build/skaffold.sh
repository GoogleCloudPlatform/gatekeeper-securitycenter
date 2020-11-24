#!/usr/bin/env bash
#
# Copyright 2020 Google LLC
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

# Skaffold uses this script to build the container image

set -euf -o pipefail

if ! [ -x "$(command -v ko)" ]; then
    pushd $(mktemp -d)
    go mod init tmp; GOFLAGS= go get github.com/google/ko/cmd/ko@v0.6.2
    popd
fi

if ! [ -x "$(command -v crane)" ]; then
    pushd $(mktemp -d)
    go mod init tmp; GOFLAGS= go get github.com/google/go-containerregistry/cmd/crane@v0.1.4
    popd
fi

image_tar=$(mktemp)

ko publish --tarball $image_tar --push=false .

crane push $image_tar $IMAGE
