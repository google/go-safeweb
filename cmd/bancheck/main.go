package main

import (
	"github.com/google/go-safeweb/cmd/bancheck/analyzers"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(analyzers.BannedFunctionAnalyzer, analyzers.BannedImportAnalyzer)
}
