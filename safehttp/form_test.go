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

package safehttp_test

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
)

func TestFormValidInt(t *testing.T) {
	stringMaxInt := strconv.FormatInt(math.MaxInt64, 10)
	tests := []struct {
		name string
		req  *http.Request
		want int64
	}{
		{
			name: "Valid int in POST non-multipart request",
			req: func() *http.Request {
				postReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza="+stringMaxInt))
				postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return postReq
			}(),
			want: math.MaxInt64,
		},
		{
			name: "Valid int in POST multipart request",
			req: func() *http.Request {
				multipartReqBody := "--123\r\n" +
					"Content-Disposition: form-data; name=\"pizza\"\r\n" +
					"\r\n" +
					stringMaxInt + "\r\n" +
					"--123--\r\n"
				multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
				multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
				return multipartReq
			}(),
			want: math.MaxInt64,
		},
	}

	for _, test := range tests {
		mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
		mux.Handle("/", safehttp.MethodPost, safehttp.HandlerFunc(func(rw *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
			var form *safehttp.Form
			if !strings.HasPrefix(ir.Header.Get("Content-Type"), "multipart/form-data") {
				var err error
				form, err = ir.PostForm()
				if err != nil {
					t.Fatalf(`ir.PostForm: got %v, want nil`, err)
				}
			} else {
				mf, err := ir.MultipartForm(32 << 20)
				if err != nil {
					t.Fatalf(`ir.MultipartForm: got %v, want nil`, err)
				}
				form = &mf.Form
			}

			got := form.Int64("pizza", 0)
			if err := form.Err(); err != nil {
				t.Errorf(`form.Error: got %v, want nil`, err)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("form.Int64:  mismatch (-want +got): \n%s", diff)
			}
			return safehttp.NotWritten
		}))

		rec := newResponseRecorder(&strings.Builder{})
		mux.ServeHTTP(rec, test.req)

		if want := safehttp.StatusNoContent; rec.status != want {
			t.Errorf("response status: got %v, want %v", rec.status, want)
		}
	}
}

func TestFormInvalidInt(t *testing.T) {
	tests := []struct {
		name string
		reqs []*http.Request
		want int64
	}{
		{
			name: "Overflow integer in request",
			reqs: []*http.Request{
				func() *http.Request {
					postReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=9223372036854775810"))
					postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					return postReq
				}(),
				func() *http.Request {
					multipartReqBody := "--123\r\n" +
						"Content-Disposition: form-data; name=\"pizza\"\r\n" +
						"\r\n" +
						"9223372036854775810\r\n" +
						"--123--\r\n"
					multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
					multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
					return multipartReq
				}(),
			},
			want: 0,
		},
		{
			name: "Not an integer in request",
			reqs: []*http.Request{
				func() *http.Request {
					postReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=diavola"))
					postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					return postReq

				}(),
				func() *http.Request {
					multipartReqBody := "--123\r\n" +
						"Content-Disposition: form-data; name=\"pizza\"\r\n" +
						"\r\n" +
						"diavola\r\n" +
						"--123--\r\n"
					multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
					multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
					return multipartReq
				}(),
			},
			want: 0,
		},
	}

	for _, test := range tests {
		for _, req := range test.reqs {
			mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
			mux.Handle("/", safehttp.MethodPost, safehttp.HandlerFunc(func(rw *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
				var form *safehttp.Form
				if !strings.HasPrefix(ir.Header.Get("Content-Type"), "multipart/form-data") {
					var err error
					form, err = ir.PostForm()
					if err != nil {
						t.Fatalf(`ir.PostForm: got %v, want nil`, err)
					}
				} else {
					mf, err := ir.MultipartForm(32 << 20)
					if err != nil {
						t.Fatalf(`ir.MultipartForm: got %v, want nil`, err)
					}
					form = &mf.Form
				}

				got := form.Int64("pizza", 0)
				if diff := cmp.Diff(test.want, got); diff != "" {
					t.Errorf("form.Int64:  mismatch (-want +got): \n%s", diff)
				}
				if form.Err() == nil {
					t.Errorf("form.Err: got nil, want error")
				}

				return safehttp.NotWritten
			}))

			rec := newResponseRecorder(&strings.Builder{})
			mux.ServeHTTP(rec, req)

			if want := safehttp.StatusNoContent; rec.status != want {
				t.Errorf("response status: got %v, want %v", rec.status, want)
			}
		}
	}
}

