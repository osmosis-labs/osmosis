package types

import "strings"

const (
	ModuleName = "rate-limited-ibc" // IBC at the end to avoid conflicts with the ibc prefix

)

// RouterKey is the message route. Can only contain
// alphanumeric characters.
var RouterKey = strings.ReplaceAll(ModuleName, "-", "")
