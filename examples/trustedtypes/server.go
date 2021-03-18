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
	"log"
	"net"
	"net/http"

	"github.com/google/safehtml/template"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/csp"
	"github.com/google/go-safeweb/safehttp/plugins/htmlinject"
)

func main() {
	host := "localhost"
	port := "8080"
	addr := net.JoinHostPort(host, port)
	var mc safehttp.ServeMuxConfig
	mc.Intercept(csp.Default(""))

	safeTmplSrc, _ := template.TrustedSourceFromConstantDir("", template.TrustedSource{}, "safe.html")
	safeTmpl := template.Must(htmlinject.LoadFiles(nil, htmlinject.LoadConfig{}, safeTmplSrc))
	mc.Handle("/safe", safehttp.MethodGet, safehttp.HandlerFunc(handleTemplate(safeTmpl)))
	
	unsafeTmplSrc, _ := template.TrustedSourceFromConstantDir("", template.TrustedSource{}, "unsafe.html")
	unsafeTmpl := template.Must(htmlinject.LoadFiles(nil, htmlinject.LoadConfig{}, unsafeTmplSrc))
	mc.Handle("/unsafe", safehttp.MethodGet, safehttp.HandlerFunc(handleTemplate(unsafeTmpl)))

	log.Printf("Visit http://%s\n", addr)
	log.Printf("Listening on %s...\n", addr)
	log.Fatal(http.ListenAndServe(addr, mc.Mux()))
}

func handleTemplate(tmpl safehttp.Template) func(w safehttp.ResponseWriter, req *safehttp.IncomingRequest) safehttp.Result {
	return func(w safehttp.ResponseWriter, req *safehttp.IncomingRequest) safehttp.Result {
		return safehttp.ExecuteTemplate(w, tmpl, nil)
	}
}
