package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v15/app/keepers"
	"github.com/osmosis-labs/osmosis/v15/app/upgrades"
)

const (
	// DAI/OSMO pool ID
	// https://app.osmosis.zone/pool/674
	// Note, new concentrated liquidity pool
	// swap fee is initialized to be the same as the balancers pool swap fee of 0.2%.
	daiOsmoPoolId = uint64(674)
	// Denom0 translates to a base asset while denom1 to a quote asset
	// We want quote asset to be DAI so that when the limit orders on ticks
	// are implemented, we have tick spacing in terms of DAI as the quote.
	desiredDenom0 = "uosmo"
	// TODO: confirm pre-launch.
	tickSpacing = 1
)

var (
	// TODO: confirm pre-launch.
	exponentAtPriceOne = sdk.OneInt().Neg()

	DAIIBCDenom  = "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7"
	USDCIBCDenom = "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858"

	// authorized_quote_denoms quote assets that can be used as token1
	// when creating a pool. We limit the quote assets to a small set
	// for the purposes of having convinient price increments stemming
	// from tick to price conversion. These increments are in a human
	// understandeable magnitude only for token1 as a quote.
	authorizedQuoteDenoms []string = []string{
		"uosmo",
		DAIIBCDenom,
		USDCIBCDenom,
	}
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

		// Although parameters are set on InitGenesis() in RunMigrations(), we reset them here
		// for visibility of the final configuration.
		defaultConcentratedLiquidityParams := keepers.ConcentratedLiquidityKeeper.GetParams(ctx)
		defaultConcentratedLiquidityParams.AuthorizedQuoteDenoms = authorizedQuoteDenoms
		keepers.ConcentratedLiquidityKeeper.SetParams(ctx, defaultConcentratedLiquidityParams)

		if err := createCanonicalConcentratedLiquidityPoolAndMigrationLink(ctx, daiOsmoPoolId, desiredDenom0, keepers); err != nil {
			return nil, err
		}

		return migrations, nil
	}
}
