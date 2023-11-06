package v21

import (
	"cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	auctiontypes "github.com/skip-mev/block-sdk/x/auction/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/osmosis-labs/osmosis/v20/app/keepers"
	"github.com/osmosis-labs/osmosis/v20/app/upgrades"
	concentratedliquiditytypes "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/types"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v20/x/cosmwasmpool/types"
	gammtypes "github.com/osmosis-labs/osmosis/v20/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v20/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v20/x/lockup/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v20/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v20/x/protorev/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v20/x/superfluid/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v20/x/tokenfactory/types"
	twaptypes "github.com/osmosis-labs/osmosis/v20/x/twap/types"

	// SDK v47 modules
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
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// UNFORKINGNOTE: If we don't manually set this to 2, the gov modules doesn't go through its necessary migrations to version 4
		fromVM[govtypes.ModuleName] = 2
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

			// wasm
			case wasmtypes.ModuleName:
				keyTable = wasmtypes.ParamKeyTable() //nolint:staticcheck

			// POB
			case auctiontypes.ModuleName:
				// already SDK v47
				continue

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
		baseapp.MigrateParams(ctx, baseAppLegacySS, &keepers.ConsensusParamsKeeper)

		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// set POB params
		err = setAuctionParams(ctx, keepers)
		if err != nil {
			return nil, err
		}

		return migrations, nil
	}
}

func setAuctionParams(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	pobAddr := keepers.AccountKeeper.GetModuleAddress(auctiontypes.ModuleName)

	auctionParams := auctiontypes.Params{
		MaxBundleSize:          2,
		EscrowAccountAddress:   pobAddr,
		ReserveFee:             sdk.Coin{Denom: "uosmo", Amount: sdk.NewInt(1_000_000)},
		MinBidIncrement:        sdk.Coin{Denom: "uosmo", Amount: sdk.NewInt(1_000_000)},
		FrontRunningProtection: true,
		ProposerFee:            math.LegacyNewDecWithPrec(25, 2),
	}
	return keepers.AuctionKeeper.SetParams(ctx, auctionParams)
}
