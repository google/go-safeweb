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
package config

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestBannedImportConfig(t *testing.T) {
	tests := []struct {
		desc  string
		files map[string]string // fake workspace files
		want  BannedApis
	}{
		{
			desc: "file with empty definitions",
			files: map[string]string{
				"file.json": `
				{}
				`,
			},
			want: BannedApis{},
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
			want: BannedApis{},
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
							"allowedDir": "subdirs/vetted/..."
						}]
					}]
				}
				`,
			},
			want: BannedApis{
				"legacyconversions": {{
					Name: "legacyconversions",
					Msg:  "Sample message",
					Exemptions: []Exemption{
						{
							Justification: "My justification",
							AllowedDir:    "subdirs/vetted/...",
						},
					},
				}},
			},
		},
		{
			desc: "multiple files",
			files: map[string]string{
				"file1.json": `
				{
					"imports": [{
						"name": "import1"
					}]
				}
				`,
				"file2.json": `
				{
					"imports": [{
						"name": "import2"
					}]
				}
				`,
			},
			want: BannedApis{
				"import1": {{Name: "import1"}},
				"import2": {{Name: "import2"}},
			},
		},
		{
			desc: "duplicate definitions",
			files: map[string]string{
				"file1.json": `
				{
					"imports": [{
						"name": "import",
						"msg": "Banned by team x",
						"exemptions": [{
							"justification": "My justification",
							"allowedDir": "subdirs/vetted/..."
						}]
					}]
				}
				`,
				"file2.json": `
				{
					"imports": [{
						"name": "import",
						"msg": "Banned by team y",
						"exemptions": [{
							"justification": "#yolo",
							"allowedDir": "otherdir/legacy/..."
						}]
					}]
				}
				`,
			},
			want: BannedApis{
				"import": {
					{
						Name: "import",
						Msg:  "Banned by team x",
						Exemptions: []Exemption{{
							Justification: "My justification",
							AllowedDir:    "subdirs/vetted/..."}},
					},
					{
						Name: "import",
						Msg:  "Banned by team y",
						Exemptions: []Exemption{{
							Justification: "#yolo",
							AllowedDir:    "otherdir/legacy/..."}},
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

			imports, err := ReadBannedImports(files)

			if err != nil {
				t.Errorf("ReadBannedImports() got err: %v want: nil", err)
			}
			if diff := cmp.Diff(imports, test.want); diff != "" {
				t.Errorf("config mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBannedFunctionConfig(t *testing.T) {
	tests := []struct {
		desc  string
		files map[string]string // fake workspace files
		want  BannedApis
	}{
		{
			desc: "file with empty definitions",
			files: map[string]string{
				"file.json": `
				{}
				`,
			},
			want: BannedApis{},
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
			want: BannedApis{},
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
							"allowedDir": "subdirs/vetted/..."
						}]
					}]
				}
				`,
			},
			want: BannedApis{
				"safehttp.NewServeMuxConfig": {{
					Name: "safehttp.NewServeMuxConfig",
					Msg:  "Sample message",
					Exemptions: []Exemption{
						{
							Justification: "My justification",
							AllowedDir:    "subdirs/vetted/...",
						},
					},
				}},
			},
		},
		{
			desc: "multiple files",
			files: map[string]string{
				"file1.json": `
				{
					"functions": [{
						"name": "function1"
					}]
				}
				`,
				"file2.json": `
				{
					"functions": [{
						"name": "function2"
					}]
				}
				`,
			},
			want: BannedApis{
				"function1": {{Name: "function1"}},
				"function2": {{Name: "function2"}},
			},
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
							"allowedDir": "subdirs/vetted/..."
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
							"allowedDir": "otherdir/legacy/..."
						}]
					}]
				}
				`,
			},
			want: BannedApis{
				"function": {
					{
						Name: "function",
						Msg:  "Banned by team x",
						Exemptions: []Exemption{
							{
								Justification: "My justification",
								AllowedDir:    "subdirs/vetted/...",
							},
						},
					},
					{
						Name: "function",
						Msg:  "Banned by team y",
						Exemptions: []Exemption{
							{
								Justification: "#yolo",
								AllowedDir:    "otherdir/legacy/...",
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

			fns, err := ReadBannedFunctions(files)

			if err != nil {
				t.Errorf("Read() got err: %v want: nil", err)
			}
			if diff := cmp.Diff(fns, test.want); diff != "" {
				t.Errorf("config mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
func TestConfigErrors(t *testing.T) {
	tests := []struct {
		desc     string
		files    map[string]string // fake workspace files
		fileName string            // file name to read
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
			fns, errFns := ReadBannedFunctions([]string{file})
			imports, errImports := ReadBannedImports([]string{file})

			if fns != nil {
				t.Errorf("ReadBannedFunctions(%q) got %v, wanted nil",
					fns, test.fileName)
			}
			if imports != nil {
				t.Errorf("ReadBannedImports(%q) got %v, wanted nil",
					fns, test.fileName)
			}
			if errFns == nil {
				t.Errorf("ReadBannedFunctions(%q) succeeded but wanted error",
					test.fileName)
			}
			if errImports == nil {
				t.Errorf("ReadBannedImports(%q) succeeded but wanted error",
					test.fileName)
			}
		})
	}
}
