package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper/internal/events"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func (k Keeper) applyJoinPoolStateChange(ctx sdk.Context, pool types.PoolI, joiner sdk.AccAddress, numShares sdk.Int, joinCoins sdk.Coins) error {
	err := k.bankKeeper.SendCoins(ctx, joiner, pool.GetAddress(), joinCoins)
	if err != nil {
		return err
	}

	err = k.MintPoolShareToAccount(ctx, pool, joiner, numShares)
	if err != nil {
		return err
	}

	err = k.SetPool(ctx, pool)
	if err != nil {
		return err
	}

	events.EmitAddLiquidityEvent(ctx, joiner, pool.GetId(), joinCoins)
	k.hooks.AfterJoinPool(ctx, joiner, pool.GetId(), joinCoins, numShares)
	k.RecordTotalLiquidityIncrease(ctx, joinCoins)
	return nil
}

func (k Keeper) applyExitPoolStateChange(ctx sdk.Context, pool types.PoolI, exiter sdk.AccAddress, numShares sdk.Int, exitCoins sdk.Coins) error {
	err := k.bankKeeper.SendCoins(ctx, pool.GetAddress(), exiter, exitCoins)
	if err != nil {
		return err
	}

	err = k.BurnPoolShareFromAccount(ctx, pool, exiter, numShares)
	if err != nil {
		return err
	}

	err = k.SetPool(ctx, pool)
	if err != nil {
		return err
	}

	events.EmitRemoveLiquidityEvent(ctx, exiter, pool.GetId(), exitCoins)
	k.hooks.AfterExitPool(ctx, exiter, pool.GetId(), numShares, exitCoins)
	k.RecordTotalLiquidityDecrease(ctx, exitCoins)
	return nil
}

// MintPoolShareToAccount attempts to mint shares of a GAMM denomination to the
// specified address returning an error upon failure. Shares are minted using
// the x/gamm module account.
func (k Keeper) MintPoolShareToAccount(ctx sdk.Context, pool types.PoolI, addr sdk.AccAddress, amount sdk.Int) error {
	amt := sdk.NewCoins(sdk.NewCoin(types.GetPoolShareDenom(pool.GetId()), amount))

	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, amt)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, amt)
	if err != nil {
		return err
	}

	return nil
}

// BurnPoolShareFromAccount burns `amount` of the given pools shares held by `addr`.
func (k Keeper) BurnPoolShareFromAccount(ctx sdk.Context, pool types.PoolI, addr sdk.AccAddress, amount sdk.Int) error {
	amt := sdk.Coins{
		sdk.NewCoin(types.GetPoolShareDenom(pool.GetId()), amount),
	}

	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, amt)
	if err != nil {
		return err
	}

	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, amt)
	if err != nil {
		return err
	}

	return nil
}
