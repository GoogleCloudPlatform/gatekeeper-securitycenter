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

package version

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	writer = os.Stdout

	// Version is provided by ldflags at compile time
	Version = "(devel)"

	// Cmd is the version sub-command
	Cmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return printVersion(Version)
		},
	}
)

func printVersion(version string) error {
	// if version == "" {
	// 	if buildInfo, ok := debug.ReadBuildInfo(); ok {
	// 		version = buildInfo.Main.Version
	// 	} else {
	// 		return fmt.Errorf("could not get build information")
	// 	}
	// }
	_, err := fmt.Fprintln(writer, version)
	return err
}
