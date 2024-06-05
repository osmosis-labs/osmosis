package v26

import (
	"context"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v25/app/keepers"
	"github.com/osmosis-labs/osmosis/v25/app/upgrades"

	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/osmosis-labs/osmosis/osmomath"
)

const (
	mainnetChainID = "osmosis-1"
	// Edgenet is to function exactly the same as mainnet, and expected
	// to be state-exported from mainnet state.
	edgenetChainID = "edgenet"
	// Testnet will have its own state. Contrary to mainnet, we would
	// like to migrate all testnet pools at once.
	testnetChainID = "osmo-test-5"
	// E2E chain IDs which we expect to migrate all pools similar to testnet.
	e2eChainIDA = "osmo-test-a"
	e2eChainIDB = "osmo-test-b"
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

		// UNFORKING v2 TODO: I think there is just one new gov param that is not registered, which is why this is needed. Need to figure out what it is rather than re-setting all params.
		// Set all gov params explicitly. E2E had issues when this was not done, so setting this here to ensure no issues on mainnet.
		var newGovParams govv1.Params
		if ctx.ChainID() == mainnetChainID || ctx.ChainID() == edgenetChainID {
			newGovParams = govv1.NewParams(sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(1600000000))), sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(5000000000))), time.Second*1209600, time.Second*432000, time.Second*86400,
				"0.200000000000000000", "0.500000000000000000", "0.667000000000000000", "0.334000000000000000", "0.250000000000000000", "0.500000000000000000", "", false, false, true, "0.010000000000000000")
		} else if ctx.ChainID() == testnetChainID {
			newGovParams = govv1.NewParams(sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(1600000000))), sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(5000000000))), time.Second*1209600, time.Second*432000, time.Second*86400,
				"0.200000000000000000", "0.500000000000000000", "0.667000000000000000", "0.334000000000000000", "0.250000000000000000", "0.500000000000000000", "", false, false, true, "0.010000000000000000")
		} else if ctx.ChainID() == e2eChainIDA || ctx.ChainID() == e2eChainIDB {
			newGovParams = govv1.NewParams(sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(10000000))), sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.NewInt(50000000))), time.Second*1209600, time.Second*12, time.Second*11,
				"0.200000000000000000", "0.500000000000000000", "0.667000000000000000", "0.334000000000000000", "0.250000000000000000", "0.500000000000000000", "", false, false, true, "0.010000000000000000")
		}
		err = keepers.GovKeeper.Params.Set(ctx, newGovParams)
		if err != nil {
			return nil, err
		}

		// We now want to treat denom pair taker fees as uni-directional. In other words, if the pair is set as `denom0: uosmo` and `denom1: usdc``, then the taker fee override is only
		// applied if uosmo is the token in and usdc is the token out. In this example, if usdc is the token in and uosmo is the token out, then the default taker fee is applied.
		// We therefore iterate over all existing taker fee overrides and create the opposite pair, treating all the existing overrides as bi-directional.
		allTradingPairTakerFees, err := keepers.PoolManagerKeeper.GetAllTradingPairTakerFees(ctx)
		if err != nil {
			return nil, err
		}
		for _, tradingPairTakerFee := range allTradingPairTakerFees {
			// Create the opposite pair. This is why DenomOfTokenOut is in the DenomOfTokenIn position and vice versa.
			keepers.PoolManagerKeeper.SetDenomPairTakerFee(ctx, tradingPairTakerFee.DenomOfTokenOut, tradingPairTakerFee.DenomOfTokenIn, tradingPairTakerFee.TakerFee)
		}

		return migrations, nil
	}
}
