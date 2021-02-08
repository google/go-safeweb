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

// Pacakge defaults provides ready to use, safe, pre-configured instances of safehttp types.
package defaults

import (
	"github.com/google/go-safeweb/safehttp"
	"github.com/google/go-safeweb/safehttp/plugins/coop"
	"github.com/google/go-safeweb/safehttp/plugins/cors"
	"github.com/google/go-safeweb/safehttp/plugins/csp"
	"github.com/google/go-safeweb/safehttp/plugins/fetchmetadata"
	"github.com/google/go-safeweb/safehttp/plugins/hostcheck"
	"github.com/google/go-safeweb/safehttp/plugins/hsts"
	"github.com/google/go-safeweb/safehttp/plugins/staticheaders"
	"github.com/google/go-safeweb/safehttp/plugins/xsrf/xsrfhtml"
)

// ServeMuxConfig creates a safe and ready to use ServeMuxConfig with all necessary security interceptors installed.
func ServeMuxConfig(xsrfKey string, hosts, allowedCORSOrigins []string) *safehttp.ServeMuxConfig {
	var c safehttp.ServeMuxConfig

	// TODO(clap): add a report group once we support reporting.
	c.Intercept(coop.Default(""))
	if len(allowedCORSOrigins) > 0 {
		// TODO(clap): find a way to make this plugin communicate to the fetchmetadata one which endpoints are protected by CORS.
		c.Intercept(cors.Default(allowedCORSOrigins...))
	}
	// TODO(clap): add a report-uri once we support reporting.
	// TODO(clap): find a way to make the FramingPolicy here work together with the Framing plugin once we have it.
	c.Intercept(csp.Default(""))
	c.Intercept(fetchmetadata.NewInterceptor())
	c.Intercept(hostcheck.New(hosts...))
	c.Intercept(hsts.Default())
	c.Intercept(staticheaders.Interceptor{})
	c.Intercept(&xsrfhtml.Interceptor{SecretAppKey: xsrfKey})
	return &c
}
