package v17

import (
	"errors"
	"fmt"
	"time"

	ibchookstypes "github.com/osmosis-labs/osmosis/x/ibc-hooks/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	cltypes "github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v17/x/gamm/types"
	gammmigration "github.com/osmosis-labs/osmosis/v17/x/gamm/types/migration"
	superfluidtypes "github.com/osmosis-labs/osmosis/v17/x/superfluid/types"

	"github.com/osmosis-labs/osmosis/v17/app/keepers"
	"github.com/osmosis-labs/osmosis/v17/app/upgrades"
	"github.com/osmosis-labs/osmosis/v17/x/protorev/types"
)

const (
	mainnetChainID = "osmosis-1"
	e2eChainA      = "osmo-test-a"
	e2eChainB      = "osmo-test-b"
)

var notEnoughLiquidityForSwapErr = errorsmod.Wrapf(gammtypes.ErrInvalidMathApprox, "token amount must be positive")

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		var assetPairs []AssetPair
		poolLinks := []gammmigration.BalancerToConcentratedPoolLink{}
		fullRangeCoinsUsed := sdk.Coins{}

		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Set the asset pair list depending on the chain ID.
		if ctx.ChainID() == mainnetChainID || ctx.ChainID() == e2eChainA || ctx.ChainID() == e2eChainB {
			// Upgrades specific balancer pools to concentrated liquidity pools and links them to their CL equivalent.
			ctx.Logger().Info(fmt.Sprintf("Chain ID is %s, running mainnet upgrade handler", ctx.ChainID()))
			assetPairs = InitializeAssetPairs(ctx, keepers)
		} else {
			// Upgrades all existing balancer pools to concentrated liquidity pools and links them to their CL equivalent.
			ctx.Logger().Info(fmt.Sprintf("Chain ID is %s, running testnet upgrade handler", ctx.ChainID()))
			assetPairs, err = InitializeAssetPairsTestnet(ctx, keepers)
		}
		if err != nil {
			return nil, err
		}

		communityPoolAddress := keepers.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)

		for _, assetPair := range assetPairs {
			clPoolDenom, clPoolId, err := createCLPoolWithCommunityPoolPosition(ctx, keepers, assetPair.LinkedClassicPool, assetPair.BaseAsset, assetPair.SpreadFactor, communityPoolAddress, &poolLinks, &fullRangeCoinsUsed)
			if errors.Is(err, notEnoughLiquidityForSwapErr) {
				continue
			} else if err != nil {
				return nil, err
			}

			if assetPair.Superfluid {
				ctx.Logger().Info(fmt.Sprintf("gammPoolId %d is superfluid enabled, enabling %s as a superfluid asset", assetPair.LinkedClassicPool, clPoolDenom))
				err := authorizeSuperfluid(ctx, keepers, clPoolDenom)
				if err != nil {
					return nil, err
				}
			}

			err = manuallySetTWAPRecords(ctx, keepers, clPoolId)
			if err != nil {
				return nil, err
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

		// Set ibc-hooks params
		keepers.IBCHooksKeeper.SetParams(ctx, ibchookstypes.DefaultParams())

		// Reset the pool weights upon upgrade. This will add support for CW pools on ProtoRev.
		keepers.ProtoRevKeeper.SetInfoByPoolType(ctx, types.DefaultPoolTypeInfo)

		return migrations, nil
	}
}