func TestFormValidUint(t *testing.T) {
	stringMaxUint := strconv.FormatUint(math.MaxUint64, 10)
	tests := []struct {
		name string
		req  *http.Request
		want uint64
	}{
		{
			name: "Valid unsigned integer in POST non-multipart request",
			req: func() *http.Request {
				postReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza="+stringMaxUint))
				postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return postReq
			}(),
			want: math.MaxUint64,
		},
		{
			name: "Valid unsigned integer in POST multipart request",
			req: func() *http.Request {
				multipartReqBody := "--123\r\n" +
					"Content-Disposition: form-data; name=\"pizza\"\r\n" +
					"\r\n" +
					stringMaxUint + "\r\n" +
					"--123--\r\n"
				multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
				multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
				return multipartReq
			}(),
			want: math.MaxUint64,
		},
	}

	for _, test := range tests {
		mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
		mux.Handle("/", safehttp.MethodPost, safehttp.HandlerFunc(func(rw *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
			var form *safehttp.Form
			if !strings.HasPrefix(ir.Header.Get("Content-Type"), "multipart/form-data") {
				var err error
				form, err = ir.PostForm()
				if err != nil {
					t.Fatalf(`ir.PostForm: got %v, want nil`, err)
				}
			} else {
				mf, err := ir.MultipartForm(32 << 20)
				if err != nil {
					t.Fatalf(`ir.MultipartForm: got %v, want nil`, err)
				}
				form = &mf.Form
			}

			got := form.Uint64("pizza", 0)
			if err := form.Err(); err != nil {
				t.Errorf(`form.Error: got %v, want nil`, err)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("form.Uint64:  mismatch (-want +got): \n%s", diff)
			}
			return safehttp.NotWritten
		}))

		rec := newResponseRecorder(&strings.Builder{})
		mux.ServeHTTP(rec, test.req)

		if want := safehttp.StatusNoContent; rec.status != want {
			t.Errorf("response status: got %v, want %v", rec.status, want)
		}
	}
}

func TestFormInvalidUint(t *testing.T) {
	tests := []struct {
		name string
		reqs []*http.Request
		want uint64
	}{
		{
			name: "Not an unsigned integer",
			reqs: []*http.Request{
				func() *http.Request {
					postReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=-1"))
					postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					return postReq
				}(),
				func() *http.Request {
					multipartReqBody := "--123\r\n" +
						"Content-Disposition: form-data; name=\"pizza\"\r\n" +
						"\r\n" +
						"-1\r\n" +
						"--123--\r\n"
					multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
					multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
					return multipartReq
				}(),
			},
			want: 0,
		},
		{
			name: "Overflow unsigned integer",
			reqs: []*http.Request{
				func() *http.Request {
					postReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=18446744073709551630"))
					postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					return postReq
				}(),
				func() *http.Request {
					multipartReqBody := "--123\r\n" +
						"Content-Disposition: form-data; name=\"pizza\"\r\n" +
						"\r\n" +
						"18446744073709551630\r\n" +
						"--123--\r\n"
					multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
					multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
					return multipartReq
				}(),
			},
			want: 0,
		},
	}
	for _, test := range tests {
		for _, req := range test.reqs {
			mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
			mux.Handle("/", safehttp.MethodPost, safehttp.HandlerFunc(func(rw *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
				var form *safehttp.Form
				if !strings.HasPrefix(ir.Header.Get("Content-Type"), "multipart/form-data") {
					var err error
					form, err = ir.PostForm()
					if err != nil {
						t.Fatalf(`ir.PostForm: got %v, want nil`, err)
					}
				} else {
					mf, err := ir.MultipartForm(32 << 20)
					if err != nil {
						t.Fatalf(`ir.MultipartForm: got %v, want nil`, err)
					}
					form = &mf.Form
				}

				got := form.Uint64("pizza", 0)
				if diff := cmp.Diff(test.want, got); diff != "" {
					t.Errorf("form.Uint64:  mismatch (-want +got): \n%s", diff)
				}
				if form.Err() == nil {
					t.Errorf("form.Err: got nil, want error")
				}
				return safehttp.NotWritten
			}))

			rec := newResponseRecorder(&strings.Builder{})
			mux.ServeHTTP(rec, req)

			if want := safehttp.StatusNoContent; rec.status != want {
				t.Errorf("response status: got %v, want %v", rec.status, want)
			}
		}
	}
}

