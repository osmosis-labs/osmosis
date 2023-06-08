package accum

import "fmt"

const (
	modulePrefix      = "accum"
	accumulatorPrefix = "acc"
	positionPrefix    = "pos"
	keySeparator      = "||" // needs to be different from other modules.

	accumPrefixKey    = modulePrefix + keySeparator + accumulatorPrefix + keySeparator
	positionPrefixKey = modulePrefix + keySeparator + positionPrefix + keySeparator
)

// formatAccumPrefix returns the key prefix used
// specifically for accumulator values in the KVStore.
// Returns "accum||acc||{accumName}" as bytes.
func formatAccumPrefixKey(accumName string) []byte {
	return []byte(fmt.Sprintf(accumPrefixKey+"%s", accumName))
}

// FormatPositionPrefixKey returns the key prefix used
// specifically for position values in the KVStore.
// Returns "accum||pos||{accumName}||{name}" as bytes.
// We use a different key separator, namely `||`, to separate the accumulator name and the position name.
// This is because we require that accumName does not contain this as a substring.
func FormatPositionPrefixKey(accumName, name string) []byte {
	return []byte(fmt.Sprintf(positionPrefixKey+"%s"+keySeparator+"%s", accumName, name))
}
