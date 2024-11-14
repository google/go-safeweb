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

//go:build go1.16
// +build go1.16

package safehttp

import (
	"embed"
	"net/http"
)

func FileServerEmbed(fs embed.FS) Handler {
	fileServer := http.FileServer(http.FS(fs))
	return HandlerFunc(func(rw ResponseWriter, req *IncomingRequest) Result {
		fsrw := &fileServerResponseWriter{flight: rw.(*flight), header: http.Header{}}
		fileServer.ServeHTTP(fsrw, req.req)
		return fsrw.result
	})
}
