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

// Package requestparsing contains tests to verify the
// request parsing behavior of `net/http` in Go's standard library.
// Note: All tests use verbatim newline characters in their
// requests instead of using multiline strings to ensure that
// \r and \n end up in exactly the right places.
package requestparsing
