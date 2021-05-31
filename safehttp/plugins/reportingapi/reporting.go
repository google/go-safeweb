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

// Package reportingapi is an implementation of the Report-To header described
// in https://www.w3.org/TR/reporting/#header.
//
// It allows for setting reporting groups to use in conjuction with COOP and CSP.
package reportingapi

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-safeweb/safehttp"
)

// DefaultMaxAge is used as default cache duration for report groups and will make them last 7 days.
const DefaultMaxAge = 7 * 24 * 60 * 60

// ReportToHeaderKey is the HTTP header key for the Reporting API.
const ReportToHeaderKey = "Report-To"

// Endpoint is the Go representation of the endpoints values as specified
// in https://www.w3.org/TR/reporting/#endpoints-member
type Endpoint struct {
	// URL defines the location of the endpoint.
	// The URL that the member's value represents MUST be potentially trustworthy. Non-secure endpoints will be ignored.
	URL string `json:"url"`
	// Priority defines which failover class the endpoint belongs to.
	Priority uint `json:"priority,omitempty"`
	// Weight defines load balancing for the failover class that the endpoint belongs to.
	Weight uint `json:"weight,omitempty"`
}

// Group is the Go representation of the Report-To header values as specified
// in https://www.w3.org/TR/reporting/#header.
type Group struct {
	// Name associates a name with the set of endpoints.
	// This is serialized as the "group" field in the Report-To header.
	// If left empty, the endpoint group will be given the name "default".
	Name string `json:"group,omitempty"`
	// IncludeSubdomains enables this endpoint group for all subdomains of the current origin’s host.
	IncludeSubdomains bool `json:"include_subdomains,omitempty"`
	// MaxAge defines the endpoint group’s lifetime, as a non-negative integer number of seconds.
	// A value of 0 will cause the endpoint group to be removed from the user agent’s reporting cache.
	MaxAge uint `json:"max_age"`
	// Endpoints is the list of endpoints that belong to this endpoint group.
	Endpoints []Endpoint `json:"endpoints"`
}

// NewGroup creates a new Group with MaxAge set to DefaultMaxage and all
// optional values set to their default values.
func NewGroup(name string, url string, otherUrls ...string) Group {
	var es []Endpoint
	for _, u := range append([]string{url}, otherUrls...) {
		es = append(es, Endpoint{URL: u})
	}
	return Group{
		Name:      name,
		MaxAge:    DefaultMaxAge,
		Endpoints: es,
	}
}

// Interceptor is the interceptor for the Report-To header.
type Interceptor struct {
	values []string
}

// NewInterceptor instantiates a new Interceptor for the given groups.
func NewInterceptor(groups ...Group) Interceptor {
	var i Interceptor
	for _, r := range groups {
		buf, err := json.Marshal(r)
		if err != nil {
			// This path should not be possible.
			panic(fmt.Sprintf("marshalling report: %#v, %v", r, err))
		}
		i.values = append(i.values, string(buf))
	}
	return i
}

// Before adds all the configured Report-To header values as separate headers.
func (i Interceptor) Before(w safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	for _, v := range i.values {
		w.Header().Add(ReportToHeaderKey, v)
	}
	return safehttp.NotWritten()
}

// Commit is a no-op, required to satisfy the safehttp.Interceptor interface.
func (i Interceptor) Commit(w safehttp.ResponseHeadersWriter, r *safehttp.IncomingRequest, resp safehttp.Response, _ safehttp.InterceptorConfig) {
}
