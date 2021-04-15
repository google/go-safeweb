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

// +build go1.16

package htmlinject

import (
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"

	"github.com/google/safehtml/template"
	"github.com/google/safehtml/template/uncheckedconversions"
)

// LoadGlobEmbed is like LoadGlob but works on an embedded filesystem.
func LoadGlobEmbed(tpl *template.Template, lcfg LoadConfig, pattern template.TrustedSource, fsys embed.FS) (*template.Template, error) {
	return loadGlobFS(tpl, lcfg, pattern, fsys)
}

// TODO(clap): the rest of this file is a copy-paste of a lot of code. Unify this code path with that
// once we decide to drop support for Go version before 1.16 and we can use fs everywhere.

func loadGlobFS(tpl *template.Template, lcfg LoadConfig, pattern template.TrustedSource, fsys embed.FS) (*template.Template, error) {
	filenames, err := fs.Glob(fsys, pattern.String())
	if err != nil {
		return nil, err
	}
	if len(filenames) == 0 {
		return nil, fmt.Errorf("pattern matches no files: %#q", pattern.String())
	}
	var tts []template.TrustedSource
	for _, fn := range filenames {
		// The pattern expanded from a trusted source, so the expansion is still trusted.
		tts = append(tts, uncheckedconversions.TrustedSourceFromStringKnownToSatisfyTypeContract(fn))
	}
	return loadFilesFS(tpl, lcfg, fsys, tts...)
}

func loadFilesFS(tpl *template.Template, lcfg LoadConfig, fsys fs.FS, filenames ...template.TrustedSource) (*template.Template, error) {
	// The naming juggling below is quite odd but is kept for consistency.
	if len(filenames) == 0 {
		return nil, fmt.Errorf("no files named in call to LoadFiles")
	}
	for _, fnts := range filenames {
		fn := fnts.String()
		b, err := ioutil.ReadFile(fn)
		if err != nil {
			return nil, err
		}
		name := filepath.Base(fn)
		var t *template.Template
		if tpl == nil {
			tpl = template.New(name)
		}
		if name == tpl.Name() {
			t = tpl
		} else {
			t = tpl.New(name)
		}
		// We are loading a file from a TrustedSource, so this conversion is safe.
		tts := uncheckedconversions.TrustedTemplateFromStringKnownToSatisfyTypeContract(string(b))
		_, err = LoadTrustedTemplate(t, lcfg, tts)
		if err != nil {
			return nil, err
		}
	}
	return tpl, nil
}
