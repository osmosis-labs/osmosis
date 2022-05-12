package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func (k Keeper) ApplyUpdateParam(ctx sdk.Context, update types.UpdatePoolParam) error {
	pool, err := k.GetPoolAndPoke(ctx, update.PoolId)
	if err != nil {
		return err
	}
	pool.ApplyUpdateParam(ctx, update)
	return k.SetPool(ctx, pool)
}

func (k Keeper) HandleUpdatePoolParamsProposal(ctx sdk.Context, p *types.UpdatePoolParamsProposal) error {
	for _, update := range p.Updates {
		if err := k.ApplyUpdateParam(ctx, update); err != nil {
			return err
		}
	}
	return nil
}
