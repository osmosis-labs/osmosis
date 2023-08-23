package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
}

type ProtorevKeeper interface {
	GetPoolForDenomPair(ctx sdk.Context, baseDenom, denomToMatch string) (uint64, error)
}
