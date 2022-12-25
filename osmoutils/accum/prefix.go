package accum

import "fmt"

// formatAccumPrefix returns the key prefix used for any
// accum module values to be stored in the KVStore.
// Returns "accum/{key}" as bytes.
func formatModulePrefixKey(key string) []byte {
	return []byte(fmt.Sprintf("accum/%s", key))
}

// formatAccimPrefix returns the key prefix used
// specifically for accumulator values in the KVStore.
// Returns "accum/acc/{name}" as bytes.
func formatAccumPrefixKey(name string) []byte {
	return formatModulePrefixKey(fmt.Sprintf("acc/%s", name))
}

// formatPositionPrefixKey returns the key prefix used
// specifically for position values in the KVStore.
// Returns "accum/pos/{address}" as bytes.
func formatPositionPrefixKey(address string) []byte {
	return formatAccumPrefixKey(fmt.Sprintf("pos/%s", address))
}
