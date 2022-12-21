package templates

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFilePathFromImportPath(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		importPath       string
		expectedFilePath string
	}{
		"standard": {importPath: "github.com/osmosis-labs/osmosis/v13/x/twap", expectedFilePath: "x/twap"},
	}
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			filePath := ParseFilePathFromImportPath(test.importPath)
			require.Equal(t, test.expectedFilePath, filePath)
		})
	}
}
