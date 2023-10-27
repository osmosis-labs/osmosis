package v21

import (
	"fmt"

	"cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	buildertypes "github.com/skip-mev/pob/x/builder/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/osmosis-labs/osmosis/v20/app/keepers"
	"github.com/osmosis-labs/osmosis/v20/app/upgrades"
	protorevtypes "github.com/osmosis-labs/osmosis/v20/x/protorev/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v20/x/superfluid/types"

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
		// UNFORKINGNOTE: Add the remaining modules to this
		fromVM[govtypes.ModuleName] = 1
		baseAppLegacySS := keepers.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
		// https://github.com/cosmos/cosmos-sdk/pull/12363/files
		// Set param key table for params module migration
		for _, subspace := range keepers.ParamsKeeper.GetSubspaces() {
			subspace := subspace
			fmt.Printf("subspace: %+v\n", subspace.Name())

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
			case buildertypes.ModuleName:
				// already SDK v47
				continue

			// osmosis modules
			case protorevtypes.ModuleName:
				keyTable = protorevtypes.ParamKeyTable() //nolint:staticcheck
			case superfluidtypes.ModuleName:
				keyTable = superfluidtypes.ParamKeyTable() //nolint:staticcheck

			default:
				continue
			}

			// 			// "Normal" keepers
			// AccountKeeper                *authkeeper.AccountKeeper
			// BankKeeper                   bankkeeper.BaseKeeper
			// AuthzKeeper                  *authzkeeper.Keeper
			// StakingKeeper                *stakingkeeper.Keeper
			// DistrKeeper                  *distrkeeper.Keeper
			// DowntimeKeeper               *downtimedetector.Keeper
			// SlashingKeeper               *slashingkeeper.Keeper
			// IBCKeeper                    *ibckeeper.Keeper
			// IBCHooksKeeper               *ibchookskeeper.Keeper
			// ICAHostKeeper                *icahostkeeper.Keeper
			// ICQKeeper                    *icqkeeper.Keeper
			// TransferKeeper               *ibctransferkeeper.Keeper
			// EvidenceKeeper               *evidencekeeper.Keeper
			// GAMMKeeper                   *gammkeeper.Keeper
			// TwapKeeper                   *twap.Keeper
			// LockupKeeper                 *lockupkeeper.Keeper
			// EpochsKeeper                 *epochskeeper.Keeper
			// IncentivesKeeper             *incentiveskeeper.Keeper
			// ProtoRevKeeper               *protorevkeeper.Keeper
			// MintKeeper                   *mintkeeper.Keeper
			// PoolIncentivesKeeper         *poolincentiveskeeper.Keeper
			// TxFeesKeeper                 *txfeeskeeper.Keeper
			// SuperfluidKeeper             *superfluidkeeper.Keeper
			// GovKeeper                    *govkeeper.Keeper
			// WasmKeeper                   *wasm.Keeper
			// ContractKeeper               *wasmkeeper.PermissionedKeeper
			// TokenFactoryKeeper           *tokenfactorykeeper.Keeper
			// PoolManagerKeeper            *poolmanager.Keeper
			// ValidatorSetPreferenceKeeper *valsetpref.Keeper
			// ConcentratedLiquidityKeeper  *concentratedliquidity.Keeper
			// CosmwasmPoolKeeper           *cosmwasmpool.Keeper

			if !subspace.HasKeyTable() {
				subspace.WithKeyTable(keyTable)
			}
		}

		// Migrate Tendermint consensus parameters from x/params module to a deprecated x/consensus module.
		// The old params module is required to still be imported in your app.go in order to handle this migration.
		baseapp.MigrateParams(ctx, baseAppLegacySS, &keepers.ConsensusParamsKeeper)

		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// x/POB
		pobAddr := keepers.AccountKeeper.GetModuleAddress(buildertypes.ModuleName)

		builderParams := buildertypes.DefaultGenesisState().GetParams()
		builderParams.EscrowAccountAddress = pobAddr
		builderParams.MaxBundleSize = 4
		builderParams.FrontRunningProtection = false
		builderParams.MinBidIncrement.Denom = keepers.StakingKeeper.BondDenom(ctx)
		builderParams.MinBidIncrement.Amount = math.NewInt(1000000)
		builderParams.ReserveFee.Denom = keepers.StakingKeeper.BondDenom(ctx)
		builderParams.ReserveFee.Amount = math.NewInt(1000000)
		if err := keepers.BuildKeeper.SetParams(ctx, builderParams); err != nil {
			return nil, err
		}

		return migrations, nil
	}
}
