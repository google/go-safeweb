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

// Package web is an example package maintained by security experts in a
// development team.
//
// This makes it possible to restrict the usage of net/http package methods used
// for starting an HTTP server, providing a safe way to do it instead.
package web

import (
	"fmt"
	"net/http"

	"github.com/google/go-safeweb/safehttp/plugins/coop"
	"github.com/google/go-safeweb/safehttp/plugins/cors"
	"github.com/google/go-safeweb/safehttp/plugins/csp"
	"github.com/google/go-safeweb/safehttp/plugins/fetchmetadata"
	"github.com/google/go-safeweb/safehttp/plugins/hostcheck"
	"github.com/google/go-safeweb/safehttp/plugins/hsts"
	"github.com/google/go-safeweb/safehttp/plugins/staticheaders"

	"github.com/google/go-safeweb/safehttp"
)

// NewMuxConfig returns a ServeMuxConfig with a set of interceptors already
// installed for security reasons.
// These include:
//
//  - Cross-Origin-Opener-Policy
//  - Content-Security-Policy
//  - Fetch Metadata
//
// Warning: XSRF protection is currently missing due to
// https://github.com/google/go-safeweb/issues/171.
func NewMuxConfig() *safehttp.ServeMuxConfig {
	c := &safehttp.ServeMuxConfig{}

	c.Intercept(coop.Default(""))
	c.Intercept(csp.Default(""))
	c.Intercept(fetchmetadata.NewInterceptor(""))
	c.Intercept(staticheaders.Interceptor{})
	return c
}

// ListenAndServe starts an HTTP server on the given address.
// In addition to the security mechanisms applied by NewMuxConfig, it also adds:
//
//  - HSTS
//  - CORS
//  - Host checking (against DNS rebinding and request smuggling)
//
// If you need to start an HTTP on a localhost for development purposes, you'll
// likely need to disable some of these protections. Use ListenAndServeDev
// instead.
func ListenAndServe(addr string, mc *safehttp.ServeMuxConfig) error {
	mc.Intercept(hsts.Default())
	mc.Intercept(cors.Default())
	mc.Intercept(hostcheck.New(addr))
	return http.ListenAndServe(addr, mc.Mux())
}

// ListenAndServeDev starts a HTTP server on localhost:port for development
// purposes. Most notably, it doesn't include some of the security mechanisms
// that ListenAndServe provides.
//
// Important: the host checking plugin will accept only requests coming to
// localhost:port, not e.g. 127.0.0.1:port.
func ListenAndServeDev(port int, mc *safehttp.ServeMuxConfig) error {
	addr := fmt.Sprintf("localhost:%d", port)
	mc.Intercept(hostcheck.New(addr))
	return http.ListenAndServe(addr, mc.Mux())
}
