// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package safehttp

// Config provide additional configurations to Interceptors when needed.
type Config interface {
	// Apply applies the configurations to an Interceptor if it's of the
	// appropriate type. The bool indicates whether the configuration took
	// effect. If that's the case, the modified interceptor should be returned
	// and otherwise Apply should return a nil Interceptor.
	Apply(Interceptor) (Interceptor, bool)
}
