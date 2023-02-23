package types

import "strings"

const (
	ModuleName = "rate-limited-ibc" // IBC at the end to avoid conflicts with the ibc prefix

)

var (
	// RouterKey is the message route. Can only contain
	// alphanumeric characters.
	RouterKey = strings.ReplaceAll(ModuleName, "-", "")
)
