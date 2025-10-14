package v30

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	auctiontypes "github.com/skip-mev/block-sdk/v2/x/auction/types"

	"github.com/osmosis-labs/osmosis/v31/app/keepers"
	"github.com/osmosis-labs/osmosis/v31/app/upgrades"
	poolmanager "github.com/osmosis-labs/osmosis/v31/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v31/x/poolmanager/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)

		setAuthorizedQuoteDenomsAsCommunityPoolDenomWhitelist(sdkCtx, keepers.PoolManagerKeeper)

		err = transferTopOfBlockAuctionFundsToCommunityPool(sdkCtx, keepers.AccountKeeper, keepers.BankKeeper, keepers.DistrKeeper)
		if err != nil {
			return nil, err
		}

		return migrations, nil
	}
}

// community_pool_denom_whitelist is a newly introduced parameter in poolmanager to decoupled
// the community pool denom whitelist from the authorized quote denoms.
// This migration copies the authorized quote denoms to the community pool denom whitelist to keep
// the existing behavior the same.
func setAuthorizedQuoteDenomsAsCommunityPoolDenomWhitelist(ctx sdk.Context, poolManagerKeeper *poolmanager.Keeper) {
	// Get the authorized quote denoms directly from the parameter store
	// to avoid unmarshaling issues with the new TakerFeeParams field
	var authorizedQuoteDenoms []string
	poolManagerKeeper.GetParam(ctx, types.KeyAuthorizedQuoteDenoms, &authorizedQuoteDenoms)

	// Set the community pool denom whitelist to the same value
	poolManagerKeeper.SetParam(ctx, types.KeyCommunityPoolDenomWhitelist, authorizedQuoteDenoms)
}

// see: https://daodao.zone/dao/osmosis/proposals/958
func transferTopOfBlockAuctionFundsToCommunityPool(ctx sdk.Context, accountKeeper *authkeeper.AccountKeeper, bankKeeper *bankkeeper.BaseKeeper, distrKeeper *distrkeeper.Keeper) error {
	// https://www.mintscan.io/osmosis/address/osmo1j4yzhgjm00ch3h0p9kel7g8sp6g045qfnc9kmc
	auctionModuleAccountAddr := accountKeeper.GetModuleAccount(ctx, auctiontypes.ModuleName).GetAddress()
	usdcDenom := "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"

	totalUsdcBalance := bankKeeper.GetBalance(ctx, auctionModuleAccountAddr, usdcDenom)
	return distrKeeper.FundCommunityPool(
		ctx,
		sdk.NewCoins(totalUsdcBalance),
		auctionModuleAccountAddr,
	)
}
