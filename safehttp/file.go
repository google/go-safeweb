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
	"mime"
	"net/http"
	"path/filepath"
)

// FileResponse is a file response. It is a wrapper around the http.FileServer.
type FileResponse struct {
	Root    http.FileSystem
	Request *IncomingRequest
}

// ContentType returns the Content-Type based solely on the extension of the
// filepath. If none is found, the Content-Type is set to
// "application/octet-stream; charset=binary".
func (fs *FileResponse) ContentType() string {
	ct := mime.TypeByExtension(filepath.Ext(fs.Request.URL.Path()))
	if ct == "" {
		return "application/octet-stream; charset=binary"
	}
	return ct
}

// Writes writes the response. Write does not override the Content-Type when
// successfuly serving files. It does override it for redirects and errors.
//
// To safely set the Content-Type based on the filepath extension, set the
// header to the result of ContentType, before calling Write.
func (fs *FileResponse) Write(rw http.ResponseWriter) {
	// Note: FileServer does not override the Content-Type when successfuly
	// serving files. It does override it for redirects and errors.
	http.FileServer(fs.Root).ServeHTTP(rw, fs.Request.req)
}

// ServeFiles is a wrapper around the http.FileServer.
func ServeFiles(root http.FileSystem) Handler {
	return HandlerFunc(func(rw ResponseWriter, req *IncomingRequest) Result {
		return rw.Write(FileResponse{root, req})
	})
}
