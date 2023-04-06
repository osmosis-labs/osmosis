package accum

import "fmt"

const (
	modulePrefix      = "accum"
	accumulatorPrefix = "acc"
	positionPrefix    = "pos"
)

// formatAccumPrefix returns the key prefix used for any
// accum module values to be stored in the KVStore.
// Returns "accum/{key}" as bytes.
func formatModulePrefixKey(key string) []byte {
	return []byte(fmt.Sprintf("%s/%s", modulePrefix, key))
}

// formatAccumPrefix returns the key prefix used
// specifically for accumulator values in the KVStore.
// Returns "accum/acc/{name}" as bytes.
func formatAccumPrefixKey(name string) []byte {
	return formatModulePrefixKey(fmt.Sprintf("%s/%s", accumulatorPrefix, name))
}

// FormatPositionPrefixKey returns the key prefix used
// specifically for position values in the KVStore.
// Returns "accum/pos/{accumName}/{name}" as bytes.
func FormatPositionPrefixKey(accumName, name string) []byte {
	return formatAccumPrefixKey(fmt.Sprintf("%s/%s/%s", positionPrefix, accumName, name))
}
