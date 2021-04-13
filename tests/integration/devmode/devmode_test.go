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

package devmode_test

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/defaults"
	"github.com/google/safehtml"
)

func TestDevMode(t *testing.T) {
	safehttp.UseLocalDev()
	t.Run("can load in dev mode", func(t *testing.T) {
		const resp = "response"
		cfg, _ := defaults.ServeMuxConfig([]string{"test.host.example"}, "test-xsrf-key")
		cfg.Handle("/test", "GET", safehttp.HandlerFunc(func(w safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
			form, err := r.URL.Query()
			if err != nil {
				t.Errorf("Cannot parse GET form: %v", err)
			}
			b := form.Bool("test", false)
			if !b {
				t.Error("test parameter, got false, want true")
			}
			w.SetCookie(safehttp.NewCookie("test", "insecure"))
			return w.Write(safehtml.HTMLEscaped(resp))
		}))
		mux := cfg.Mux()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "https://test.host.example/test?test=true", nil)
		mux.ServeHTTP(w, r)
		if w.Result().StatusCode != 200 {
			t.Errorf("Status: got %d, want 200", w.Result().StatusCode)
		}
		got, err := ioutil.ReadAll(w.Result().Body)
		if err != nil {
			t.Errorf("Read body: got %v", err)
		}
		if bytes.Compare(got, []byte(resp)) != 0 {
			t.Errorf("body: got %q, want %q", got, resp)
		}
		cs := w.Result().Cookies()
		if len(cs) == 0 {
			t.Errorf("got no cookies, wanted cookies")
		}
		for _, c := range cs {
			if c.Secure {
				t.Errorf("got secure cookie %q, should have not been secure", c.Raw)
			}
		}
	})
}
