package v26

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v25/app/keepers"
	"github.com/osmosis-labs/osmosis/v25/app/upgrades"
)

const (
	minDepositRatio = "0.010000000000000000"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Set minDepositRatio to 1%
		newGovParams, err := keepers.GovKeeper.Params.Get(ctx)
		if err != nil {
			return nil, err
		}
		newGovParams.MinDepositRatio = minDepositRatio

		err = keepers.GovKeeper.Params.Set(ctx, newGovParams)
		if err != nil {
			return nil, err
		}

		// We now want to treat denom pair taker fees as uni-directional. In other words, if the pair is set as `denom0: uosmo` and `denom1: usdc`, then the taker fee override is only
		// applied if uosmo is the token in and usdc is the token out. In this example, if usdc is the token in and uosmo is the token out, then the default taker fee is applied.
		// We therefore iterate over all existing taker fee overrides and create the opposite pair, treating all the existing overrides as bi-directional.
		allTradingPairTakerFees, err := keepers.PoolManagerKeeper.GetAllTradingPairTakerFees(ctx)
		if err != nil {
			return nil, err
		}
		for _, tradingPairTakerFee := range allTradingPairTakerFees {
			// Create the opposite pair. This is why TokenOutDenom is in the TokenInDenom position and vice versa.
			keepers.PoolManagerKeeper.SetDenomPairTakerFee(ctx, tradingPairTakerFee.TokenOutDenom, tradingPairTakerFee.TokenInDenom, tradingPairTakerFee.TakerFee)
		}

		return migrations, nil
	}
}
