package app

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	icq "github.com/cosmos/ibc-apps/modules/async-icq/v8"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/market"
	markettypes "github.com/osmosis-labs/osmosis/v27/x/market/types"
	"github.com/osmosis-labs/osmosis/v27/x/oracle"
	oracletypes "github.com/osmosis-labs/osmosis/v27/x/oracle/types"
	"github.com/osmosis-labs/osmosis/v27/x/treasury"
	treasurytypes "github.com/osmosis-labs/osmosis/v27/x/treasury/types"

	ibcwasm "github.com/cosmos/ibc-go/modules/light-clients/08-wasm"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	ibchost "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ibctestingtypes "github.com/cosmos/ibc-go/v8/testing/types"

	packetforward "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"

	ibchookstypes "github.com/osmosis-labs/osmosis/x/ibc-hooks/types"

	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"

	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v8/types"

	downtimemodule "github.com/osmosis-labs/osmosis/v27/x/downtime-detector/module"
	downtimetypes "github.com/osmosis-labs/osmosis/v27/x/downtime-detector/types"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"

	ibc_hooks "github.com/osmosis-labs/osmosis/x/ibc-hooks"

	"cosmossdk.io/x/evidence"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
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
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	"github.com/skip-mev/block-sdk/v2/x/auction"
	auctiontypes "github.com/skip-mev/block-sdk/v2/x/auction/types"

	"github.com/osmosis-labs/osmosis/osmoutils/partialord"
	smartaccount "github.com/osmosis-labs/osmosis/v27/x/smart-account"
	smartaccounttypes "github.com/osmosis-labs/osmosis/v27/x/smart-account/types"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	_ "github.com/osmosis-labs/osmosis/v27/client/docs/statik"
	"github.com/osmosis-labs/osmosis/v27/simulation/simtypes"
	concentratedliquidity "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/clmodule"
	concentratedliquiditytypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	cwpoolmodule "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/module"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
	"github.com/osmosis-labs/osmosis/v27/x/epochs"
	epochstypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v27/x/gamm"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/ibcratelimitmodule"
	ibcratelimittypes "github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/types"
	"github.com/osmosis-labs/osmosis/v27/x/incentives"
	incentivestypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	"github.com/osmosis-labs/osmosis/v27/x/lockup"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/mint"
	minttypes "github.com/osmosis-labs/osmosis/v27/x/mint/types"
	poolincentives "github.com/osmosis-labs/osmosis/v27/x/pool-incentives"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
	poolmanager "github.com/osmosis-labs/osmosis/v27/x/poolmanager/module"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v27/x/protorev"
	protorevtypes "github.com/osmosis-labs/osmosis/v27/x/protorev/types"
	stablestakingincentives "github.com/osmosis-labs/osmosis/v27/x/stable-staking-incentives"
	stablestakingincentivestypes "github.com/osmosis-labs/osmosis/v27/x/stable-staking-incentives/types"
	superfluid "github.com/osmosis-labs/osmosis/v27/x/superfluid"
	superfluidtypes "github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
	"github.com/osmosis-labs/osmosis/v27/x/tokenfactory"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"
	"github.com/osmosis-labs/osmosis/v27/x/twap/twapmodule"
	twaptypes "github.com/osmosis-labs/osmosis/v27/x/twap/types"
	"github.com/osmosis-labs/osmosis/v27/x/txfees"
	txfeestypes "github.com/osmosis-labs/osmosis/v27/x/txfees/types"
	valsetpreftypes "github.com/osmosis-labs/osmosis/v27/x/valset-pref/types"
	valsetprefmodule "github.com/osmosis-labs/osmosis/v27/x/valset-pref/valpref-module"
)

