package osmocli

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/cosmos/cosmos-sdk/version"
)

type LongMetadata struct {
	BinaryName    string
	CommandPrefix string
}

func NewLongMetadata(moduleName string) LongMetadata {
	commandPrefix := fmt.Sprintf("$ %s q %s", version.AppName, moduleName)
	return LongMetadata{
		BinaryName:    version.AppName,
		CommandPrefix: commandPrefix,
	}
}

func FormatLongDescription(longString string, moduleName string) string {
	template, err := template.New("long_description").Parse(longString)
	if err != nil {
		panic("incorrectly configured long message")
	}
	bld := strings.Builder{}
	arg := NewLongMetadata(moduleName)
	err = template.Execute(&bld, arg)
	if err != nil {
		panic("incorrectly configured long message")
	}
	return strings.TrimSpace(bld.String())
}
