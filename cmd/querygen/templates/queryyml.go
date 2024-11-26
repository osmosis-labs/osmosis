package templates

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type QueryYml struct {
	// Keeper struct descriptor
	Keeper Keeper `yaml:"keeper"`
	// Path to client folder e.g. "github.com/osmosis-labs/osmosis/v27/x/twap/client"
	ClientPath string `yaml:"client_path"`
	// list of all queries, key is the query name, e.g. `GetArithmeticTwap`
	Queries map[string]YmlQueryDescriptor `yaml:"queries"`

	protoPath string
}

type Keeper struct {
	// e.g. github.com/osmosis-labs/osmosis/v27/x/twap
	Path string `yaml:"path"`
	// e.g. Keeper
	Struct string `yaml:"struct"`
}

type YmlQueryDescriptor struct {
	ProtoWrapper *ProtoWrapperDescriptor `yaml:"proto_wrapper,omitempty"`
	Cli          *CliDescriptor
}

type ProtoWrapperDescriptor struct {
	DefaultValues map[string]string `yaml:"default_values"`
	QueryFunc     string            `yaml:"query_func"`
	Response      string            `yaml:"response"`
}

type CliDescriptor struct{}

func ReadYmlFile(filepath string) (QueryYml, error) {
	content, err := os.ReadFile(filepath) // the file is inside the local directory
	if err != nil {
		return QueryYml{}, err
	}
	var query QueryYml
	err = yaml.Unmarshal(content, &query)
	if err != nil {
		return QueryYml{}, err
	}
	query.protoPath = filepath
	return query, nil
}

// input is of form github.com/osmosis-labs/osmosis/vXX/{PATH}
// returns PATH
func ParseFilePathFromImportPath(importPath string) string {
	splits := strings.Split(importPath, "/")
	pathSplits := splits[4:]
	return strings.Join(pathSplits, "/")
}
