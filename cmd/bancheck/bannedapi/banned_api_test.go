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

package bannedapi

import (
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestBannedAPIAnalyzer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc  string
		files map[string]string
	}{
		{
			desc: "No banned APIs",
			files: map[string]string{
				"config.json": `
				{}
				`,
				"main/test.go": `
				package main;
				func main() {}
				`,
			},
		},
		{
			desc: "Banned APIs exist",
			files: map[string]string{
				"config.json": `
				{
					"functions": [
						{
							"name": "fmt.Printf",
							"msg": "Banned by team A"
						}
					],
					"imports": [
						{
							"name": "fmt",
							"msg": "Banned by team A"
						}
					]
				}
				`,
				"main/test.go": `
				package main

				import "fmt" // want "Banned API found \"fmt\". Additional info: Banned by team A"

				func main() {
					fmt.Printf("Hello") // want "Banned API found \"fmt.Printf\". Additional info: Banned by team A"
				}
				`,
			},
		},
		{
			desc: "Banned APIs in exempted package",
			files: map[string]string{
				"config.json": `
				{
					"functions": [
						{
							"name": "fmt.Printf",
							"msg": "Banned by team A",
							"exemptions": [
								{
									"justification": "#yolo",
									"allowedPkg": "main"
								}
							]
						}
					],
					"imports": [
						{
							"name": "fmt",
							"msg": "Banned by team A",
							"exemptions": [
								{
									"justification": "#yolo",
									"allowedPkg": "main"
								}
							]
						}
					]
				}
				`,
				"main/test.go": `
				package main

				import "fmt"

				func main() {
					fmt.Printf("Hello")
				}
				`,
			},
		},
		{
			desc: "Banned renamed import",
			files: map[string]string{
				"config.json": `
				{
					"imports": [
						{
							"name": "fmt",
							"msg": "Banned by team A"
						}
					]
				}
				`,
				"main/test.go": `
				package main

				import renamed "fmt" // want "Banned API found \"fmt\". Additional info: Banned by team A"

				func main() {
					renamed.Printf("Hello")
				}
				`,
			},
		},
		{
			desc: "Banned function from renamed import",
			files: map[string]string{
				"config.json": `
				{
					"functions": [
						{
							"name": "fmt.Printf",
							"msg": "Banned by team A"
						}
					]
				}
				`,
				"main/test.go": `
				package main

				import renamed "fmt"

				func main() {
					renamed.Printf("Hello") // want "Banned API found \"fmt.Printf\". Additional info: Banned by team A"
				}
				`,
			},
		},
		{
			desc: "Package and function name collission",
			files: map[string]string{
				"config.json": `
				{
					"functions": [
						{
							"name": "fmt.Printf",
							"msg": "Banned by team A"
						}
					]
				}
				`,
				"main/test.go": `
				package main

				type Foo struct{}

				func (f *Foo) Printf(txt string) {}

				func main() {
					var fmt = &Foo{}
					fmt.Printf("Hello")
				}
				`,
			},
		},
		{
			desc: "Banned API from multiple config files",
			files: map[string]string{
				"team_a_config.json": `
				{
					"functions": [
						{
							"name": "fmt.Printf",
							"msg": "Banned by team A"
						}
					]
				}
				`,
				"team_b_config.json": `
				{
					"functions": [
						{
							"name": "fmt.Printf",
							"msg": "Banned by team B"
						}
					]
				}
				`,
				"main/test.go": `
				package main

				import "fmt"

				func main() {
					fmt.Printf("Hello") // want "Banned API found \"fmt.Printf\". Additional info: Banned by team A" "Banned API found \"fmt.Printf\". Additional info: Banned by team B"
				}
				`,
			},
		},
		{
			desc: "Banned API in one of many config files",
			files: map[string]string{
				"team_a_config.json": `
				{
					"functions": [
						{
							"name": "fmt.Printf",
							"msg": "Banned by team A",
							"exemptions": [
								{
									"justification": "#yolo",
									"allowedPkg": "main"
								}
							]
						}
					]
				}
				`,
				"team_b_config.json": `
				{
					"functions": [
						{
							"name": "fmt.Printf",
							"msg": "Banned by team B for realz"
						}
					]
				}
				`,
				"main/test.go": `
				package main

				import "fmt"

				func main() {
					fmt.Printf("Hello") // want "Banned API found \"fmt.Printf\". Additional info: Banned by team B for realz"
				}
				`,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			dir, cleanup, err := analysistest.WriteFiles(test.files)
			if err != nil {
				t.Fatalf("WriteFiles() returned err: %v", err)
			}
			defer cleanup()

			var configFiles []string
			for name := range test.files {
				if strings.HasSuffix(name, "config.json") {
					configFiles = append(configFiles, filepath.Join(dir, "src", name))
				}
			}

			a := NewAnalyzer()
			a.Flags.Set("configs", strings.Join(configFiles, ","))
			analysistest.Run(t, dir, a, "main")
		})
	}
}
