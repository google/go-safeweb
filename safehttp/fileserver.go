package safehttp

import (
	"errors"
	"net/http"
)

// FileServer returns a handler that serves HTTP requests with the contents of
// the file system rooted at root.
func FileServer(root string) Handler {
	fileServer := http.FileServer(http.Dir(root))

	return HandlerFunc(func(rw ResponseWriter, req *IncomingRequest) Result {
		fileServerRW := &fileServerResponseWriter{flight: rw.(*flight), header: http.Header{}}
		fileServer.ServeHTTP(fileServerRW, req.req)
		return fileServerRW.result
	})
}

type fileServerResponseWriter struct {
	flight *flight
	result Result

	// We don't allow direct access to the flight's underlying http.Header. We
	// just copy over the contents on a call to WriteHeader, with the exception
	// of the Content-Type header.
	header http.Header

	// Once WriteHeader is called, any subsequent calls to it are no-ops.
	committed bool

	// If the first call to WriteHeader is not a 200 OK, we call
	// flight.WriteError with a 404 StatusCode and make further calls to Write
	// no-ops in order to not leak information about the filesystem.
	errored bool
}

func (fileServerRW *fileServerResponseWriter) Header() http.Header {
	return fileServerRW.header
}

func (fileServerRW *fileServerResponseWriter) Write(b []byte) (int, error) {
	if !fileServerRW.committed {
		fileServerRW.WriteHeader(int(StatusOK))
	}

	if fileServerRW.errored {
		// Let the framework handle the error
		return 0, errors.New("discarded")
	}
	return fileServerRW.flight.rw.Write(b)
}

func (fileServerRW *fileServerResponseWriter) WriteHeader(statusCode int) {
	if fileServerRW.committed {
		// We've already committed to a response. The headers and status code
		// were written. Ignore this call.
		return
	}
	fileServerRW.committed = true

	// Note: Add or Set might panic if a header has been already claimed. This
	// is intended behavior.
	headers := fileServerRW.flight.Header()
	for k, v := range fileServerRW.header {
		if len(v) == 0 {
			continue
		}
		if k == "Content-Type" {
			// Skip setting the Content-Type. The Dispatcher handles it.
			continue
		}
		headers.Del(k)
		for _, vv := range v {
			headers.Add(k, vv)
		}
	}

	if statusCode != int(StatusOK) {
		fileServerRW.errored = true
		// We are writing 404 for every error to avoid leaking information about
		// the filesystem.
		fileServerRW.result = fileServerRW.flight.WriteError(StatusNotFound)
		return
	}

	fileServerRW.result = fileServerRW.flight.Write(FileServerResponse{
		Path:        fileServerRW.flight.req.URL.Path(),
		contentType: contentType(fileServerRW.header),
	})
}

// FileServerResponse represents a FileServer response.
type FileServerResponse struct {
	// The URL path.
	Path string

	// private, to not allow modifications
	contentType string
}

// ContentType is the Content-Type of the response.
func (resp FileServerResponse) ContentType() string {
	return resp.contentType
}

func contentType(h http.Header) string {
	if len(h["Content-Type"]) > 0 {
		return h["Content-Type"][0]
	}
	// Content-Type should have been set by the http.FileServer.
	return "application/octet-stream; charset=utf-8"
}
