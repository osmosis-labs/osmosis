package templates

import (
	"path/filepath"
	"strings"
)

type XYml struct {
	// Path to simtypes e.g. "github.com/osmosis-labs/osmosis/v15/simulation"
	SimtypesPath string `yaml:"simtypes_path"`
	// Path to module e.g. "github.com/osmosis-labs/osmosis/v15/x/testmodule"
	ModulePath string `yaml:"module_path"`

	ModuleName string `yaml:"module_name"`
}

// input is of form github.com/osmosis-labs/osmosis/vXX/{PATH}
// returns PATH
func ParseFilePathFromImportPath(importPath string) string {
	splits := strings.Split(importPath, "/")
	pathSplits := splits[4:]
	return strings.Join(pathSplits, "/")
}

// input is of form cmd/modulegen/templates/x/{PATH}
// returns PATH folder name and the full go file PATH
func ParseXFilePath(filePath string) (string, string) {
	dir := filepath.Dir(filePath)
	folderPath, err := filepath.Rel("cmd/modulegen/templates/x", dir)
	if err != nil {
		panic(err)
	}
	goFilePath := strings.Replace(filepath.Join(folderPath, filepath.Base(filePath[:len(filePath)-4]+"go")), "_template", "", 1)
	return folderPath, goFilePath
}
