package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type LockI interface {
	GetOwner() string
	Amount() sdk.Coins
}
