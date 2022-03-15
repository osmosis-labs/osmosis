package wasm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/v7/app/wasm/types"
	gammkeeper "github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
)

type QueryPlugin struct {
	gammKeeper *gammkeeper.Keeper
}

// NewQueryPlugin constructor
func NewQueryPlugin(
	gammK *gammkeeper.Keeper,
) *QueryPlugin {
	return &QueryPlugin{
		gammKeeper: gammK,
	}
}

func (qp QueryPlugin) GetPoolState(ctx sdk.Context, poolId uint64) (*types.PoolState, error) {
	poolData, err := qp.gammKeeper.GetPool(ctx, poolId)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gamm get pool")
	}
	var poolState types.PoolState
	poolAssets := poolData.GetAllPoolAssets()
	for _, asset := range poolAssets {
		poolState.Assets = append(poolState.Assets, asset.Token)
	}
	poolState.Shares = poolData.GetTotalShares()

	return &poolState, nil
}
