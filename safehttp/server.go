// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   https://www.apache.org/licenses/LICENSE-2.0
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
	"errors"
	"net"
	"net/http"
	"time"
)

// Server is a safe wrapper for a standard HTTP server.
// The zero value is safe and ready to use and will apply safe defaults on serving.
// Changing any of the fields after the server has been started is a no-op.
type Server struct {
	// Addr optionally specifies the TCP address for the server to listen on,
	// in the form "host:port". If empty, ":http" (port 80) is used.
	// The service names are defined in RFC 6335 and assigned by IANA.
	// See net.Dial for details of the address format.
	Addr string

	// Mux is the ServeMux to use for the current server. A nil Mux is invalid.
	Mux *ServeMux

	// TODO(empijei): potentially consider exposing ReadHeaderTimeout for
	// fine-grained handling (e.g. websocket endpoints).

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled.
	IdleTimeout time.Duration

	// MaxHeaderBytes controls the maximum number of bytes the
	// server will read parsing the request header's keys and
	// values, including the request line. It does not limit the
	// size of the request body.
	MaxHeaderBytes int

	// TLSConfig optionally provides a TLS configuration for use
	// by ServeTLS and ListenAndServeTLS. Note that this value is
	// cloned on serving, so it's not possible to modify the
	// configuration with methods like tls.Config.SetSessionTicketKeys.
	//
	// When the server is started the cloned configuration will be changed
	// to set the minimum TLS version to 1.2 and to prefer Server Ciphers.
	TLSConfig *tls.Config

	// OnShutdown is a slice of functions to call on Shutdown.
	// This can be used to gracefully shutdown connections that have undergone
	// ALPN protocol upgrade or that have been hijacked.
	// These functions should start protocol-specific graceful shutdown, but
	// should not wait for shutdown to complete.
	OnShutdown []func()

	// DisableKeepAlives controls whether HTTP keep-alives should be disabled.
	DisableKeepAlives bool

	srv     *http.Server
	started bool
}

func (s *Server) buildStd() error {
	if s.started {
		return errors.New("server already started")
	}
	if s.srv != nil {
		// Server was already built
		return nil
	}
	if s.Mux == nil {
		return errors.New("building server without a mux")
	}

	srv := &http.Server{
		Addr:           s.Addr,
		Handler:        s.Mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 10 * 1024,
	}
	if s.ReadTimeout != 0 {
		srv.ReadTimeout = s.ReadTimeout
	}
	if s.WriteTimeout != 0 {
		srv.WriteTimeout = s.WriteTimeout
	}
	if s.IdleTimeout != 0 {
		srv.IdleTimeout = s.IdleTimeout
	}
	if s.MaxHeaderBytes != 0 {
		srv.MaxHeaderBytes = s.MaxHeaderBytes
	}
	if s.TLSConfig != nil {
		cfg := s.TLSConfig.Clone()
		cfg.MinVersion = tls.VersionTLS12
		cfg.PreferServerCipherSuites = true
		srv.TLSConfig = cfg
	}
	for _, f := range s.OnShutdown {
		srv.RegisterOnShutdown(f)
	}
	if s.DisableKeepAlives {
		srv.SetKeepAlivesEnabled(false)
	}
	s.srv = srv
	return nil
}

// Clone returns an unstarted deep copy of Server that can be re-configured and re-started.
func (s *Server) Clone() *Server {
	cln := *s
	cln.started = false
	cln.TLSConfig = s.TLSConfig.Clone()
	cln.srv = nil
	return &cln
}

// ListenAndServe is a wrapper for https://golang.org/pkg/net/http/#Server.ListenAndServe
func (s *Server) ListenAndServe() error {
	if err := s.buildStd(); err != nil {
		return err
	}
	s.started = true
	return s.srv.ListenAndServe()
}

// ListenAndServeTLS is a wrapper for https://golang.org/pkg/net/http/#Server.ListenAndServeTLS
func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	if err := s.buildStd(); err != nil {
		return err
	}
	s.started = true
	return s.srv.ListenAndServeTLS(certFile, keyFile)
}

// Serve is a wrapper for https://golang.org/pkg/net/http/#Server.Serve
func (s *Server) Serve(l net.Listener) error {
	if err := s.buildStd(); err != nil {
		return err
	}
	s.started = true
	return s.srv.Serve(l)
}

// ServeTLS is a wrapper for https://golang.org/pkg/net/http/#Server.ServeTLS
func (s *Server) ServeTLS(l net.Listener, certFile, keyFile string) error {
	if err := s.buildStd(); err != nil {
		return err
	}
	s.started = true
	return s.srv.ServeTLS(l, certFile, keyFile)
}

// Shutdown is a wrapper for https://golang.org/pkg/net/http/#Server.Shutdown
func (s *Server) Shutdown(ctx context.Context) error {
	if !s.started {
		return errors.New("shutting down unstarted server")
	}
	s.srv.SetKeepAlivesEnabled(false)
	return s.srv.Shutdown(ctx)
}

// Close is a wrapper for https://golang.org/pkg/net/http/#Server.Close
func (s *Server) Close() error {
	if !s.started {
		return errors.New("closing unstarted server")
	}
	return s.srv.Close()
}
