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
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/google/go-safeweb/internal/requesttesting"

	"github.com/google/go-cmp/cmp"
)

func TestContentLengthTransferEncoding(t *testing.T) {
	type testWant struct {
		headers          map[string][]string
		contentLength    int64
		transferEncoding []string
		body             string
	}

	var tests = []struct {
		name    string
		request []byte
		want    testWant
	}{
		{
			name: "ContentLength",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 5\r\n" +
				"\r\n" +
				"ABCDE\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{"Content-Length": []string{"5"}},
				contentLength:    5,
				transferEncoding: nil,
				body:             "ABCDE",
			},
		},
		{
			name: "ContentButNoContentLength",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"\r\n" +
				"ABCDE\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{},
				contentLength:    0,
				transferEncoding: nil,
				body:             "",
			},
		},
		{
			name: "TransferEncodingChunked",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n\r\n"),
			want: testWant{
				headers:          map[string][]string{},
				contentLength:    -1,
				transferEncoding: []string{"chunked"},
				body:             "ABCDE",
			},
		},
		{
			name: "ContentLengthAndTransferEncoding",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 11\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{},
				contentLength:    -1,
				transferEncoding: []string{"chunked"},
				body:             "ABCDE",
			},
		},
		{
			name: "TransferEncodingListChunkedIdentity",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 11\r\n" +
				"Transfer-Encoding: chunked, identity\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{},
				contentLength:    -1,
				transferEncoding: []string{"chunked"},
				body:             "ABCDE",
			},
		},
		{
			name: "TransferEncodingCasingOrdering1",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"transfer-Encoding: identity\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{},
				contentLength:    -1,
				transferEncoding: []string{"chunked"},
				body:             "ABCDE",
			},
		},
		{
			name: "TransferEncodingCasingOrdering2",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"transfer-Encoding: chunked\r\n" +
				"Transfer-Encoding: identity\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: testWant{
				headers:          map[string][]string{},
				contentLength:    -1,
				transferEncoding: []string{"chunked"},
				body:             "ABCDE",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, func(r *http.Request) {
				if diff := cmp.Diff(tt.want.headers, map[string][]string(r.Header)); diff != "" {
					t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(tt.want.transferEncoding, r.TransferEncoding); diff != "" {
					t.Errorf("r.TransferEncoding mismatch (-want +got):\n%s", diff)
				}

				if r.ContentLength != tt.want.contentLength {
					t.Errorf("r.ContentLength got: %v want: %v", r.ContentLength, tt.want.contentLength)
				}

				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("ioutil.ReadAll(r.Body) got err: %v", err)
				}

				if got := string(body); got != tt.want.body {
					t.Errorf("r.Body got: %q want: %q", got, tt.want.body)
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

func TestContentLengthTransferEncodingStatusMessages(t *testing.T) {
	var tests = []struct {
		name    string
		request []byte
		want    string
	}{
		{
			name: "MultipleContentLength",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 5\r\n" +
				"Content-Length: 14\r\n" +
				"\r\n" +
				"ABCDE\r\n" +
				"\r\n" +
				"ABCDE\r\n" +
				"\r\n"),
			want: statusBadRequest,
		},
		{
			name: "MultipleTransferEncodingChunkedSecond",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 11\r\n" +
				"Transfer-Encoding: asdf\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: statusNotImplemented,
		},
		{
			name: "TransferEncodingListChunkedFirst",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 11\r\n" +
				"Transfer-Encoding: chunked, xyz, asdf\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: statusNotImplemented,
		},
		{
			name: "TransferEncodingListChunkedChunked",
			request: []byte("GET / HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 11\r\n" +
				"Transfer-Encoding: chunked, chunked\r\n" +
				"\r\n" +
				"5\r\n" +
				"ABCDE\r\n" +
				"0\r\n" +
				"\r\n"),
			want: statusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := requesttesting.MakeRequest(context.Background(), tt.request, nil)
			if err != nil {
				t.Fatalf("MakeRequest() got err: %v", err)
			}

			if got := extractStatus(resp); got != tt.want {
				t.Errorf("status code got: %q want: %q", got, tt.want)
			}
		})
	}
}

func TestTransferEncodingChunkSizeLength(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n" +
		"Host:localhost\r\n" +
		"Transfer-Encoding: chunked\r\n" +
		"\r\n" +
		"000000000000000000007\r\n" +
		"\r\n" +
		"ABCDE\r\n" +
		"0\r\n" +
		"\r\n")

	_, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
		if diff := cmp.Diff(map[string][]string{}, map[string][]string(r.Header)); diff != "" {
			t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]string{"chunked"}, r.TransferEncoding); diff != "" {
			t.Errorf("r.TransferEncoding mismatch (-want +got):\n%s", diff)
		}

		if r.ContentLength != -1 {
			t.Errorf("r.ContentLength got: %v want: -1", r.ContentLength)
		}

		_, err := ioutil.ReadAll(r.Body)
		if err == nil {
			t.Fatalf("ioutil.ReadAll(r.Body) got: %v want: http chunk length too large", err)
		}
	})
	if err != nil {
		t.Fatalf("MakeRequest() got err: %v", err)
	}
}

