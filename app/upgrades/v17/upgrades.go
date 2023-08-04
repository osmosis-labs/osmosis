package v17

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	cltypes "github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/types"
	gammmigration "github.com/osmosis-labs/osmosis/v17/x/gamm/types/migration"
	poolManagerTypes "github.com/osmosis-labs/osmosis/v17/x/poolmanager/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v17/x/superfluid/types"

	"github.com/osmosis-labs/osmosis/v17/app/keepers"
	"github.com/osmosis-labs/osmosis/v17/app/upgrades"
	"github.com/osmosis-labs/osmosis/v17/x/protorev/types"
)

const (
	mainnetChainID     = "osmosis-1"
	e2eChainA          = "osmo-test-a"
	e2eChainB          = "osmo-test-b"
	balancerWeight     = 1
	stableWeight       = 4
	concentratedWeight = 300
	cosmwasmWeight     = 300
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		poolLinks := []gammmigration.BalancerToConcentratedPoolLink{}
		fullRangeCoinsUsed := sdk.Coins{}

		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		communityPoolAddress := keepers.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)

		if ctx.ChainID() == mainnetChainID || ctx.ChainID() == e2eChainA || ctx.ChainID() == e2eChainB {
			// Upgrades specific balancer pools to concentrated liquidity pools and links them to their CL equivalent.
			ctx.Logger().Info(fmt.Sprintf("Chain ID is %s, running mainnet upgrade handler", ctx.ChainID()))
			err = mainnetUpgradeHandler(ctx, keepers, communityPoolAddress, &poolLinks, &fullRangeCoinsUsed)
		} else {
			// Upgrades all existing balancer pools to concentrated liquidity pools and links them to their CL equivalent.
			ctx.Logger().Info(fmt.Sprintf("Chain ID is %s, running testnet upgrade handler", ctx.ChainID()))
			err = testnetUpgradeHandler(ctx, keepers, communityPoolAddress, &poolLinks, &fullRangeCoinsUsed)
		}
		if err != nil {
			return nil, err
		}

		// Set the migration links in x/gamm.
		// This will also migrate the CFMM distribution records to point to the new CL pools.
		err = keepers.GAMMKeeper.UpdateMigrationRecords(ctx, poolLinks)
		if err != nil {
			return nil, err
		}

		// Because we had done direct sends from the community pool, we need to manually change the fee pool to reflect the change in balance.

		// Remove coins we used from the community pool to make the CL positions
		feePool := keepers.DistrKeeper.GetFeePool(ctx)
		newPool, negative := feePool.CommunityPool.SafeSub(sdk.NewDecCoinsFromCoins(fullRangeCoinsUsed...))
		if negative {
			return nil, fmt.Errorf("community pool cannot be negative: %s", newPool)
		}

		// Update and set the new fee pool
		feePool.CommunityPool = newPool
		keepers.DistrKeeper.SetFeePool(ctx, feePool)

		// Reset the pool weights upon upgrade. This will add support for CW pools on ProtoRev.
		keepers.ProtoRevKeeper.SetPoolWeights(ctx, types.PoolWeights{
			BalancerWeight:     balancerWeight,
			StableWeight:       stableWeight,
			ConcentratedWeight: concentratedWeight,
			CosmwasmWeight:     cosmwasmWeight,
		})

		return migrations, nil
	}
}

// mainnetUpgradeHandler creates CL pools for all balancer pools defined in the asset pairs struct. It also links the CL pools to their balancer pool counterpart, creates a full range position with the community pool,
// authorizes superfluid for the CL pool if the balancer pool is superfluid enabled, and manually sets the TWAP records for the CL pool.
func mainnetUpgradeHandler(ctx sdk.Context, keepers *keepers.AppKeepers, communityPoolAddress sdk.AccAddress, poolLinks *[]gammmigration.BalancerToConcentratedPoolLink, fullRangeCoinsUsed *sdk.Coins) error {
	assetPairs := InitializeAssetPairs(ctx, keepers)

	for _, assetPair := range assetPairs {
		clPoolDenom, clPoolId, err := createCLPoolWithCommunityPoolPosition(ctx, keepers, assetPair.LinkedClassicPool, assetPair.BaseAsset, assetPair.SpreadFactor, communityPoolAddress, poolLinks, fullRangeCoinsUsed)
		if err != nil {
			return err
		}

		err = authorizeSuperfluidIfEnabled(ctx, keepers, assetPair.LinkedClassicPool, clPoolDenom)
		if err != nil {
			return err
		}

		err = manuallySetTWAPRecords(ctx, keepers, clPoolId)
		if err != nil {
			return err
		}
	}

	return nil
}

