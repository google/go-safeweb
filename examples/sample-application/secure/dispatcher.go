// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build go1.16
// +build go1.16

package secure

import (
	"net/http"

	"github.com/google/go-safeweb/safehttp"

	"github.com/google/go-safeweb/examples/sample-application/secure/responses"
	"github.com/google/go-safeweb/examples/sample-application/secure/templates"
)

// dispatcher is a custom dispatcher implementation. See
// https://pkg.go.dev/github.com/google/go-safeweb/safehttp#hdr-Dispatcher.
type dispatcher struct {
	// No need for a Write method, the default dispatcher knows how to write all
	// non-error responses we use in this project.
	safehttp.DefaultDispatcher
}

func (d dispatcher) Error(rw http.ResponseWriter, resp safehttp.ErrorResponse) error {
	if ce, ok := resp.(responses.Error); ok {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(int(ce.Code()))
		return templates.All.ExecuteTemplate(rw, "error.tpl.html", ce.Message)
	}
	// Calling the default dispatcher in case we have no custom responses that match.
	// This is strongly advised.
	return d.DefaultDispatcher.Error(rw, resp)
}
