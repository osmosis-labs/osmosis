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
		"standard": {importPath: "github.com/osmosis-labs/osmosis/v15/x/twap", expectedFilePath: "x/twap"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			filePath := ParseFilePathFromImportPath(test.importPath)
			require.Equal(t, test.expectedFilePath, filePath)
		})
	}
}

func TestParseXFilePath(t *testing.T) {
	tests := map[string]struct {
		importPath         string
		expectedFilePath   string
		expectedFolderPath string
	}{
		"standard": {
			importPath: "cmd/modulegen/templates/x/client/cli/flags_template.tmpl", 
			expectedFilePath: "client/cli/flags.go",
			expectedFolderPath: "client/cli",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			folderPath, filePath := ParseXFilePath(test.importPath)
			require.Equal(t, test.expectedFilePath, filePath)
			require.Equal(t, test.expectedFolderPath, folderPath)
		})
	}
}
