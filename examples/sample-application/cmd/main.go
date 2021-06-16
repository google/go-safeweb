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

// +build go1.16

package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/google/go-safeweb/safehttp"

	"github.com/google/go-safeweb/examples/sample-application/secure"
	"github.com/google/go-safeweb/examples/sample-application/server"
	"github.com/google/go-safeweb/examples/sample-application/storage"
)

var (
	port = flag.Int("port", 8080, "Port for the HTTP server")
	dev  = flag.Bool("dev", false, "Run in dev mode")
)

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)
	flag.Parse()
	safehttp.UseLocalDev() // TODO(clap): remove this
	if *dev {
		safehttp.UseLocalDev()
	}
	db := storage.NewDB()

	addr := net.JoinHostPort("localhost", strconv.Itoa(*port))
	mux := secure.NewMuxConfig(db, addr).Mux()
	server.Load(db, mux)

	log.Printf("Listening on %q", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
