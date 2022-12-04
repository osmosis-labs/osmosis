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
	Short         string

	// Newline Example:
	ExampleHeader string
}

func NewLongMetadata(moduleName string) *LongMetadata {
	commandPrefix := fmt.Sprintf("$ %s q %s", version.AppName, moduleName)
	return &LongMetadata{
		BinaryName:    version.AppName,
		CommandPrefix: commandPrefix,
	}
}

func (m *LongMetadata) WithShort(short string) *LongMetadata {
	m.Short = short
	return m
}

func FormatLongDesc(longString string, meta *LongMetadata) string {
	template, err := template.New("long_description").Parse(longString)
	if err != nil {
		panic("incorrectly configured long message")
	}
	bld := strings.Builder{}
	meta.ExampleHeader = "\n\nExample:"
	err = template.Execute(&bld, meta)
	if err != nil {
		panic("incorrectly configured long message")
	}
	return strings.TrimSpace(bld.String())
}

func FormatLongDescDirect(longString string, moduleName string) string {
	return FormatLongDesc(longString, NewLongMetadata(moduleName))
}