// testnetUpgradeHandler creates CL pools for all existing balancer pools. It also links the CL pools to their balancer pool counterpart, creates a full range position with the community pool,
// authorizes superfluid for the CL pool if the balancer pool is superfluid enabled, and manually sets the TWAP records for the CL pool.
func testnetUpgradeHandler(ctx sdk.Context, keepers *keepers.AppKeepers, communityPoolAddress sdk.AccAddress, poolLinks *[]gammmigration.BalancerToConcentratedPoolLink, fullRangeCoinsUsed *sdk.Coins) error {
	// Retrieve all GAMM pools on the testnet.
	pools, err := keepers.GAMMKeeper.GetPools(ctx)
	if err != nil {
		return err
	}

	for _, pool := range pools {
		skipPool, gammPoolId, baseAsset, spreadFactor, err := testnetParsePoolRecord(ctx, pool, keepers)
		if err != nil {
			return err
		}
		if skipPool {
			continue
		}

		clPoolDenom, clPoolId, err := createCLPoolWithCommunityPoolPosition(ctx, keepers, gammPoolId, baseAsset, spreadFactor, communityPoolAddress, poolLinks, fullRangeCoinsUsed)
		if err != nil {
			return err
		}

		err = authorizeSuperfluidIfEnabled(ctx, keepers, gammPoolId, clPoolDenom)
		if err != nil {
			return err
		}

		err = manuallySetTWAPRecords(ctx, keepers, clPoolId)
		if err != nil {
			return err
		}
	}
	return nil
}

// createCLPoolWithCommunityPoolPosition creates a CL pool for a given balancer pool and adds a full range position with the community pool.
// There must be 1 OSMO worth baseAsset in the community pool for this to work.
func createCLPoolWithCommunityPoolPosition(ctx sdk.Context, keepers *keepers.AppKeepers, gammPoolId uint64, baseAsset string, spreadFactor sdk.Dec, communityPoolAddress sdk.AccAddress, poolLinks *[]gammmigration.BalancerToConcentratedPoolLink, fullRangeCoinsUsed *sdk.Coins) (clPoolDenom string, clPoolId uint64, err error) {
	// Create a concentrated liquidity pool for asset pair.
	clPool, err := keepers.GAMMKeeper.CreateConcentratedPoolFromCFMM(ctx, gammPoolId, baseAsset, spreadFactor, TickSpacing)
	if err != nil {
		return "", 0, err
	}
	clPoolId = clPool.GetId()
	clPoolDenom = cltypes.GetConcentratedLockupDenomFromPoolId(clPoolId)

	// Add the pool link to the list of pool links (we set them all at once later)
	*poolLinks = append(*poolLinks, gammmigration.BalancerToConcentratedPoolLink{
		BalancerPoolId: gammPoolId,
		ClPoolId:       clPoolId,
	})

	// Swap 0.1 OSMO for baseAsset from the community pool.
	osmoIn := sdk.NewCoin(QuoteAsset, sdk.NewInt(100000))
	linkedClassicPool, err := keepers.PoolManagerKeeper.GetPool(ctx, gammPoolId)
	if err != nil {
		return "", 0, err
	}
	respectiveBaseAssetInt, err := keepers.GAMMKeeper.SwapExactAmountIn(ctx, communityPoolAddress, linkedClassicPool, osmoIn, baseAsset, sdk.ZeroInt(), linkedClassicPool.GetSpreadFactor(ctx))
	if err != nil {
		return "", 0, err
	}

	respectiveBaseAsset := sdk.NewCoin(baseAsset, respectiveBaseAssetInt)

	// Create a full range position via the community pool with the funds we calculated above.
	fullRangeCoins := sdk.NewCoins(respectiveBaseAsset, osmoIn)
	_, _, _, _, err = keepers.ConcentratedLiquidityKeeper.CreateFullRangePosition(ctx, clPoolId, communityPoolAddress, fullRangeCoins)
	if err != nil {
		return "", 0, err
	}

	// Track the coins used to create the full range position (we manually update the fee pool later all at once).
	*fullRangeCoinsUsed = fullRangeCoinsUsed.Add(sdk.NewCoins(osmoIn)...)

	return clPoolDenom, clPoolId, nil
}

