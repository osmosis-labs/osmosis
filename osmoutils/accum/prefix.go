package accum

import "fmt"

const (
	modulePrefix      = "accum"
	accumulatorPrefix = "acc"
	positionPrefix    = "pos"
)

// formatAccumPrefix returns the key prefix used
// specifically for accumulator values in the KVStore.
// Returns "accum/acc/{name}" as bytes.
func formatAccumPrefixKey(name string) []byte {
	return []byte(fmt.Sprintf(modulePrefix+"/"+accumulatorPrefix+"/%s", name))
}

// FormatPositionPrefixKey returns the key prefix used
// specifically for position values in the KVStore.
// Returns "accum/acc/pos/{accumName}/{name}" as bytes.
func FormatPositionPrefixKey(accumName, name string) []byte {
	return []byte(fmt.Sprintf(modulePrefix+"/"+accumulatorPrefix+"/"+positionPrefix+"/%s/%s", accumName, name))
}
