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

// echo implements a simple echo server which listents on localhost:8080.
//
// Endpoints:
//  - /echo?message=YourMessage
//  - /uptime
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/safehtml"
	"github.com/google/safehtml/template"

	"github.com/google/go-safeweb/examples/echo/security/web"
	"github.com/google/go-safeweb/safehttp"
)

var start time.Time

func main() {
	port := 8080
	addr := fmt.Sprintf("localhost:%d", port)
	m := web.NewMuxConfigDev(port).Mux()

	m.Handle("/echo", safehttp.MethodGet, safehttp.HandlerFunc(echo))
	m.Handle("/uptime", safehttp.MethodGet, safehttp.HandlerFunc(uptime))

	start = time.Now()
	log.Printf("Visit http://%s\n", addr)
	log.Printf("Listening on %s...\n", addr)
	log.Fatal(http.ListenAndServe(addr, m))
}

func echo(w safehttp.ResponseWriter, req *safehttp.IncomingRequest) safehttp.Result {
	q, err := req.URL.Query()
	if err != nil {
		return w.WriteError(safehttp.StatusBadRequest)
	}
	x := q.String("message", "")
	if len(x) == 0 {
		return w.WriteError(safehttp.StatusBadRequest)
	}
	return w.Write(safehtml.HTMLEscaped(x))
}

var uptimeTmpl *template.Template = template.Must(template.New("uptime").Parse(
	`<h1>Uptime: {{ .Uptime }}</h1>
{{- if .EasterEgg }}<h1>You've found an easter egg using "{{ .EasterEgg }}". Congrats!</h1>{{ end -}}`))

func uptime(w safehttp.ResponseWriter, req *safehttp.IncomingRequest) safehttp.Result {
	var x struct {
		Uptime    time.Duration
		EasterEgg string
	}
	x.Uptime = time.Now().Sub(start)

	// Easter egg handling.
	q, err := req.URL.Query()
	if err != nil {
		return w.WriteError(safehttp.StatusBadRequest)
	}
	if egg := q.String("easter_egg", ""); len(egg) != 0 {
		x.EasterEgg = egg
	}

	return safehttp.ExecuteTemplate(w, uptimeTmpl, x)
}
