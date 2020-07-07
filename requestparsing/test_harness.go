package requestparsing

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
)

type fakeHandler struct {
	// callback is an anonymous function that verifies properties of the HTTP
	// request and response provided in the test cases.
	callback func(*http.Request)
}

func (h *fakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.callback(r)
	io.WriteString(w, "Hello world!")
}

type fakeListener struct {
	closeOnce      sync.Once
	connChannel    chan net.Conn
	serverEndpoint io.Closer
	clientEndpoint net.Conn
}

// newFakeListener creates an instance of fakeListener. This will pass requests
// to the HTTP server as part of the testing harness.
func newFakeListener() *fakeListener {
	s2c, c2s := net.Pipe()
	c := make(chan net.Conn, 1)
	c <- s2c
	return &fakeListener{
		connChannel:    c,
		serverEndpoint: s2c,
		clientEndpoint: c2s,
	}
}

// Accept passes a network connection to the HTTP server to
// enable bidirectional communication with the client. It will return an error
// if Accept is called after the listener was closed
func (l *fakeListener) Accept() (net.Conn, error) {
	ch, ok := <-l.connChannel
	if !ok {
		return nil, errors.New("Listener closed")
	}
	return ch, nil
}

// Close will close the two network connections and the listener
func (l *fakeListener) Close() (err error) {
	l.closeOnce.Do(func() {
		err = l.serverEndpoint.Close()
		if err != nil {
			return
		}
		err = l.clientEndpoint.Close()
		if err != nil {
			return
		}
		close(l.connChannel)
	})
	return err
}

// Addr returns the network address of the client endpoint.
func (l *fakeListener) Addr() net.Addr {
	return l.clientEndpoint.LocalAddr()
}

// SendRequest writes a request to the client endpoint connection. This will
// be passed to the server through the listener. The function blocks until the
// server has finished reading the message.
func (l *fakeListener) SendRequest(request []byte) error {
	wrote, err := l.clientEndpoint.Write(request)
	if requestLen := len(request); wrote != requestLen {
		return errors.New("client connection failed to write the entire request")
	}
	return err
}

// readResponse reads the response from the clientEndpoint connection, sent by
// the listening server. It will block until the server has sent a response
// or has timed out.
func (l *fakeListener) readResponse() ([]byte, error) {
	// TODO(maramihali@, grenfeldt@): refactor this
	bytes := make([]byte, 4096)
	n, err := l.clientEndpoint.Read(bytes)
	if n == 4096 {
		return nil, errors.New("response larger than 4096 bytes")
	}
	return bytes[:n], err
}

// makeRequest instantiates a new http.Server, sends the request provided as
// argument and returns the response
func makeRequest(ctx context.Context, req []byte, callbackFun func(*http.Request)) ([]byte, error) {
	listener := newFakeListener()
	defer listener.Close()

	handler := &fakeHandler{callback: callbackFun}
	server := &http.Server{Handler: handler}
	go server.Serve(listener)
	defer server.Close()

	if err := listener.SendRequest(req); err != nil {
		return nil, err
	}

	resp, err := listener.readResponse()
	if err != nil {
		return nil, err
	}

	return resp, server.Shutdown(ctx)
}
