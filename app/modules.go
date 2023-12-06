package app

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	icq "github.com/cosmos/ibc-apps/modules/async-icq/v7"

	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibchost "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ibctestingtypes "github.com/cosmos/ibc-go/v7/testing/types"

	packetforward "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"

	ibchookstypes "github.com/osmosis-labs/osmosis/x/ibc-hooks/types"

	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v7/types"

	downtimemodule "github.com/osmosis-labs/osmosis/v21/x/downtime-detector/module"
	downtimetypes "github.com/osmosis-labs/osmosis/v21/x/downtime-detector/types"

	ibc_hooks "github.com/osmosis-labs/osmosis/x/ibc-hooks"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
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

	"github.com/osmosis-labs/osmosis/osmoutils/partialord"
	appparams "github.com/osmosis-labs/osmosis/v21/app/params"
	_ "github.com/osmosis-labs/osmosis/v21/client/docs/statik"
	"github.com/osmosis-labs/osmosis/v21/simulation/simtypes"
	concentratedliquidity "github.com/osmosis-labs/osmosis/v21/x/concentrated-liquidity/clmodule"
	concentratedliquiditytypes "github.com/osmosis-labs/osmosis/v21/x/concentrated-liquidity/types"
	cwpoolmodule "github.com/osmosis-labs/osmosis/v21/x/cosmwasmpool/module"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v21/x/cosmwasmpool/types"
	"github.com/osmosis-labs/osmosis/v21/x/gamm"
	gammtypes "github.com/osmosis-labs/osmosis/v21/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v21/x/ibc-rate-limit/ibcratelimitmodule"
	ibcratelimittypes "github.com/osmosis-labs/osmosis/v21/x/ibc-rate-limit/types"
	"github.com/osmosis-labs/osmosis/v21/x/incentives"
	incentivestypes "github.com/osmosis-labs/osmosis/v21/x/incentives/types"
	"github.com/osmosis-labs/osmosis/v21/x/lockup"
	lockuptypes "github.com/osmosis-labs/osmosis/v21/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v21/x/mint"
	minttypes "github.com/osmosis-labs/osmosis/v21/x/mint/types"
	poolincentives "github.com/osmosis-labs/osmosis/v21/x/pool-incentives"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v21/x/pool-incentives/types"
	poolmanager "github.com/osmosis-labs/osmosis/v21/x/poolmanager/module"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v21/x/protorev"
	protorevtypes "github.com/osmosis-labs/osmosis/v21/x/protorev/types"
	superfluid "github.com/osmosis-labs/osmosis/v21/x/superfluid"
	superfluidtypes "github.com/osmosis-labs/osmosis/v21/x/superfluid/types"
	"github.com/osmosis-labs/osmosis/v21/x/tokenfactory"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v21/x/tokenfactory/types"
	"github.com/osmosis-labs/osmosis/v21/x/twap/twapmodule"
	twaptypes "github.com/osmosis-labs/osmosis/v21/x/twap/types"
	"github.com/osmosis-labs/osmosis/v21/x/txfees"
	txfeestypes "github.com/osmosis-labs/osmosis/v21/x/txfees/types"
	valsetpreftypes "github.com/osmosis-labs/osmosis/v21/x/valset-pref/types"
	valsetprefmodule "github.com/osmosis-labs/osmosis/v21/x/valset-pref/valpref-module"
	"github.com/osmosis-labs/osmosis/x/epochs"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

