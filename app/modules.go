package app

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v3/modules/core"
	ibchost "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	ica "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	"github.com/osmosis-labs/bech32-ibc/x/bech32ibc"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"
	"github.com/osmosis-labs/bech32-ibc/x/bech32ics20"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	appparams "github.com/osmosis-labs/osmosis/v10/app/params"
	_ "github.com/osmosis-labs/osmosis/v10/client/docs/statik"
	"github.com/osmosis-labs/osmosis/v10/osmoutils/partialord"
	"github.com/osmosis-labs/osmosis/v10/x/epochs"
	epochstypes "github.com/osmosis-labs/osmosis/v10/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v10/x/gamm"
	gammtypes "github.com/osmosis-labs/osmosis/v10/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v10/x/incentives"
	incentivestypes "github.com/osmosis-labs/osmosis/v10/x/incentives/types"
	"github.com/osmosis-labs/osmosis/v10/x/lockup"
	lockuptypes "github.com/osmosis-labs/osmosis/v10/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v10/x/mint"
	minttypes "github.com/osmosis-labs/osmosis/v10/x/mint/types"
	poolincentives "github.com/osmosis-labs/osmosis/v10/x/pool-incentives"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v10/x/pool-incentives/types"
	superfluid "github.com/osmosis-labs/osmosis/v10/x/superfluid"
	superfluidtypes "github.com/osmosis-labs/osmosis/v10/x/superfluid/types"
	"github.com/osmosis-labs/osmosis/v10/x/tokenfactory"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v10/x/tokenfactory/types"
	"github.com/osmosis-labs/osmosis/v10/x/txfees"
	txfeestypes "github.com/osmosis-labs/osmosis/v10/x/txfees/types"
)