func TestFormValidString(t *testing.T) {
	tests := []struct {
		name string
		reqs []*http.Request
		want []string
	}{
		{
			name: "Valid string in POST non-multipart request",
			reqs: func() []*http.Request {
				reqs := []*http.Request{
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=diavola")),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=ăȚâȘî")),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=\x64\x69\x61\x76\x6f\x6c\x61")),
				}
				for _, req := range reqs {
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				}
				return reqs
			}(),
			want: []string{"diavola", "ăȚâȘî", "diavola"},
		},
		{
			name: "Valid string in POST multipart request",
			reqs: func() []*http.Request {
				slice := []string{"diavola", "ăȚâȘî", "diavola"}
				var reqs []*http.Request
				for _, val := range slice {
					multipartReqBody := "--123\r\n" +
						"Content-Disposition: form-data; name=\"pizza\"\r\n" +
						"\r\n" +
						val + "\r\n" +
						"--123--\r\n"
					multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
					multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
					reqs = append(reqs, multipartReq)

				}
				return reqs
			}(),
			want: []string{"diavola", "ăȚâȘî", "diavola"},
		},
	}

	for _, test := range tests {
		for idx, req := range test.reqs {
			mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
			mux.Handle("/", safehttp.MethodPost, safehttp.HandlerFunc(func(rw *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
				var form *safehttp.Form
				if !strings.HasPrefix(ir.Header.Get("Content-Type"), "multipart/form-data") {
					var err error
					form, err = ir.PostForm()
					if err != nil {
						t.Fatalf(`ir.PostForm: got %v, want nil`, err)
					}
				} else {
					mf, err := ir.MultipartForm(32 << 20)
					if err != nil {
						t.Fatalf(`ir.MultipartForm: got %v, want nil`, err)
					}
					form = &mf.Form
				}

				got := form.String("pizza", "")
				if diff := cmp.Diff(test.want[idx], got); diff != "" {
					t.Errorf("form.String: diff (-want +got): \n%s", diff)
				}
				if err := form.Err(); err != nil {
					t.Errorf("form.Err: got %v, want nil", err)
				}
				return safehttp.NotWritten
			}))

			rec := newResponseRecorder(&strings.Builder{})
			mux.ServeHTTP(rec, req)

			if want := safehttp.StatusNoContent; rec.status != want {
				t.Errorf("response status: got %v, want %v", rec.status, want)
			}
		}
	}
}

