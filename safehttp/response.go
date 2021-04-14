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

// ErrorResponse is an HTTP error response. The Dispatcher is responsible for
// determining whether it is safe.
type ErrorResponse interface {
	Code() StatusCode
}

// JSONResponse should encapsulate a valid JSON object that will be serialised
// and written to the http.ResponseWriter using a JSON encoder.
type JSONResponse struct {
	Data interface{}
}

// WriteJSON creates a JSONResponse from the data object and calls the Write
// function of the ResponseWriter, passing the response. The data object should
// be valid JSON, otherwise an error will occur.
func WriteJSON(w ResponseWriter, data interface{}) Result {
	return w.Write(JSONResponse{data})
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

	// ExecuteTemplate applies the named associated template to the specified data
	// object and writes the output to the io.Writer.
	//
	// ExecuteTemplate returns an error if applying the data object to the
	// Template fails or if an error occurs while writing the result to the
	// io.Writer.
	ExecuteTemplate(wr io.Writer, name string, data interface{}) error
}

// TemplateResponse bundles a Template with its data and names to function
// mappings to be passed together to the commit phase.
type TemplateResponse struct {
	Template Template
	Name     string
	Data     interface{}
	FuncMap  map[string]interface{}
}

// ExecuteTemplate creates a TemplateResponse from the provided Template and its
// data and calls the Write function of the ResponseWriter, passing the
// response.
// Leaving name empty is valid if the template does not have associated templates.
func ExecuteTemplate(w ResponseWriter, t Template, name string, data interface{}) Result {
	return w.Write(&TemplateResponse{Template: t, Name: name, Data: data, FuncMap: nil})
}

// ExecuteTemplateWithFuncs creates a TemplateResponse from the provided
// Template, its data and the name to function mappings and calls the Write
// function of the ResponseWriter, passing the response.
// Leaving name empty is valid if the template does not have associated templates.
func ExecuteTemplateWithFuncs(w ResponseWriter, t Template, name string, data interface{}, fm map[string]interface{}) Result {
	return w.Write(&TemplateResponse{Template: t, Name: name, Data: data, FuncMap: fm})
}

// NoContentResponse is sent to the commit phase when it's initiated from
// ResponseWriter.NoContent.
type NoContentResponse struct{}
