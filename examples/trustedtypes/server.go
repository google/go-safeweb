// Copyright 2022 Google LLC
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
//   - /safe#<script>alert(1)</script>
//   - /unsafe#<script>alert(1)</script>
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
	mb, addr := newServeMuxConfig()
	mux := mb.Mux()

	safeTmpl, _ := loadTemplate("safe.html")
	mux.Handle("/safe", safehttp.MethodGet, safehttp.HandlerFunc(handleTemplate(safeTmpl)))

	unsafeTmpl, _ := loadTemplate("unsafe.html")
	mux.Handle("/unsafe", safehttp.MethodGet, safehttp.HandlerFunc(handleTemplate(unsafeTmpl)))

	log.Printf("Visit http://%s\n", addr)
	log.Printf("Listening on %s...\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func newServeMuxConfig() (*safehttp.ServeMuxConfig, string) {
	host := "localhost"
	port := "8080"
	addr := net.JoinHostPort(host, port)
	mc := safehttp.NewServeMuxConfig(nil)
	for _, i := range csp.Default("") {
		mc.Intercept(i)
	}
	return mc, addr
}

func loadTemplate(src string) (*template.Template, error) {
	tmplSrc, err := template.TrustedSourceFromConstantDir("", template.TrustedSource{}, src)
	if err != nil {
		return nil, err
	}

	tmpl := template.Must(htmlinject.LoadFiles(nil, htmlinject.LoadConfig{}, tmplSrc))
	return tmpl, nil
}

func handleTemplate(tmpl safehttp.Template) func(w safehttp.ResponseWriter, req *safehttp.IncomingRequest) safehttp.Result {
	return func(w safehttp.ResponseWriter, req *safehttp.IncomingRequest) safehttp.Result {
		return safehttp.ExecuteTemplate(w, tmpl, nil)
	}
}
