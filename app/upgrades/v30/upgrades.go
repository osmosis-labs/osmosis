package v30

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/app/keepers"
	"github.com/osmosis-labs/osmosis/v30/app/upgrades"
	poolmanager "github.com/osmosis-labs/osmosis/v30/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v30/x/poolmanager/types"
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
