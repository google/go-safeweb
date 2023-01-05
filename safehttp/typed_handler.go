package safehttp

type TypedHandler[CustomResponseWriter ResponseWriter] interface {
	ServeHTTP(CustomResponseWriter, *IncomingRequest) Result
}

type HandlerAdapter[CustomResponseWriter ResponseWriter] struct {
	CustomResponseWriterFactory func(ResponseWriter) CustomResponseWriter
}

func (ha *HandlerAdapter[CustomResponseWriter]) Adapt(h TypedHandler[CustomResponseWriter]) Handler {
	return &adaptedHandler[CustomResponseWriter]{rwFactory: ha.CustomResponseWriterFactory, h: h}
}

func (ha *HandlerAdapter[CustomResponseWriter]) AdaptFunc(f func(CustomResponseWriter, *IncomingRequest) Result) HandlerFunc {
	return HandlerFunc(func(w ResponseWriter, req *IncomingRequest) Result {
		return f(ha.CustomResponseWriterFactory(w), req)
	})
}

type adaptedHandler[CustomResponseWriter ResponseWriter] struct {
	rwFactory func(ResponseWriter) CustomResponseWriter
	h TypedHandler[CustomResponseWriter]
}

func (ah *adaptedHandler[CustomResponseWriter]) ServeHTTP(rw ResponseWriter, req *IncomingRequest) Result {
	return ah.h.ServeHTTP(ah.rwFactory(rw), req)
}