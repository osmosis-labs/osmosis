package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/osmosis-labs/osmosis/v17/app/keepers"
	clmodel "github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/model"
	gammkeeper "github.com/osmosis-labs/osmosis/v17/x/gamm/keeper"
	gammmigration "github.com/osmosis-labs/osmosis/v17/x/gamm/types/migration"
	"github.com/osmosis-labs/osmosis/v17/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v17/x/poolmanager/types"
)

// createConcentratedPoolFromCFMM creates a new concentrated liquidity pool with the desiredDenom0 token as the
// token 0, links it with an existing CFMM pool, and returns the created pool.
// It uses pool manager module account as the creator of the pool.
// Returns error if desired denom 0 is not in associated with the CFMM pool.
// Returns error if CFMM pool does not have exactly 2 denoms.
// Returns error if pool creation fails.
func createConcentratedPoolFromCFMM(ctx sdk.Context, cfmmPoolIdToLinkWith uint64, desiredDenom0 string, accountKeeper authkeeper.AccountKeeper, gammKeeper gammkeeper.Keeper, poolmanagerKeeper poolmanager.Keeper) (poolmanagertypes.PoolI, error) {
	cfmmPool, err := gammKeeper.GetCFMMPool(ctx, cfmmPoolIdToLinkWith)
	if err != nil {
		return nil, err
	}

	poolmanagerModuleAcc := accountKeeper.GetModuleAccount(ctx, poolmanagertypes.ModuleName)
	poolCreatorAddress := poolmanagerModuleAcc.GetAddress()

	poolLiquidity := cfmmPool.GetTotalPoolLiquidity(ctx)
	if len(poolLiquidity) != 2 {
		return nil, ErrMustHaveTwoDenoms
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
		return nil, NoDesiredDenomInPoolError{desiredDenom0}
	}

	// Swap fee is 0.2%, which is an authorized spread factor.
	spreadFactor := cfmmPool.GetSpreadFactor(ctx)

	createPoolMsg := clmodel.NewMsgCreateConcentratedPool(poolCreatorAddress, desiredDenom0, denom1, TickSpacing, spreadFactor)
	concentratedPool, err := poolmanagerKeeper.CreateConcentratedPoolAsPoolManager(ctx, createPoolMsg)
	if err != nil {
		return nil, err
	}

	return concentratedPool, nil
}

// createCanonicalConcentratedLiquidityPoolAndMigrationLink creates a new concentrated liquidity pool from an existing CFMM pool.
// This method calls OverwriteMigrationRecords, which creates a migration link between the CFMM/CL pool as well as migrates the
// gauges and distribution records from the CFMM pool to the new CL pool.
// Returns error if fails to create concentrated liquidity pool from CFMM pool.
func createCanonicalConcentratedLiquidityPoolAndMigrationLink(ctx sdk.Context, cfmmPoolId uint64, desiredDenom0 string, keepers *keepers.AppKeepers) (poolmanagertypes.PoolI, error) {
	concentratedPool, err := createConcentratedPoolFromCFMM(ctx, cfmmPoolId, desiredDenom0, *keepers.AccountKeeper, *keepers.GAMMKeeper, *keepers.PoolManagerKeeper)
	if err != nil {
		return nil, err
	}

	// Set the migration link in x/gamm.
	// This will also migrate the CFMM distribution records to point to the new CL pool.
	err = keepers.GAMMKeeper.OverwriteMigrationRecordsAndRedirectDistrRecords(ctx, gammmigration.MigrationRecords{
		BalancerToConcentratedPoolLinks: []gammmigration.BalancerToConcentratedPoolLink{
			{
				BalancerPoolId: cfmmPoolId,
				ClPoolId:       concentratedPool.GetId(),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return concentratedPool, nil
}