func TestFormValidFloat64(t *testing.T) {
	stringMaxFloat := strconv.FormatFloat(math.MaxFloat64, 'f', 6, 64)
	stringNegativeFloat := strconv.FormatFloat(-math.SmallestNonzeroFloat64, 'f', 324, 64)
	tests := []struct {
		name string
		reqs []*http.Request
		want []float64
	}{
		{
			name: "Valid floats in POST non-multipart request",
			reqs: func() []*http.Request {
				reqs := []*http.Request{
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza="+stringMaxFloat)),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza="+stringNegativeFloat)),
				}
				for _, req := range reqs {
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				}
				return reqs
			}(),
			want: []float64{math.MaxFloat64, -math.SmallestNonzeroFloat64},
		},
		{
			name: "Valid floats in POST multipart request",
			reqs: func() []*http.Request {
				slice := []string{stringMaxFloat, stringNegativeFloat}
				var reqs []*http.Request
				for _, val := range slice {
					multipartReqBody := "--123\r\n" +
						"Content-Disposition: form-data; name=\"pizza\"\r\n" +
						"\r\n" +
						val + "\r\n" +
						"--123--\r\n"
					multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
					multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
					reqs = append(reqs, multipartReq)

				}
				return reqs
			}(),
			want: []float64{math.MaxFloat64, -math.SmallestNonzeroFloat64},
		},
	}

	for _, test := range tests {
		for idx, req := range test.reqs {
			mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
			mux.Handle("/", safehttp.MethodPost, safehttp.HandlerFunc(func(rw *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
				var form *safehttp.Form
				if !strings.HasPrefix(ir.Header.Get("Content-Type"), "multipart/form-data") {
					var err error
					form, err = ir.PostForm()
					if err != nil {
						t.Fatalf(`ir.PostForm: got %v, want nil`, err)
					}
				} else {
					mf, err := ir.MultipartForm(32 << 20)
					if err != nil {
						t.Fatalf(`ir.MultipartForm: got %v, want nil`, err)
					}
					form = &mf.Form
				}

				got := form.Float64("pizza", 0.0)
				if diff := cmp.Diff(test.want[idx], got); diff != "" {
					t.Errorf("form.Float64: diff (-want +got): \n%s", diff)
				}
				if err := form.Err(); err != nil {
					t.Errorf("form.Err: got %v, want nil", err)
				}
				return safehttp.NotWritten
			}))

			rec := newResponseRecorder(&strings.Builder{})
			mux.ServeHTTP(rec, req)

			if want := safehttp.StatusNoContent; rec.status != want {
				t.Errorf("response status: got %v, want %v", rec.status, want)
			}
		}
	}
}

func TestFormInvalidFloat64(t *testing.T) {
	tests := []struct {
		name string
		reqs []*http.Request
		want float64
	}{
		{
			name: "Not a float64 in request",
			reqs: []*http.Request{
				func() *http.Request {
					postReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=diavola"))
					postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					return postReq
				}(),
				func() *http.Request {
					multipartReqBody := "--123\r\n" +
						"Content-Disposition: form-data; name=\"pizza\"\r\n" +
						"\r\n" +
						"diavola\r\n" +
						"--123--\r\n"
					multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
					multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
					return multipartReq
				}(),
			},
			want: 0.0,
		},
		{
			name: "Overflow float64 in request",
			reqs: []*http.Request{
				func() *http.Request {
					postReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=1.797693134862315708145274237317043567981e309"))
					postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					return postReq
				}(),
				func() *http.Request {
					multipartReqBody := "--123\r\n" +
						"Content-Disposition: form-data; name=\"pizza\"\r\n" +
						"\r\n" +
						"1.797693134862315708145274237317043567981e309\r\n" +
						"--123--\r\n"
					multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
					multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
					return multipartReq
				}(),
			},
			want: 0.0,
		},
	}
	for _, test := range tests {
		for _, req := range test.reqs {
			mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
			mux.Handle("/", safehttp.MethodPost, safehttp.HandlerFunc(func(rw *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
				var form *safehttp.Form
				if !strings.HasPrefix(ir.Header.Get("Content-Type"), "multipart/form-data") {
					var err error
					form, err = ir.PostForm()
					if err != nil {
						t.Fatalf(`ir.PostForm: got %v, want nil`, err)
					}
				} else {
					mf, err := ir.MultipartForm(32 << 20)
					if err != nil {
						t.Fatalf(`ir.MultipartForm: got %v, want nil`, err)
					}
					form = &mf.Form
				}

				got := form.Float64("pizza", 0.0)
				if diff := cmp.Diff(test.want, got); diff != "" {
					t.Errorf("form.Float64:  mismatch (-want +got): \n%s", diff)
				}
				if form.Err() == nil {
					t.Errorf("form.Err: got nil, want error")
				}

				return safehttp.NotWritten
			}))

			rec := newResponseRecorder(&strings.Builder{})
			mux.ServeHTTP(rec, req)

			if want := safehttp.StatusNoContent; rec.status != want {
				t.Errorf("response status: got %v, want %v", rec.status, want)
			}
		}
	}
}

