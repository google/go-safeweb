// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package htmlinject

import (
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	safetemplate "github.com/google/safehtml/template"
	"github.com/google/safehtml/template/uncheckedconversions"
)

func ExampleTransform() {
	const in = `
<html>
<head>
<link rel=preload as="script" src="gopher.js">
</head>
<body>
{{.Content}}
<script type="application/javascript">alert("script")</script>
<form>
First name:<br>
<input type="text" name="firstname"><br>
Last name:<br>
<input type="text" name="lastname">
</form>
</body>
</html>
`
	got, err := Transform(strings.NewReader(in), CSPNoncesDefault, XSRFTokensDefault)
	if err != nil {
		// handle error
		panic(err)
	}
	template.Must(template.New("example transform").Funcs(map[string]interface{}{
		"XSRFToken": func() string { return "XSRFToken-secret" },
		"CSPNonce":  func() string { return "CSPNonce-secret" },
	}).Parse(got)).Execute(os.Stdout, map[string]string{"Content": "This is some content"})
	// Output:
	// <html>
	// <head>
	// <link nonce="CSPNonce-secret" rel=preload as="script" src="gopher.js">
	// </head>
	// <body>
	// This is some content
	// <script nonce="CSPNonce-secret" type="application/javascript">alert("script")</script>
	// <form><input type="hidden" name="xsrf-token" value="XSRFToken-secret">
	// First name:<br>
	// <input type="text" name="firstname"><br>
	// Last name:<br>
	// <input type="text" name="lastname">
	// </form>
	// </body>
	// </html>
}

func BenchmarkTransform(b *testing.B) {
	b.ReportAllocs()
	var (
		config = []TransformConfig{CSPNoncesDefault, XSRFTokensDefault}
		in     = `
<html>
<head>
<link rel=preload as="script" src="gopher.js">
</head>
<body>
<script type="application/javascript">alert("script")</script>
<form>
  First name:<br>
  <input type="text" name="firstname"><br>
  Last name:<br>
  <input type="text" name="lastname">
</form>
</body>
</html>
`
		want = `
<html>
<head>
<link nonce="{{CSPNonce}}" rel=preload as="script" src="gopher.js">
</head>
<body>
<script nonce="{{CSPNonce}}" type="application/javascript">alert("script")</script>
<form><input type="hidden" name="xsrf-token" value="{{XSRFToken}}">
  First name:<br>
  <input type="text" name="firstname"><br>
  Last name:<br>
  <input type="text" name="lastname">
</form>
</body>
</html>
`
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err := Transform(strings.NewReader(in), config...)
		if err != nil {
			b.Fatalf("Transform: got err %q, didn't want one", err)
		}
		if got != want {
			b.Errorf("got %q, want %q", got, want)
		}
	}
}

var tests = []struct {
	name      string
	xsrf, csp bool
	in, want  string
}{
	{
		name: "nothing to change",
		in: `
<html>
<header><title>This is title</title></header>
<body>
Hello world
</body>
</html>
`,
		want: `
<html>
<header><title>This is title</title></header>
<body>
Hello world
</body>
</html>
`,
	},
	{
		name: "add CSP nonces",
		csp:  true,
		in: `
<html>
<head>
<link rel=preload as="script" src="gopher.js">
<style>
h1 {
  border: 5px solid yellow;
}
</style>
</head>
<body>
<script type="application/javascript">alert("script")</script>
</body>
</html>
`,
		want: `
<html>
<head>
<link nonce="{{CSPNonce}}" rel=preload as="script" src="gopher.js">
<style nonce="{{CSPNonce}}">
h1 {
  border: 5px solid yellow;
}
</style>
</head>
<body>
<script nonce="{{CSPNonce}}" type="application/javascript">alert("script")</script>
</body>
</html>
`,
	},
	{
		name: "add XSRF protection",
		xsrf: true,
		in: `
<form>
  First name:<br>
  <input type="text" name="firstname"><br>
  Last name:<br>
  <input type="text" name="lastname">
</form>
`,
		want: `
<form><input type="hidden" name="xsrf-token" value="{{XSRFToken}}">
  First name:<br>
  <input type="text" name="firstname"><br>
  Last name:<br>
  <input type="text" name="lastname">
</form>
`,
	},
	{
		name: "all configs",
		xsrf: true,
		csp:  true,
		in: `
<html>
<head>
<link rel="stylesheet" href="styles.css">
<link rel=preload as="script" src="gopher.js">
</head>
<body>
<script type="application/javascript">alert("script")</script>
<form>
  First name:<br>
  <input type="text" name="firstname"><br>
  Last name:<br>
  <input type="text" name="lastname">
</form>
</body>
</html>
`,
		want: `
<html>
<head>
<link rel="stylesheet" href="styles.css">
<link nonce="{{CSPNonce}}" rel=preload as="script" src="gopher.js">
</head>
<body>
<script nonce="{{CSPNonce}}" type="application/javascript">alert("script")</script>
<form><input type="hidden" name="xsrf-token" value="{{XSRFToken}}">
  First name:<br>
  <input type="text" name="firstname"><br>
  Last name:<br>
  <input type="text" name="lastname">
</form>
</body>
</html>
`,
	},
}

