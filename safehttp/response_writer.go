package safehttp

// ResponseWriter TODO
type ResponseWriter struct {
	d Dispatcher
}

// Result TODO
type Result struct{}

// Write TODO
func (w *ResponseWriter) Write(resp Response) Result {
	err := w.d.Write(resp)
	if err != nil {
		panic("error")
	}
	return Result{}
}

// WriteTemplate TODO
func (w *ResponseWriter) WriteTemplate(t Template, data interface{}) Result {
	err := w.d.ExecuteTemplate(t, data)
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
	Write(resp Response) error
	ExecuteTemplate(t Template, data interface{}) error
}
