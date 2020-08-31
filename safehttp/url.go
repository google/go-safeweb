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

package safehttp

import (
	"net/url"
)

// URL represents a parsed URL (technically, a URI reference).
type URL struct {
	url *url.URL
}

// Query parses the query string in the URL and returns a form
// containing its values. The returned error describes the first
// decoding error encountered, if any.
func (u URL) Query() (Form, error) {
	v, err := url.ParseQuery(u.url.RawQuery)
	if err != nil {
		return Form{}, err
	}
	return Form{values: map[string][]string(v)}, nil
}

// String reassembles the URL into a valid URL string.
//
// The method uses the net/url.EscapedPath method to obtain the path.
// See the net/url.EscapedPath method for more details.
func (u URL) String() string {
	// The escaping is perfomed by u.url.String()
	return u.url.String()
}

// Host returns the host or the host:port of the URL.
func (u URL) Host() string {
	return u.url.Host
}

// Hostname returns the host of the URL, stripping any valid
// port number if present.
//
// If the result is enclosed in square brackets, as literal IPv6
// addresses are, the square brackets are removed from the result.
func (u URL) Hostname() string {
	return u.url.Hostname()
}

// Port returns the port part of the URL. If the
// host doesn't contain a valid port number, Port returns an
// empty string.
func (u URL) Port() string {
	return u.url.Port()
}

// Path returns the path of the URL.
//
// Note that the path is stored in decoded form: /%47%6f%2f
// becomes /Go/. A consequence is that it is impossible to tell
// which slashes in the path were slashes in the rawURL and which
// were %2f.
func (u URL) Path() string {
	return u.url.Path
}

// Parse parses a rawURL string into a URL structure.
//
// The rawURl may be relative (a path, without a host) or absolute (starting
// with a scheme). Trying to parse a hostname and path without a scheme is
// invalid but may not necessarily return an error, due to parsing ambiguities.
func Parse(rawurl string) (*URL, error) {
	parsed, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	return &URL{url: parsed}, nil
}
