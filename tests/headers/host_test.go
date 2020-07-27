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

package headers

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-safeweb/internal/requesttesting"
)

func TestHostHeader(t *testing.T) {
	var tests = []struct {
		name    string
		request []byte
		want    string
	}{
		{
			name: "Basic",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"\r\n"),
			want: "localhost:8080",
		},
		{
			// https://tools.ietf.org/html/rfc7230#section-5.3.2
			name: "AbsoluteFormURL",
			request: []byte("GET http://y.com/asdf HTTP/1.1\r\n" +
				"Host: x.com\r\n" +
				"\r\n"),
			want: "y.com",
		},
		{
			// https://tools.ietf.org/html/rfc7230#section-5.3.3
			name: "AuthorityForm",
			request: []byte("GET y.com:123/asdf HTTP/1.1\r\n" +
				"Host: x.com\r\n" +
				"\r\n"),
			want: "x.com",
		},
		{
			name: "NoDoubleSlash",
			request: []byte("GET http:y.com/asdf HTTP/1.1\r\n" +
				"Host: x.com\r\n" +
				"\r\n"),
			want: "x.com",
		},
		{
			name: "NoSchemaOnlyDoubleSlash",
			request: []byte("GET //y.com/asdf HTTP/1.1\r\n" +
				"Host: x.com\r\n" +
				"\r\n"),
			want: "x.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, func(r *http.Request) {
				if len(r.Header) != 0 {
					t.Errorf("len(r.Header) got: %v want: 0", len(r.Header))
				}

				if r.Host != tt.want {
					t.Errorf("r.Host got: %q want: %q", r.Host, tt.want)
				}
			})
			if err != nil {
				t.Fatalf("MakeRequest() got err: %v", err)
			}

			if got, want := extractStatus(resp), statusOK; got != want {
				t.Errorf("status code got: %q want: %q", got, want)
			}
		})
	}
}

func TestHostHeaderMultiple(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: x.com\r\n" +
		"Host: y.com\r\n" +
		"\r\n")

	resp, err := requesttesting.MakeRequest(context.Background(), request, nil)
	if err != nil {
		t.Fatalf("MakeRequest() got err: %v", err)
	}

	if got, want := extractStatus(resp), statusTooManyHostHeaders; got != want {
		t.Errorf("status code got: %q want: %q", got, want)
	}
}

func TestAbsoluteFormURLInvalidSchema(t *testing.T) {
	// When sending a request using the absolute
	// form as the request target, any schema is currently
	// accepted.
	//
	// The desired behavior would instead be that only
	// http or https are accepted as schemas and that
	// the server responds with a 400 (Bad Request) when
	// it receives anything else.

	request := []byte("GET x://y.com/asdf HTTP/1.1\r\n" +
		"Host: x.com\r\n" +
		"\r\n")

	t.Run("Current behavior", func(t *testing.T) {
		resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
			if len(r.Header) != 0 {
				t.Errorf("len(r.Header) got: %v want: 0", len(r.Header))
			}

			if want := "y.com"; r.Host != want {
				t.Errorf("r.Host got: %q want: %q", r.Host, want)
			}
		})
		if err != nil {
			t.Fatalf("MakeRequest() got err: %v want: nil", err)
		}

		if got, want := extractStatus(resp), statusOK; got != want {
			t.Errorf("status code got: %q want: %q", got, want)
		}
	})

	t.Run("Desired behavior", func(t *testing.T) {
		t.Skip()
		resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
			t.Error("Expected handler to not be called!")
		})
		if err != nil {
			t.Fatalf("MakeRequest() got err: %v want: nil", err)
		}

		if got, want := extractStatus(resp), statusBadRequest; got != want {
			t.Errorf("status code got: %q want: %q", got, want)
		}
	})
}
