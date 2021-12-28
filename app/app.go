package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"
	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	transfer "github.com/cosmos/ibc-go/v2/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v2/modules/core"
	ibcclientclient "github.com/cosmos/ibc-go/v2/modules/core/02-client/client"
	ibchost "github.com/cosmos/ibc-go/v2/modules/core/24-host"
	"github.com/gorilla/mux"

	"github.com/osmosis-labs/osmosis/app/keepers"
	appparams "github.com/osmosis-labs/osmosis/app/params"
	v4 "github.com/osmosis-labs/osmosis/app/upgrades/v4"
	v5 "github.com/osmosis-labs/osmosis/app/upgrades/v5"
	_ "github.com/osmosis-labs/osmosis/client/docs/statik"
	"github.com/osmosis-labs/osmosis/x/claim"
	claimtypes "github.com/osmosis-labs/osmosis/x/claim/types"
	"github.com/osmosis-labs/osmosis/x/epochs"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	"github.com/osmosis-labs/osmosis/x/gamm"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
	"github.com/osmosis-labs/osmosis/x/incentives"
	incentivestypes "github.com/osmosis-labs/osmosis/x/incentives/types"
	"github.com/osmosis-labs/osmosis/x/lockup"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/osmosis-labs/osmosis/x/mint"
	minttypes "github.com/osmosis-labs/osmosis/x/mint/types"
	poolincentives "github.com/osmosis-labs/osmosis/x/pool-incentives"
	poolincentivesclient "github.com/osmosis-labs/osmosis/x/pool-incentives/client"
	poolincentivestypes "github.com/osmosis-labs/osmosis/x/pool-incentives/types"
	"github.com/osmosis-labs/osmosis/x/txfees"
	txfeestypes "github.com/osmosis-labs/osmosis/x/txfees/types"

	"github.com/osmosis-labs/bech32-ibc/x/bech32ibc"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"
	"github.com/osmosis-labs/bech32-ibc/x/bech32ics20"
)

const appName = "OsmosisApp"

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			paramsclient.ProposalHandler, distrclient.ProposalHandler, upgradeclient.ProposalHandler, upgradeclient.CancelProposalHandler,
			poolincentivesclient.UpdatePoolIncentivesHandler,
			ibcclientclient.UpdateClientProposalHandler, ibcclientclient.UpgradeProposalHandler,
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		ibc.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
		vesting.AppModuleBasic{},
		gamm.AppModuleBasic{},
		txfees.AppModuleBasic{},
		incentives.AppModuleBasic{},
		lockup.AppModuleBasic{},
		poolincentives.AppModuleBasic{},
		epochs.AppModuleBasic{},
		claim.AppModuleBasic{},
		bech32ibc.AppModuleBasic{},
	)

	// module accounts that are allowed to receive tokens
	allowedReceivingModAcc = map[string]bool{
		distrtypes.ModuleName: true,
	}
)

var _ App = (*OsmosisApp)(nil)

// Osmosis extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type OsmosisApp struct {
	*baseapp.BaseApp
	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*sdk.KVStoreKey
	tkeys   map[string]*sdk.TransientStoreKey
	memKeys map[string]*sdk.MemoryStoreKey

	// keepers
	Keepers keepers.AppKeepers

	// the module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager

	// module migration manager
	configurator module.Configurator
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".osmosisd")
}

