package osmoutils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// NewModuleAddressWithPrefix returns a new module address with the given prefix and identifier.
func NewModuleAddressWithPrefix(moduleName, prefix string, identifier []byte) sdk.AccAddress {
	key := append([]byte(prefix), identifier...)
	return address.Module(moduleName, key)
}
