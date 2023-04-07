// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reportingapi_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/reportingapi"
	"github.com/google/go-safeweb/safehttp/safehttptest"
)

var sortStringSlices = cmpopts.SortSlices(func(a, b string) bool { return a < b })

func TestJSONTags(t *testing.T) {
	var tests = []struct {
		name string
		g    reportingapi.Group
		want string
	}{
		{
			name: "Defaults",
			g: reportingapi.Group{
				Name:      "group-name",
				Endpoints: []reportingapi.Endpoint{{URL: "https://foo.bar"}},
			},
			want: `{"group":"group-name","max_age":0,"endpoints":[{"url":"https://foo.bar"}]}`,
		},
		{
			name: "All fields",
			g: reportingapi.Group{
				Name:              "name",
				IncludeSubdomains: true,
				MaxAge:            100,
				Endpoints: []reportingapi.Endpoint{
					{
						URL:      "https://foo.bar",
						Priority: 2,
						Weight:   3,
					},
				},
			},
			want: `{"group":"name","include_subdomains":true,"max_age":100,"endpoints":[{"url":"https://foo.bar","priority":2,"weight":3}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(&tt.g)
			if err != nil {
				t.Errorf("Got error %v, did not want an error", err)
			}
			if got := string(got); got != tt.want {
				t.Errorf("json.Marshal(%v): got %s want %s", tt.g, got, tt.want)
			}
		})
	}
}

func TestBefore(t *testing.T) {
	var tests = []struct {
		name string
		cfg  []reportingapi.Group
		want []string
	}{
		{
			name: "single",
			cfg:  []reportingapi.Group{reportingapi.NewGroup("test", "https://fuffa.buffa/reporting")},
			want: []string{`{"group":"test","max_age":604800,"endpoints":[{"url":"https://fuffa.buffa/reporting"}]}`},
		},
		{
			name: "multiple endpoints",
			cfg:  []reportingapi.Group{reportingapi.NewGroup("test", "https://fuffa.buffa/reporting1", "https://fuffa.buffa/reporting2")},
			want: []string{`{"group":"test","max_age":604800,"endpoints":[{"url":"https://fuffa.buffa/reporting1"},{"url":"https://fuffa.buffa/reporting2"}]}`},
		},
		{
			name: "multiple headers",
			cfg: []reportingapi.Group{
				reportingapi.NewGroup("test1", "https://fuffa.buffa/reporting1"),
				reportingapi.NewGroup("test2", "https://fuffa.buffa/reporting2"),
			},
			want: []string{
				`{"group":"test1","max_age":604800,"endpoints":[{"url":"https://fuffa.buffa/reporting1"}]}`,
				`{"group":"test2","max_age":604800,"endpoints":[{"url":"https://fuffa.buffa/reporting2"}]}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRW, rr := safehttptest.NewFakeResponseWriter()
			req := safehttptest.NewRequest(safehttp.MethodGet, "/", nil)
			i := reportingapi.NewInterceptor(tt.cfg...)
			i.Before(fakeRW, req, nil)
			got := rr.Header().Values("Report-To")
			if diff := cmp.Diff(tt.want, got, sortStringSlices); diff != "" {
				t.Errorf("Report-To headers: -want +got %s", diff)
			}
		})
	}
}
