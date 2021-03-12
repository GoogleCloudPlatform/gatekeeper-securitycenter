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

module github.com/googlecloudplatform/gatekeeper-securitycenter

go 1.16

require (
	cloud.google.com/go v0.79.0
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/stdr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.5
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.16.0
	google.golang.org/api v0.41.0
	google.golang.org/genproto v0.0.0-20210310155132-4ce2db91004e
	google.golang.org/grpc v1.36.0
	google.golang.org/protobuf v1.25.0
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
)
