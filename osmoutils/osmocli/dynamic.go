package osmocli

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

type Descriptor interface {
	GetCustomFlagOverrides() map[string]string
	AttachToUse(str string)
}

// fields that are not provided as arguments
var nonAttachableFields []string = []string{"sender", "pagination", "owner", "admin"}

// attachFieldsToUse extracts fields from reqP proto message and dynamically appends them into Use field
func attachFieldsToUse[reqP proto.Message](desc Descriptor) {
	req := osmoutils.MakeNew[reqP]()
	v := reflect.ValueOf(req).Type().Elem() // get underlying non-pointer struct
	var useField string
	for i := 0; i < v.NumField(); i++ {
		fn := pascalToKebab(v.Field(i).Name)

		// if a field is parsed from a flag, skip it
		if desc.GetCustomFlagOverrides()[fn] != "" || osmoutils.Contains(nonAttachableFields, fn) {
			continue
		}

		useField += fmt.Sprintf(" [%s]", fn)
	}

	desc.AttachToUse(useField)
}

// pascalToKebab converts PascalCase string to kebab-case string
func pascalToKebab(s string) string {
	reg := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	s = reg.ReplaceAllString(s, `${1}-${2}`)

	// Convert everything to lowercase
	return strings.ToLower(s)
}
