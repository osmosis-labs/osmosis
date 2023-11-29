package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	store "github.com/cosmos/cosmos-sdk/store/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/osmosis-labs/osmosis/osmoutils"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"
	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	sqslog "github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v20/app/keepers"
	"github.com/osmosis-labs/osmosis/v20/app/upgrades"
	v10 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v10"
	v11 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v11"
	v12 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v12"
	v13 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v13"
	v14 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v14"
	v15 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v15"
	v16 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v16"
	v17 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v17"
	v18 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v18"
	v19 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v19"
	v20 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v20"
	v3 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v3"
	v4 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v4"
	v5 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v5"
	v6 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v6"
	v7 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v7"
	v8 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v8"
	v9 "github.com/osmosis-labs/osmosis/v20/app/upgrades/v9"
	_ "github.com/osmosis-labs/osmosis/v20/client/docs/statik"
	"github.com/osmosis-labs/osmosis/v20/ingest"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs"

	redischaininfoingester "github.com/osmosis-labs/osmosis/v20/ingest/sqs/chain_info/ingester/redis"
	redispoolsingester "github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/ingester/redis"
)

const (
	appName = "OsmosisApp"

	// Environment variable configurations
	// TODO: replace all SQS environment variables with a config file
	ENV_NAME_INGEST_TYPE                             = "INGEST_TYPE"
	ENV_NAME_INGEST_SQS_DBHOST                       = "INGEST_SQS_DBHOST"
	ENV_NAME_INGEST_SQS_DBPORT                       = "INGEST_SQS_DBPORT"
	ENV_NAME_INGEST_SQS_SERVER_ADDRESS               = "INGEST_SQS_SERVER_ADDRESS"
	ENV_NAME_INGEST_SQS_SERVER_TIMEOUT_DURATION_SECS = "INGEST_SQS_SERVER_TIMEOUT_DURATION_SECS"
	ENV_NAME_INGEST_SQS_LOGGER_FILENAME              = "INGEST_SQS_LOGGER_FILENAME"
	ENV_NAME_INGEST_SQS_LOGGER_IS_PRODUCTION         = "INGEST_SQS_LOGGER_IS_PRODUCTION"
	ENV_NAME_INGEST_SQS_LOGGER_LEVEL                 = "INGEST_SQS_LOGGER_LEVEL"
	ENV_NAME_GRPC_GATEWAY_ENDPOINT                   = "ENV_NAME_GRPC_GATEWAY_ENDPOINT"
	ENV_VALUE_INGESTER_SQS                           = "sqs"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(keepers.AppModuleBasics...)

	// module account permissions
	maccPerms = moduleAccountPermissions

	// module accounts that are allowed to receive tokens.
	allowedReceivingModAcc = map[string]bool{}

	// TODO: Refactor wasm items into a wasm.go file
	// WasmProposalsEnabled enables all x/wasm proposals when it's value is "true"
	// and EnableSpecificWasmProposals is empty. Otherwise, all x/wasm proposals
	// are disabled.
	WasmProposalsEnabled = "true"

	// EnableSpecificWasmProposals, if set, must be comma-separated list of values
	// that are all a subset of "EnableAllProposals", which takes precedence over
	// WasmProposalsEnabled.
	//
	// See: https://github.com/CosmWasm/wasmd/blob/02a54d33ff2c064f3539ae12d75d027d9c665f05/x/wasm/internal/types/proposal.go#L28-L34
	EnableSpecificWasmProposals = ""

	// EmptyWasmOpts defines a type alias for a list of wasm options.
	EmptyWasmOpts []wasm.Option

	// _ sdksimapp.App = (*OsmosisApp)(nil)

	Upgrades = []upgrades.Upgrade{v4.Upgrade, v5.Upgrade, v7.Upgrade, v9.Upgrade, v11.Upgrade, v12.Upgrade, v13.Upgrade, v14.Upgrade, v15.Upgrade, v16.Upgrade, v17.Upgrade, v18.Upgrade, v19.Upgrade, v20.Upgrade}
	Forks    = []upgrades.Fork{v3.Fork, v6.Fork, v8.Fork, v10.Fork}
)