func TestFormValidBool(t *testing.T) {
	tests := []struct {
		name string
		reqs []*http.Request
		want []bool
	}{
		{
			name: "Valid booleans in POST non-multipart request",
			reqs: func() []*http.Request {
				reqs := []*http.Request{
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=true")),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=false")),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=True")),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=False")),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=TRUE")),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=FALSE")),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=t")),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=f")),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=T")),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=F")),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=1")),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=0")),
				}
				for _, req := range reqs {
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				}
				return reqs
			}(),
			want: []bool{true, false, true, false, true, false, true, false, true, false, true, false},
		},
		{
			name: "Valid booleans in POST multipart request",
			reqs: func() []*http.Request {
				slice := []string{"true", "false", "True", "False", "TRUE", "FALSE", "t", "f", "T", "F", "1", "0"}
				var reqs []*http.Request
				for _, val := range slice {
					multipartReqBody := "--123\r\n" +
						"Content-Disposition: form-data; name=\"pizza\"\r\n" +
						"\r\n" +
						val + "\r\n" +
						"--123--\r\n"
					multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
					multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
					reqs = append(reqs, multipartReq)

				}
				return reqs
			}(),
			want: []bool{true, false, true, false, true, false, true, false, true, false, true, false},
		},
	}

	for _, test := range tests {
		for idx, req := range test.reqs {
			mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
			mux.Handle("/", safehttp.MethodPost, safehttp.HandlerFunc(func(rw *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
				var form *safehttp.Form
				if !strings.HasPrefix(ir.Header.Get("Content-Type"), "multipart/form-data") {
					var err error
					form, err = ir.PostForm()
					if err != nil {
						t.Fatalf(`ir.PostForm: got %v, want nil`, err)
					}
				} else {
					mf, err := ir.MultipartForm(32 << 20)
					if err != nil {
						t.Fatalf(`ir.MultipartForm: got %v, want nil`, err)
					}
					form = &mf.Form
				}

				got := form.Bool("pizza", false)
				if diff := cmp.Diff(test.want[idx], got); diff != "" {
					t.Errorf("form.Bool: diff (-want +got): \n%s", diff)
				}
				if err := form.Err(); err != nil {
					t.Errorf("form.Err: got %v, want nil", err)
				}

				return safehttp.NotWritten
			}))

			rec := newResponseRecorder(&strings.Builder{})
			mux.ServeHTTP(rec, req)

			if want := safehttp.StatusNoContent; rec.status != want {
				t.Errorf("response status: got %v, want %v", rec.status, want)
			}
		}
	}
}

func TestFormInvalidBool(t *testing.T) {
	tests := []struct {
		name string
		reqs []*http.Request
		want bool
	}{
		{
			name: "Invalid booleans in POST non-multipart request",
			reqs: func() []*http.Request {
				reqs := []*http.Request{
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=TruE")),
					httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=potato")),
				}
				for _, req := range reqs {
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				}
				return reqs
			}(),
			want: false,
		},
		{
			name: "Invalid booleans in POST multipart request",
			reqs: func() []*http.Request {
				slice := []string{"TruE", "potato"}
				var reqs []*http.Request
				for _, val := range slice {
					multipartReqBody := "--123\r\n" +
						"Content-Disposition: form-data; name=\"pizza\"\r\n" +
						"\r\n" +
						val + "\r\n" +
						"--123--\r\n"
					multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
					multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
					reqs = append(reqs, multipartReq)

				}
				return reqs
			}(),
			want: false,
		},
	}

	for _, test := range tests {
		for _, req := range test.reqs {
			mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
			mux.Handle("/", safehttp.MethodPost, safehttp.HandlerFunc(func(rw *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
				var form *safehttp.Form
				if !strings.HasPrefix(ir.Header.Get("Content-Type"), "multipart/form-data") {
					var err error
					form, err = ir.PostForm()
					if err != nil {
						t.Fatalf(`ir.PostForm: got %v, want nil`, err)
					}
				} else {
					mf, err := ir.MultipartForm(32 << 20)
					if err != nil {
						t.Fatalf(`ir.MultipartForm: got %v, want nil`, err)
					}
					form = &mf.Form
				}

				got := form.Bool("pizza", false)
				if diff := cmp.Diff(test.want, got); diff != "" {
					t.Errorf("form.Bool:  mismatch (-want +got): \n%s", diff)
				}
				if form.Err() == nil {
					t.Errorf("form.Err: got nil, want error")
				}
				return safehttp.NotWritten
			}))

			rec := newResponseRecorder(&strings.Builder{})
			mux.ServeHTTP(rec, req)

			if want := safehttp.StatusNoContent; rec.status != want {
				t.Errorf("response status: got %v, want %v", rec.status, want)
			}
		}
	}
}

