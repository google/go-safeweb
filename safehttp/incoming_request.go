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
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// IncomingRequest TODO
type IncomingRequest struct {
	req       *http.Request
	Header    Header
	parseOnce sync.Once
}

func newIncomingRequest(req *http.Request) IncomingRequest {
	return IncomingRequest{req: req, Header: newHeader(req.Header)}
}

// QueryForm parses the query parameters provided in the request. It returns
// the parsed parameters as a Form object, if no error occured, or the parsing error otherwise.
func (r *IncomingRequest) QueryForm() (f *Form, err error) {
	r.parseOnce.Do(func() {
		if r.req.Method != "GET" {
			f, err = nil, fmt.Errorf("got request method %s, want GET", r.req.Method)
		}
		if err := r.req.ParseForm(); err != nil {
			f = nil
		}
	})
	f, err = &Form{values: r.req.Form}, nil
	return
}

// PostForm parses the form parameters provided in the body of a POST, PATCH or
// PUT request that does not have Content-Type: multipart/form-data. It returns
// the parsed parameters as a Form object, if no error occured, or the parsing error otherwise.
func (r *IncomingRequest) PostForm() (f *Form, err error) {
	r.parseOnce.Do(func() {
		if r.req.Method != "POST" && r.req.Method != "PATCH" && r.req.Method != "PUT" {
			f, err = nil, fmt.Errorf("got request method %s, want POST/PATCH/PUT", r.req.Method)
		}

		if ct := r.req.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			f, err = nil, fmt.Errorf("invalid method called for Content-Type: %s, want MultipartForm", ct)
		}
		if err := r.req.ParseForm(); err != nil {
			f = &Form{}
		}
	})
	f, err = &Form{values: r.req.PostForm}, nil
	return
}

// MultipartForm parses the form parameters provided in the body of a POST,
// PATCH or PUT request that has Content-Type set tomultipart/form-data. It
// returns a MultipartForm object containing the parsed form parameter and
// files, if no error occured, or the parsing error otherwise.
func (r *IncomingRequest) MultipartForm(maxMemory int64) (f *MultipartForm, err error) {
	r.parseOnce.Do(func() {
		const defaultMaxMemory = 32 << 20
		if r.req.Method != "POST" && r.req.Method != "PATCH" && r.req.Method != "PUT" {
			f, err = nil, fmt.Errorf("got request method %s, want POST/PATCH/PUT", r.req.Method)
		}

		if ct := r.req.Header.Get("Content-Type"); !strings.HasPrefix(ct, "multipart/form-data") {
			f, err = nil, fmt.Errorf("invalid method called for Content-Type: %s, want PostForm", ct)
		}
		if maxMemory < 0 || maxMemory > defaultMaxMemory {
			maxMemory = defaultMaxMemory
		}

		if err := r.req.ParseMultipartForm(maxMemory); err != nil {
			f = &MultipartForm{}
		}
	})
	f, err = &MultipartForm{
		Form: Form{
			values: r.req.MultipartForm.Value,
		},
		file: r.req.MultipartForm.File}, nil
	return
}
