package gov

import (
	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func HandleSetSwapFeeProposal(ctx sdk.Context, k keeper.Keeper, p *types.SetSwapFeeProposal) error {
	poolId := p.Content.PoolId
	pool, err := k.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return err
	}
	err = pool.SetSwapFee(ctx, p.Content.SwapFee)
	if err != nil {
		return err
	}
	err = k.SetPool(ctx, pool)
	if err != nil {
		return err
	}

	event := sdk.NewEvent(
		types.TypeEvtSetSwapFee,
		sdk.NewAttribute(types.AttributeKeySwapFee, p.Content.SwapFee.String()),
	)
	ctx.EventManager().EmitEvent(event)
	return nil
}

func HandleSetExitFeeProposal(ctx sdk.Context, k keeper.Keeper, p *types.SetExitFeeProposal) error {
	poolId := p.Content.PoolId
	pool, err := k.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return err
	}
	err = pool.SetExitFee(ctx, p.Content.ExitFee)
	if err != nil {
		return err
	}
	err = k.SetPool(ctx, pool)
	if err != nil {
		return err
	}

	event := sdk.NewEvent(
		types.TypeEvtSetExitFee,
		sdk.NewAttribute(types.AttributeKeySwapFee, p.Content.ExitFee.String()),
	)
	ctx.EventManager().EmitEvent(event)
	return nil
}
