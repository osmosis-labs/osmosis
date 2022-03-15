package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type PoolState struct {
	Assets []sdk.Coin
	Shares sdk.Coin
}
