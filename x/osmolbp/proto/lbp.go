package proto

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetAddress returns the address of an lbp.
// If the lbp address is not bech32 valid, it panics.
func (lbp LBP) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(lbp.Address)
	if err != nil {
		panic(fmt.Sprintf("could not bech32 decode address of lbp with id: %d", lbp.Id))
	}
	return addr
}
