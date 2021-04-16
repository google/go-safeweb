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

package staticheaders_test

import (
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/staticheaders"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

func TestPlugin(t *testing.T) {
	req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
	fakeRW, rr := safehttptest.NewFakeResponseWriter()

	p := staticheaders.Interceptor{}
	p.Before(fakeRW, req, nil)

	if got, want := rr.Code, safehttp.StatusOK; got != int(want) {
		t.Errorf("rr.Code got: %v want: %v", got, want)
	}
}
