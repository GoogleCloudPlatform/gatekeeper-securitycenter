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
	"github.com/spf13/pflag"
)

// Flag interface to implement for each flag
type Flag interface {
	// Add the flag to be parsed
	Add(*pflag.FlagSet)
	// Validate the flag value
	Validate() error
}

// New creates a new flag validator.
// The flags will be validated in the order they were added.
func New(flags ...Flag) *Flags {
	return &Flags{
		flags: flags,
	}
}

// Flags for command line flags
type Flags struct {
	flags []Flag
}

// AddToFlagSet adds the
func (p *Flags) AddToFlagSet(flagSet *pflag.FlagSet) {
	for _, f := range p.flags {
		f.Add(flagSet)
	}
}

// Validate command-line flags
func (p *Flags) Validate() error {
	for _, f := range p.flags {
		err := f.Validate()
		if err != nil && err != SkipValidation {
			return err
		}
		if err != nil && err == SkipValidation {
			// return now and skip remaining flag validation.
			return nil
		}
	}
	return nil
}
