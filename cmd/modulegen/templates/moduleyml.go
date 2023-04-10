package templates

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type ModuleYml struct {
	// Path to simtypes e.g. "github.com/osmosis-labs/osmosis/v15/simulation"
	SimtypesPath string `yaml:"simtypes_path"`
	// Path to module e.g. "github.com/osmosis-labs/osmosis/v15/x/testmodule"
	ModulePath string `yaml:"module_path"`

	ModuleName string `yaml:"module_name"`

	filePath string
}

// type YmlModuleDescriptor struct {
// 	// ProtoWrapper *ProtoWrapperDescriptor `yaml:"proto_wrapper,omitempty"`
// 	Cli          *CliDescriptor
// }

func ReadYmlFile(filepath string) (ModuleYml, error) {
	content, err := os.ReadFile(filepath) // the file is inside the local directory
	if err != nil {
		return ModuleYml{}, err
	}

	var module ModuleYml
	err = yaml.Unmarshal(content, &module)

	if err != nil {
		return ModuleYml{}, err
	}

	module.filePath = filepath
	return module, nil
}

// input is of form github.com/osmosis-labs/osmosis/vXX/{PATH}
// returns PATH
func ParseFilePathFromImportPath(importPath string) string {
	splits := strings.Split(importPath, "/")
	pathSplits := splits[4:]
	return strings.Join(pathSplits, "/")
}