// moduleAccountPermissions defines module account permissions
// TODO: Having to input nil's here is unacceptable, we need a way to automatically derive this.
var moduleAccountPermissions = map[string][]string{
	authtypes.FeeCollectorName:               nil,
	distrtypes.ModuleName:                    nil,
	ibchookstypes.ModuleName:                 nil,
	icatypes.ModuleName:                      nil,
	icqtypes.ModuleName:                      nil,
	minttypes.ModuleName:                     {authtypes.Minter, authtypes.Burner},
	minttypes.DeveloperVestingModuleAcctName: nil,
	stakingtypes.BondedPoolName:              {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName:           {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:                      {authtypes.Burner},
	ibctransfertypes.ModuleName:              {authtypes.Minter, authtypes.Burner},
	gammtypes.ModuleName:                     {authtypes.Minter, authtypes.Burner},
	incentivestypes.ModuleName:               {authtypes.Minter, authtypes.Burner},
	protorevtypes.ModuleName:                 {authtypes.Minter, authtypes.Burner},
	lockuptypes.ModuleName:                   {authtypes.Minter, authtypes.Burner},
	poolincentivestypes.ModuleName:           nil,
	stablestakingincentivestypes.ModuleName:  nil,
	superfluidtypes.ModuleName:               {authtypes.Minter, authtypes.Burner},
	txfeestypes.ModuleName:                   nil,
	txfeestypes.NonNativeTxFeeCollectorName:  nil,
	wasmtypes.ModuleName:                     {authtypes.Burner},
	tokenfactorytypes.ModuleName:             {authtypes.Minter, authtypes.Burner},
	valsetpreftypes.ModuleName:               {authtypes.Staking},
	poolmanagertypes.ModuleName:              nil,
	markettypes.ModuleName:                   {authtypes.Minter, authtypes.Burner},
	treasurytypes.ModuleName:                 {authtypes.Minter, authtypes.Burner},
	oracletypes.ModuleName:                   nil,
	cosmwasmpooltypes.ModuleName:             nil,
	auctiontypes.ModuleName:                  nil,
	smartaccounttypes.ModuleName:             nil,
}

// appModules return modules to initialize module manager.
func appModules(
	app *SymphonyApp,
	encodingConfig appparams.EncodingConfig,
	skipGenesisInvariants bool,
) []module.AppModule {
	appCodec := encodingConfig.Marshaler

	return []module.AppModule{
		genutil.NewAppModule(
			app.AccountKeeper,
			app.StakingKeeper,
			app.BaseApp,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, *app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(*app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		gov.NewAppModule(appCodec, app.GovKeeper, *app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, *app.MintKeeper, app.AccountKeeper, app.BankKeeper),
		slashing.NewAppModule(appCodec, *app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName), app.interfaceRegistry),
		distr.NewAppModule(appCodec, *app.DistrKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		downtimemodule.NewAppModule(*app.DowntimeKeeper),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper, addresscodec.NewBech32Codec(appparams.Bech32PrefixAccAddr)),
		wasm.NewAppModule(appCodec, app.WasmKeeper, app.StakingKeeper, *app.AccountKeeper, app.BankKeeper, app.BaseApp.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		evidence.NewAppModule(*app.EvidenceKeeper),
		authzmodule.NewAppModule(appCodec, *app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		ibc.NewAppModule(app.IBCKeeper),
		ibcwasm.NewAppModule(*app.IBCWasmClientKeeper),
		ica.NewAppModule(app.ICAControllerKeeper, app.ICAHostKeeper),
		params.NewAppModule(*app.ParamsKeeper),
		consensus.NewAppModule(appCodec, *app.AppKeepers.ConsensusParamsKeeper),
		app.RawIcs20TransferAppModule,
		gamm.NewAppModule(appCodec, *app.GAMMKeeper, app.AccountKeeper, app.BankKeeper),
		poolmanager.NewAppModule(*app.PoolManagerKeeper, app.GAMMKeeper),
		market.NewAppModule(*app.MarketKeeper, app.AccountKeeper, app.BankKeeper, app.OracleKeeper),
		oracle.NewAppModule(appCodec, *app.OracleKeeper, app.AccountKeeper, app.BankKeeper),
		treasury.NewAppModule(appCodec, *app.TreasuryKeeper),
		twapmodule.NewAppModule(*app.TwapKeeper),
		concentratedliquidity.NewAppModule(appCodec, *app.ConcentratedLiquidityKeeper),
		protorev.NewAppModule(appCodec, *app.ProtoRevKeeper, app.AccountKeeper, app.BankKeeper, app.EpochsKeeper, app.GAMMKeeper),
		txfees.NewAppModule(*app.TxFeesKeeper),
		incentives.NewAppModule(*app.IncentivesKeeper, app.AccountKeeper, app.BankKeeper, app.EpochsKeeper),
		lockup.NewAppModule(*app.LockupKeeper, app.AccountKeeper, app.BankKeeper),
		poolincentives.NewAppModule(*app.PoolIncentivesKeeper),
		stablestakingincentives.NewAppModule(*app.StableStakingIncentivesKeeper),
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
		auction.NewAppModule(appCodec, *app.AuctionKeeper),
		smartaccount.NewAppModule(appCodec, *app.SmartAccountKeeper),
	}
}

// orderBeginBlockers returns the order of BeginBlockers, by module name.
func orderBeginBlockers(allModuleNames []string) []string {
	ord := partialord.NewPartialOrdering(allModuleNames)
	// Upgrades should be run VERY first
	// Epochs is set to be next right now, this in principle could change to come later / be at the end,
	// but would have to be a holistic change with other pipelines taken into account.
	// Epochs must come before staking, because txfees epoch hook sends fees to the auth "fee collector"
	// module account, which is then distributed to stakers. If staking comes before epochs, then the
	// funds will not be distributed to stakers as expected.
	ord.FirstElements(epochstypes.ModuleName, capabilitytypes.ModuleName)

	// Staking ordering
	// TODO: Perhaps this can be relaxed, left to future work to analyze.
	ord.Sequence(distrtypes.ModuleName, slashingtypes.ModuleName, evidencetypes.ModuleName, stakingtypes.ModuleName)
	// superfluid must come after distribution & epochs.
	// TODO: we actually set it to come after staking, since that's what happened before, and want to minimize chance of break.
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

	// only Symphony modules with endblock code are: twap, crisis, govtypes, staking
	// we don't care about the relative ordering between them.
	return ord.TotalOrdering()
}

// OrderInitGenesis returns module names in order for init genesis calls.
func OrderInitGenesis(allModuleNames []string) []string {
	// NOTE: The genutils module must occur after staking so that pools are
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
		markettypes.ModuleName,
		oracletypes.ModuleName,
		treasurytypes.ModuleName,
		protorevtypes.ModuleName,
		twaptypes.ModuleName,
		txfeestypes.ModuleName,
		smartaccounttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		ibctransfertypes.ModuleName,
		consensusparamtypes.ModuleName,
		poolincentivestypes.ModuleName,
		stablestakingincentivestypes.ModuleName,
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
		ibcwasmtypes.ModuleName,
		// ibc_hooks after auth keeper
		ibchookstypes.ModuleName,
		icqtypes.ModuleName,
		packetforwardtypes.ModuleName,
		cosmwasmpooltypes.ModuleName,
		auctiontypes.ModuleName,
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

func (app *SymphonyApp) GetAccountKeeper() simtypes.AccountKeeper {
	return app.AppKeepers.AccountKeeper
}

func (app *SymphonyApp) GetBankKeeper() simtypes.BankKeeper {
	return app.AppKeepers.BankKeeper
}

// Required for ibctesting
func (app *SymphonyApp) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return *app.AppKeepers.StakingKeeper // Dereferencing the pointer
}
func (app *SymphonyApp) GetSDKStakingKeeper() stakingkeeper.Keeper {
	return *app.AppKeepers.StakingKeeper // Dereferencing the pointer
}

func (app *SymphonyApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.AppKeepers.IBCKeeper // This is a *ibckeeper.Keeper
}

func (app *SymphonyApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.AppKeepers.ScopedIBCKeeper
}

func (app *SymphonyApp) GetPoolManagerKeeper() simtypes.PoolManagerKeeper {
	return app.AppKeepers.PoolManagerKeeper
}

func (app *SymphonyApp) GetTxConfig() client.TxConfig {
	return GetEncodingConfig().TxConfig
}
