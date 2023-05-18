package v16

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/osmosis-labs/osmosis/v15/app/keepers"
	clkeeper "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
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

	// Swap fee is 0.2%, which is an authorized swap fee.
	swapFee := cfmmPool.GetSwapFee(ctx)

	createPoolMsg := clmodel.NewMsgCreateConcentratedPool(poolCreatorAddress, desiredDenom0, denom1, TickSpacing, swapFee)
	concentratedPool, err := poolmanagerKeeper.CreateConcentratedPoolAsPoolManager(ctx, createPoolMsg)
	if err != nil {
		return nil, err
	}

	return concentratedPool, nil
}

// createCanonicalConcentratedLiquidityPoolAndMigrationLink creates a new concentrated liquidity pool from an existing
// CFMM pool, and migrates the gauges and distribution records from the CFMM pool to the new CL pool.
// Additionally, it creates a migration link between the CFMM pool and the CL pool and stores it in x/gamm.
// Returns error if fails to create concentrated liquidity pool from CFMM pool.
// Returns error if fails to get gauges for CFMM pool.
// Returns error if fails to get gauge for the concentrated liquidity pool.
func createCanonicalConcentratedLiquidityPoolAndMigrationLink(ctx sdk.Context, cfmmPoolId uint64, desiredDenom0 string, keepers *keepers.AppKeepers) (poolmanagertypes.PoolI, error) {
	concentratedPool, err := createConcentratedPoolFromCFMM(ctx, cfmmPoolId, desiredDenom0, *keepers.AccountKeeper, *keepers.GAMMKeeper, *keepers.PoolManagerKeeper)
	if err != nil {
		return nil, err
	}

	// Get CFMM gauges
	cfmmGauges, err := keepers.PoolIncentivesKeeper.GetGaugesForCFMMPool(ctx, cfmmPoolId)
	if err != nil {
		return nil, err
	}

	if len(cfmmGauges) == 0 {
		return nil, ErrNoGaugeToRedirect
	}

	// Get longest gauge duration from balancer.
	longestDurationGauge := cfmmGauges[0]
	for i := 1; i < len(cfmmGauges); i++ {
		if cfmmGauges[i].DistributeTo.Duration > longestDurationGauge.DistributeTo.Duration {
			longestDurationGauge = cfmmGauges[i]
		}
	}

	// Get concentrated liquidity gauge duration.
	distributionEpochDuration := keepers.IncentivesKeeper.GetEpochInfo(ctx).Duration

	// Get concentrated gauge correspondng to the distribution epoch duration.
	concentratedGaugeId, err := keepers.PoolIncentivesKeeper.GetPoolGaugeId(ctx, concentratedPool.GetId(), distributionEpochDuration)
	if err != nil {
		return nil, err
	}

	// Iterate through all the distr records, and redirect the old balancer gauge to the new concentrated gauge.
	distrInfo := keepers.PoolIncentivesKeeper.GetDistrInfo(ctx)
	for i, distrRecord := range distrInfo.Records {
		if distrRecord.GaugeId == longestDurationGauge.Id {
			distrInfo.Records[i].GaugeId = concentratedGaugeId
		}
	}

	// Set the new distr info.
	keepers.PoolIncentivesKeeper.SetDistrInfo(ctx, distrInfo)

	// Set the migration link in x/gamm.
	keepers.GAMMKeeper.OverwriteMigrationRecords(ctx, gammtypes.MigrationRecords{
		BalancerToConcentratedPoolLinks: []gammtypes.BalancerToConcentratedPoolLink{
			{
				BalancerPoolId: cfmmPoolId,
				ClPoolId:       concentratedPool.GetId(),
			},
		},
	})

	return concentratedPool, nil
}

func createFullRangePositionNoSend(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, coins sdk.Coins, clKeeper clkeeper.Keeper) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, err error) {
	// Check that exactly two coins are provided.
	if len(coins) != 2 {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, types.NumCoinsError{NumCoins: len(coins)}
	}

	concentratedPool, err := clKeeper.GetConcentratedPoolById(ctx, poolId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, err
	}

	// Defense in depth, ensure coins provided match the pool's token denominations.
	if coins.AmountOf(concentratedPool.GetToken0()).LTE(sdk.ZeroInt()) {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, types.Amount0IsNegativeError{Amount0: coins.AmountOf(concentratedPool.GetToken0())}
	}
	if coins.AmountOf(concentratedPool.GetToken1()).LTE(sdk.ZeroInt()) {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, types.Amount1IsNegativeError{Amount1: coins.AmountOf(concentratedPool.GetToken1())}
	}
	if len(coins) != 2 {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, types.NumCoinsError{NumCoins: len(coins)}
	}

	// Create a full range (min to max tick) concentrated liquidity position.
	positionId, amount0, amount1, liquidity, joinTime, _, _, err = clKeeper.CreatePositionNoSendUnsafe(ctx, concentratedPool.GetId(), owner, coins, sdk.ZeroInt(), sdk.ZeroInt(), types.MinTick, types.MaxTick)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, err
	}

	return positionId, amount0, amount1, liquidity, joinTime, nil
}
