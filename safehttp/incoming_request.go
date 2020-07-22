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
)

// IncomingRequest TODO
type IncomingRequest struct {
	req    *http.Request
	Header Header
}

func newIncomingRequest(req *http.Request) IncomingRequest {
	return IncomingRequest{req: req, Header: newHeader(req.Header)}
}

// QueryForm TODO
func (r IncomingRequest) QueryForm() (*Form, error) {
	if r.req.Method != "GET" {
		return nil, fmt.Errorf("got request method %s, want GET", r.req.Method)
	}
	err := r.req.ParseForm()
	if err != nil {
		return nil, err
	}
	return &Form{values: r.req.Form}, nil
}

// PostForm TODO
func (r IncomingRequest) PostForm() (*Form, error) {
	if r.req.Method != "POST" && r.req.Method != "PATCH" && r.req.Method != "PUT" {
		return nil, fmt.Errorf("got request method %s, want POST/PATCH/PUT", r.req.Method)
	}

	if ct := r.req.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
		return nil, fmt.Errorf("invalid method called for Content-Type: %s, want MultipartForm", ct)
	}
	err := r.req.ParseForm()
	if err != nil {
		return &Form{}, err
	}
	return &Form{values: r.req.PostForm}, nil
}

// MultipartForm TODO
func (r IncomingRequest) MultipartForm(maxMemory int64) (*MultipartForm, error) {
	const defaultMaxMemory = 32 << 20
	if r.req.Method != "POST" && r.req.Method != "PATCH" && r.req.Method != "PUT" {
		return nil, fmt.Errorf("got request method %s, want POST/PATCH/PUT", r.req.Method)
	}

	if ct := r.req.Header.Get("Content-Type"); !strings.HasPrefix(ct, "multipart/form-data") {
		return nil, fmt.Errorf("invalid method called for Content-Type: %s, want PostForm", ct)
	}
	if maxMemory < 0 || maxMemory > defaultMaxMemory {
		maxMemory = defaultMaxMemory
	}
	err := r.req.ParseMultipartForm(maxMemory)
	if err != nil {
		return &MultipartForm{}, err
	}
	mf := &MultipartForm{
		Form: &Form{
			values: r.req.MultipartForm.Value,
		},
		file: r.req.MultipartForm.File}
	return mf, nil
}