func TestFormInvalidSlice(t *testing.T) {
	// Parsing behaviour of native types that are supported is not included as
	// it was tested in previous tests.
	tests := []struct {
		name string
		reqs []*http.Request
		got  interface{}
		want interface{}
	}{
		{
			name: "Request with multiple types in slice",
			reqs: []*http.Request{
				func() *http.Request {
					postReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=true&pizza=1.3"))
					postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					return postReq
				}(),
				func() *http.Request {
					multipartReqBody := "--123\r\n" +
						"Content-Disposition: form-data; name=\"pizza\"\r\n" +
						"\r\n" +
						"true\r\n" +
						"--123\r\n" +
						"Content-Disposition: form-data; name=\"pizza\"\r\n" +
						"\r\n" +
						"1.3\r\n" +
						"--123--\r\n"
					multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
					multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
					return multipartReq
				}(),
			},
		},
		{
			name: "Unsupported slice type",
			reqs: []*http.Request{
				func() *http.Request {
					postReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=true&pizza=1.3"))
					postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					return postReq
				}(),
				func() *http.Request {
					multipartReqBody := "--123\r\n" +
						"Content-Disposition: form-data; name=\"pizza\"\r\n" +
						"\r\n" +
						"true\r\n" +
						"--123\r\n" +
						"Content-Disposition: form-data; name=\"pizza\"\r\n" +
						"\r\n" +
						"1.3\r\n" +
						"--123--\r\n"
					multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
					multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
					return multipartReq
				}(),
			},
		},
	}

	for _, test := range tests {
		for _, req := range test.reqs {
			mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
			mux.Handle("/", safehttp.MethodPost, safehttp.HandlerFunc(func(rw *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
				var form *safehttp.Form
				if !strings.HasPrefix(ir.Header.Get("Content-Type"), "multipart/form-data") {
					var err error
					form, err = ir.PostForm()
					if err != nil {
						t.Fatalf(`ir.PostForm: got %v, want nil`, err)
					}
				} else {
					mf, err := ir.MultipartForm(32 << 20)
					if err != nil {
						t.Fatalf(`ir.MultipartForm: got %v, want nil`, err)
					}
					form = &mf.Form
				}

				form.Slice("pizza", test.got)
				if diff := cmp.Diff(test.want, test.got); diff != "" {
					t.Errorf("form.Slice: diff (-want +got): \n%v", diff)
				}
				if form.Err() == nil {
					t.Errorf("form.Err: got nil, want error")
				}

				return safehttp.NotWritten
			}))

			rec := newResponseRecorder(&strings.Builder{})
			mux.ServeHTTP(rec, req)

			if want := safehttp.StatusNoContent; rec.status != want {
				t.Errorf("response status: got %v, want %v", rec.status, want)
			}
		}
	}
}

