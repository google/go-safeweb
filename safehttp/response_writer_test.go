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
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-safeweb/safehttp"
)

func TestResponseWriterSetCookie(t *testing.T) {
	rr := newResponseRecorder(&strings.Builder{})
	rw := safehttp.NewResponseWriter(testDispatcher{}, rr, nil)

	c := safehttp.NewCookie("foo", "bar")
	err := rw.SetCookie(c)
	if err != nil {
		t.Errorf("rw.SetCookie(c) got: %v want: nil", err)
	}

	wantHeaders := map[string][]string{
		"Set-Cookie": {"foo=bar; HttpOnly; Secure; SameSite=Lax"},
	}
	if diff := cmp.Diff(wantHeaders, map[string][]string(rr.header)); diff != "" {
		t.Errorf("rr.header mismatch (-want +got):\n%s", diff)
	}
}

func TestResponseWriterSetInvalidCookie(t *testing.T) {
	rr := newResponseRecorder(&strings.Builder{})
	rw := safehttp.NewResponseWriter(testDispatcher{}, rr, nil)

	c := safehttp.NewCookie("f=oo", "bar")
	err := rw.SetCookie(c)
	if err == nil {
		t.Error("rw.SetCookie(c) got: nil want: error")
	}
}
