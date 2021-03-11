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

package flag

import (
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"k8s.io/client-go/util/homedir"
)

// Kubeconfig used to connect to cluster
// Ref: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/
type Kubeconfig struct {
	value string
}

func (k *Kubeconfig) Add(flags *pflag.FlagSet) {
	var kubeconfigDefault string
	home := homedir.HomeDir()
	if home != "" {
		kubeconfigHomePath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(kubeconfigHomePath); err == nil {
			kubeconfigDefault = kubeconfigHomePath
		}
	}
	flags.StringVar(&k.value, "kubeconfig", kubeconfigDefault,
		"(optional) absolute path to the kubeconfig file, leave blank to use in-cluster config")
}

func (k *Kubeconfig) Validate() error {
	return nil
}

func (k *Kubeconfig) Value() string {
	return k.value
}
