package e2eTesting

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Account keeps a genesis account data.
type Account struct {
	Address sdk.AccAddress
	PrivKey cryptotypes.PrivKey
}
