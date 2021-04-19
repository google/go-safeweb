package config

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestBannedFunctionAnalyzer(t *testing.T) {
	tests := []struct {
		desc   string            // describes the test case
		files  map[string]string // fake workspace files
		config *Config           // the expected config
	}{
		{
			desc: "file with empty definitions",
			files: map[string]string{
				"file.json": `
				{}
				`,
			},
			config: &Config{Imports: []BannedImport{}, Functions: []BannedFunction{}},
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
			config: &Config{Imports: []BannedImport{}, Functions: []BannedFunction{}},
		},
		{
			desc: "file with banned import",
			files: map[string]string{
				"file.json": `
				{
					"imports": [{
						"name": "github.com/google/go-safeweb/safesql/legacyconversions",
						"msg": "Sample message",
						"exemptions": [{
							"justification": "My justification",
							"allowedDir": "mycompany.com/my/subdirs/vetted/..."
						}]
					}]
				}
				`,
			},
			config: &Config{
				Imports: []BannedImport{
					{
						Name: "github.com/google/go-safeweb/safesql/legacyconversions",
						Msg:  "Sample message",
						Exemptions: []Exemption{
							{
								Justification: "My justification",
								AllowedDir:    "mycompany.com/my/subdirs/vetted/...",
							},
						},
					},
				},
				Functions: []BannedFunction{},
			},
		},
		{
			desc: "file with banned function",
			files: map[string]string{
				"file.json": `
				{
					"functions": [{
						"name": "github.com/google/go-safeweb/safehttp.NewServeMuxConfig",
						"msg": "Sample message",
						"exemptions": [{
							"justification": "My justification",
							"allowedDir": "mycompany.com/my/subdirs/vetted/..."
						}]
					}]
				}
				`,
			},
			config: &Config{
				Imports: []BannedImport{},
				Functions: []BannedFunction{
					{
						Name: "github.com/google/go-safeweb/safehttp.NewServeMuxConfig",
						Msg:  "Sample message",
						Exemptions: []Exemption{
							{
								Justification: "My justification",
								AllowedDir:    "mycompany.com/my/subdirs/vetted/...",
							},
						},
					},
				},
			},
		},
		{
			desc: "multiple files",
			files: map[string]string{
				"file1.json": `
				{
					"functions": [{
						"name": "function1"
					}],
					"imports": [{
						"name": "import1"
					}]
				}
				`,
				"file2.json": `
				{
					"functions": [{
						"name": "function2"
					}],
					"imports": [{
						"name": "import2"
					}]
				}
				`,
			},
			config: &Config{
				Imports: []BannedImport{
					{Name: "import1"},
					{Name: "import2"},
				},
				Functions: []BannedFunction{
					{Name: "function1"},
					{Name: "function2"},
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

			config, error := Read(files)

			if error != nil {
				t.Errorf("Read() got err: %v want: nil", error)
			}
			if diff := cmp.Diff(config, test.config); diff != "" {
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
			cfg, error := Read([]string{file})

			if cfg != nil {
				t.Errorf("Read(%q) returned a config but wanted nil", test.fileName)
			}
			if error == nil {
				t.Errorf("Read(%q) succeeded but wanted error", test.fileName)
			}
		})
	}
}
