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

// ValidationError is a type used to define sentinel errors that alter the way flags are validated.
type ValidationError string

func (e ValidationError) Error() string {
	return string(e)
}

// SkipValidation skips subsequent flag validation.
// Use with flags such as DryRun to skip validation of flags that are not used in dry-run mode.
const SkipValidation = ValidationError("skipping validation")
