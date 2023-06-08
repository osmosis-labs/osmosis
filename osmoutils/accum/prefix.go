package accum

import "fmt"

const (
	modulePrefix      = "accum"
	accumulatorPrefix = "acc"
	positionPrefix    = "pos"
)

// formatAccumPrefix returns the key prefix used
// specifically for accumulator values in the KVStore.
// Returns "accum/acc/{accumName}" as bytes.
func formatAccumPrefixKey(accumName string) []byte {
	return []byte(fmt.Sprintf(modulePrefix+"/"+accumulatorPrefix+"/%s", accumName))
}

// FormatPositionPrefixKey returns the key prefix used
// specifically for position values in the KVStore.
// Returns "accum/pos/{accumName}/{name}" as bytes.
func FormatPositionPrefixKey(accumName, name string) []byte {
	return []byte(fmt.Sprintf(modulePrefix+"/"+positionPrefix+"/%s/%s", accumName, name))
}
