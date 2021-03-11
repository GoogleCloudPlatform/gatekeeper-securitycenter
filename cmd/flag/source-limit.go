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
	"fmt"

	"github.com/spf13/pflag"
)

const defaultSourceLimit = 100

// SourceLimit defines the max number of sources to retrieve from the Security Command Center API
// when listing sources.
type SourceLimit struct {
	value int
}

func (l *SourceLimit) Add(flags *pflag.FlagSet) {
	flags.IntVar(&l.value, "source-limit", defaultSourceLimit,
		"(optional) The max number of sources to return")
}

func (l *SourceLimit) Validate() error {
	if l.value < 1 || l.value > 1000 {
		return fmt.Errorf("invalid source-limit, must be between 1 and 1000")
	}
	return nil
}

func (l *SourceLimit) Value() int {
	return l.value
}
