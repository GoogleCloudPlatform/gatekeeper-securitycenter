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

package logging

import (
	"log"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

// CreateStdLog creates a logr.Logger using Go's standard log package
func CreateStdLog(name string) logr.Logger {
	verbosity := 1
	if strings.ToLower(os.Getenv("DEBUG")) == "true" {
		verbosity = 2
	}
	stdr.SetVerbosity(verbosity)
	return stdr.New(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)).WithName(name)
}
