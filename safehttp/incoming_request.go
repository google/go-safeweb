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
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// IncomingRequest represents an HTTP request received by the server.
type IncomingRequest struct {
	// Header is the collection of HTTP headers.
	//
	// The Host header is removed from this struct and can be retrieved using Host()
	Header Header
	// TLS is set just like this TLS field of the net/http.Request. For more information
	// see https://pkg.go.dev/net/http?tab=doc#Request.
	TLS *tls.ConnectionState
	req *http.Request

	// The fields below are kept as pointers to allow cloning through
	// IncomingRequest.WithContext. Otherwise, we'd need to copy locks.
	postParseOnce      *sync.Once
	multipartParseOnce *sync.Once
}

// NewIncomingRequest creates an IncomingRequest
// from the underlying http.Request.
func NewIncomingRequest(req *http.Request) *IncomingRequest {
	if req == nil {
		return nil
	}
	req = req.WithContext(context.WithValue(req.Context(),
		flightValuesCtxKey{}, flightValues{m: make(map[interface{}]interface{})}))
	return &IncomingRequest{
		req:                req,
		Header:             NewHeader(req.Header),
		TLS:                req.TLS,
		postParseOnce:      &sync.Once{},
		multipartParseOnce: &sync.Once{},
	}
}

// Body returns the request body reader. It is always non-nil but will return
// EOF immediately when no body is present.
func (r *IncomingRequest) Body() io.ReadCloser {
	return r.req.Body
}

// Host returns the host the request is targeted to. This value comes from the
// Host header.
func (r *IncomingRequest) Host() string {
	return r.req.Host
}

// Method returns the HTTP method of the IncomingRequest.
func (r *IncomingRequest) Method() string {
	return r.req.Method
}

// PostForm parses the form parameters provided in the body of a POST, PATCH or
// PUT request that does not have Content-Type: multipart/form-data. It returns
// the parsed form parameters as a Form object. If a parsing
// error occurs it will return it, together with a nil Form. Unless we expect
// the header Content-Type: multipart/form-data in a POST request, this method
// should  always be used for forms in POST requests.
func (r *IncomingRequest) PostForm() (*Form, error) {
	var err error
	r.postParseOnce.Do(func() {
		if m := r.req.Method; m != MethodPost && m != MethodPatch && m != MethodPut {
			err = fmt.Errorf("got request method %s, want POST/PATCH/PUT", m)
			return
		}

		if ct := r.req.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			err = fmt.Errorf("invalid method called for Content-Type: %s", ct)
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
// returns a MultipartForm object containing the parsed form parameters and
// file uploads (if any) or the parsing error together with a nil MultipartForm
// otherwise.
//
// If the parsed request body is larger than maxMemory, up to maxMemory bytes
// will be stored in main memory, with the rest stored on disk in temporary
// files.
func (r *IncomingRequest) MultipartForm(maxMemory int64) (*MultipartForm, error) {
	var err error
	r.multipartParseOnce.Do(func() {
		if m := r.req.Method; m != MethodPost && m != MethodPatch && m != MethodPut {
			err = fmt.Errorf("got request method %s, want POST/PATCH/PUT", m)
			return
		}

		if ct := r.req.Header.Get("Content-Type"); !strings.HasPrefix(ct, "multipart/form-data") {
			err = fmt.Errorf("invalid method called for Content-Type: %s", ct)
			return
		}
		err = r.req.ParseMultipartForm(maxMemory)
	})
	if err != nil {
		return nil, err
	}
	return newMulipartForm(r.req.MultipartForm), nil
}

// Cookie returns the named cookie provided in the request or
// net/http.ErrNoCookie if not found. If multiple cookies match the given name,
// only one cookie will be returned.
func (r *IncomingRequest) Cookie(name string) (*Cookie, error) {
	c, err := r.req.Cookie(name)
	if err != nil {
		return nil, err
	}
	return &Cookie{wrapped: c}, nil
}

// Cookies parses and returns the HTTP cookies sent with the request.
func (r *IncomingRequest) Cookies() []*Cookie {
	cl := r.req.Cookies()
	res := make([]*Cookie, 0, len(cl))
	for _, c := range cl {
		res = append(res, &Cookie{wrapped: c})
	}
	return res
}

// Context returns the context of a safehttp.IncomingRequest. This is always
// non-nil and will default to the background context. The context of a
// safehttp.IncomingRequest is the context of the underlying http.Request.
//
// The context is cancelled when the client's connection
// closes, the request is canceled (with HTTP/2), or when the ServeHTTP method
// returns.
func (r *IncomingRequest) Context() context.Context {
	return r.req.Context()
}

// WithContext returns a shallow copy of the request with its context changed to
// ctx. The provided ctx must be non-nil.
//
// This is similar to the net/http.Request.WithContext method.
func (r *IncomingRequest) WithContext(ctx context.Context) *IncomingRequest {
	r2 := new(IncomingRequest)
	*r2 = *r
	r2.req = r2.req.WithContext(ctx)
	return r2
}

// URL specifies the URL that is parsed from the Request-Line. For most requests,
// only URL.Path() will return a non-empty result. (See RFC 7230, Section 5.3)
func (r *IncomingRequest) URL() *URL {
	return &URL{url: r.req.URL}
}

func rawRequest(r *IncomingRequest) *http.Request {
	return r.req
}