func TestTransform(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := []TransformConfig{}
			if tt.csp {
				cfg = append(cfg, CSPNoncesDefault)
			}
			if tt.xsrf {
				cfg = append(cfg, XSRFTokensDefault)

			}
			got, err := Transform(strings.NewReader(tt.in), cfg...)
			if err != nil {
				t.Fatalf("Transform: got err %q, didn't want one", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("-want +got %s", diff)
			}
		})
	}
}

func TestLoadTrustedTemplateWithDefaultConfig(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := LoadConfig{DisableCSP: !tt.csp, DisableXSRF: !tt.xsrf}
			gotTpl, err := LoadTrustedTemplate(nil, cfg, uncheckedconversions.TrustedTemplateFromStringKnownToSatisfyTypeContract(tt.in))
			if err != nil {
				t.Fatalf("LoadTrustedTemplate: got err %q", err)
			}
			// Test that whatever we provide is clonable or injection won't be possible.
			gotTpl, err = gotTpl.Clone()
			if err != nil {
				t.Fatalf("Clone loaded template: got err %q", err)
			}
			var sb strings.Builder
			// Make functions return the source code that calls them and compare the results with the source.
			// Sadly there is no good way to take the parsed sources out of a text/template.Template.
			err = gotTpl.Funcs(map[string]interface{}{
				XSRFTokensDefaultFuncName: func() string { return "{{" + XSRFTokensDefaultFuncName + "}}" },
				CSPNoncesDefaultFuncName:  func() string { return "{{" + CSPNoncesDefaultFuncName + "}}" },
			}).Execute(&sb, nil)
			if err != nil {
				t.Fatalf("Execute: got err %q", err)
			}
			if diff := cmp.Diff(tt.want, sb.String()); diff != "" {
				t.Errorf("-want +got %s", diff)
			}
		})
	}
}

func TestLoadGlob(t *testing.T) {
	tpl, err := LoadGlob(nil, LoadConfig{}, safetemplate.TrustedSourceFromConstant("testdata/*.tpl"))

	// Test that whatever we provide is clonable or injection won't be possible.
	tpl, err = tpl.Clone()
	if err != nil {
		t.Fatalf("Clone loaded template: got err %q", err)
	}
	tpl = tpl.Funcs(map[string]interface{}{
		XSRFTokensDefaultFuncName: func() string { return "{{" + XSRFTokensDefaultFuncName + "}}" },
		CSPNoncesDefaultFuncName:  func() string { return "{{" + CSPNoncesDefaultFuncName + "}}" },
	})
	if got, want := len(tpl.Templates()), 2; got != want {
		t.Fatalf("Loaded templates: got %d want %d %s", got, want, tpl.DefinedTemplates())
	}
	for _, inner := range tpl.Templates() {
		t.Run(inner.Name(), func(t *testing.T) {
			var sb strings.Builder
			err := tpl.ExecuteTemplate(&sb, inner.Name(), nil)
			if err != nil {
				t.Fatalf("Executing: %v", err)
			}
			stripExt := strings.TrimSuffix(inner.Name(), filepath.Ext(inner.Name()))
			b, err := ioutil.ReadFile("testdata/" + stripExt + ".want")
			if err != nil {
				t.Fatalf("Reading '.want' file: %v", err)
			}
			if got, want := sb.String(), string(b); got != want {
				t.Errorf("got: %v, want: %v", got, want)
			}
		})
	}
}
