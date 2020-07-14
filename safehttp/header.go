package safehttp

import (
	"io"
	"net/http"
	"net/textproto"
)

var disallowedHeaders = map[string]bool{"Set-Cookie": true}

// Header represents the key-value pairs in an HTTP header.
// The keys will be in canonical form, as returned by
// textproto.CanonicalMIMEHeaderKey.
type Header struct {
	wrappedHeader http.Header
	immutable     map[string]bool
}

// NewHeader creates a new empty header.
func NewHeader() Header {
	return Header{wrappedHeader: http.Header{}, immutable: map[string]bool{}}
}

// MarkImmutable marks the header with the name `name` as immutable.
// This header is now read-only. `name` is canonicalized using
// textproto.CanonicalMIMEHeaderKey first.
func (h Header) MarkImmutable(name string) {
	name = textproto.CanonicalMIMEHeaderKey(name)
	if disallowedHeaders[name] {
		return
	}
	h.immutable[name] = true
}

// Set sets the header with the name `name` to the value of `value`.
// `name` is canonicalized using textproto.CanonicalMIMEHeaderKey first.
// If this headers is not immutable, this function removes all other
// values currently associated with this header before setting the new
// value. Returns an error when applied on immutable headers.
func (h Header) Set(name, value string) error {
	name = textproto.CanonicalMIMEHeaderKey(name)
	if disallowedHeaders[name] {
		return headerIsDisallowedError{name: name}
	}
	if h.immutable[name] {
		return headerIsImmutableError{name: name}
	}
	h.wrappedHeader.Set(name, value)
	return nil
}

// Add adds a new header with the name `name` and the value `value` to
// the collection of headers. `name` is canonicalized using
// textproto.CanonicalMIMEHeaderKey first. Returns an error when applied
// on immutable headers.
func (h Header) Add(name, value string) error {
	name = textproto.CanonicalMIMEHeaderKey(name)
	if disallowedHeaders[name] {
		return headerIsDisallowedError{name: name}
	}
	if h.immutable[name] {
		return headerIsImmutableError{name: name}
	}
	h.wrappedHeader.Add(name, value)
	return nil
}

// Del deletes all headers with name `name`. `name` is canonicalized using
// textproto.CanonicalMIMEHeaderKey first. Returns an error when applied
// on immutable headers.
func (h Header) Del(name string) error {
	name = textproto.CanonicalMIMEHeaderKey(name)
	if disallowedHeaders[name] {
		return headerIsDisallowedError{name: name}
	}
	if h.immutable[name] {
		return headerIsImmutableError{name: name}
	}
	h.wrappedHeader.Del(name)
	return nil
}

// Get returns the value of the first header with the name `name`.
// `name` is canonicalized using textproto.CanonicalMIMEHeaderKey first.
// If no header exists with the name `name` then "" is returned.
func (h Header) Get(name string) string {
	return h.wrappedHeader.Get(name)
}

// Values returns all the values of all the headers with the name `name`.
// `name` is canonicalized using textproto.CanonicalMIMEHeaderKey first.
// If no header exists with the name `name` then nil is returned.
func (h Header) Values(name string) []string {
	return h.wrappedHeader.Values(name)
}

// Write writes the headers in wire format to the writer.
func (h Header) Write(writer io.Writer) error {
	return h.wrappedHeader.Write(writer)
}

// WriteSubset writes the headers in wire format to the writer. Excludes
// header names in `exclude` while writing to wire format. Header names in
// `exclude` are first canonicalized using textproto.CanonicalMIMEHeaderKey.
func (h Header) WriteSubset(writer io.Writer, exclude map[string]bool) error {
	newExclude := map[string]bool{}
	for k, v := range exclude {
		newExclude[textproto.CanonicalMIMEHeaderKey(k)] = v
	}
	return h.wrappedHeader.WriteSubset(writer, newExclude)
}

// SetCookie adds the cookie provided as a Set-Cookie header in the header
// collection.
// TODO: Replace http.Cookie with safehttp.Cookie.
func (h Header) SetCookie(cookie *http.Cookie) {
	if v := cookie.String(); v != "" {
		h.wrappedHeader.Add("Set-Cookie", v)
	}
}

type headerIsImmutableError struct {
	name string
}

func (err headerIsImmutableError) Error() string {
	return "The header with name \"" + err.name + "\" is immutable."
}

type headerIsDisallowedError struct {
	name string
}

func (err headerIsDisallowedError) Error() string {
	return "The header with name \"" + err.name + "\" is disallowed."
}
