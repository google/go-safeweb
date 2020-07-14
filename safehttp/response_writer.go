package safehttp

import (
	"net/http"
)

// ResponseWriter TODO
type ResponseWriter struct {
	d  Dispatcher
	rw http.ResponseWriter
}

// Result TODO
type Result struct{}

// Write TODO
func (w *ResponseWriter) Write(resp Response) Result {
	err := w.d.Write(w.rw, resp)
	if err != nil {
		panic("error")
	}
	return Result{}
}

// WriteTemplate TODO
func (w *ResponseWriter) WriteTemplate(t Template, data interface{}) Result {
	err := w.d.ExecuteTemplate(w.rw, t, data)
	if err != nil {
		panic("error")
	}
	return Result{}
}

// ServerError TODO
func (w *ResponseWriter) ServerError(code StatusCode, resp Response) Result {
	return Result{}
}

// Dispatcher TODO
type Dispatcher interface {
	Write(rw http.ResponseWriter, resp Response) error
	ExecuteTemplate(rw http.ResponseWriter, t Template, data interface{}) error
}
