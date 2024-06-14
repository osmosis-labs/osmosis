package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StakingKeeper interface {
	BondDenom(ctx context.Context) (string, error)
}

type TxFeesKeeperI interface {
	IsFeeToken(ctx sdk.Context, denom string) (bool, error)
}
