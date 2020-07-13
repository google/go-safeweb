package safehttp

import "io"

// Response TODO
type Response interface{}

// Template TODO
type Template interface {
	Execute(wr io.Writer, data interface{}) error
}
