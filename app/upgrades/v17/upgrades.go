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
	superfluidtypes "github.com/osmosis-labs/osmosis/v17/x/superfluid/types"

	"github.com/osmosis-labs/osmosis/v17/app/keepers"
	"github.com/osmosis-labs/osmosis/v17/app/upgrades"
	"github.com/osmosis-labs/osmosis/v17/x/protorev/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Get community pool address.
		communityPoolAddress := keepers.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)

		// fullRangeCoinsUsed tracks the coins we use in the below for loop from the community pool to create the full range position for each new pool.
		fullRangeCoinsUsed := sdk.NewCoins()

		poolLinks := []gammmigration.BalancerToConcentratedPoolLink{}

		assetPairs := InitializeAssetPairs(ctx, keepers)

		for _, assetPair := range assetPairs {
			// Create a concentrated liquidity pool for asset pair.
			clPool, err := keepers.GAMMKeeper.CreateConcentratedPoolFromCFMM(ctx, assetPair.LinkedClassicPool, assetPair.BaseAsset, assetPair.SpreadFactor, TickSpacing)
			if err != nil {
				return nil, err
			}
			clPoolId := clPool.GetId()
			clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(clPoolId)

			// Add the pool link to the list of pool links (we set them all at once later)
			poolLinks = append(poolLinks, gammmigration.BalancerToConcentratedPoolLink{
				BalancerPoolId: assetPair.LinkedClassicPool,
				ClPoolId:       clPoolId,
			})

			// Determine the amount of baseAsset that can be bought with 1 OSMO.
			oneOsmo := sdk.NewCoin(QuoteAsset, sdk.NewInt(1000000))
			linkedClassicPool, err := keepers.PoolManagerKeeper.GetPool(ctx, assetPair.LinkedClassicPool)
			if err != nil {
				return nil, err
			}
			respectiveBaseAsset, err := keepers.GAMMKeeper.CalcOutAmtGivenIn(ctx, linkedClassicPool, oneOsmo, assetPair.BaseAsset, sdk.ZeroDec())
			if err != nil {
				return nil, err
			}

			// Create a full range position via the community pool with the funds we calculated above.
			fullRangeCoins := sdk.NewCoins(respectiveBaseAsset, oneOsmo)
			_, actualBaseAmtUsed, actualQuoteAmtUsed, _, err := keepers.ConcentratedLiquidityKeeper.CreateFullRangePosition(ctx, clPoolId, communityPoolAddress, fullRangeCoins)
			if err != nil {
				return nil, err
			}

			// Track the coins used to create the full range position (we manually update the fee pool later all at once).
			fullRangeCoinsUsed = fullRangeCoinsUsed.Add(sdk.NewCoins(sdk.NewCoin(QuoteAsset, actualQuoteAmtUsed), sdk.NewCoin(assetPair.BaseAsset, actualBaseAmtUsed))...)

			// If pair was previously superfluid enabled, add the cl pool's full range denom as an authorized superfluid asset.
			if assetPair.Superfluid {
				superfluidAsset := superfluidtypes.SuperfluidAsset{
					Denom:     clPoolDenom,
					AssetType: superfluidtypes.SuperfluidAssetTypeConcentratedShare,
				}
				err = keepers.SuperfluidKeeper.AddNewSuperfluidAsset(ctx, superfluidAsset)
				if err != nil {
					return nil, err
				}
			}

			clPoolTwapRecords, err := keepers.TwapKeeper.GetAllMostRecentRecordsForPool(ctx, clPoolId)
			if err != nil {
				return nil, err
			}

			for _, twapRecord := range clPoolTwapRecords {
				twapRecord.LastErrorTime = time.Time{}
				keepers.TwapKeeper.StoreNewRecord(ctx, twapRecord)
			}
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
		keepers.ProtoRevKeeper.SetInfoByPoolType(ctx, types.DefaultPoolTypeInfo)

		return migrations, nil
	}
}