func TestMultipleTransferEncodingChunkedFirst(t *testing.T) {
	// When faced with two Transfer-Encoding headers, net/http
	// currently only looks at the first one and ignores the
	// rest.
	//
	// RFC 7230 doesn’t say anything specific about
	// having multiple TE headers. It only says that, in general
	// a recipient may combine the value of multiple headers
	// with the same name:
	// “ [...] A recipient MAY combine multiple header fields
	//   with the same field name into one "field-name: field-value"
	//   pair, without changing the semantics of the message, by
	//   appending each subsequent field value to the combined
	//   field value in order, separated by a comma.  The order in
	//   which header fields with the same field name are received
	//   is therefore significant to the interpretation of the
	//   combined field value; a proxy MUST NOT change the order
	//   of these field values when forwarding a message. [...] “
	// - RFC 7230 Section 3.2.4
	//
	// net/http should probably reject requests with two TE headers
	// just as it does with requests having two CL headers.
	// There is no legitimate reason for having two TE headers
	// since the TE header already supports having multiple
	// encodings in the same header delimited by commas.
	//
	// The desired behavior would therefore be to respond with
	// 400 (Bad Request) when a request with two TE headers
	// is received.

	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Length: 11\r\n" +
		"Transfer-Encoding: chunked\r\n" +
		"Transfer-Encoding: asdf\r\n" +
		"\r\n" +
		"5\r\n" +
		"ABCDE\r\n" +
		"0\r\n" +
		"\r\n")

	t.Run("Current behavior", func(t *testing.T) {
		resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
			if diff := cmp.Diff(map[string][]string{}, map[string][]string(r.Header)); diff != "" {
				t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff([]string{"chunked"}, r.TransferEncoding); diff != "" {
				t.Errorf("r.TransferEncoding mismatch (-want +got):\n%s", diff)
			}

			if want := int64(-1); r.ContentLength != want {
				t.Errorf("r.ContentLength got: %v want: %v", r.ContentLength, want)
			}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ioutil.ReadAll(r.Body) got err: %v", err)
			}

			if got, want := string(body), "ABCDE"; got != want {
				t.Errorf("r.Body got: %q want: %q", got, want)
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

func TestTransferEncodingIdentity(t *testing.T) {
	// net/http currently accepts the identity transfer encoding and
	// interprets it as if there was no transfer encoding at all.
	//
	// But RFC 7230 has deprecated and removed the identity encoding.
	// The net/http package should therefore remove support for this
	// encoding. The Desired behavior is therefore to return a
	// 501 (Not Implemented).

	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Length: 11\r\n" +
		"Transfer-Encoding: identity\r\n" +
		"\r\n" +
		"5\r\n" +
		"ABCDE\r\n" +
		"0\r\n" +
		"\r\n")

	t.Run("Current behavior", func(t *testing.T) {
		resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
			wantHeaders := map[string][]string{"Content-Length": []string{"11"}}
			if diff := cmp.Diff(wantHeaders, map[string][]string(r.Header)); diff != "" {
				t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff([]string(nil), r.TransferEncoding); diff != "" {
				t.Errorf("r.TransferEncoding mismatch (-want +got):\n%s", diff)
			}

			if want := int64(11); r.ContentLength != want {
				t.Errorf("r.ContentLength got: %v want: %v", r.ContentLength, want)
			}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ioutil.ReadAll(r.Body) got err: %v", err)
			}

			if got, want := string(body), "5\r\nABCDE\r\n0"; got != want {
				t.Errorf("r.Body got: %q want: %q", got, want)
			}
		})
		if err != nil {
			t.Fatalf("MakeRequest() got err: %v", err)
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
			t.Fatalf("MakeRequest() got err: %v", err)
		}

		if got, want := extractStatus(resp), statusNotImplemented; got != want {
			t.Errorf("status code got: %q want: %q", got, want)
		}
	})
}

func TestTransferEncodingListIdentityFirst(t *testing.T) {
	// There is a bug in the handling of lists of transfer
	// codings that contain the identity encoding. Any transfer
	// coding that is put after identity in the list of
	// transfer codings is ignored. Even if it doesn't
	// exists.
	//
	// The desired behavior would instead be to return
	// 501 (Not Implemented) when faced with an unrecognized
	// transfer coding.

	request := []byte("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Content-Length: 11\r\n" +
		"Transfer-Encoding: identity, xyz, asdf\r\n" +
		"\r\n" +
		"5\r\n" +
		"ABCDE\r\n" +
		"0\r\n" +
		"\r\n")

	t.Run("Current behavior", func(t *testing.T) {
		resp, err := requesttesting.MakeRequest(context.Background(), request, func(r *http.Request) {
			wantHeaders := map[string][]string{"Content-Length": []string{"11"}}
			if diff := cmp.Diff(wantHeaders, map[string][]string(r.Header)); diff != "" {
				t.Errorf("r.Header mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff([]string(nil), r.TransferEncoding); diff != "" {
				t.Errorf("r.TransferEncoding mismatch (-want +got):\n%s", diff)
			}

			if want := int64(11); r.ContentLength != want {
				t.Errorf("r.ContentLength got: %v want: %v", r.ContentLength, want)
			}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ioutil.ReadAll(r.Body) got err: %v", err)
			}

			if got, want := string(body), "5\r\nABCDE\r\n0"; got != want {
				t.Errorf("r.Body got: %q want: %q", got, want)
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

		if got, want := extractStatus(resp), statusNotImplemented; got != want {
			t.Errorf("status code got: %q want: %q", got, want)
		}
	})
}
