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

import "github.com/spf13/pflag"

// Cluster name or other identifier
type Cluster struct {
	value string
}

func (c *Cluster) Add(flags *pflag.FlagSet) {
	flags.StringVar(&c.value, "cluster", "",
		"(optional) name or other identifier for the cluster, added to findings")
}

func (c *Cluster) Validate() error {
	return nil
}

func (c *Cluster) Value() string {
	return c.value
}
