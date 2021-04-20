package config

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestBannedImportConfig(t *testing.T) {
	tests := []struct {
		desc    string            // describes the test case
		files   map[string]string // fake workspace files
		imports BannedIdents      // the expected imports
	}{
		{
			desc: "file with empty definitions",
			files: map[string]string{
				"file.json": `
				{}
				`,
			},
			imports: BannedIdents{},
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
			imports: BannedIdents{},
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
			imports: BannedIdents{
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
			imports: BannedIdents{
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
			imports: BannedIdents{
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
				t.Fatalf("Test %s: WriteFiles() returned err: %v", test.desc, err)
			}
			defer cleanup()
			files := make([]string, 0)
			for f := range test.files {
				path := filepath.Join(dir, "src", f)
				files = append(files, path)
			}

			imports, error := ReadBannedImports(files)

			if error != nil {
				t.Errorf("Read() got err: %v want: nil", error)
			}
			if diff := cmp.Diff(imports, test.imports); diff != "" {
				t.Errorf("config mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBannedFunctionConfig(t *testing.T) {
	tests := []struct {
		desc      string            // describes the test case
		files     map[string]string // fake workspace files
		functions BannedIdents      // the expected imports
	}{
		{
			desc: "file with empty definitions",
			files: map[string]string{
				"file.json": `
				{}
				`,
			},
			functions: BannedIdents{},
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
			functions: BannedIdents{},
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
			functions: BannedIdents{
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
			functions: BannedIdents{
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
			functions: BannedIdents{
				"function": {
					{
						Name: "function",
						Msg:  "Banned by team x",
						Exemptions: []Exemption{{
							Justification: "My justification",
							AllowedDir:    "subdirs/vetted/..."}},
					},
					{
						Name: "function",
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
				t.Fatalf("Test %s: WriteFiles() returned err: %v", test.desc, err)
			}
			defer cleanup()
			files := make([]string, 0)
			for f := range test.files {
				path := filepath.Join(dir, "src", f)
				files = append(files, path)
			}

			fns, error := ReadBannedFunctions(files)

			if error != nil {
				t.Errorf("Read() got err: %v want: nil", error)
			}
			if diff := cmp.Diff(fns, test.functions); diff != "" {
				t.Errorf("config mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
func TestConfigErrors(t *testing.T) {
	tests := []struct {
		desc     string            // describes the test case
		files    map[string]string // fake workspace files
		fileName string            // file name to read
	}{
		{
			desc:     "file does not exist",
			files:    map[string]string{},
			fileName: "nonexistent",
		},
		{
			desc:     "file is a directory",
			files:    map[string]string{},
			fileName: "",
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
				t.Fatalf("Test %s: WriteFiles() returned err: %v", test.desc, err)
			}
			defer cleanup()

			file := filepath.Join(dir, "src", test.fileName)
			fns, errorFns := ReadBannedFunctions([]string{file})
			imports, errorImports := ReadBannedImports([]string{file})

			if fns != nil {
				t.Errorf("ReadBannedFunctions(%q) returned a config but wanted nil", test.fileName)
			}
			if imports != nil {
				t.Errorf("ReadBannedImports(%q) returned a config but wanted nil", test.fileName)
			}
			if errorFns == nil {
				t.Errorf("ReadBannedFunctions(%q) succeeded but wanted error", test.fileName)
			}
			if errorImports == nil {
				t.Errorf("ReadBannedImports(%q) succeeded but wanted error", test.fileName)
			}
		})
	}
}
