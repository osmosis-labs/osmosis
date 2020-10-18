package pool

import (
	"github.com/c-osmosis/osmosis/x/gamm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type lp struct {
	denom      string
	bankKeeper bankkeeper.Keeper
}

func (p lp) pushPoolShare(ctx sdk.Context, to sdk.AccAddress, amount sdk.Int) error {
	lp := sdk.Coin{Denom: p.denom, Amount: amount}
	return p.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, to, sdk.Coins{lp})
}

func (p lp) pullPoolShare(ctx sdk.Context, from sdk.AccAddress, amount sdk.Int) error {
	lp := sdk.Coin{Denom: p.denom, Amount: amount}
	return p.bankKeeper.SendCoinsFromAccountToModule(ctx, from, types.ModuleName, sdk.Coins{lp})
}

func (p lp) mintPoolShare(ctx sdk.Context, amount sdk.Int) error {
	lp := sdk.Coin{Denom: p.denom, Amount: amount}
	return p.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.Coins{lp})
}

func (p lp) burnPoolShare(ctx sdk.Context, amount sdk.Int) error {
	lp := sdk.Coin{Denom: p.denom, Amount: amount}
	return p.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.Coins{lp})
}
