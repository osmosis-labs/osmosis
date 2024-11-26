package templates

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFilePathFromImportPath(t *testing.T) {
	tests := map[string]struct {
		importPath       string
		expectedFilePath string
	}{
		"standard": {importPath: "github.com/osmosis-labs/osmosis/v27/x/twap", expectedFilePath: "x/twap"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			filePath := ParseFilePathFromImportPath(test.importPath)
			require.Equal(t, test.expectedFilePath, filePath)
		})
	}
}
