package templates

import (
	"path/filepath"
	"strings"
)

type ProtoYml struct {
	// Path to module e.g. "github.com/osmosis-labs/osmosis/v15/x/testmodule"
	ModulePath string `yaml:"module_path"`

	ModuleName string `yaml:"module_name"`
}

// input is of form cmd/modulegen/templates/proto/{PATH}
// returns PATH folder and go file PATH
func ParseProtoFilePath(filePath string) (string, string) {
	dir := filepath.Dir(filePath)
	folderPath, err := filepath.Rel("cmd/modulegen/templates/proto", dir)
	if err != nil {
		panic(err)
	}
	var protoFilePath string
	if strings.Contains(filePath, "yml") {
		// parse query.yml file
		protoFilePath = strings.Replace(filepath.Join(folderPath, filepath.Base(filePath[:len(filePath)-4]+"yml")), "_yml_template", "", 1)
	} else {
		protoFilePath = strings.Replace(filepath.Join(folderPath, filepath.Base(filePath[:len(filePath)-4]+"proto")), "_template", "", 1)
	}

	return folderPath, protoFilePath
}