// authorizeSuperfluidIfEnabled authorizes superfluid for a CL pool if the balancer pool is superfluid enabled.
func authorizeSuperfluidIfEnabled(ctx sdk.Context, keepers *keepers.AppKeepers, gammPoolId uint64, clPoolDenom string) (err error) {
	// If pair was previously superfluid enabled, add the cl pool's full range denom as an authorized superfluid asset.
	poolShareDenom := fmt.Sprintf("gamm/pool/%d", gammPoolId)
	_, err = keepers.SuperfluidKeeper.GetSuperfluidAsset(ctx, poolShareDenom)
	if err == nil {
		superfluidAsset := superfluidtypes.SuperfluidAsset{
			Denom:     clPoolDenom,
			AssetType: superfluidtypes.SuperfluidAssetTypeConcentratedShare,
		}
		err = keepers.SuperfluidKeeper.AddNewSuperfluidAsset(ctx, superfluidAsset)
		if err != nil {
			return err
		}
	}
	return nil
}

// manuallySetTWAPRecords manually sets the TWAP records for a CL pool. This prevents a panic when the CL pool is first used.
func manuallySetTWAPRecords(ctx sdk.Context, keepers *keepers.AppKeepers, clPoolId uint64) error {
	clPoolTwapRecords, err := keepers.TwapKeeper.GetAllMostRecentRecordsForPool(ctx, clPoolId)
	if err != nil {
		return err
	}

	for _, twapRecord := range clPoolTwapRecords {
		twapRecord.LastErrorTime = time.Time{}
		keepers.TwapKeeper.StoreNewRecord(ctx, twapRecord)
	}
	return nil
}

// testnetParsePoolRecord parses a pool record and returns whether or not to skip the pool, the pool's gammPoolId, the pool's base asset, and the pool's spread factor.
func testnetParsePoolRecord(ctx sdk.Context, pool poolManagerTypes.PoolI, keepers *keepers.AppKeepers) (bool, uint64, string, sdk.Dec, error) {
	// We only want to upgrade balancer pools.
	if pool.GetType() != poolManagerTypes.Balancer {
		return true, 0, "", sdk.Dec{}, nil
	}

	gammPoolId := pool.GetId()
	cfmmPool, err := keepers.GAMMKeeper.GetCFMMPool(ctx, gammPoolId)
	if err != nil {
		return true, 0, "", sdk.Dec{}, err
	}

	poolCoins := cfmmPool.GetTotalPoolLiquidity(ctx)

	// We only want to upgrade pools paired with OSMO. OSMO will be the quote asset.
	quoteAsset, baseAsset := "", ""
	for _, coin := range poolCoins {
		if coin.Denom == QuoteAsset {
			quoteAsset = coin.Denom
		} else {
			baseAsset = coin.Denom
		}
	}
	if quoteAsset == "" || baseAsset == "" {
		return true, 0, "", sdk.Dec{}, nil
	}

	// Set the spread factor to the same spread factor the GAMM pool was.
	// If its spread factor is not authorized, set it to the first authorized non-zero spread factor.
	spreadFactor := cfmmPool.GetSpreadFactor(ctx)
	authorizedSpreadFactors := keepers.ConcentratedLiquidityKeeper.GetParams(ctx).AuthorizedSpreadFactors
	spreadFactorAuthorized := false
	for _, authorizedSpreadFactor := range authorizedSpreadFactors {
		if authorizedSpreadFactor.Equal(spreadFactor) {
			spreadFactorAuthorized = true
			break
		}
	}
	if !spreadFactorAuthorized {
		spreadFactor = authorizedSpreadFactors[1]
	}
	return false, gammPoolId, baseAsset, spreadFactor, nil
}
