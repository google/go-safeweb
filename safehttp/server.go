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
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// Server is a safe wrapper for a standard HTTP server.
type Server struct {
	srv *http.Server
}

// NewServer constructs a new Server that serves the given multiplexer at the given address.
func NewServer(addr string, mux *ServeMux) Server {
	return Server{
		srv: &http.Server{
			Addr:           addr,
			Handler:        mux,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   5 * time.Second,
			IdleTimeout:    120 * time.Second,
			MaxHeaderBytes: 10 * 1024, // 10KB max of headers
		},
	}
}

// SetTLSConfig adds some safe defaults and sets the underlying server TLS configuration.
func (s Server) SetTLSConfig(cfg *tls.Config) {
	cfg = cfg.Clone()
	cfg.MinVersion = tls.VersionTLS12
	cfg.PreferServerCipherSuites = true
	s.srv.TLSConfig = cfg
}

func (s Server) SetReadTimeout(d time.Duration) {
	s.srv.ReadTimeout = d
}

func (s Server) SetWriteTimeout(d time.Duration) {
	s.srv.WriteTimeout = d
}

func (s Server) SetIdleTimeout(d time.Duration) {
	s.srv.IdleTimeout = d
}

func (s Server) SetMaxHeaderBytes(n int) {
	s.srv.MaxHeaderBytes = n
}

// Close is a wrapper for https://golang.org/pkg/net/http/#Server.Close
func (s Server) Close() error {
	return s.srv.Close()
}

// ListenAndServe is a wrapper for https://golang.org/pkg/net/http/#Server.ListenAndServe
func (s Server) ListenAndServe() error {
	return s.srv.ListenAndServe()
}

// ListenAndServeTLS is a wrapper for https://golang.org/pkg/net/http/#Server.ListenAndServeTLS
func (s Server) ListenAndServeTLS(certFile, keyFile string) error {
	return s.srv.ListenAndServeTLS(certFile, keyFile)
}

// RegisterOnShutdown is a wrapper for https://golang.org/pkg/net/http/#Server.RegisterOnShutdown
func (s Server) RegisterOnShutdown(f func()) {
	s.srv.RegisterOnShutdown(f)
}

// Serve is a wrapper for https://golang.org/pkg/net/http/#Server.Serve
func (s Server) Serve(l net.Listener) error {
	return s.srv.Serve(l)
}

// ServeTLS is a wrapper for https://golang.org/pkg/net/http/#Server.ServeTLS
func (s Server) ServeTLS(l net.Listener, certFile, keyFile string) error {
	return s.srv.ServeTLS(l, certFile, keyFile)
}

// SetKeepAlivesEnabled is a wrapper for https://golang.org/pkg/net/http/#Server.SetKeepAlivesEnabled
func (s Server) SetKeepAlivesEnabled(v bool) {
	s.srv.SetKeepAlivesEnabled(v)
}

// Shutdown is a wrapper for https://golang.org/pkg/net/http/#Server.Shutdown
func (s Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