// moduleAccountPermissions defines module account permissions
// TODO: Having to input nil's here is unacceptable, we need a way to automatically derive this.
var moduleAccountPermissions = map[string][]string{
	authtypes.FeeCollectorName:                    nil,
	distrtypes.ModuleName:                         nil,
	ibchookstypes.ModuleName:                      nil,
	icatypes.ModuleName:                           nil,
	icqtypes.ModuleName:                           nil,
	minttypes.ModuleName:                          {authtypes.Minter, authtypes.Burner},
	minttypes.DeveloperVestingModuleAcctName:      nil,
	stakingtypes.BondedPoolName:                   {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName:                {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:                           {authtypes.Burner},
	ibctransfertypes.ModuleName:                   {authtypes.Minter, authtypes.Burner},
	gammtypes.ModuleName:                          {authtypes.Minter, authtypes.Burner},
	incentivestypes.ModuleName:                    {authtypes.Minter, authtypes.Burner},
	protorevtypes.ModuleName:                      {authtypes.Minter, authtypes.Burner},
	lockuptypes.ModuleName:                        {authtypes.Minter, authtypes.Burner},
	poolincentivestypes.ModuleName:                nil,
	superfluidtypes.ModuleName:                    {authtypes.Minter, authtypes.Burner},
	txfeestypes.ModuleName:                        nil,
	txfeestypes.FeeCollectorForStakingRewardsName: nil,
	txfeestypes.FeeCollectorForCommunityPoolName:  nil,
	wasmtypes.ModuleName:                          {authtypes.Burner},
	tokenfactorytypes.ModuleName:                  {authtypes.Minter, authtypes.Burner},
	valsetpreftypes.ModuleName:                    {authtypes.Staking},
	poolmanagertypes.ModuleName:                   nil,
	cosmwasmpooltypes.ModuleName:                  nil,
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
		auth.NewAppModule(appCodec, *app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(*app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, *app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		gov.NewAppModule(appCodec, app.GovKeeper, *app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, *app.MintKeeper, app.AccountKeeper, app.BankKeeper),
		slashing.NewAppModule(appCodec, *app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName)),
		distr.NewAppModule(appCodec, *app.DistrKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		downtimemodule.NewAppModule(*app.DowntimeKeeper),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper),
		wasm.NewAppModule(appCodec, app.WasmKeeper, app.StakingKeeper, *app.AccountKeeper, app.BankKeeper, app.BaseApp.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		evidence.NewAppModule(*app.EvidenceKeeper),
		authzmodule.NewAppModule(appCodec, *app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		ibc.NewAppModule(app.IBCKeeper),
		ica.NewAppModule(nil, app.ICAHostKeeper),
		params.NewAppModule(*app.ParamsKeeper),
		app.RawIcs20TransferAppModule,
		gamm.NewAppModule(appCodec, *app.GAMMKeeper, app.AccountKeeper, app.BankKeeper),
		poolmanager.NewAppModule(*app.PoolManagerKeeper, app.GAMMKeeper),
		twapmodule.NewAppModule(*app.TwapKeeper),
		concentratedliquidity.NewAppModule(appCodec, *app.ConcentratedLiquidityKeeper),
		protorev.NewAppModule(appCodec, *app.ProtoRevKeeper, app.AccountKeeper, app.BankKeeper, app.EpochsKeeper, app.GAMMKeeper),
		txfees.NewAppModule(*app.TxFeesKeeper),
		incentives.NewAppModule(*app.IncentivesKeeper, app.AccountKeeper, app.BankKeeper, app.EpochsKeeper),
		lockup.NewAppModule(*app.LockupKeeper, app.AccountKeeper, app.BankKeeper),
		poolincentives.NewAppModule(*app.PoolIncentivesKeeper),
		epochs.NewAppModule(*app.EpochsKeeper),
		superfluid.NewAppModule(
			*app.SuperfluidKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.StakingKeeper,
			app.LockupKeeper,
			app.GAMMKeeper,
			app.EpochsKeeper,
			app.ConcentratedLiquidityKeeper,
		),
		tokenfactory.NewAppModule(*app.TokenFactoryKeeper, app.AccountKeeper, app.BankKeeper),
		valsetprefmodule.NewAppModule(appCodec, *app.ValidatorSetPreferenceKeeper),
		ibcratelimitmodule.NewAppModule(*app.RateLimitingICS4Wrapper),
		ibc_hooks.NewAppModule(app.AccountKeeper, *app.IBCHooksKeeper),
		icq.NewAppModule(*app.AppKeepers.ICQKeeper, app.GetSubspace(icqtypes.ModuleName)),
		packetforward.NewAppModule(app.PacketForwardKeeper, app.GetSubspace(packetforwardtypes.ModuleName)),
		cwpoolmodule.NewAppModule(appCodec, *app.CosmwasmPoolKeeper),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
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
	// We leave downtime-detector un-constrained.
	// every remaining module's begin block is a no-op.
	return ord.TotalOrdering()
}

// OrderEndBlockers returns EndBlockers (crisis, govtypes, staking) with no relative order.
func OrderEndBlockers(allModuleNames []string) []string {
	ord := partialord.NewPartialOrdering(allModuleNames)

	// Staking must be after gov.
	ord.FirstElements(govtypes.ModuleName)
	ord.LastElements(stakingtypes.ModuleName)

	// only Osmosis modules with endblock code are: twap, crisis, govtypes, staking
	// we don't care about the relative ordering between them.
	return ord.TotalOrdering()
}

// OrderInitGenesis returns module names in order for init genesis calls.
func OrderInitGenesis(allModuleNames []string) []string {
	// NOTE: The genutils moodule must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	return []string{
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		downtimetypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		ibchost.ModuleName,
		icatypes.ModuleName,
		gammtypes.ModuleName,
		poolmanagertypes.ModuleName,
		protorevtypes.ModuleName,
		twaptypes.ModuleName,
		txfeestypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		ibctransfertypes.ModuleName,
		consensusparamtypes.ModuleName,
		poolincentivestypes.ModuleName,
		superfluidtypes.ModuleName,
		tokenfactorytypes.ModuleName,
		valsetpreftypes.ModuleName,
		incentivestypes.ModuleName,
		epochstypes.ModuleName,
		lockuptypes.ModuleName,
		authz.ModuleName,
		concentratedliquiditytypes.ModuleName,
		ibcratelimittypes.ModuleName,
		// wasm after ibc transfer
		wasmtypes.ModuleName,
		// ibc_hooks after auth keeper
		ibchookstypes.ModuleName,
		icqtypes.ModuleName,
		packetforwardtypes.ModuleName,
		cosmwasmpooltypes.ModuleName,
	}
}

// ModuleAccountAddrs returns all the app's module account addresses.
func ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

func (app *OsmosisApp) GetAccountKeeper() simtypes.AccountKeeper {
	return app.AppKeepers.AccountKeeper
}

func (app *OsmosisApp) GetBankKeeper() simtypes.BankKeeper {
	return app.AppKeepers.BankKeeper
}

// Required for ibctesting
func (app *OsmosisApp) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return *app.AppKeepers.StakingKeeper // Dereferencing the pointer
}
func (app *OsmosisApp) GetSDKStakingKeeper() stakingkeeper.Keeper {
	return *app.AppKeepers.StakingKeeper // Dereferencing the pointer
}

func (app *OsmosisApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.AppKeepers.IBCKeeper // This is a *ibckeeper.Keeper
}

func (app *OsmosisApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.AppKeepers.ScopedIBCKeeper
}

func (app *OsmosisApp) GetPoolManagerKeeper() simtypes.PoolManagerKeeper {
	return app.AppKeepers.PoolManagerKeeper
}

func (app *OsmosisApp) GetTxConfig() client.TxConfig {
	return MakeEncodingConfig().TxConfig
}