// GetWasmEnabledProposals parses the WasmProposalsEnabled and
// EnableSpecificWasmProposals values to produce a list of enabled proposals to
// pass into the application.
func GetWasmEnabledProposals() []wasm.ProposalType {
	if EnableSpecificWasmProposals == "" {
		if WasmProposalsEnabled == "true" {
			return wasm.EnableAllProposals
		}

		return wasm.DisableAllProposals
	}

	chunks := strings.Split(EnableSpecificWasmProposals, ",")

	proposals, err := wasm.ConvertToProposals(chunks)
	if err != nil {
		panic(err)
	}

	return proposals
}

// OsmosisApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type OsmosisApp struct {
	*baseapp.BaseApp
	keepers.AppKeepers

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry
	invCheckPeriod    uint

	mm           *module.Manager
	configurator module.Configurator
	homePath     string
}

// init sets DefaultNodeHome to default osmosisd install location.
func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".osmosisd")
}

// initReusablePackageInjections injects data available within osmosis into the reusable packages.
// This is done to ensure they can be built without depending on at compilation time and thus imported by other chains
// This should always be called before any other function to avoid inconsistent data
func initReusablePackageInjections() {
	// Inject ClawbackVestingAccount account type into osmoutils
	osmoutils.OsmoUtilsExtraAccountTypes = map[reflect.Type]struct{}{
		reflect.TypeOf(&vestingtypes.ClawbackVestingAccount{}): {},
	}
}

// overrideWasmVariables overrides the wasm variables to:
//   - allow for larger wasm files
func overrideWasmVariables() {
	// Override Wasm size limitation from WASMD.
	wasmtypes.MaxWasmSize = 3 * 1024 * 1024
	wasmtypes.MaxProposalWasmSize = wasmtypes.MaxWasmSize
}