// moduleAccountPermissions defines module account permissions
// TODO: Having to input nil's here is unacceptable, we need a way to automatically derive this.
var moduleAccountPermissions = map[string][]string{
	authtypes.FeeCollectorName:               nil,
	distrtypes.ModuleName:                    nil,
	icatypes.ModuleName:                      nil,
	minttypes.ModuleName:                     {authtypes.Minter, authtypes.Burner},
	minttypes.DeveloperVestingModuleAcctName: nil,
	stakingtypes.BondedPoolName:              {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName:           {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:                      {authtypes.Burner},
	ibctransfertypes.ModuleName:              {authtypes.Minter, authtypes.Burner},
	gammtypes.ModuleName:                     {authtypes.Minter, authtypes.Burner},
	incentivestypes.ModuleName:               {authtypes.Minter, authtypes.Burner},
	lockuptypes.ModuleName:                   {authtypes.Minter, authtypes.Burner},
	poolincentivestypes.ModuleName:           nil,
	superfluidtypes.ModuleName:               {authtypes.Minter, authtypes.Burner},
	txfeestypes.ModuleName:                   nil,
	txfeestypes.NonNativeFeeCollectorName:    nil,
	wasm.ModuleName:                          {authtypes.Burner},
	tokenfactorytypes.ModuleName:             {authtypes.Minter, authtypes.Burner},
}

// appModules return modules to initialize module manager.
func appModules(
	app *OsmosisApp,
	encodingConfig appparams.EncodingConfig,
	skipGenesisInvariants bool,
) []module.AppModule {
	appCodec := encodingConfig.Marshaler

	return []module.AppModule{
		genutil.NewAppModule(
			app.AccountKeeper,
			app.StakingKeeper,
			app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, *app.AccountKeeper, nil),
		vesting.NewAppModule(*app.AccountKeeper, app.BankKeeper),
		bech32ics20.NewAppModule(appCodec, *app.Bech32ICS20Keeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants),
		gov.NewAppModule(appCodec, *app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, *app.MintKeeper, app.AccountKeeper, app.BankKeeper),
		slashing.NewAppModule(appCodec, *app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper),
		distr.NewAppModule(appCodec, *app.DistrKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper),
		staking.NewAppModule(appCodec, *app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		upgrade.NewAppModule(*app.UpgradeKeeper),
		wasm.NewAppModule(appCodec, app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		evidence.NewAppModule(*app.EvidenceKeeper),
		authzmodule.NewAppModule(appCodec, *app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		ibc.NewAppModule(app.IBCKeeper),
		ica.NewAppModule(nil, app.ICAHostKeeper),
		params.NewAppModule(*app.ParamsKeeper),
		app.TransferModule,
		gamm.NewAppModule(appCodec, *app.GAMMKeeper, app.AccountKeeper, app.BankKeeper),
		txfees.NewAppModule(appCodec, *app.TxFeesKeeper),
		incentives.NewAppModule(appCodec, *app.IncentivesKeeper, app.AccountKeeper, app.BankKeeper, app.EpochsKeeper),
		lockup.NewAppModule(appCodec, *app.LockupKeeper, app.AccountKeeper, app.BankKeeper),
		poolincentives.NewAppModule(appCodec, *app.PoolIncentivesKeeper),
		epochs.NewAppModule(appCodec, *app.EpochsKeeper),
		superfluid.NewAppModule(
			appCodec,
			*app.SuperfluidKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.StakingKeeper,
			app.LockupKeeper,
			app.GAMMKeeper,
			app.EpochsKeeper,
		),
		tokenfactory.NewAppModule(appCodec, *app.TokenFactoryKeeper, app.AccountKeeper, app.BankKeeper),
		bech32ibc.NewAppModule(appCodec, *app.Bech32IBCKeeper),
	}
}

// orderBeginBlockers returns the order of BeginBlockers, by module name.
func orderBeginBlockers(allModuleNames []string) []string {
	ord := partialord.NewPartialOrdering(allModuleNames)
	// Upgrades should be run VERY first
	// Epochs is set to be next right now, this in principle could change to come later / be at the end.
	// But would have to be a holistic change with other pipelines taken into account.
	ord.FirstElements(upgradetypes.ModuleName, epochstypes.ModuleName, capabilitytypes.ModuleName)

	// Staking ordering
	// TODO: Perhaps this can be relaxed, left to future work to analyze.
	ord.Sequence(distrtypes.ModuleName, slashingtypes.ModuleName, evidencetypes.ModuleName, stakingtypes.ModuleName)
	// superfluid must come after distribution & epochs.
	// TODO: we actually set it to come after staking, since thats what happened before, and want to minimize chance of break.
	ord.After(superfluidtypes.ModuleName, stakingtypes.ModuleName)
	// TODO: This can almost certainly be un-constrained, but we keep the constraint to match prior functionality.
	// IBChost came after staking, before superfluid.
	// TODO: Come back and delete this line after testing the base change.
	ord.Sequence(stakingtypes.ModuleName, ibchost.ModuleName, superfluidtypes.ModuleName)
	// every remaining module's begin block is a no-op.
	return ord.TotalOrdering()
}

func OrderEndBlockers(allModuleNames []string) []string {
	ord := partialord.NewPartialOrdering(allModuleNames)
	// Epochs must run after all other end blocks
	ord.LastElements(epochstypes.ModuleName)
	// txfees auto-swap code should occur before any potential gamm end block code.
	ord.Before(txfeestypes.ModuleName, gammtypes.ModuleName)
	// only remaining modules that aren;t no-ops are: crisis & govtypes
	// we don't care about the relative ordering between them.

	return ord.TotalOrdering()
}

// OrderInitGenesis returns module names in order for init genesis calls.
func OrderInitGenesis(allModuleNames []string) []string {
	return []string{
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		ibchost.ModuleName,
		icatypes.ModuleName,
		gammtypes.ModuleName,
		txfeestypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		ibctransfertypes.ModuleName,
		bech32ibctypes.ModuleName, // comes after ibctransfertypes
		poolincentivestypes.ModuleName,
		superfluidtypes.ModuleName,
		tokenfactorytypes.ModuleName,
		incentivestypes.ModuleName,
		epochstypes.ModuleName,
		lockuptypes.ModuleName,
		authz.ModuleName,
		// wasm after ibc transfer
		wasm.ModuleName,
	}
}

// simulationModules returns modules for simulation manager
func simulationModules(
	app *OsmosisApp,
	encodingConfig appparams.EncodingConfig,
	skipGenesisInvariants bool,
) []module.AppModuleSimulation {
	appCodec := encodingConfig.Marshaler

	return []module.AppModuleSimulation{
		auth.NewAppModule(appCodec, *app.AccountKeeper, authsims.RandomGenesisAccounts),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		authzmodule.NewAppModule(appCodec, *app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		gamm.NewAppModule(appCodec, *app.GAMMKeeper, app.AccountKeeper, app.BankKeeper),
		txfees.NewAppModule(appCodec, *app.TxFeesKeeper),
		gov.NewAppModule(appCodec, *app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, *app.MintKeeper, app.AccountKeeper, app.BankKeeper),
		slashing.NewAppModule(appCodec, *app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper),
		distr.NewAppModule(appCodec, *app.DistrKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper),
		staking.NewAppModule(appCodec, *app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		params.NewAppModule(*app.ParamsKeeper),
		evidence.NewAppModule(*app.EvidenceKeeper),
		wasm.NewAppModule(appCodec, app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		incentives.NewAppModule(appCodec, *app.IncentivesKeeper, app.AccountKeeper, app.BankKeeper, app.EpochsKeeper),
		lockup.NewAppModule(appCodec, *app.LockupKeeper, app.AccountKeeper, app.BankKeeper),
		poolincentives.NewAppModule(appCodec, *app.PoolIncentivesKeeper),
		epochs.NewAppModule(appCodec, *app.EpochsKeeper),
		superfluid.NewAppModule(
			appCodec,
			*app.SuperfluidKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.StakingKeeper,
			app.LockupKeeper,
			app.GAMMKeeper,
			app.EpochsKeeper,
		),
		tokenfactory.NewAppModule(appCodec, *app.TokenFactoryKeeper, app.AccountKeeper, app.BankKeeper),
		app.TransferModule,
	}
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *OsmosisApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}
