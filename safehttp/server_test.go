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
	"bufio"
	"bytes"
	"net/http"
	"net/http/httputil"
	"testing"
	"time"

	"github.com/google/go-safeweb/internal/requesttesting"
	"github.com/google/safehtml"
)

func TestServer(t *testing.T) {
	l := requesttesting.NewFakeListener()
	defer l.Close()

	var cfg ServeMuxConfig
	cfg.Handle("/", "GET", HandlerFunc(func(w ResponseWriter, r *IncomingRequest) Result {
		return w.Write(safehtml.HTMLEscaped("response"))
	}))
	rtim := 10 * time.Second
	s := Server{
		Mux:         cfg.Mux(),
		ReadTimeout: rtim,
	}
	go s.Serve(l)
	defer s.Close()

	req := []byte("GET / HTTP/1.1\r\nHost: pkg.go.dev\r\n\r\n")
	if err := l.SendRequest(req); err != nil {
		t.Fatalf("Sending request: %v", err)
	}

	respBuf := make([]byte, 1024)
	n, err := l.ReadResponse(respBuf)
	if err != nil {
		t.Fatalf("Reading response: %v", err)
	}
	respBuf = respBuf[:n]
	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(respBuf)), &http.Request{})
	resp.Header.Del("Date") // This will change every time

	got, err := httputil.DumpResponse(resp, true)
	if err != nil {
		t.Errorf("Cannot dump Respnse: %v", err)
	}
	want := []byte("HTTP/1.1 200 OK\r\nContent-Length: 8\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n\r\nresponse")

	if !bytes.Equal(got, want) {
		t.Errorf("Respnse: got %q, want %q", got, want)
	}
	if s.srv.ReadTimeout != rtim {
		t.Errorf("Builder did not set ReadTimeout: got %v want %v", s.srv.ReadTimeout, rtim)
	}
	if s.srv.WriteTimeout != 5*time.Second {
		t.Errorf("Builder did not set WriteTimeout: got %v want %v", s.srv.WriteTimeout, 5*time.Second)
	}
}
