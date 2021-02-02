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

// httpcompose implements a simple server which composes net/http and safehttp muxes.
//
// Endpoints:
//  - /foo/test?msg=<script>alert(1)</script>
//  - /bar/test?msg=<script>alert(1)</script>
//  - /safe/test?msg=<script>alert(1)</script>
package main

import (
	"fmt"
	"html"
	"log"
	"net/http"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
)

func main() {
	fooHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		q := r.URL.Query().Get("msg")
		fmt.Fprintf(w, "Foo, %q", html.EscapeString(q))
	}

	barHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		q := r.URL.Query().Get("msg")
		// Insecure: Forgot html.EscapeString. XSS is possible.
		fmt.Fprintf(w, "Bar, %q", q)
	}

	secureHandler := func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		form, err := r.URL.Query()
		if err != nil {
			panic(err)
		}
		q := form.String("msg", "defaultstr")
		return w.Write(safehtml.HTMLEscaped(q))
	}

	fooMux := http.NewServeMux()
	fooMux.HandleFunc("/test", fooHandler)
	barMux := http.NewServeMux()
	barMux.HandleFunc("/test", barHandler)

	mc := &safehttp.ServeMuxConfig{}
	mc.Handle("/test", safehttp.MethodGet, safehttp.HandlerFunc(secureHandler))

	topMux := http.NewServeMux()
	topMux.Handle("/foo/", http.StripPrefix("/foo", fooMux))
	topMux.Handle("/bar/", http.StripPrefix("/bar", barMux))
	topMux.Handle("/safe/", http.StripPrefix("/safe", mc.Mux()))

	log.Println("Visit http://localhost:8080")
	log.Println("Listening on localhost:8080...")
	http.ListenAndServe("localhost:8080", topMux)
}