func TestFormErrorHandling(t *testing.T) {
	tests := []struct {
		name string
		req  *http.Request
	}{
		{
			name: "Post form errors",
			req: func() *http.Request {
				postReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizzaInt=diavola&pizzaBool=true&pizzaUint=-13"))
				postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return postReq
			}(),
		},
		{
			name: "Multipart form errors",
			req: func() *http.Request {
				multipartReqBody := "--123\r\n" +
					"Content-Disposition: form-data; name=\"pizzaInt\"\r\n" +
					"\r\n" +
					"diavola\r\n" +
					"--123\r\n" +
					"Content-Disposition: form-data; name=\"pizzaBool\"\r\n" +
					"\r\n" +
					"true\r\n" +
					"--123\r\n" +
					"Content-Disposition: form-data; name=\"pizzaUint\"\r\n" +
					"\r\n" +
					"-13\r\n" +
					"--123--\r\n"
				multipartReq := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(multipartReqBody))
				multipartReq.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
				return multipartReq
			}(),
		},
	}

	for _, tc := range tests {
		mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
		mux.Handle("/", safehttp.MethodPost, safehttp.HandlerFunc(func(rw *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
			var form *safehttp.Form
			if !strings.HasPrefix(ir.Header.Get("Content-Type"), "multipart/form-data") {
				var err error
				form, err = ir.PostForm()
				if err != nil {
					t.Fatalf(`ir.PostForm: got %v, want nil`, err)
				}
			} else {
				mf, err := ir.MultipartForm(32 << 20)
				if err != nil {
					t.Fatalf(`ir.MultipartForm: got %v, want nil`, err)
				}
				form = &mf.Form
			}

			gotInt64 := form.Int64("pizzaInt", 0)
			if diff := cmp.Diff(int64(0), gotInt64); diff != "" {
				t.Errorf("form.Int64: diff (-want +got): \n%s", diff)
			}
			prevErr := form.Err()
			if prevErr == nil {
				t.Errorf("form.Err: got nil, want error")
			}

			gotBool := form.Bool("pizzaBool", false)
			if diff := cmp.Diff(true, gotBool); diff != "" {
				t.Errorf("form.Bool: diff (-want +got): \n%s", diff)
			}
			// We expect the same error here becase calling form.Bool succeeds
			if err := form.Err(); prevErr.Error() != err.Error() {
				t.Errorf("form.Err: want %v, got %v", prevErr, err)
			}

			gotUint64 := form.Uint64("pizzaUint", 0)
			if diff := cmp.Diff(uint64(0), gotUint64); diff != "" {
				t.Errorf("form.Uint64: diff (-want +got): \n%s", diff)
			}
			if form.Err() == nil {
				t.Errorf("form.Err: got nil, want error")
			}

			return safehttp.NotWritten
		}))

		rec := newResponseRecorder(&strings.Builder{})
		mux.ServeHTTP(rec, tc.req)

		if want := safehttp.StatusNoContent; rec.status != want {
			t.Errorf("response status: got %v, want %v", rec.status, want)
		}
	}
}

