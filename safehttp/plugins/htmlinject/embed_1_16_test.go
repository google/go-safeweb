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

//go:build go1.16
// +build go1.16

package htmlinject

import (
	"embed"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	safetemplate "github.com/google/safehtml/template"
)

//go:embed testdata
var embedTestdata embed.FS

func TestLoadGlobEmbed(t *testing.T) {
	tpl, err := LoadGlobEmbed(nil, LoadConfig{}, safetemplate.TrustedSourceFromConstant("testdata/*.tpl"), embedTestdata)

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
