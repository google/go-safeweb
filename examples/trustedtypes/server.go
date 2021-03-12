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

// Implements a simple server presenting DOM XSS protection with Trusted Types.
//
// Endpoints:
//  - /safe#<script>alert(1)</script>
//  - /unsafe#<script>alert(1)</script>
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/safehtml/template"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/csp"
	"github.com/google/go-safeweb/safehttp/plugins/hostcheck"
	"github.com/google/go-safeweb/safehttp/plugins/htmlinject"
)

func main() {
	host := "localhost"
	port := 8080
	addr := fmt.Sprintf("%s:%d", host, port)
	var mc safehttp.ServeMuxConfig
	mc.Intercept(csp.Default(""))
	mc.Intercept(hostcheck.New(addr))

	mc.Handle("/safe", safehttp.MethodGet, safehttp.HandlerFunc(safe))
	mc.Handle("/unsafe", safehttp.MethodGet, safehttp.HandlerFunc(unsafe))

	log.Printf("Visit http://%s:%d\n", host, port)
	log.Printf("Listening on %s:%d...\n", host, port)
	log.Fatal(http.ListenAndServe(addr, mc.Mux()))
}

func safe(w safehttp.ResponseWriter, req *safehttp.IncomingRequest) safehttp.Result {
	tmplSrc, _ := template.TrustedSourceFromConstantDir("", template.TrustedSource{}, "safe.html")
	tmpl, _ := htmlinject.LoadFiles(nil, htmlinject.LoadConfig{}, tmplSrc)
	return safehttp.ExecuteTemplate(w, tmpl, nil)
}

func unsafe(w safehttp.ResponseWriter, req *safehttp.IncomingRequest) safehttp.Result {
	tmplSrc, _ := template.TrustedSourceFromConstantDir("", template.TrustedSource{}, "unsafe.html")
	tmpl, _ := htmlinject.LoadFiles(nil, htmlinject.LoadConfig{}, tmplSrc)
	return safehttp.ExecuteTemplate(w, tmpl, nil)
}
