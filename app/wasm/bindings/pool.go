package wasmbindings

import sdk "github.com/cosmos/cosmos-sdk/types"

type PoolAssets struct {
	Assets []sdk.Coin
	Shares sdk.Coin
}
