// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package headers

import (
	"bytes"
	"strings"
)

const (
	statusOK               = "HTTP/1.1 200 OK"
	statusNotImplemented   = "HTTP/1.1 501 Not Implemented"
	statusBadRequestPrefix = "HTTP/1.1 400 Bad Request"
)

func extractStatus(response []byte) string {
	position := bytes.IndexByte(response, '\r')
	if position == -1 {
		return ""
	}
	return string(response[:position])
}

func matchStatus(got, want string) bool {
	return strings.HasPrefix(got, want)
}