// NewOsmosis returns a reference to an initialized Osmosis.
func NewOsmosisApp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, skipUpgradeHeights map[int64]bool,
	homePath string, invCheckPeriod uint, encodingConfig appparams.EncodingConfig, appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp),
) *OsmosisApp {

	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(appName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := keepers.AllKVStoreKeys()
	tkeys := keepers.AllTStoreKeys()
	memKeys := keepers.AllMemStoreKeys()

	app := &OsmosisApp{
		BaseApp:           bApp,
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		Keepers:           keepers.NewBlankKeepers(),
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	keepers.InitSpecialKeepers(&app.Keepers,
		app.BaseApp, encodingConfig, cdc, skipUpgradeHeights, homePath, invCheckPeriod)

	app.setupUpgrades()

	keepers.InitNormalKeepers(&app.Keepers, app.BaseApp, encodingConfig, cdc, app.BlockedAddrs())
	keepers.SetupHooks(&app.Keepers)

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	var skipGenesisInvariants = cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		genutil.NewAppModule(
			app.Keepers.AccountKeeper, app.Keepers.StakingKeeper, app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, *app.Keepers.AccountKeeper, nil),
		vesting.NewAppModule(*app.Keepers.AccountKeeper, app.Keepers.BankKeeper),
		// TODO: This is where I left off in changes
		bech32ics20.NewAppModule(appCodec, *app.Keepers.Bech32ICS20Keeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper),
		transferModule,
		claim.NewAppModule(appCodec, *app.ClaimKeeper),
		gamm.NewAppModule(appCodec, app.GAMMKeeper, app.AccountKeeper, app.BankKeeper),
		txfees.NewAppModule(appCodec, app.TxFeesKeeper),
		incentives.NewAppModule(appCodec, app.IncentivesKeeper, app.AccountKeeper, app.BankKeeper, app.EpochsKeeper),
		lockup.NewAppModule(appCodec, app.LockupKeeper, app.AccountKeeper, app.BankKeeper),
		poolincentives.NewAppModule(appCodec, app.PoolIncentivesKeeper),
		epochs.NewAppModule(appCodec, app.EpochsKeeper),
		bech32ibc.NewAppModule(appCodec, app.Bech32IBCKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		// Upgrades should be run _very_ first
		upgradetypes.ModuleName,
		// Note: epochs' begin should be "real" start of epochs, we keep epochs beginblock at the beginning
		epochstypes.ModuleName,
		minttypes.ModuleName, poolincentivestypes.ModuleName, distrtypes.ModuleName, slashingtypes.ModuleName,
		evidencetypes.ModuleName, stakingtypes.ModuleName, ibchost.ModuleName, capabilitytypes.ModuleName,
	)
	app.mm.SetOrderEndBlockers(
		lockuptypes.ModuleName,
		crisistypes.ModuleName, govtypes.ModuleName, stakingtypes.ModuleName, claimtypes.ModuleName,
		authz.ModuleName,
		// Note: epochs' endblock should be "real" end of epochs, we keep epochs endblock at the end
		epochstypes.ModuleName,
	)

	// NOTE: The genutils moodule must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName, authtypes.ModuleName, banktypes.ModuleName, distrtypes.ModuleName, stakingtypes.ModuleName,
		slashingtypes.ModuleName, govtypes.ModuleName, minttypes.ModuleName, crisistypes.ModuleName,
		ibchost.ModuleName,
		gammtypes.ModuleName,
		txfeestypes.ModuleName,
		genutiltypes.ModuleName, evidencetypes.ModuleName, ibctransfertypes.ModuleName,
		bech32ibctypes.ModuleName, // comes after ibctransfertypes
		poolincentivestypes.ModuleName,
		claimtypes.ModuleName,
		incentivestypes.ModuleName,
		epochstypes.ModuleName,
		lockuptypes.ModuleName,
		authz.ModuleName,
	)

	app.mm.RegisterInvariants(&app.CrisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), encodingConfig.Amino)
	app.configurator = module.NewConfigurator(app.AppCodec(), app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	app.sm = module.NewSimulationManager(
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		gamm.NewAppModule(appCodec, app.GAMMKeeper, app.AccountKeeper, app.BankKeeper),
		txfees.NewAppModule(appCodec, app.TxFeesKeeper),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		params.NewAppModule(app.ParamsKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		incentives.NewAppModule(appCodec, app.IncentivesKeeper, app.AccountKeeper, app.BankKeeper, app.EpochsKeeper),
		lockup.NewAppModule(appCodec, app.LockupKeeper, app.AccountKeeper, app.BankKeeper),
		poolincentives.NewAppModule(appCodec, app.PoolIncentivesKeeper),
		epochs.NewAppModule(appCodec, app.EpochsKeeper),
		transferModule,
	)

	app.sm.RegisterStoreDecoders()

	// add test gRPC service for testing gRPC queries in isolation
	testdata.RegisterQueryServer(app.GRPCQueryRouter(), testdata.QueryImpl{})

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(
		NewAnteHandler(
			app.AccountKeeper, app.BankKeeper,
			app.TxFeesKeeper, app.GAMMKeeper,
			ante.DefaultSigVerificationGasConsumer,
			encodingConfig.TxConfig.SignModeHandler(),
			app.IBCKeeper.ChannelKeeper,
		),
	)
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper

	return app
}

// MakeCodecs constructs the *std.Codec and *codec.LegacyAmino instances used by
// simapp. It is useful for tests and clients who do not want to construct the
// full simapp
func MakeCodecs() (codec.Codec, *codec.LegacyAmino) {
	config := MakeEncodingConfig()
	return config.Marshaler, config.Amino
}

// Name returns the name of the App
func (app *OsmosisApp) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *OsmosisApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	BeginBlockForks(ctx, app)
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *OsmosisApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *OsmosisApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())

	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *OsmosisApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *OsmosisApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range keepers.MaccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlockedAddrs returns all the app's module account addresses that are not
// allowed to receive external tokens.
func (app *OsmosisApp) BlockedAddrs() map[string]bool {
	blockedAddrs := make(map[string]bool)
	for acc := range keepers.MaccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = !allowedReceivingModAcc[acc]
	}

	// We block all OFAC-blocked ETH addresses from receiving tokens as well
	// The list is sourced from: https://www.treasury.gov/ofac/downloads/sanctions/1.0/sdn_advanced.xml
	ofacRawEthAddrs := []string{
		"0x7F367cC41522cE07553e823bf3be79A889DEbe1B",
		"0xd882cfc20f52f2599d84b8e8d58c7fb62cfe344b",
		"0x901bb9583b24d97e995513c6778dc6888ab6870e",
		"0xa7e5d5a720f06526557c513402f2e6b5fa20b008",
		"0x8576acc5c05d6ce88f4e49bf65bdf0c62f91353c",
		"0x1da5821544e25c636c1417ba96ade4cf6d2f9b5a",
		"0x7Db418b5D567A4e0E8c59Ad71BE1FcE48f3E6107",
		"0x72a5843cc08275C8171E582972Aa4fDa8C397B2A",
		"0x7F19720A857F834887FC9A7bC0a0fBe7Fc7f8102",
		"0x9f4cda013e354b8fc285bf4b9a60460cee7f7ea9",
		"0x3cbded43efdaf0fc77b9c55f6fc9988fcc9b757d",
		"0x2f389ce8bd8ff92de3402ffce4691d17fc4f6535",
		"0x19aa5fe80d33a56d56c78e82ea5e50e5d80b4dff",
		"0xe7aa314c77f4233c18c6cc84384a9247c0cf367b",
		"0x308ed4b7b49797e1a98d3818bff6fe5385410370",
		"0x2f389ce8bd8ff92de3402ffce4691d17fc4f6535",
		"0x19aa5fe80d33a56d56c78e82ea5e50e5d80b4dff",
		"0x67d40EE1A85bf4a4Bb7Ffae16De985e8427B6b45",
		"0x6f1ca141a28907f78ebaa64fb83a9088b02a8352",
		"0x6acdfba02d390b97ac2b2d42a63e85293bcc160e",
		"0x48549a34ae37b12f6a30566245176994e17c6b4a",
		"0x5512d943ed1f7c8a43f3435c85f7ab68b30121b0",
		"0xc455f7fd3e0e12afd51fba5c106909934d8a0e4a",
		"0xfec8a60023265364d066a1212fde3930f6ae8da7",
	}
	for _, addr := range ofacRawEthAddrs {
		blockedAddrs[addr] = true
		blockedAddrs[strings.ToLower(addr)] = true
	}

	return blockedAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *OsmosisApp) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