func TestFormValidSlicePost(t *testing.T) {
	tests := []struct {
		req         *http.Request
		placeholder interface{}
		want        interface{}
	}{
		{
			req: func() *http.Request {
				req := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=-8&pizza=9&pizza=-100"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			}(),
			placeholder: &[]int64{},
			want:        &[]int64{-8, 9, -100},
		},
		{
			req: func() *http.Request {
				req := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=8&pizza=9&pizza=10"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			}(),
			placeholder: &[]uint64{},
			want:        &[]uint64{8, 9, 10},
		},
		{
			req: func() *http.Request {
				req := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=margeritta&pizza=diavola&pizza=calzone"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			}(),
			placeholder: &[]string{},
			want:        &[]string{"margeritta", "diavola", "calzone"},
		},
		{
			req: func() *http.Request {
				req := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=1.3&pizza=8.9&pizza=-4.1"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			}(),
			placeholder: &[]float64{},
			want:        &[]float64{1.3, 8.9, -4.1},
		},
		{
			req: func() *http.Request {
				req := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader("pizza=true&pizza=false&pizza=T"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			}(),
			placeholder: &[]bool{},
			want:        &[]bool{true, false, true},
		},
	}

	for _, tc := range tests {
		mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
		mux.Handle("/", safehttp.MethodPost, safehttp.HandlerFunc(func(rw *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
			form, err := ir.PostForm()
			if err != nil {
				t.Fatalf(`ir.PostForm: got err %v, want nil`, err)
			}

			form.Slice("pizza", tc.placeholder)
			if err := form.Err(); err != nil {
				t.Errorf(`form.Err: got err %v, want nil`, err)
			}

			if diff := cmp.Diff(tc.want, tc.placeholder); diff != "" {
				t.Errorf("form.Slice: diff (-want +got): \n%v", diff)
			}

			return safehttp.NotWritten
		}))

		rec := newResponseRecorder(&strings.Builder{})
		mux.ServeHTTP(rec, tc.req)

		if respStatus, want := rec.status, safehttp.StatusNoContent; respStatus != want {
			t.Errorf("response status: got %v, want %v", respStatus, want)
		}
	}

}

func TestFormValidSliceMultipart(t *testing.T) {
	multipartBody := template.Must(template.New("multipart").Parse(
		"{{ range . -}}" +
			"--123\r\n" +
			"Content-Disposition: form-data; name=\"pizza\"\r\n" +
			"\r\n" +
			"{{ . }}\r\n" +
			"{{ end }}" +
			"--123--\r\n"))

	tests := []struct {
		req         *http.Request
		placeholder interface{}
		want        interface{}
	}{
		{
			req: func() *http.Request {
				b := &strings.Builder{}
				multipartBody.Execute(b, []string{"-8", "9", "-100"})
				req := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(b.String()))
				req.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
				return req
			}(),
			placeholder: &[]int64{},
			want:        &[]int64{-8, 9, -100},
		},
		{
			req: func() *http.Request {
				b := &strings.Builder{}
				multipartBody.Execute(b, []string{"8", "9", "10"})
				req := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(b.String()))
				req.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
				return req
			}(),
			placeholder: &[]uint64{},
			want:        &[]uint64{8, 9, 10},
		},
		{
			req: func() *http.Request {
				b := &strings.Builder{}
				multipartBody.Execute(b, []string{"margeritta", "diavola", "calzone"})
				req := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(b.String()))
				req.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
				return req
			}(),
			placeholder: &[]string{},
			want:        &[]string{"margeritta", "diavola", "calzone"},
		},
		{
			req: func() *http.Request {
				b := &strings.Builder{}
				multipartBody.Execute(b, []string{"1.3", "8.9", "-4.1"})
				req := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(b.String()))
				req.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
				return req
			}(),
			placeholder: &[]float64{},
			want:        &[]float64{1.3, 8.9, -4.1},
		},
		{
			req: func() *http.Request {
				b := &strings.Builder{}
				multipartBody.Execute(b, []string{"true", "false", "true"})
				req := httptest.NewRequest(safehttp.MethodPost, "http://foo.com/", strings.NewReader(b.String()))
				req.Header.Set("Content-Type", `multipart/form-data; boundary="123"`)
				return req
			}(),
			placeholder: &[]bool{},
			want:        &[]bool{true, false, true},
		},
	}

	for _, tc := range tests {
		mux := safehttp.NewServeMux(testDispatcher{}, "foo.com")
		mux.Handle("/", safehttp.MethodPost, safehttp.HandlerFunc(func(rw *safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
			form, err := ir.MultipartForm(32 << 20)
			if err != nil {
				t.Fatalf(`ir.MultipartForm: got err %v, want nil`, err)
			}

			form.Slice("pizza", tc.placeholder)
			if err := form.Err(); err != nil {
				t.Errorf(`form.Err: got err %v, want nil`, err)
			}

			if diff := cmp.Diff(tc.want, tc.placeholder); diff != "" {
				t.Errorf("form.Slice: diff (-want +got): \n%v", diff)
			}
			return safehttp.NotWritten
		}))

		rec := newResponseRecorder(&strings.Builder{})
		mux.ServeHTTP(rec, tc.req)

		if respStatus, want := rec.status, safehttp.StatusNoContent; respStatus != want {
			t.Errorf("response status: got %v, want %v", respStatus, want)
		}
	}
}
