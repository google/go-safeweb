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
// the parsed query parameters as a Form object, if no error occurred. If a parsing
// error occurs it will return it, together with a nil Form.
func (r *IncomingRequest) QueryForm() (*Form, error) {
	var err error
	r.parseOnce.Do(func() {
		if r.req.Method != "GET" {
			err = fmt.Errorf("got request method %s, want GET", r.req.Method)
			return
		}
		err = r.req.ParseForm()
	})
	if err != nil {
		return nil, err
	}
	return &Form{values: r.req.Form}, nil
}

// PostForm parses the form parameters provided in the body of a POST, PATCH or
// PUT request that does not have Content-Type: multipart/form-data. It returns
// the parsed form parameters as a Form object, if no error occurred. If a parsing
// error occurs it will return it, together with a nil Form. Unless we expect the
// header Content-Type: multipart/form-data in a POST request, this method should
// always be used for forms in POST requests.
func (r *IncomingRequest) PostForm() (*Form, error) {
	var err error
	r.parseOnce.Do(func() {
		if r.req.Method != "POST" && r.req.Method != "PATCH" && r.req.Method != "PUT" {
			err = fmt.Errorf("got request method %s, want POST/PATCH/PUT", r.req.Method)
			return
		}

		if ct := r.req.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			err = fmt.Errorf("invalid method called for Content-Type: %s, want MultipartForm", ct)
			return
		}
		err = r.req.ParseForm()
	})
	if err != nil {
		return nil, err
	}
	return &Form{values: r.req.PostForm}, nil
}

// MultipartForm parses the form parameters provided in the body of a POST,
// PATCH or PUT request that has Content-Type set to multipart/form-data. It
// returns a MultipartForm object containing the parsed form parameter and
// files, if no error occurred, or the parsing error together with a nil
// MultipartForm otherwise. This method should  only be used when the user expects
// a POST request with the Content-Type: multipart/form-data header.
func (r *IncomingRequest) MultipartForm(maxMemory int64) (*MultipartForm, error) {
	var err error
	r.parseOnce.Do(func() {
		// Ensures no more than 32 MB are stored in memory when a form file is
		// passed as part of the request. If this is bigger than 32 MB, the rest
		// will be stored on disk.
		const defaultMaxMemory = 32 << 20
		if r.req.Method != "POST" && r.req.Method != "PATCH" && r.req.Method != "PUT" {
			err = fmt.Errorf("got request method %s, want POST/PATCH/PUT", r.req.Method)
			return
		}

		if ct := r.req.Header.Get("Content-Type"); !strings.HasPrefix(ct, "multipart/form-data") {
			err = fmt.Errorf("invalid method called for Content-Type: %s, want PostForm", ct)
			return
		}
		if maxMemory < 0 || maxMemory > defaultMaxMemory {
			maxMemory = defaultMaxMemory
		}

		err = r.req.ParseMultipartForm(maxMemory)
	})
	if err != nil {
		return nil, err
	}
	return &MultipartForm{
			Form: Form{
				values: r.req.MultipartForm.Value,
			}},
		nil
}
