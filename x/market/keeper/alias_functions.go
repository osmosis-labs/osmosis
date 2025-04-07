package keeper

import (
	"context"
	"github.com/osmosis-labs/osmosis/v27/x/market/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GetMarketAccount returns market ModuleAccount
func (k Keeper) GetMarketAccount(ctx context.Context) authtypes.ModuleAccountI {
	return k.AccountKeeper.GetModuleAccount(ctx, types.ModuleName)
}
