package templates

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFilePath(t *testing.T) {
	tests := map[string]struct {
		importPath         string
		expectedFolderPath string
		expectedGoFilePath string
	}{
		"standard": {
			importPath:         "cmd/modulegen/templates/x/client/cli/tx.yml",
			expectedFolderPath: "client/cli",
			expectedGoFilePath: "client/cli/tx.go",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			folderPath, goFilePath := ParseXFilePath(test.importPath)
			require.Equal(t, test.expectedFolderPath, folderPath)
			require.Equal(t, test.expectedGoFilePath, goFilePath)
		})
	}
}
