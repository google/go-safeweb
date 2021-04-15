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
//  - /foo?msg=<script>alert(1)</script>
//  - /bar?msg=<script>alert(1)</script>
//  - /secure?msg=<script>alert(1)</script>
//  - /insecure?msg=<script>alert(1)</script>
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

	fooHandleFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		q := r.URL.Query().Get("msg")
		fmt.Fprintf(w, "Foo, %q", html.EscapeString(q))
	}

	barHandleFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		q := r.URL.Query().Get("msg")
		// Insecure: Forgot html.EscapeString. XSS is possible.
		fmt.Fprintf(w, "Bar, %q", q)
	}

	safehttpHandler := func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		form, err := r.URL.Query()
		if err != nil {
			panic(err)
		}
		q := form.String("msg", "defaultstr")
		return w.Write(safehtml.HTMLEscaped(q))
	}

	insecureMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			barHandleFunc(w, r)
			next.ServeHTTP(w, r)
		})
	}

	mc := safehttp.NewServeMuxConfig(nil)
	// Generate http.Handler using a safehttp.ServeMuxConfig.
	httpHandlerFromSafehttp := mc.StdHandler(safehttp.MethodGet, safehttp.HandlerFunc(safehttpHandler))
	// Note: no mux generation through Mux() call.

	http.HandleFunc("/foo", fooHandleFunc)
	http.HandleFunc("/bar", barHandleFunc)
	// Register safehttp handler that was converted to net/http handler.
	// Safe as long as no insecure middleware wraps the handler further.
	http.Handle("/secure", httpHandlerFromSafehttp)
	// Insecure: safehttp handler wrapped by net/http handler cannot provide any guarantees.
	http.Handle("/insecure", insecureMiddleware(httpHandlerFromSafehttp))

	log.Println("Visit http://localhost:8080")
	log.Println("Listening on localhost:8080...")
	http.ListenAndServe("localhost:8080", nil)

}
