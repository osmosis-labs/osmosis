package templates

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseProtoFilePath(t *testing.T) {
	tests := map[string]struct {
		importPath         string
		expectedFilePath   string
		expectedFolderPath string
	}{
		"standard": {
			importPath: "cmd/modulegen/templates/proto/genesis_template.tmpl", 
			expectedFilePath: "genesis.proto",
			expectedFolderPath: ".",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			folderPath, filePath := ParseProtoFilePath(test.importPath)
			require.Equal(t, test.expectedFilePath, filePath)
			require.Equal(t, test.expectedFolderPath, folderPath)
		})
	}
}