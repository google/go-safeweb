package analyzers

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestBannedImportAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), BannedImportAnalyzer)
}
