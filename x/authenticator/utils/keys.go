package utils

import (
	"fmt"
	"strings"
)

const KeySeparator = "|"

// BuildKey creates a key by concatenating the provided elements with the key separator.
func BuildKey(elements ...interface{}) []byte {
	strElements := make([]string, len(elements))
	for i, element := range elements {
		strElements[i] = fmt.Sprint(element)
	}
	return []byte(strings.Join(strElements, KeySeparator) + KeySeparator)
}
