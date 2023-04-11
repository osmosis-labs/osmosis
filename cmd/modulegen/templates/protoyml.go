package templates

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ProtoYml struct {
	// Path to module e.g. "github.com/osmosis-labs/osmosis/v15/x/testmodule"
	ModulePath string `yaml:"module_path"`

	ModuleName string `yaml:"module_name"`

	// import path to proto e.g. "cosmos/base/v1beta1/coin.proto"
	ImportPath map[string]ImportPathDescriptor `yaml:"import_path"`

	// list of all params, key is the param name, e.g. `AuthorizedTickSpacing`
	Params map[string]YmlParamDescriptor `yaml:"params"`

	filePath string
}

type YmlParamDescriptor struct {
	// e.g. authorized_tick_spacing
	Name string `yaml:"name"`

	// e.g. repeated uint64
	Type string `yaml:"type"`

	Id uint `yaml:"id"`

	Tags map[string]TagDescriptor `yaml:"tags"`
}

type TagDescriptor struct {
	// e.g. nullable
	Name string `yaml:"tag_name"`
	// e.g. false
	Val string `yaml:"tag_val"`
}

type ImportPathDescriptor struct {
	// e.g. cosmos/base/v1beta1/coin.proto
	Name string `yaml:"name"`
}

func ReadProtoYmlFile(filepath string) (ProtoYml, error) {
	content, err := os.ReadFile(filepath) // the file is inside the local directory
	if err != nil {
		return ProtoYml{}, err
	}

	var module ProtoYml
	err = yaml.Unmarshal(content, &module)

	if err != nil {
		return ProtoYml{}, err
	}

	module.filePath = filepath
	return module, nil
}
