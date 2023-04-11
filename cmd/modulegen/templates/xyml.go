package templates

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type XYml struct {
	// Path to simtypes e.g. "github.com/osmosis-labs/osmosis/v15/simulation"
	SimtypesPath string `yaml:"simtypes_path"`
	// Path to module e.g. "github.com/osmosis-labs/osmosis/v15/x/testmodule"
	ModulePath string `yaml:"module_path"`

	ModuleName string `yaml:"module_name"`

	// list of all queries, key is the query name, e.g. `GetArithmeticTwap`
	Queries map[string]YmlQueryDescriptor `yaml:"queries"`

	filePath string
}

type YmlQueryDescriptor struct {
	ProtoWrapper *ProtoWrapperDescriptor `yaml:"proto_wrapper,omitempty"`
	Cli          *CliDescriptor
}

type YmlTxDescriptor struct {
}

type ProtoWrapperDescriptor struct {
	DefaultValues map[string]string `yaml:"default_values"`
	QueryFunc     string            `yaml:"query_func"`
	Response      string            `yaml:"response"`
}

type CliDescriptor struct{}

func ReadXYmlFile(filepath string) (XYml, error) {
	content, err := os.ReadFile(filepath) // the file is inside the local directory
	if err != nil {
		return XYml{}, err
	}

	var module XYml
	err = yaml.Unmarshal(content, &module)

	if err != nil {
		return XYml{}, err
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

// input is of form cmd/modulegen/templates/x/{PATH}
// returns PATH folder and go file PATH
func ParseXFilePath(filePath string) (string, string) {
	dir := filepath.Dir(filePath)
	folderPath, err := filepath.Rel("cmd/modulegen/templates/x", dir)
	if err != nil {
		panic(err)
	}
	goFilePath := filepath.Join(folderPath, filepath.Base(filePath[:len(filePath)-4]+".go"))
	return folderPath, goFilePath
}
