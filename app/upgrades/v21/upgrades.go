package v21

import (
	"context"

	wasmv2 "github.com/CosmWasm/wasmd/x/wasm/migrations/v2"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v8/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v30/app/keepers"
	appparams "github.com/osmosis-labs/osmosis/v30/app/params"
	"github.com/osmosis-labs/osmosis/v30/app/upgrades"
	concentratedliquiditytypes "github.com/osmosis-labs/osmosis/v30/x/concentrated-liquidity/types"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v30/x/cosmwasmpool/types"
	gammtypes "github.com/osmosis-labs/osmosis/v30/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v30/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v30/x/lockup/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v30/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v30/x/poolmanager/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v30/x/protorev/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v30/x/superfluid/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v30/x/tokenfactory/types"
	twaptypes "github.com/osmosis-labs/osmosis/v30/x/twap/types"

	// SDK v47 modules
	upgradetypes "cosmossdk.io/x/upgrade/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		// I spent a very long time trying to figure out how to test this in a non hacky way.
		// TL;DR, on mainnet, we run a fork of v0.43, so we should be starting at version 2.
		// Without this change, since we unfork to the primary repo, we start at version 5, which
		// wouldn't allow us to run each migration.
		//
		// Now, starting from 2 only works on mainnet because the legacysubspace is set.
		// Because the legacysubspace is not set in the gotest, we can't simply run these migrations without setting the legacysubspace.
		// This legacysubspace can only be set at the initChain level, so it isn't clear to me how to directly set this in the test.
		if ctx.ChainID() != TestingChainId {
			fromVM[govtypes.ModuleName] = 2
		}
		baseAppLegacySS := keepers.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())

		// https://github.com/cosmos/cosmos-sdk/pull/12363/files
		// Set param key table for params module migration
		for _, subspace := range keepers.ParamsKeeper.GetSubspaces() {
			subspace := subspace

			var keyTable paramstypes.KeyTable
			switch subspace.Name() {
			// sdk
			case authtypes.ModuleName:
				keyTable = authtypes.ParamKeyTable() //nolint:staticcheck
			case banktypes.ModuleName:
				keyTable = banktypes.ParamKeyTable() //nolint:staticcheck
			case stakingtypes.ModuleName:
				keyTable = stakingtypes.ParamKeyTable() //nolint:staticcheck
			case minttypes.ModuleName:
				keyTable = minttypes.ParamKeyTable() //nolint:staticcheck
			case distrtypes.ModuleName:
				keyTable = distrtypes.ParamKeyTable() //nolint:staticcheck
			case slashingtypes.ModuleName:
				keyTable = slashingtypes.ParamKeyTable() //nolint:staticcheck
			case govtypes.ModuleName:
				keyTable = govv1.ParamKeyTable() //nolint:staticcheck
			case crisistypes.ModuleName:
				keyTable = crisistypes.ParamKeyTable() //nolint:staticcheck

			// ibc types
			case ibctransfertypes.ModuleName:
				keyTable = ibctransfertypes.ParamKeyTable() //nolint:staticcheck
			case icahosttypes.SubModuleName:
				keyTable = icahosttypes.ParamKeyTable() //nolint:staticcheck
			case icacontrollertypes.SubModuleName:
				keyTable = icacontrollertypes.ParamKeyTable() //nolint:staticcheck
			case icqtypes.ModuleName:
				keyTable = icqtypes.ParamKeyTable() //nolint:staticcheck

			// wasm
			case wasmtypes.ModuleName:
				keyTable = wasmv2.ParamKeyTable() //nolint:staticcheck

			// osmosis modules
			case protorevtypes.ModuleName:
				keyTable = protorevtypes.ParamKeyTable() //nolint:staticcheck
			case superfluidtypes.ModuleName:
				keyTable = superfluidtypes.ParamKeyTable() //nolint:staticcheck
			// downtime doesn't have params
			case gammtypes.ModuleName:
				keyTable = gammtypes.ParamKeyTable() //nolint:staticcheck
			case twaptypes.ModuleName:
				keyTable = twaptypes.ParamKeyTable() //nolint:staticcheck
			case lockuptypes.ModuleName:
				keyTable = lockuptypes.ParamKeyTable() //nolint:staticcheck
			// epochs doesn't have params (it really should imo)
			case incentivestypes.ModuleName:
				keyTable = incentivestypes.ParamKeyTable() //nolint:staticcheck
			case poolincentivestypes.ModuleName:
				keyTable = poolincentivestypes.ParamKeyTable() //nolint:staticcheck
			// txfees doesn't have params
			case tokenfactorytypes.ModuleName:
				keyTable = tokenfactorytypes.ParamKeyTable() //nolint:staticcheck
			case poolmanagertypes.ModuleName:
				keyTable = poolmanagertypes.ParamKeyTable() //nolint:staticcheck
			// valsetpref doesn't have params
			case concentratedliquiditytypes.ModuleName:
				keyTable = concentratedliquiditytypes.ParamKeyTable() //nolint:staticcheck
			case cosmwasmpooltypes.ModuleName:
				keyTable = cosmwasmpooltypes.ParamKeyTable() //nolint:staticcheck

			default:
				continue
			}

			if !subspace.HasKeyTable() {
				subspace.WithKeyTable(keyTable)
			}
		}

		// Migrate Tendermint consensus parameters from x/params module to a deprecated x/consensus module.
		// The old params module is required to still be imported in your app.go in order to handle this migration.
		err := baseapp.MigrateParams(ctx, baseAppLegacySS, keepers.ConsensusParamsKeeper.ParamsStore)
		if err != nil {
			return nil, err
		}

		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Set expedited proposal param:
		govParams, err := keepers.GovKeeper.Params.Get(ctx)
		if err != nil {
			return nil, err
		}
		govParams.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(5000000000)))
		govParams.MinInitialDepositRatio = "0.250000000000000000"
		err = keepers.GovKeeper.Params.Set(ctx, govParams)
		if err != nil {
			return nil, err
		}

		// Set CL param:
		keepers.ConcentratedLiquidityKeeper.SetParam(ctx, concentratedliquiditytypes.KeyHookGasLimit, concentratedliquiditytypes.DefaultContractHookGasLimit)

		// Add protorev to the taker fee exclusion list:
		protorevModuleAccount := keepers.AccountKeeper.GetModuleAccount(ctx, protorevtypes.ModuleName)
		poolManagerParams := keepers.PoolManagerKeeper.GetParams(ctx)
		poolManagerParams.TakerFeeParams.ReducedFeeWhitelist = append(poolManagerParams.TakerFeeParams.ReducedFeeWhitelist, protorevModuleAccount.GetAddress().String())
		keepers.PoolManagerKeeper.SetParams(ctx, poolManagerParams)

		// Since we are now tracking all protocol rev, we set the accounting height to the current block height for each module
		// that generates protocol rev.
		keepers.PoolManagerKeeper.SetTakerFeeTrackerStartHeight(ctx, ctx.BlockHeight())
		// keepers.TxFeesKeeper.SetTxFeesTrackerStartHeight(ctx, ctx.BlockHeight())
		// We start the cyclic arb tracker from the value it currently is at since it has been tracking since inception (without a start height).
		// This will allow us to display the amount of cyclic arb profits that have been generated from a certain block height.
		allCyclicArbProfits := keepers.ProtoRevKeeper.GetAllProfits(ctx)
		allCyclicArbProfitsCoins := osmoutils.ConvertCoinArrayToCoins(allCyclicArbProfits)
		keepers.ProtoRevKeeper.SetCyclicArbProfitTrackerValue(ctx, allCyclicArbProfitsCoins)
		keepers.ProtoRevKeeper.SetCyclicArbProfitTrackerStartHeight(ctx, ctx.BlockHeight())

		return migrations, nil
	}
}
