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

// Package requesttesting provides a harness and other test utilities for
// verifying the behaviour of the net/http package in Go's standard library.
package requesttesting

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
)

// AssertHandler is used to assert properties about the http.Request that it receives in using a callback function.
type AssertHandler struct {
	callback func(*http.Request)
}

func (h *AssertHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.callback(r)
	io.WriteString(w, "Hello world!")
}

// FakeListener creates a custom listener that avoids opening a socket in order
// to establish communication between an HTTP client and server
type FakeListener struct {
	closeOnce      sync.Once
	channel        chan net.Conn
	serverEndpoint io.Closer
	clientEndpoint net.Conn
}

// NewFakeListener creates an instance of fakeListener.
func NewFakeListener() *FakeListener {
	s2c, c2s := net.Pipe()
	c := make(chan net.Conn, 1)
	c <- s2c
	return &FakeListener{
		channel:        c,
		serverEndpoint: s2c,
		clientEndpoint: c2s,
	}
}

// Accept passes a network connection to the HTTP server to enable bidirectional communication with the client.
// It will return an error if Accept is called after the listener was closed.
func (l *FakeListener) Accept() (net.Conn, error) {
	ch, ok := <-l.channel
	if !ok {
		return nil, errors.New("Listener closed")
	}
	return ch, nil
}

// Close will close the two network connections and the listener.
func (l *FakeListener) Close() error {
	l.closeOnce.Do(func() {
		close(l.channel)
	})
	err := l.serverEndpoint.Close()
	err2 := l.clientEndpoint.Close()
	if err2 != nil {
		return err2
	}
	return err
}

// Addr returns the network address of the client endpoint.
func (l *FakeListener) Addr() net.Addr {
	return l.clientEndpoint.LocalAddr()
}

// SendRequest writes a request to the client endpoint connection. This will be passed to the server through the listener.
// The function blocks until the server has finished reading the message.
func (l *FakeListener) sendRequest(request []byte) error {
	n, err := l.clientEndpoint.Write(request)

	if err != nil {
		return err
	}
	if n != len(request) {
		return errors.New("client connection failed to write the entire request")
	}
	return nil
}

// readResponse reads the response from the clientEndpoint connection, sent by the listening server.
// It will block until the server has sent a response.
func (l *FakeListener) readResponse(bytes []byte) (int, error) {
	return l.clientEndpoint.Read(bytes)
}

// MakeRequest instantiates a new http.Server, sends the request provided as argument and returns the response.
// 'callback' will be called in the http.Handler with the http.Request that the handler receives.
// The size of the response is limited to 4096 bytes. If the response received is larger, an error will be returned.
func MakeRequest(ctx context.Context, req []byte, callback func(*http.Request)) ([]byte, error) {
	listener := NewFakeListener()
	defer listener.Close()

	handler := &AssertHandler{callback: callback}
	server := &http.Server{Handler: handler}
	go server.Serve(listener)
	defer server.Close()

	if err := listener.sendRequest(req); err != nil {
		return nil, err
	}

	resp := make([]byte, 4096)
	n, err := listener.readResponse(resp)
	if err != nil {
		return nil, err
	}
	if n == 4096 {
		return nil, errors.New("response larger than or equal to 4096 bytes")
	}

	return resp[:n], server.Shutdown(ctx)
}
