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

package xsrf_test

import (
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/xsrf"
	"github.com/google/safehtml"
	"github.com/google/safehtml/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testDispatcher struct{}

func (testDispatcher) Write(rw http.ResponseWriter, resp safehttp.Response) error {
	switch x := resp.(type) {
	case safehtml.HTML:
		_, err := rw.Write([]byte(x.String()))
		return err
	default:
		panic("not a safe response type")
	}
}

func (testDispatcher) ExecuteTemplate(rw http.ResponseWriter, t safehttp.Template, data interface{}) error {
	switch x := t.(type) {
	case *template.Template:
		return x.Execute(rw, data)
	default:
		panic("not a safe response type")
	}
}

type responseRecorder struct {
	header http.Header
	writer io.Writer
	status int
}

func newResponseRecorder(w io.Writer) *responseRecorder {
	return &responseRecorder{
		header: http.Header{},
		writer: w,
		status: http.StatusOK,
	}
}

func (r *responseRecorder) Header() http.Header {
	return r.header
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	return r.writer.Write(data)
}

type testStorageService struct{}

func (s testStorageService) GetUserID() (string, error) {
	return "potato", nil
}

func TestXSRFToken(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		host       string
		path       string
		wantStatus int
	}{
		{
			name:       "Valid token",
			target:     "http://foo.com/pizza",
			host:       "foo.com",
			path:       "/pizza",
			wantStatus: 200,
		},
		{
			name:       "Invalid host in token generation",
			target:     "http://foo.com/pizza",
			host:       "bar.com",
			path:       "/pizza",
			wantStatus: 403,
		},
		{
			name:       "Invalid path in token generation",
			target:     "http://foo.com/pizza",
			host:       "foo.com",
			path:       "/spaghetti",
			wantStatus: 403,
		},
	}
	for _, test := range tests {
		p := xsrf.NewPlugin("1234", testStorageService{})
		tok, err := p.GenerateToken(test.host, test.path)
		if err != nil {
			t.Fatalf("p.GenerateToken: got %v, want nil", err)
		}
		b := strings.NewReader(xsrf.TokenKey + "=" + tok)
		req := httptest.NewRequest("POST", test.target, b)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		m := safehttp.NewMachinery(p.Before, &testDispatcher{})
		rec := newResponseRecorder(&strings.Builder{})
		m.HandleRequest(rec, req)
		if want := test.wantStatus; rec.status != want {
			t.Errorf("response status: got %v, want %v", rec.status, want)
		}

	}
}