// NewOsmosisApp returns a reference to an initialized Osmosis.
func NewOsmosisApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	appOpts servertypes.AppOptions,
	wasmOpts []wasm.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *OsmosisApp {
	initReusablePackageInjections() // This should run before anything else to make sure the variables are properly initialized
	overrideWasmVariables()
	encodingConfig := GetEncodingConfig()
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry
	wasmEnabledProposals := GetWasmEnabledProposals()

	bApp := baseapp.NewBaseApp(appName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	app := &OsmosisApp{
		AppKeepers:        keepers.AppKeepers{},
		BaseApp:           bApp,
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
	}

	app.homePath = homePath
	dataDir := filepath.Join(homePath, "data")
	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	// Uncomment this for debugging contracts. In the future this could be made into a param passed by the tests
	// wasmConfig.ContractDebugMode = true
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}
	app.InitSpecialKeepers(
		appCodec,
		bApp,
		wasmDir,
		cdc,
		invCheckPeriod,
		skipUpgradeHeights,
		homePath,
	)
	app.setupUpgradeStoreLoaders()
	app.InitNormalKeepers(
		appCodec,
		bApp,
		maccPerms,
		dataDir,
		wasmDir,
		wasmConfig,
		wasmEnabledProposals,
		wasmOpts,
		app.BlockedAddrs(),
	)

	isIngestManagerEnabled := os.Getenv(ENV_NAME_INGEST_TYPE) == ENV_VALUE_INGESTER_SQS
	app.IngestManager = ingest.NewIngestManager()
	if isIngestManagerEnabled {
		dbHost := os.Getenv(ENV_NAME_INGEST_SQS_DBHOST)
		dbPort := os.Getenv(ENV_NAME_INGEST_SQS_DBPORT)
		grpcAddress := os.Getenv(ENV_NAME_GRPC_GATEWAY_ENDPOINT)
		if grpcAddress == "" {
			grpcAddress = "http://localhost:26657"
		}

		sidecarQueryServerAddress := os.Getenv(ENV_NAME_INGEST_SQS_SERVER_ADDRESS)
		sidecarQueryServerTimeoutDuration, err := strconv.Atoi(os.Getenv(ENV_NAME_INGEST_SQS_SERVER_TIMEOUT_DURATION_SECS))
		if err != nil {
			panic(fmt.Sprintf("error while parsing timeout duration: %s", err))
		}

		// logger configs
		loggerFileName := os.Getenv(ENV_NAME_INGEST_SQS_LOGGER_FILENAME)
		isProductionLoggerStr := os.Getenv(ENV_NAME_INGEST_SQS_LOGGER_IS_PRODUCTION)
		isProductionLogger := isProductionLoggerStr == "true"
		logLevel := os.Getenv(ENV_NAME_INGEST_SQS_LOGGER_LEVEL)

		// logger
		logger, err := sqslog.NewLogger(isProductionLogger, loggerFileName, logLevel)
		if err != nil {
			panic(fmt.Sprintf("error while creating logger: %s", err))
		}
		logger.Info("Starting sidecar query server")

		// TODO: move to config file
		routerConfig := domain.RouterConfig{
			PreferredPoolIDs:          []uint64{},
			MaxPoolsPerRoute:          4,
			MaxRoutes:                 5,
			MaxSplitRoutes:            3,
			MaxSplitIterations:        10,
			MinOSMOLiquidity:          10000, // 10_000 OSMO
			RouteUpdateHeightInterval: 0,
			RouteCacheEnabled:         false,
		}

		// Create sidecar query server
		sidecarQueryServer, err := sqs.NewSideCarQueryServer(appCodec, routerConfig, dbHost, dbPort, sidecarQueryServerAddress, grpcAddress, sidecarQueryServerTimeoutDuration, logger)
		if err != nil {
			panic(fmt.Sprintf("error while creating sidecar query server: %s", err))
		}

		txManager := sidecarQueryServer.GetTxManager()

		// Create pools ingester
		poolsIngester := redispoolsingester.NewPoolIngester(sidecarQueryServer.GetPoolsRepository(), sidecarQueryServer.GetRouterRepository(), sidecarQueryServer.GetTokensUseCase(), txManager, routerConfig, app.GAMMKeeper, app.ConcentratedLiquidityKeeper, app.CosmwasmPoolKeeper, app.BankKeeper, app.ProtoRevKeeper, app.PoolManagerKeeper)
		poolsIngester.SetLogger(sidecarQueryServer.GetLogger())

		chainInfoingester := redischaininfoingester.NewChainInfoIngester(sidecarQueryServer.GetChainInfoRepository(), txManager)
		chainInfoingester.SetLogger(sidecarQueryServer.GetLogger())

		// Create sqs ingester that encapsulates all ingesters.
		sqsIngester := sqs.NewSidecarQueryServerIngester(poolsIngester, chainInfoingester, txManager)

		// Set the sidecar query server ingester to the ingest manager.
		app.IngestManager.SetIngester(sqsIngester)
	}

	// TODO: There is a bug here, where we register the govRouter routes in InitNormalKeepers and then
	// call setupHooks afterwards. Therefore, if a gov proposal needs to call a method and that method calls a
	// hook, we will get a nil pointer dereference error due to the hooks in the keeper not being
	// setup yet. I will refrain from creating an issue in the sdk for now until after we unfork to 0.47,
	// because I believe the concept of Routes is going away.
	// https://github.com/osmosis-labs/osmosis/issues/6580
	app.SetupHooks()

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: All module / keeper changes should happen prior to this module.NewManager line being called.
	// However in the event any changes do need to happen after this call, ensure that that keeper
	// is only passed in its keeper form (not de-ref'd anywhere)
	//
	// Generally NewAppModule will require the keeper that module defines to be passed in as an exact struct,
	// but should take in every other keeper as long as it matches a certain interface. (So no need to be de-ref'd)
	//
	// Any time a module requires a keeper de-ref'd thats not its native one,
	// its code-smell and should probably change. We should get the staking keeper dependencies fixed.
	app.mm = module.NewManager(appModules(app, encodingConfig, skipGenesisInvariants)...)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	// NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)

	// Tell the app's module manager how to set the order of BeginBlockers, which are run at the beginning of every block.
	app.mm.SetOrderBeginBlockers(orderBeginBlockers(app.mm.ModuleNames())...)

	// Tell the app's module manager how to set the order of EndBlockers, which are run at the end of every block.
	app.mm.SetOrderEndBlockers(OrderEndBlockers(app.mm.ModuleNames())...)

	app.mm.SetOrderInitGenesis(OrderInitGenesis(app.mm.ModuleNames())...)

	app.mm.RegisterInvariants(app.CrisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), encodingConfig.Amino)
	app.configurator = module.NewConfigurator(app.AppCodec(), app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	app.setupUpgradeHandlers()

	// app.sm.RegisterStoreDecoders()

	// add test gRPC service for testing gRPC queries in isolation
	testdata.RegisterQueryServer(app.GRPCQueryRouter(), testdata.QueryImpl{})

	// initialize stores
	app.MountKVStores(app.GetKVStoreKey())
	app.MountTransientStores(app.GetTransientStoreKey())
	app.MountMemoryStores(app.GetMemoryStoreKey())

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(
		NewAnteHandler(
			appOpts,
			wasmConfig,
			app.GetKey(wasm.StoreKey),
			app.AccountKeeper,
			app.BankKeeper,
			app.TxFeesKeeper,
			app.GAMMKeeper,
			ante.DefaultSigVerificationGasConsumer,
			encodingConfig.TxConfig.SignModeHandler(),
			app.IBCKeeper,
		),
	)
	app.SetPostHandler(NewPostHandler(app.ProtoRevKeeper))
	app.SetEndBlocker(app.EndBlocker)

	// Register snapshot extensions to enable state-sync for wasm.
	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), app.WasmKeeper),
		)
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}

		ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})

		// Initialize pinned codes in wasmvm as they are not persisted there
		if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			tmos.Exit(fmt.Sprintf("failed initialize pinned codes %s", err))
		}
	}

	return app
}

