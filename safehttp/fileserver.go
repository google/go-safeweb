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

package safehttp

import (
	"errors"
	"net/http"
)

// FileServer returns a handler that serves HTTP requests with the contents of
// the file system rooted at root.
func FileServer(root string) Handler {
	fileServer := http.FileServer(http.Dir(root))

	return HandlerFunc(func(rw ResponseWriter, req *IncomingRequest) Result {
		fsrw := &fileServerResponseWriter{flight: rw.(*flight), header: http.Header{}}
		fileServer.ServeHTTP(fsrw, req.req)
		return fsrw.result
	})
}

type fileServerResponseWriter struct {
	flight *flight
	result Result

	// We don't allow direct access to the flight's underlying http.Header. We
	// just copy over the contents on a call to WriteHeader, with the exception
	// of the Content-Type header.
	header http.Header

	// Once WriteHeader is called, any subsequent calls to it are no-ops.
	committed bool

	// If the first call to WriteHeader is not a 200 OK, we call
	// flight.WriteError with a 404 StatusCode and make further calls to Write
	// no-ops in order to not leak information about the filesystem.
	errored bool
}

func (fsrw *fileServerResponseWriter) Header() http.Header {
	return fsrw.header
}

func (fsrw *fileServerResponseWriter) Write(b []byte) (int, error) {
	if !fsrw.committed {
		fsrw.WriteHeader(int(StatusOK))
	}

	if fsrw.errored {
		// Let the framework handle the error.
		return 0, errors.New("discarded")
	}
	return fsrw.flight.rw.Write(b)
}

func (fsrw *fileServerResponseWriter) WriteHeader(statusCode int) {
	if fsrw.committed {
		// We've already committed to a response. The headers and status code
		// were written. Ignore this call.
		return
	}
	fsrw.committed = true

	headers := fsrw.flight.Header()
	ct := "application/octet-stream; charset=utf-8"
	if len(fsrw.header["Content-Type"]) > 0 {
		ct = fsrw.header["Content-Type"][0]
	}
	// Content-Type should have been set by the http.FileServer.
	// Note: Add or Set might panic if a header has been already claimed. This
	// is intended behavior.
	for k, v := range fsrw.header {
		if len(v) == 0 {
			continue
		}
		if k == "Content-Type" {
			// Skip setting the Content-Type. The Dispatcher handles it.
			continue
		}
		if headers.IsClaimed(k) {
			continue
		}
		headers.Del(k)
		for _, vv := range v {
			headers.Add(k, vv)
		}
	}

	if statusCode != int(StatusOK) {
		fsrw.errored = true
		// We are writing 404 for every error to avoid leaking information about
		// the filesystem.
		fsrw.result = fsrw.flight.WriteError(StatusNotFound)
		return
	}

	fsrw.result = fsrw.flight.Write(FileServerResponse{
		Path:        fsrw.flight.req.URL().Path(),
		contentType: ct,
	})
}

// FileServerResponse represents a FileServer response.
type FileServerResponse struct {
	// The URL path.
	Path string

	// private, to not allow modifications
	contentType string
}

// ContentType is the Content-Type of the response.
func (resp FileServerResponse) ContentType() string {
	return resp.contentType
}
