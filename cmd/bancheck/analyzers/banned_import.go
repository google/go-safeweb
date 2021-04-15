package analyzers

import (
	"errors"

	"golang.org/x/tools/go/analysis"
)

var BannedImportAnalyzer = &analysis.Analyzer{
	Name: "bannedimport",
	Doc:  "Checks for usage of banned imports and reports them",
	Run:  checkBannedImports,
}

func checkBannedImports(pass *analysis.Pass) (interface{}, error) {
	return nil, errors.New("not implemented yet")
}
