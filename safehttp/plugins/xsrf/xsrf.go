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

package xsrf

import (
	"fmt"
	"github.com/google/go-safeweb/safehttp"

	// TODO(@empijei, @kele, @mattiasgrenfeldt, @mihalimara22): decide whether
	// we want to depend on this package or reimplement thefunctionality
	"golang.org/x/net/xsrftoken"
)

const (
	// TokenKey is the form key used when sending the token as part of POST
	// request.
	TokenKey = "xsrf-token"
)

// StorageService  contains information about the users of the web application,
// including their IDs, needed in generating the XSRF token.
type StorageService interface {
	// GetUserID returns the ID of the user making the request based on the
	// incoming request. If an error occurs, it returns it together with an
	// empty string.
	// TODO(@mihalimara22): add a *safehttp.IncomingRequest as a parameter to
	// this function once the method for this is exported.
	GetUserID() (string, error)
}

// Plugin implements XSRF protection.
// TODO(@mihalimara22): Add Fetch Metadata support
type Plugin struct {
	appKey  string
	storage StorageService
}

// NewPlugin creates a new XSRF plugin.It requires an application key and a
// storage service. The appKey uniquely identifies each registered service and
// should have high entropy. The storage service supports retrieving ID's of the
// application's users. Both the appKey and user ID are used in the XSRF
// token generation algorithm.
func NewPlugin(appKey string, s StorageService) *Plugin {
	return &Plugin{
		appKey:  appKey,
		storage: s,
	}
}

// GenerateToken generates a cryptographically safe XSRF token per user, using
// their ID and the request host and path.
func (p *Plugin) GenerateToken(host string, path string) (string, error) {
	userID, err := p.storage.GetUserID()
	if err != nil {
		return "", fmt.Errorf("couldn't retrive the user ID: %v", err)
	}
	return xsrftoken.Generate(p.appKey, userID, host+path), nil
}

// validateToken validates the XSRF token. This should be present in all
// requests as the value of form parameter xsrf-token.
func (p *Plugin) validateToken(r *safehttp.IncomingRequest) safehttp.StatusCode {
	userID, err := p.storage.GetUserID()
	if err != nil {
		return safehttp.StatusUnauthorized
	}
	f, err := r.PostForm()
	if err != nil {
		mf, err := r.MultipartForm(32 << 20)
		if err != nil {
			return safehttp.StatusBadRequest
		}
		f = &mf.Form
	}
	tok := f.String(TokenKey, "")
	if f.Err() != nil || tok == "" {
		return safehttp.StatusForbidden
	}
	if ok := xsrftoken.Valid(tok, p.appKey, userID, r.Host()+r.Path()); !ok {
		return safehttp.StatusForbidden
	}
	return 0
}

// Before should be executed before directing the request to the handler. The
// function applies checks to the Incoming Request to ensure this is not part
// of a Cross-Site Request Forgery.
func (p *Plugin) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
	if status := p.validateToken(r); status != 0 {
		return w.ClientError(status)
	}
	return safehttp.Result{}
}
