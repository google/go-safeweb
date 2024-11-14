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

package config

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestReadConfigs(t *testing.T) {
	tests := []struct {
		desc  string
		files map[string]string
		want  *Config
	}{
		{
			desc: "file with empty definitions",
			files: map[string]string{
				"file.json": `
				{}
				`,
			},
			want: &Config{},
		},
		{
			desc: "file with unknown field",
			files: map[string]string{
				"file.json": `
				{
					"unknown": 1
				}
				`,
			},
			want: &Config{},
		},
		{
			desc: "file with banned import",
			files: map[string]string{
				"file.json": `
				{
					"imports": [{
						"name": "legacyconversions",
						"msg": "Sample message",
						"exemptions": [{
							"justification": "My justification",
							"allowedPkg": "subdirs/vetted/..."
						}]
					}]
				}
				`,
			},
			want: &Config{Imports: []BannedAPI{
				{
					Name: "legacyconversions",
					Msg:  "Sample message",
					Exemptions: []Exemption{
						{
							Justification: "My justification",
							AllowedPkg:    "subdirs/vetted/...",
						},
					},
				}},
			},
		},
		{
			desc: "multiple files with imports",
			files: map[string]string{
				"file1.json": `
				{
					"imports": [{
						"name": "import1",
						"msg": "msg1" 
					}]
				}
				`,
				"file2.json": `
				{
					"imports": [{
						"name": "import2",
						"msg": "msg2" 
					}]
				}
				`,
			},
			want: &Config{Imports: []BannedAPI{
				{Name: "import1", Msg: "msg1"}, {Name: "import2", Msg: "msg2"},
			}},
		},
		{
			desc: "file with banned function",
			files: map[string]string{
				"file.json": `
				{
					"functions": [{
						"name": "safehttp.NewServeMuxConfig",
						"msg": "Sample message",
						"exemptions": [{
							"justification": "My justification",
							"allowedPkg": "subdirs/vetted/..."
						}]
					}]
				}
				`,
			},
			want: &Config{Functions: []BannedAPI{{
				Name: "safehttp.NewServeMuxConfig",
				Msg:  "Sample message",
				Exemptions: []Exemption{
					{
						Justification: "My justification",
						AllowedPkg:    "subdirs/vetted/...",
					},
				},
			}}},
		},
		{
			desc: "file with banned imports and functions",
			files: map[string]string{
				"file.json": `
				{
					"imports": [{
						"name": "legacyconversions",
						"msg": "Sample message",
						"exemptions": [{
							"justification": "My justification",
							"allowedPkg": "subdirs/vetted/..."
						}]
					}],
					"functions": [{
						"name": "safehttp.NewServeMuxConfig",
						"msg": "Sample message",
						"exemptions": [{
							"justification": "My justification",
							"allowedPkg": "subdirs/vetted/..."
						}]
					}]
				}
				`,
			},
			want: &Config{
				Imports: []BannedAPI{
					{
						Name: "legacyconversions",
						Msg:  "Sample message",
						Exemptions: []Exemption{
							{
								Justification: "My justification",
								AllowedPkg:    "subdirs/vetted/...",
							},
						},
					}},
				Functions: []BannedAPI{{
					Name: "safehttp.NewServeMuxConfig",
					Msg:  "Sample message",
					Exemptions: []Exemption{
						{
							Justification: "My justification",
							AllowedPkg:    "subdirs/vetted/...",
						},
					},
				}},
			},
		},
		{
			desc: "multiple files with functions",
			files: map[string]string{
				"file1.json": `
				{
					"functions": [{
						"name": "function1",
						"msg": "msg1" 
					}]
				}
				`,
				"file2.json": `
				{
					"functions": [{
						"name": "function2",
						"msg": "msg2" 
					}]
				}
				`,
			},
			want: &Config{Functions: []BannedAPI{
				{Name: "function1", Msg: "msg1"}, {Name: "function2", Msg: "msg2"},
			}},
		},
		{
			desc: "duplicate definitions",
			files: map[string]string{
				"file1.json": `
				{
					"functions": [{
						"name": "function",
						"msg": "Banned by team x",
						"exemptions": [{
							"justification": "My justification",
							"allowedPkg": "subdirs/vetted/..."
						}]
					}]
				}
				`,
				"file2.json": `
				{
					"functions": [{
						"name": "function",
						"msg": "Banned by team y",
						"exemptions": [{
							"justification": "#yolo",
							"allowedPkg": "otherdir/legacy/..."
						}]
					}]
				}
				`,
			},
			want: &Config{
				Functions: []BannedAPI{
					{
						Name: "function",
						Msg:  "Banned by team x",
						Exemptions: []Exemption{
							{
								Justification: "My justification",
								AllowedPkg:    "subdirs/vetted/...",
							},
						},
					},
					{
						Name: "function",
						Msg:  "Banned by team y",
						Exemptions: []Exemption{
							{
								Justification: "#yolo",
								AllowedPkg:    "otherdir/legacy/...",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			dir, cleanup, err := analysistest.WriteFiles(test.files)
			if err != nil {
				t.Fatalf("WriteFiles() returned err: %v", err)
			}
			defer cleanup()
			var files []string
			for f := range test.files {
				path := filepath.Join(dir, "src", f)
				files = append(files, path)
			}

			cfg, err := ReadConfigs(files)

			if err != nil {
				t.Errorf("ReadConfigs() got err: %v want: nil", err)
			}
			if diff := cmp.Diff(cfg, test.want, cmpopts.SortSlices(BannedAPICmp)); diff != "" {
				t.Errorf("config mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

var BannedAPICmp = func(b1, b2 BannedAPI) bool { return b1.Msg < b2.Msg }

func TestConfigErrors(t *testing.T) {
	tests := []struct {
		desc     string
		files    map[string]string
		fileName string
	}{
		{
			desc:     "file does not exist",
			files:    map[string]string{},
			fileName: "nonexistent",
		},
		{
			desc: "file is a directory",
			files: map[string]string{
				"dir/file.json": ``,
			},
			fileName: "dir",
		},
		{
			desc: "file has invalid contents",
			files: map[string]string{
				"file.json": `
				{"imports":"this should be an object"}
				`,
			},
			fileName: "file.json",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			dir, cleanup, err := analysistest.WriteFiles(test.files)
			if err != nil {
				t.Fatalf("WriteFiles() got err: %v", err)
			}
			defer cleanup()

			file := filepath.Join(dir, "src", test.fileName)
			cfg, err := ReadConfigs([]string{file})

			if cfg != nil {
				t.Errorf("ReadConfigs() got %v, wanted nil", cfg)
			}
			if err == nil {
				t.Errorf("ReadConfigs() got %v, wanted error", cfg)
			}
		})
	}
}
