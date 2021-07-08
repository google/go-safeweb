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

// Package main contains the CLI used for detecting risky APIs.
// See https://pkg.go.dev/github.com/google/go-safeweb/safehttp#hdr-Restricting_Risky_APIs
// for a high level overview.
//
// Overview
//
// Bancheck is a program that allows you to define risky APIs and check for their usage.
// It can be used as part of the CI/CD pipeline to avoid common pitfalls and prevent
// potentially vulnerable code from being deployed. Under the hood it uses the go/analysis
// package https://pkg.go.dev/golang.org/x/tools/go/analysis which provides all the tools that
// are needed for static code analysis. The tool resolves fully qualified function
// and import names and checks them against a config file that defines risky APIs.
//
// Usage
//
// Apart from the standard https://pkg.go.dev/golang.org/x/tools/go/analysis#Analyzer flags
// the command requires a config flag where a list of config files should be provided.
// You can find a sample usage below.
//
// Config
//
// Config lets you specify which APIs should be banned, explain why they are risky to use
// and allow a list of packages for which the check should be skipped.
// The structure of a config can be found in go-safeweb/cmd/bancheck/config/config.go.
//
// Note: It is possible to have colliding config files e.g. one config file bans a usage of
// an API but another one exempts it. The tool applies checks from each config file separately
// i.e. one warning will still be returned.
//
// Example config:
//  {
// 		"functions": [
// 			{
// 				"name": "fmt.Printf",
// 				"msg": "Banned by team A"
// 			}
// 		],
// 		"imports": [
// 			{
// 				"name": "fmt",
// 				"msg": "Banned by team A",
//				"exemptions": [
//					{
//						"justification": "#yolo",
//						"allowedPkg": "main"
//					}
//				]
// 			}
// 		]
//  }
//
// Example
//
// The example below shows a simple use case where "fmt" package and "fmt.Printf" function were banned
// by two separate teams.
//
// main.go
//  package main
//
//  import "fmt"
//
//  func main() {
//  	fmt.Printf("Hello")
//  }
//
// config.json
//  {
//  	"functions": [
// 			{
// 				"name": "fmt.Printf",
// 	   			"msg": "Banned by team A"
// 	  		}
// 	 	],
//   	"imports": [
// 	  		{
// 	   			"name": "fmt",
// 	   			"msg": "Banned by team B"
// 	  		}
// 	 	],
//  }
//
// CLI usage
//  $ ./bancheck -configs config.json main.go
//  /go-safeweb/cmd/bancheck/test/main.go:3:8: Banned API found "fmt". Additional info: Banned by team B
//  /go-safeweb/cmd/bancheck/test/main.go:6:6: Banned API found "fmt.Printf". Additional info: Banned by team A
package main
