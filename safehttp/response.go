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

import (
	"io"
)

// Response should encapsulate the data passed to the ResponseWriter to be
// written by the Dispatcher. Any implementation of the interface should be
// supported by the Dispatcher.
type Response interface{}

// JSONResponse should encapsulate a valid JSON object that will be serialised
// and written to the http.ResponseWriter using a JSON encoder.
type JSONResponse struct {
	Data interface{}
}

// Template implements a template.
type Template interface {
	// Execute applies data to the template and then writes the result to
	// the io.Writer.
	//
	// Execute returns an error if applying the data object to the
	// Template fails or if an error occurs while writing the result to the
	// io.Writer.
	Execute(wr io.Writer, data interface{}) error
}

// TemplateResponse bundles a Template with its data and names to function
// mappings to be passed together to the commit phase.
type TemplateResponse struct {
	Template *Template
	Data     *interface{}
	FuncMap  map[string]interface{}
}

// NoContentResponse is sent to the commit phase when it's initiated from
// ResponseWriter.NoContent.
type NoContentResponse struct{}

// ErrorResponse is sent to the on error phase when initiated from
// ResponseWriter.WriteError, encapsulating the status code with which the
// method was called.
type ErrorResponse struct {
	Code StatusCode
}
