package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/gamm/types"
)

func (k Keeper) MintPoolShareToAccount(ctx sdk.Context, poolAcc types.PoolAccountI, addr sdk.AccAddress, amount sdk.Int) error {
	amt := sdk.Coins{
		sdk.NewCoin(types.GetPoolShareDenom(poolAcc.GetId()), amount),
	}

	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, amt)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, amt)
	if err != nil {
		return err
	}

	poolAcc.AddTotalShare(amount)

	return nil
}

func (k Keeper) BurnPoolShareFromAccount(ctx sdk.Context, poolAcc types.PoolAccountI, addr sdk.AccAddress, amount sdk.Int) error {
	amt := sdk.Coins{
		sdk.NewCoin(types.GetPoolShareDenom(poolAcc.GetId()), amount),
	}

	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, amt)
	if err != nil {
		return err
	}

	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, amt)
	if err != nil {
		return err
	}

	poolAcc.SubTotalShare(amount)

	return nil
}