// MakeCodecs returns the application codec and a legacy Amino codec.
func MakeCodecs() (codec.Codec, *codec.LegacyAmino) {
	config := MakeEncodingConfig()
	return config.Marshaler, config.Amino
}

func (app *OsmosisApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// Name returns the name of the App.
func (app *OsmosisApp) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block.
func (app *OsmosisApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	BeginBlockForks(ctx, app)
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block.
func (app *OsmosisApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	app.IngestManager.ProcessBlock(ctx)

	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization.
func (app *OsmosisApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())

	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height.
func (app *OsmosisApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
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

// InterfaceRegistry returns Osmosis' InterfaceRegistry.
func (app *OsmosisApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

func (app *OsmosisApp) ModuleManager() module.Manager {
	return *app.mm
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

// RegisterTendermintService implements the Application.RegisterTendermintService
// method.
func (app *OsmosisApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.interfaceRegistry)
}

// configure store loader that checks if version == upgradeHeight and applies store upgrades
func (app *OsmosisApp) setupUpgradeStoreLoaders() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	currentHeight := app.CommitMultiStore().LastCommitID().Version

	if upgradeInfo.Height == currentHeight+1 {
		app.customPreUpgradeHandler(upgradeInfo)
	}

	for _, upgrade := range Upgrades {
		if upgradeInfo.Name == upgrade.UpgradeName {
			storeUpgrades := upgrade.StoreUpgrades
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
		}
	}
}

func (app *OsmosisApp) customPreUpgradeHandler(upgradeInfo store.UpgradeInfo) {
	switch upgradeInfo.Name {
	case "v16":
		// v16 upgrade handler
		// remove the wasm cache for cosmwasm cherry https://github.com/CosmWasm/advisories/blob/main/CWAs/CWA-2023-002.md#wasm-module-cache-issue
		err := os.RemoveAll(app.homePath + "/wasm/wasm/cache")
		if err != nil {
			panic(err)
		}
	}
}

func (app *OsmosisApp) setupUpgradeHandlers() {
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(
				app.mm,
				app.configurator,
				app.BaseApp,
				&app.AppKeepers,
			),
		)
	}
}

// RegisterSwaggerAPI registers swagger route with API Server.
func RegisterSwaggerAPI(ctx client.Context, rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticServer))
	rtr.PathPrefix("/swagger/").Handler(staticServer)
}

// GetMaccPerms returns a copy of the module account permissions.
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}

	return dupMaccPerms
}
