package v17

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	v16 "github.com/osmosis-labs/osmosis/v16/app/upgrades/v16"
	clmodel "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/model"
	gammkeeper "github.com/osmosis-labs/osmosis/v16/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v16/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

// createConcentratedPoolFromCFMM creates a new concentrated liquidity pool with the desiredDenom0 token as the
// token 0, links it with an existing CFMM pool, and returns the created pool.
// It uses pool manager module account as the creator of the pool.
// Returns error if desired denom 0 is not in associated with the CFMM pool.
// Returns error if CFMM pool does not have exactly 2 denoms.
// Returns error if pool creation fails.
func createConcentratedPoolFromCFMM(ctx sdk.Context, cfmmPoolIdToLinkWith uint64, desiredDenom0 string, spreadFactor sdk.Dec, accountKeeper authkeeper.AccountKeeper, gammKeeper gammkeeper.Keeper, poolmanagerKeeper poolmanager.Keeper) (poolmanagertypes.PoolI, error) {
	cfmmPool, err := gammKeeper.GetCFMMPool(ctx, cfmmPoolIdToLinkWith)
	if err != nil {
		return nil, err
	}

	poolmanagerModuleAcc := accountKeeper.GetModuleAccount(ctx, poolmanagertypes.ModuleName)
	poolCreatorAddress := poolmanagerModuleAcc.GetAddress()

	poolLiquidity := cfmmPool.GetTotalPoolLiquidity(ctx)
	if len(poolLiquidity) != 2 {
		return nil, v16.ErrMustHaveTwoDenoms
	}

	foundDenom0 := false
	denom1 := ""
	for _, coin := range poolLiquidity {
		if coin.Denom == desiredDenom0 {
			foundDenom0 = true
		} else {
			denom1 = coin.Denom
		}
	}

	if !foundDenom0 {
		return nil, v16.NoDesiredDenomInPoolError{DesiredDenom: desiredDenom0}
	}

	createPoolMsg := clmodel.NewMsgCreateConcentratedPool(poolCreatorAddress, desiredDenom0, denom1, TickSpacing, spreadFactor)
	concentratedPool, err := poolmanagerKeeper.CreateConcentratedPoolAsPoolManager(ctx, createPoolMsg)
	if err != nil {
		return nil, err
	}

	return concentratedPool, nil
}