// AppCodec returns Osmosis' app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *OsmosisApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns Osmosis' InterfaceRegistry
func (app *OsmosisApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *OsmosisApp) GetKey(storeKey string) *sdk.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *OsmosisApp) GetTKey(storeKey string) *sdk.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *OsmosisApp) GetMemKey(storeKey string) *sdk.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *OsmosisApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface
func (app *OsmosisApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *OsmosisApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	rpc.RegisterRoutes(clientCtx, apiSvr.Router)
	// Register legacy tx routes.
	authrest.RegisterTxRoutes(clientCtx, apiSvr.Router)
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	ModuleBasics.RegisterRESTRoutes(clientCtx, apiSvr.Router)
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(clientCtx, apiSvr.Router)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *OsmosisApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *OsmosisApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *OsmosisApp) setupUpgrades() {
	// this configures a no-op upgrade handler for the v4 upgrade,
	// which improves the lockup module's store management.
	app.Keepers.UpgradeKeeper.SetUpgradeHandler(
		v4.UpgradeName, v4.CreateUpgradeHandler(
			app.mm, app.configurator,
			&app.Keepers))

	app.Keepers.UpgradeKeeper.SetUpgradeHandler(
		v5.UpgradeName,
		v5.CreateUpgradeHandler(
			app.mm, app.configurator,
			&app.Keepers))

	upgradeInfo, err := app.Keepers.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if upgradeInfo.Name == v5.UpgradeName && !app.Keepers.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := store.StoreUpgrades{
			Added: []string{authz.ModuleName, txfees.ModuleName, bech32ibctypes.ModuleName},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(ctx client.Context, rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticServer))
	rtr.PathPrefix("/swagger/").Handler(staticServer)
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range keepers.MaccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}
