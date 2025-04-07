package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/treasury/types"
)

// GetTreasuryModuleAccount returns treasury ModuleAccount
func (k Keeper) GetTreasuryModuleAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// GetReservePoolBalance returns the amount of Melody in the reserve pool.
func (k Keeper) GetReservePoolBalance(ctx sdk.Context) sdk.Coin {
	return k.BankKeeper.GetBalance(ctx, k.GetTreasuryModuleAccount(ctx).GetAddress(), appparams.BaseCoinUnit)
}
