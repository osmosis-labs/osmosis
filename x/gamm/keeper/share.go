package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/gamm/types"
)

func (k Keeper) MintPoolShareToAccount(ctx sdk.Context, pool types.PoolI, addr sdk.AccAddress, amount sdk.Int) error {
	amt := sdk.Coins{
		sdk.NewCoin(types.GetPoolShareDenom(pool.GetId()), amount),
	}

	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, amt)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, amt)
	if err != nil {
		return err
	}

	pool.AddTotalShare(amount)

	return nil
}

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

	pool.SubTotalShare(amount)

	return nil
}
