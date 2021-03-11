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

const defaultIntervalSeconds = 120

// Interval number of seconds to sleep between each iteration of the control loop.
type Interval struct {
	value int
}

func (i *Interval) Add(flags *pflag.FlagSet) {
	flags.IntVar(&i.value, "interval", defaultIntervalSeconds,
		"(optional) control loop interval in seconds, should be greater than or equal to audit manager internal")
}

func (i *Interval) Validate() error {
	if i.value < 5 || i.value > 1000 {
		return fmt.Errorf("invalid value for findings-limit=%v, must be between 5 and 1000", i.value)
	}
	return nil
}

func (i *Interval) Value() int {
	return i.value
}
