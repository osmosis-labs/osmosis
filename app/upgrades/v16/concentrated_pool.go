package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/osmosis-labs/osmosis/v15/app/keepers"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	incentiveskeeper "github.com/osmosis-labs/osmosis/v15/x/incentives/keeper"
	incentivestypes "github.com/osmosis-labs/osmosis/v15/x/incentives/types"
	poolincentiveskeeper "github.com/osmosis-labs/osmosis/v15/x/pool-incentives/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// getGaugesForCFMMPool returns the gauges associated with the given CFMM pool ID, by first retrieving
// the lockable durations for the pool, then using them to query the pool incentives keeper for the
// gauge IDs associated with each duration, and finally using the incentives keeper to retrieve the
// actual gauges from the retrieved gauge IDs.
func getGaugesForCFMMPool(ctx sdk.Context, incentivesKeeper incentiveskeeper.Keeper, poolIncentivesKeeper poolincentiveskeeper.Keeper, poolId uint64) ([]incentivestypes.Gauge, error) {
	lockableDurations := poolIncentivesKeeper.GetLockableDurations(ctx)
	cfmmGaugeIds := make([]uint64, 0, len(lockableDurations))
	for _, duration := range lockableDurations {
		gaugeId, err := poolIncentivesKeeper.GetPoolGaugeId(ctx, poolId, duration)
		if err != nil {
			return nil, err
		}
		cfmmGaugeIds = append(cfmmGaugeIds, gaugeId)
	}

	cfmmGauges, err := incentivesKeeper.GetGaugeFromIDs(ctx, cfmmGaugeIds)
	if err != nil {
		return nil, err
	}

	return cfmmGauges, nil
}

// createConcentratedPoolFromCFMM creates a new concentrated liquidity pool with the desiredDenom0 token as the
// token 0, links it with an existing CFMM pool, and returns the created pool.
// It first creates a module account for pool manager module, and then creates the pool from
// that module account.
// Returns error if desired denom 0 is not in associated with the CFMM pool.
// Returns error if CFMM pool does not have exactly 2 denoms.
// Returns error if pool creation fails.
func createConcentratedPoolFromCFMM(ctx sdk.Context, cfmmPoolIdToLinkWith uint64, desiredDenom0 string, accountKeeper authkeeper.AccountKeeper, gammKeeper gammkeeper.Keeper, poolmanagerKeeper poolmanager.Keeper) (poolmanagertypes.PoolI, error) {
	cfmmPool, err := gammKeeper.GetCFMMPool(ctx, cfmmPoolIdToLink)
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

	// TODO: confirm pre-launch that it is the same for CL as in balancer.
	swapFee := cfmmPool.GetSwapFee(ctx)

	createPoolMsg := clmodel.NewMsgCreateConcentratedPool(poolCreatorAddress, desiredDenom0, denom1, tickSpacing, exponentAtPriceOne, swapFee)
	concentratedPool, err := poolmanagerKeeper.CreatePoolZeroLiquidityNoFee(ctx, createPoolMsg)
	if err != nil {
		return nil, err
	}

	return concentratedPool, nil
}

// createCanonicalConcentratedLiuqidityPoolAndMigrationLink creates a new concentrated liquidity pool from an existing
// CFMM pool, and migrates the gauges and distribution records from the CFMM pool to the new CL pool.
// Additionally, it creates a migration link between the CFMM pool and the CL pool and stores it in x/gamm.
// Returns error if fails to create concentrated liquidity pool from CFMM pool.
// Returns error if fails to get gauges for CFMM pool.
// Returns error if fails to get gauge for the concentrated liquidity pool.
func createCanonicalConcentratedLiuqidityPoolAndMigrationLink(ctx sdk.Context, cfmmPoolId uint64, desiredDenom0 string, keepers *keepers.AppKeepers) error {
	concentratedPool, err := createConcentratedPoolFromCFMM(ctx, cfmmPoolId, desiredDenom0, *keepers.AccountKeeper, *keepers.GAMMKeeper, *keepers.PoolManagerKeeper)
	if err != nil {
		return err
	}

	// Get CFMM gauges
	cfmmGauges, err := getGaugesForCFMMPool(ctx, *keepers.IncentivesKeeper, *keepers.PoolIncentivesKeeper, cfmmPoolId)
	if err != nil {
		return err
	}

	// Get distribution epoch duration.
	distributionEpochDuration := keepers.IncentivesKeeper.GetEpochInfo(ctx).Duration

	// Get concentrated gauge correspondng to the distribution epoch duration.
	concentratedGaugeId, err := keepers.PoolIncentivesKeeper.GetPoolGaugeId(ctx, concentratedPool.GetId(), distributionEpochDuration)
	if err != nil {
		return err
	}

	// Find the gauge that corresponds to the distribution epoch duration.
	gaugeIdToRedirectTo := uint64(0)
	foundDesiredGauge := false
	for _, cfmmGauge := range cfmmGauges {
		if cfmmGauge.DistributeTo.Duration == distributionEpochDuration {
			gaugeIdToRedirectTo = cfmmGauge.Id
			foundDesiredGauge = true
		}
	}
	if !foundDesiredGauge {
		return CouldNotFindGaugeToRedirectError{DistributionEpochDuration: distributionEpochDuration}
	}

	// Iterate through all the distr records, and redirect the old balancer gauge to the new concentrated gauge.
	distrInfo := keepers.PoolIncentivesKeeper.GetDistrInfo(ctx)
	for i, distrRecord := range distrInfo.Records {
		if distrRecord.GaugeId == gaugeIdToRedirectTo {
			distrInfo.Records[i].GaugeId = concentratedGaugeId
		}
	}

	// Set the new distr info.
	keepers.PoolIncentivesKeeper.SetDistrInfo(ctx, distrInfo)

	// Set the migration link in x/gamm.
	keepers.GAMMKeeper.SetMigrationInfo(ctx, gammtypes.MigrationRecords{
		BalancerToConcentratedPoolLinks: []gammtypes.BalancerToConcentratedPoolLink{
			{
				BalancerPoolId: cfmmPoolId,
				ClPoolId:       concentratedPool.GetId(),
			},
		},
	})

	return nil
}