// createCLPoolWithCommunityPoolPosition creates a CL pool for a given balancer pool and adds a full range position with the community pool.
func createCLPoolWithCommunityPoolPosition(ctx sdk.Context, keepers *keepers.AppKeepers, gammPoolId uint64, baseAsset string, spreadFactor sdk.Dec, communityPoolAddress sdk.AccAddress, poolLinks *[]gammmigration.BalancerToConcentratedPoolLink, fullRangeCoinsUsed *sdk.Coins) (clPoolDenom string, clPoolId uint64, err error) {
	// Check if classic pool has enough liquidity to support a 0.1 OSMO swap before creating a CL pool.
	// If not, skip the pool.
	osmoIn := sdk.NewCoin(QuoteAsset, sdk.NewInt(100000))
	linkedClassicPool, err := keepers.PoolManagerKeeper.GetPool(ctx, gammPoolId)
	if err != nil {
		return "", 0, err
	}
	_, err = keepers.GAMMKeeper.CalcOutAmtGivenIn(ctx, linkedClassicPool, osmoIn, baseAsset, spreadFactor)
	if err != nil {
		return "", 0, err
	}

	// Create a concentrated liquidity pool for asset pair.
	ctx.Logger().Info(fmt.Sprintf("Creating CL pool from poolID (%d), baseAsset (%s), spreadFactor (%s), tickSpacing (%d)", gammPoolId, baseAsset, spreadFactor, TickSpacing))
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

	// Get community pool balance before swap and position creation
	commPoolBalanceBaseAssetPre := keepers.BankKeeper.GetBalance(ctx, communityPoolAddress, baseAsset)
	commPoolBalanceQuoteAssetPre := keepers.BankKeeper.GetBalance(ctx, communityPoolAddress, QuoteAsset)
	commPoolBalancePre := sdk.NewCoins(commPoolBalanceBaseAssetPre, commPoolBalanceQuoteAssetPre)

	// Swap 0.1 OSMO for baseAsset from the community pool.
	respectiveBaseAssetInt, err := keepers.GAMMKeeper.SwapExactAmountIn(ctx, communityPoolAddress, linkedClassicPool, osmoIn, baseAsset, sdk.ZeroInt(), linkedClassicPool.GetSpreadFactor(ctx))
	if err != nil {
		return "", 0, err
	}
	ctx.Logger().Info(fmt.Sprintf("Swapped %s for %s%s from the community pool", osmoIn.String(), respectiveBaseAssetInt.String(), baseAsset))

	respectiveBaseAsset := sdk.NewCoin(baseAsset, respectiveBaseAssetInt)

	// Create a full range position via the community pool with the funds we calculated above.
	fullRangeCoins := sdk.NewCoins(respectiveBaseAsset, osmoIn)
	_, _, _, _, err = keepers.ConcentratedLiquidityKeeper.CreateFullRangePosition(ctx, clPoolId, communityPoolAddress, fullRangeCoins)
	if err != nil {
		return "", 0, err
	}

	// Get community pool balance after swap and position creation
	commPoolBalanceBaseAssetPost := keepers.BankKeeper.GetBalance(ctx, communityPoolAddress, baseAsset)
	commPoolBalanceQuoteAssetPost := keepers.BankKeeper.GetBalance(ctx, communityPoolAddress, QuoteAsset)
	commPoolBalancePost := sdk.NewCoins(commPoolBalanceBaseAssetPost, commPoolBalanceQuoteAssetPost)

	// While we can be fairly certain the diff between these two is 0.2 OSMO, if for whatever reason
	// some baseAsset dust remains in the community pool and we don't account for it, when updating the
	// fee pool balance later, we will be off by that amount and will cause a panic.
	coinsUsed := commPoolBalancePre.Sub(commPoolBalancePost)

	// Track the coins used to create the full range position (we manually update the fee pool later all at once).
	*fullRangeCoinsUsed = fullRangeCoinsUsed.Add(coinsUsed...)

	return clPoolDenom, clPoolId, nil
}

// authorizeSuperfluid authorizes superfluid for the provided CL pool.
func authorizeSuperfluid(ctx sdk.Context, keepers *keepers.AppKeepers, clPoolDenom string) (err error) {
	superfluidAsset := superfluidtypes.SuperfluidAsset{
		Denom:     clPoolDenom,
		AssetType: superfluidtypes.SuperfluidAssetTypeConcentratedShare,
	}
	return keepers.SuperfluidKeeper.AddNewSuperfluidAsset(ctx, superfluidAsset)
}

// manuallySetTWAPRecords manually sets the TWAP records for a CL pool. This prevents a panic when the CL pool is first used.
func manuallySetTWAPRecords(ctx sdk.Context, keepers *keepers.AppKeepers, clPoolId uint64) error {
	ctx.Logger().Info(fmt.Sprintf("manually setting twap record for newly created CL poolID %d", clPoolId))
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
