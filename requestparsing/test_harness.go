package requestparsing

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
)

type mockHandler struct {
	// called in the ServeHTTP function with the received request.
	callback func(*http.Request)
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.callback(r)
	io.WriteString(w, "Hello world!")
}

type mockListener struct {
	closeOnce      sync.Once
	connChannel    chan net.Conn
	serverEndpoint io.Closer
	clientEndpoint net.Conn
}

// Creates a mock listener that passes requests to the HTTP server as part of
// the test harness
func newMockListener() *mockListener {
	s2c, c2s := net.Pipe()
	c := make(chan net.Conn, 1)
	c <- s2c
	return &mockListener{
		connChannel:    c,
		serverEndpoint: s2c,
		clientEndpoint: c2s,
	}
}

// Passes an endpoint to the server to enable communication to client
func (l *mockListener) Accept() (net.Conn, error) {
	ch, ok := <-l.connChannel
	if !ok {
		return nil, errors.New("Listener closed")
	}
	return ch, nil
}

func (l *mockListener) Close() (err error) {
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

func (l *mockListener) Addr() net.Addr {
	return l.clientEndpoint.LocalAddr()
}

// SendRequest writes 'request' to the clientEndpoint connection which will //send the request to the server listening on this listener.
// Blocks until the server has read the message.
func (l *mockListener) SendRequest(request []byte) error {
	wrote, err := l.clientEndpoint.Write(request)
	if requestLen := len(request); wrote != requestLen {
		return errors.New("client connection failed to write the entire request")
	}
	return err
}

// ReadResponse reads the response from the clientEndpoint connection which is
// sent by the listening server. Blocks until the server has sent its response
// or times out.
func (l *mockListener) ReadResponse() ([]byte, error) {
	// TODO(maramihali@, grenfeldt@): refactor this
	bytes := make([]byte, 4096)
	n, err := l.clientEndpoint.Read(bytes)
	if n == 4096 {
		return nil, errors.New("response larger than 4096 bytes")
	}
	return bytes[:n], err
}

func makeRequest(ctx context.Context, req []byte, callbackFun func(*http.Request)) ([]byte, error) {

	listener := newMockListener()
	defer listener.Close()

	handler := &mockHandler{callback: callbackFun}
	server := &http.Server{Handler: handler}
	go server.Serve(listener)
	defer server.Close()

	if err := listener.SendRequest(req); err != nil {
		return nil, err
	}

	respBody, err := listener.ReadResponse()
	if err != nil {
		return nil, err
	}

	return respBody, server.Shutdown(ctx)
}
