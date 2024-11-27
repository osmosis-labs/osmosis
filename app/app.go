package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/skip-mev/block-sdk/v2/block"
	"github.com/skip-mev/block-sdk/v2/block/base"

	"cosmossdk.io/x/evidence"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/ibc-go/modules/capability"
	ibcwasmkeeper "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/keeper"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"

	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"

	"github.com/osmosis-labs/osmosis/v27/ingest/common/poolextractor"
	"github.com/osmosis-labs/osmosis/v27/ingest/common/pooltracker"
	"github.com/osmosis-labs/osmosis/v27/ingest/common/writelistener"
	"github.com/osmosis-labs/osmosis/v27/ingest/indexer"
	indexerdomain "github.com/osmosis-labs/osmosis/v27/ingest/indexer/domain"
	indexerservice "github.com/osmosis-labs/osmosis/v27/ingest/indexer/service"
	indexerwritelistener "github.com/osmosis-labs/osmosis/v27/ingest/indexer/service/writelistener"
	"github.com/osmosis-labs/osmosis/v27/ingest/sqs"
	"github.com/osmosis-labs/osmosis/v27/ingest/sqs/domain"
	poolstransformer "github.com/osmosis-labs/osmosis/v27/ingest/sqs/pools/transformer"

	sqsservice "github.com/osmosis-labs/osmosis/v27/ingest/sqs/service"
	concentratedtypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"

	commondomain "github.com/osmosis-labs/osmosis/v27/ingest/common/domain"
	commonservice "github.com/osmosis-labs/osmosis/v27/ingest/common/service"

	"github.com/osmosis-labs/osmosis/osmomath"

	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/osmoutils"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"

	"cosmossdk.io/log"
	"github.com/CosmWasm/wasmd/x/wasm"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/bytes"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/crisis"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"

	minttypes "github.com/osmosis-labs/osmosis/v27/x/mint/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v27/x/protorev/types"

	"github.com/osmosis-labs/osmosis/v27/app/keepers"
	"github.com/osmosis-labs/osmosis/v27/app/upgrades"
	v10 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v10"
	v11 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v11"
	v12 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v12"
	v13 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v13"
	v14 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v14"
	v15 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v15"
	v16 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v16"
	v17 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v17"
	v18 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v18"
	v19 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v19"
	v20 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v20"
	v21 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v21"
	v22 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v22"
	v23 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v23"
	v24 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v24"
	v25 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v25"
	v26 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v26"
	v27 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v27"
	v28 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v28"
	v3 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v3"
	v4 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v4"
	v5 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v5"
	v6 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v6"
	v7 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v7"
	v8 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v8"
	v9 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v9"
	_ "github.com/osmosis-labs/osmosis/v27/client/docs/statik"
	"github.com/osmosis-labs/osmosis/v27/x/mint"

	blocksdkabci "github.com/skip-mev/block-sdk/v2/abci"
	"github.com/skip-mev/block-sdk/v2/abci/checktx"
	"github.com/skip-mev/block-sdk/v2/block/utils"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"

	clclient "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/client"
	cwpoolclient "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/client"
	gammclient "github.com/osmosis-labs/osmosis/v27/x/gamm/client"
	incentivesclient "github.com/osmosis-labs/osmosis/v27/x/incentives/client"
	poolincentivesclient "github.com/osmosis-labs/osmosis/v27/x/pool-incentives/client"
	poolmanagerclient "github.com/osmosis-labs/osmosis/v27/x/poolmanager/client"
	superfluidclient "github.com/osmosis-labs/osmosis/v27/x/superfluid/client"
	txfeesclient "github.com/osmosis-labs/osmosis/v27/x/txfees/client"
)

const appName = "OsmosisApp"

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// module account permissions
	maccPerms = moduleAccountPermissions

	// module accounts that are allowed to receive tokens.
	allowedReceivingModAcc = map[string]bool{protorevtypes.ModuleName: true}

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
	EmptyWasmOpts []wasmkeeper.Option

	_ runtime.AppI = (*OsmosisApp)(nil)

	Upgrades = []upgrades.Upgrade{v4.Upgrade, v5.Upgrade, v7.Upgrade, v9.Upgrade, v11.Upgrade, v12.Upgrade, v13.Upgrade, v14.Upgrade, v15.Upgrade, v16.Upgrade, v17.Upgrade, v18.Upgrade, v19.Upgrade, v20.Upgrade, v21.Upgrade, v22.Upgrade, v23.Upgrade, v24.Upgrade, v25.Upgrade, v26.Upgrade, v27.Upgrade, v28.Upgrade}
	Forks    = []upgrades.Fork{v3.Fork, v6.Fork, v8.Fork, v10.Fork}

	// rpcAddressConfigName is the name of the config key that holds the RPC address.
	rpcAddressConfigName = "rpc.laddr"
)

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
	ModuleBasics module.BasicManager
	sm           *module.SimulationManager
	configurator module.Configurator
	homePath     string

	checkTxHandler checktx.CheckTx
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
	wasmOpts []wasmkeeper.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *OsmosisApp {
	// Handler OTEL configuration.
	OTELConfig := NewOTELConfigFromOptions(appOpts)
	if OTELConfig.Enabled {
		ctx := context.Background()

		res, err := resource.New(ctx, resource.WithContainer(),
			resource.WithAttributes(semconv.ServiceNameKey.String(OTELConfig.ServiceName)),
			resource.WithFromEnv(),
		)
		if err != nil {
			panic(err)
		}

		_, err = initOTELTracer(ctx, res)
		if err != nil {
			panic(err)
		}
	}

	initReusablePackageInjections() // This should run before anything else to make sure the variables are properly initialized
	overrideWasmVariables()
	encodingConfig := GetEncodingConfig()
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry
	txConfig := encodingConfig.TxConfig

	bApp := baseapp.NewBaseApp(appName, logger, db, txConfig.TxDecoder(), baseAppOptions...)
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
	ibcWasmConfig := ibcwasmtypes.WasmConfig{
		DataDir:               filepath.Join(homePath, "ibc_08-wasm"),
		SupportedCapabilities: []string{"iterator", "stargate", "abort"},
		ContractDebugMode:     false,
	}
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
		encodingConfig,
		bApp,
		maccPerms,
		dataDir,
		wasmDir,
		wasmConfig,
		wasmOpts,
		app.BlockedAddrs(),
		ibcWasmConfig,
	)

	// Initialize the config object for the SQS ingester
	sqsConfig := sqs.NewConfigFromOptions(appOpts)

	// Initialize the config object for the indexer
	indexerConfig := indexer.NewConfigFromOptions(appOpts)

	var nodeStatusChecker commonservice.NodeStatusChecker
	if sqsConfig.IsEnabled || indexerConfig.IsEnabled {
		// Note: address can be moved to config in the future if needed.
		rpcAddress, ok := appOpts.Get(rpcAddressConfigName).(string)
		if !ok {
			panic(fmt.Sprintf("failed to retrieve %s from config.toml", rpcAddressConfigName))
		}
		// Create node status checker to be used by sqs and indexer streaming services.
		nodeStatusChecker = commonservice.NewNodeStatusChecker(rpcAddress)
	}

	streamingServices := []storetypes.ABCIListener{}

	// Initialize the SQS ingester if it is enabled.
	if sqsConfig.IsEnabled {
		sqsKeepers := commondomain.PoolExtractorKeepers{
			GammKeeper:         app.GAMMKeeper,
			CosmWasmPoolKeeper: app.CosmwasmPoolKeeper,
			WasmKeeper:         app.WasmKeeper,
			BankKeeper:         app.BankKeeper,
			ProtorevKeeper:     app.ProtoRevKeeper,
			PoolManagerKeeper:  app.PoolManagerKeeper,
			ConcentratedKeeper: app.ConcentratedLiquidityKeeper,
		}

		// Create pool tracker that tracks pool updates
		// made by the write listenetrs.
		poolTracker := pooltracker.NewMemory()

		// Create pool extractor
		poolExtractor := poolextractor.New(sqsKeepers, poolTracker)

		// Create pools ingester
		poolsTransformer := poolstransformer.NewPoolTransformer(sqsKeepers, sqs.DefaultUSDCUOSMOPool)

		blockProcessStrategyManager := commondomain.NewBlockProcessStrategyManager()

		// Create sqs grpc client
		sqsGRPCClient := sqsservice.NewGRPCCLient(sqsConfig.GRPCIngestAddress, sqsConfig.GRPCIngestMaxCallSizeBytes, appCodec)

		// Create write listeners for the SQS service.
		writeListeners, storeKeyMap := getSQSServiceWriteListeners(app, appCodec, poolTracker, app.WasmKeeper)

		// Create the SQS streaming service by setting up the write listeners,
		// the SQS ingester, and the pool tracker.
		blockUpdatesProcessUtils := &commondomain.BlockUpdateProcessUtils{
			WriteListeners: writeListeners,
			StoreKeyMap:    storeKeyMap,
		}
		sqsStreamingService := sqsservice.New(blockUpdatesProcessUtils, poolExtractor, poolsTransformer, poolTracker, sqsGRPCClient, blockProcessStrategyManager, nodeStatusChecker)

		streamingServices = append(streamingServices, sqsStreamingService)
	}

	// initialize indexer if enabled
	if indexerConfig.IsEnabled {
		indexerPublisher := indexerConfig.Initialize()

		// TODO: handle graceful shutdown
		pubSubCtx := context.Background()

		// Create cold start manager
		blockProcessStrategyManager := commondomain.NewBlockProcessStrategyManager()

		// Create pool tracker that tracks pool updates
		// made by the write listenetrs.
		poolTracker := pooltracker.NewMemory()

		// Create write listeners for the indexer service.
		writeListeners, storeKeyMap := getIndexerServiceWriteListeners(pubSubCtx, app, appCodec, poolTracker, app.WasmKeeper, indexerPublisher, blockProcessStrategyManager)

		// Create keepers for the indexer service.
		keepers := indexerdomain.Keepers{
			BankKeeper:        app.BankKeeper,
			PoolManagerKeeper: app.PoolManagerKeeper,
		}

		poolKeepers := commondomain.PoolExtractorKeepers{
			GammKeeper:         app.GAMMKeeper,
			CosmWasmPoolKeeper: app.CosmwasmPoolKeeper,
			WasmKeeper:         app.WasmKeeper,
			BankKeeper:         app.BankKeeper,
			ProtorevKeeper:     app.ProtoRevKeeper,
			PoolManagerKeeper:  app.PoolManagerKeeper,
			ConcentratedKeeper: app.ConcentratedLiquidityKeeper,
		}

		// Create the indexer streaming service.
		blockUpdatesProcessUtils := &commondomain.BlockUpdateProcessUtils{
			WriteListeners: writeListeners,
			StoreKeyMap:    storeKeyMap,
		}
		poolExtractor := poolextractor.New(poolKeepers, poolTracker)
		indexerStreamingService := indexerservice.New(blockUpdatesProcessUtils, blockProcessStrategyManager, indexerPublisher, storeKeyMap, poolExtractor, poolTracker, keepers, app.GetTxConfig().TxDecoder(), nodeStatusChecker, logger)

		// Register the SQS streaming service with the app.
		streamingServices = append(streamingServices, indexerStreamingService)
	}

	// Register the SQS streaming service with the app.
	app.SetStreamingManager(
		storetypes.StreamingManager{
			ABCIListeners: streamingServices,
			StopNodeOnErr: false,
		},
	)

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
	// Any time a module requires a keeper de-ref'd that's not its native one,
	// its code-smell and should probably change. We should get the staking keeper dependencies fixed.
	app.mm = module.NewManager(appModules(app, encodingConfig, skipGenesisInvariants)...)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	// NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)

	// Upgrades from v0.50.x onwards happen in pre block
	app.mm.SetOrderPreBlockers(upgradetypes.ModuleName)

	// Tell the app's module manager how to set the order of BeginBlockers, which are run at the beginning of every block.
	app.mm.SetOrderBeginBlockers(orderBeginBlockers(app.mm.ModuleNames())...)

	// Tell the app's module manager how to set the order of EndBlockers, which are run at the end of every block.
	app.mm.SetOrderEndBlockers(OrderEndBlockers(app.mm.ModuleNames())...)

	app.mm.SetOrderInitGenesis(OrderInitGenesis(app.mm.ModuleNames())...)

	app.mm.RegisterInvariants(app.CrisisKeeper)

	app.configurator = module.NewConfigurator(app.AppCodec(), app.MsgServiceRouter(), app.GRPCQueryRouter())
	err = app.mm.RegisterServices(app.configurator)
	if err != nil {
		panic(err)
	}

	// Override the gov ModuleBasic with all the custom proposal handers, otherwise we lose them in the CLI.
	app.ModuleBasics = module.NewBasicManagerFromManager(
		app.mm,
		map[string]module.AppModuleBasic{
			"gov": gov.NewAppModuleBasic(
				[]govclient.ProposalHandler{
					paramsclient.ProposalHandler,
					poolincentivesclient.UpdatePoolIncentivesHandler,
					poolincentivesclient.ReplacePoolIncentivesHandler,
					superfluidclient.SetSuperfluidAssetsProposalHandler,
					superfluidclient.RemoveSuperfluidAssetsProposalHandler,
					superfluidclient.UpdateUnpoolWhitelistProposalHandler,
					gammclient.ReplaceMigrationRecordsProposalHandler,
					gammclient.UpdateMigrationRecordsProposalHandler,
					gammclient.CreateCLPoolAndLinkToCFMMProposalHandler,
					gammclient.SetScalingFactorControllerProposalHandler,
					clclient.CreateConcentratedLiquidityPoolProposalHandler,
					clclient.TickSpacingDecreaseProposalHandler,
					cwpoolclient.UploadCodeIdAndWhitelistProposalHandler,
					cwpoolclient.MigratePoolContractsProposalHandler,
					txfeesclient.SubmitUpdateFeeTokenProposalHandler,
					poolmanagerclient.DenomPairTakerFeeProposalHandler,
					incentivesclient.HandleCreateGroupsProposal,
				},
			),
		},
	)

	app.setupUpgradeHandlers()

	app.sm = module.NewSimulationManager(
		auth.NewAppModule(appCodec, *app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		bank.NewAppModule(appCodec, *app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		authzmodule.NewAppModule(appCodec, *app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, *app.MintKeeper, app.AccountKeeper, app.BankKeeper),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		distr.NewAppModule(appCodec, *app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		slashing.NewAppModule(appCodec, *app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName), app.interfaceRegistry),
		params.NewAppModule(*app.ParamsKeeper),
		evidence.NewAppModule(*app.EvidenceKeeper),
		wasm.NewAppModule(appCodec, app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		ibc.NewAppModule(app.IBCKeeper),
		transfer.NewAppModule(*app.TransferKeeper),
	)

	app.sm.RegisterStoreDecoders()

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.mm.Modules))

	reflectionSvc := getReflectionService()
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	app.sm.RegisterStoreDecoders()

	// initialize lanes + mempool
	mevLane, defaultLane := CreateLanes(app, txConfig)

	// create the mempool
	lanedMempool, err := block.NewLanedMempool(
		app.Logger(),
		[]block.Lane{mevLane, defaultLane},
	)
	if err != nil {
		panic(err)
	}
	// set the mempool
	app.SetMempool(lanedMempool)

	// initialize stores
	app.MountKVStores(app.GetKVStoreKey())
	app.MountTransientStores(app.GetTransientStoreKey())
	app.MountMemoryStores(app.GetMemoryStoreKey())

	anteHandler := NewAnteHandler(
		appOpts,
		wasmConfig,
		runtime.NewKVStoreService(app.GetKey(wasmtypes.StoreKey)),
		app.AccountKeeper,
		app.SmartAccountKeeper,
		app.BankKeeper,
		app.TxFeesKeeper,
		app.GAMMKeeper,
		ante.DefaultSigVerificationGasConsumer,
		encodingConfig.TxConfig.SignModeHandler(),
		app.IBCKeeper,
		BlockSDKAnteHandlerParams{
			mevLane:       mevLane,
			auctionKeeper: *app.AppKeepers.AuctionKeeper,
			txConfig:      txConfig,
		},
		appCodec,
	)

	// update ante-handlers on lanes
	opt := []base.LaneOption{
		base.WithAnteHandler(anteHandler),
	}
	mevLane.WithOptions(opt...)
	defaultLane.WithOptions(opt...)

	// ABCI handlers
	// prepare proposal
	proposalHandler := blocksdkabci.NewDefaultProposalHandler(
		app.Logger(),
		txConfig.TxDecoder(),
		txConfig.TxEncoder(),
		lanedMempool,
	)

	// we use the block-sdk's PrepareProposal logic to build blocks
	app.SetPrepareProposal(proposalHandler.PrepareProposalHandler())

	// we use a no-op ProcessProposal, this way, we accept all proposals in avoidance
	// of liveness failures due to Prepare / Process inconsistency. In other words,
	// this ProcessProposal always returns ACCEPT.
	app.SetProcessProposal(baseapp.NoOpProcessProposal())

	cacheDecoder, err := utils.NewDefaultCacheTxDecoder(txConfig.TxDecoder())
	if err != nil {
		panic(err)
	}

	// check-tx
	mevCheckTxHandler := checktx.NewMEVCheckTxHandler(
		app,
		cacheDecoder.TxDecoder(),
		mevLane,
		anteHandler,
		app.BaseApp.CheckTx,
	)

	// wrap checkTxHandler with mempool parity handler
	parityCheckTx := checktx.NewMempoolParityCheckTx(
		app.Logger(),
		lanedMempool,
		cacheDecoder.TxDecoder(),
		mevCheckTxHandler.CheckTx(),
		app,
	)

	app.SetCheckTx(parityCheckTx.CheckTx())

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(anteHandler)
	app.SetPostHandler(NewPostHandler(appCodec, app.ProtoRevKeeper, app.SmartAccountKeeper, app.AccountKeeper, encodingConfig.TxConfig.SignModeHandler()))
	app.SetEndBlocker(app.EndBlocker)
	app.SetPrecommiter(app.Precommitter)
	app.SetPrepareCheckStater(app.PrepareCheckStater)

	// Register snapshot extensions to enable state-sync for wasm.
	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), app.WasmKeeper),
		)
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}

		err = manager.RegisterExtensions(
			ibcwasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), app.IBCWasmClientKeeper),
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

		if err := ibcwasmkeeper.InitializePinnedCodes(ctx); err != nil {
			tmos.Exit(fmt.Sprintf("failed initialize pinned codes %s", err))
		}
	}

	return app
}

// getSQSServiceWriteListeners returns the write listeners for the app that are specific to the SQS service.
func getSQSServiceWriteListeners(app *OsmosisApp, appCodec codec.Codec, blockPoolUpdateTracker domain.BlockPoolUpdateTracker, wasmkeeper *wasmkeeper.Keeper) (map[storetypes.StoreKey][]commondomain.WriteListener, map[string]storetypes.StoreKey) {
	writeListeners, storeKeyMap := getPoolWriteListeners(app, appCodec, blockPoolUpdateTracker, wasmkeeper)

	// Register all applicable keys as listeners
	registerStoreKeys(app, storeKeyMap)

	return writeListeners, storeKeyMap
}

// getIndexerServiceWriteListeners returns the write listeners for the app that are specific to the indexer service.
func getIndexerServiceWriteListeners(ctx context.Context, app *OsmosisApp, appCodec codec.Codec, blockPoolUpdateTracker domain.BlockPoolUpdateTracker, wasmkeeper *wasmkeeper.Keeper, client indexerdomain.Publisher, blockProcessStrategyManager commondomain.BlockProcessStrategyManager) (map[storetypes.StoreKey][]commondomain.WriteListener, map[string]storetypes.StoreKey) {
	writeListeners, storeKeyMap := getPoolWriteListeners(app, appCodec, blockPoolUpdateTracker, wasmkeeper)

	// Add write listeners for the bank module.
	writeListeners[app.GetKey(banktypes.ModuleName)] = []commondomain.WriteListener{
		indexerwritelistener.NewBank(ctx, client, blockProcessStrategyManager),
	}

	storeKeyMap[banktypes.ModuleName] = app.GetKey(banktypes.ModuleName)

	// Register all applicable keys as listeners
	registerStoreKeys(app, storeKeyMap)

	return writeListeners, storeKeyMap
}

// getPoolWriteListeners returns the write listeners for the app that are specific to monitoring the pools.
func getPoolWriteListeners(app *OsmosisApp, appCodec codec.Codec, blockPoolUpdateTracker domain.BlockPoolUpdateTracker, wasmkeeper *wasmkeeper.Keeper) (map[storetypes.StoreKey][]commondomain.WriteListener, map[string]storetypes.StoreKey) {
	writeListeners := make(map[storetypes.StoreKey][]commondomain.WriteListener)
	storeKeyMap := make(map[string]storetypes.StoreKey)

	writeListeners[app.GetKey(concentratedtypes.ModuleName)] = []commondomain.WriteListener{
		writelistener.NewConcentrated(blockPoolUpdateTracker),
	}
	writeListeners[app.GetKey(gammtypes.StoreKey)] = []commondomain.WriteListener{
		writelistener.NewGAMM(blockPoolUpdateTracker, appCodec),
	}
	writeListeners[app.GetKey(cosmwasmpooltypes.StoreKey)] = []commondomain.WriteListener{
		writelistener.NewCosmwasmPool(blockPoolUpdateTracker, wasmkeeper),
	}
	writeListeners[app.GetKey(banktypes.StoreKey)] = []commondomain.WriteListener{
		writelistener.NewCosmwasmPoolBalance(blockPoolUpdateTracker),
	}

	storeKeyMap[concentratedtypes.ModuleName] = app.GetKey(concentratedtypes.ModuleName)
	storeKeyMap[gammtypes.StoreKey] = app.GetKey(gammtypes.StoreKey)
	storeKeyMap[cosmwasmpooltypes.StoreKey] = app.GetKey(cosmwasmpooltypes.StoreKey)
	storeKeyMap[banktypes.StoreKey] = app.GetKey(banktypes.StoreKey)

	return writeListeners, storeKeyMap
}

// registerStoreKeys register the store keys from the given store key map
// on the app's commit multi store so that the change sets from these stores are propagated
// in ListenCommit().
func registerStoreKeys(app *OsmosisApp, storeKeyMap map[string]storetypes.StoreKey) {
	// Register all applicable keys as listeners
	storeKeys := make([]storetypes.StoreKey, 0)
	for _, storeKey := range storeKeyMap {
		storeKeys = append(storeKeys, storeKey)
	}
	app.CommitMultiStore().AddListeners(storeKeys)
}

// we cache the reflectionService to save us time within tests.
var cachedReflectionService *runtimeservices.ReflectionService = nil

func getReflectionService() *runtimeservices.ReflectionService {
	if cachedReflectionService != nil {
		return cachedReflectionService
	}
	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	cachedReflectionService = reflectionSvc
	return reflectionSvc
}

// InitOsmosisAppForTestnet is broken down into two sections:
// Required Changes: Changes that, if not made, will cause the testnet to halt or panic
// Optional Changes: Changes to customize the testnet to one's liking (lower vote times, fund accounts, etc)
func InitOsmosisAppForTestnet(app *OsmosisApp, newValAddr bytes.HexBytes, newValPubKey crypto.PubKey, newOperatorAddress, upgradeToTrigger string) *OsmosisApp {
	//
	// Required Changes:
	//

	ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})
	pubkey := &ed25519.PubKey{Key: newValPubKey.Bytes()}
	pubkeyAny, err := types.NewAnyWithValue(pubkey)
	if err != nil {
		tmos.Exit(err.Error())
	}

	// STAKING
	//

	// Create Validator struct for our new validator.
	_, bz, err := bech32.DecodeAndConvert(newOperatorAddress)
	if err != nil {
		tmos.Exit(err.Error())
	}
	bech32Addr, err := bech32.ConvertAndEncode("osmovaloper", bz)
	if err != nil {
		tmos.Exit(err.Error())
	}
	newVal := stakingtypes.Validator{
		OperatorAddress: bech32Addr,
		ConsensusPubkey: pubkeyAny,
		Jailed:          false,
		Status:          stakingtypes.Bonded,
		Tokens:          osmomath.NewInt(900000000000000),
		DelegatorShares: osmomath.MustNewDecFromStr("10000000"),
		Description: stakingtypes.Description{
			Moniker: "Testnet Validator",
		},
		Commission: stakingtypes.Commission{
			CommissionRates: stakingtypes.CommissionRates{
				Rate:          osmomath.MustNewDecFromStr("0.05"),
				MaxRate:       osmomath.MustNewDecFromStr("0.1"),
				MaxChangeRate: osmomath.MustNewDecFromStr("0.05"),
			},
		},
		MinSelfDelegation: osmomath.OneInt(),
	}

	// Remove all validators from power store
	stakingKey := app.GetKey(stakingtypes.ModuleName)
	stakingStore := ctx.KVStore(stakingKey)
	iterator, err := app.StakingKeeper.ValidatorsPowerStoreIterator(ctx)
	if err != nil {
		tmos.Exit(err.Error())
	}
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Remove all valdiators from last validators store
	iterator, err = app.StakingKeeper.LastValidatorsIterator(ctx)
	if err != nil {
		tmos.Exit(err.Error())
	}
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Remove all validators from validators store
	iterator = storetypes.KVStorePrefixIterator(stakingStore, stakingtypes.ValidatorsKey)
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Remove all validators from unbonding queue
	iterator = storetypes.KVStorePrefixIterator(stakingStore, stakingtypes.ValidatorQueueKey)
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Add our validator to power and last validators store
	err = app.StakingKeeper.SetValidator(ctx, newVal)
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.StakingKeeper.SetValidatorByConsAddr(ctx, newVal)
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.StakingKeeper.SetValidatorByPowerIndex(ctx, newVal)
	if err != nil {
		tmos.Exit(err.Error())
	}
	valAddr, err := sdk.ValAddressFromBech32(newVal.GetOperator())
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.StakingKeeper.SetLastValidatorPower(ctx, valAddr, 0)
	if err != nil {
		tmos.Exit(err.Error())
	}
	if err := app.StakingKeeper.Hooks().AfterValidatorCreated(ctx, valAddr); err != nil {
		panic(err)
	}

	// DISTRIBUTION
	//

	// Initialize records for this validator across all distribution stores
	valAddr, err = sdk.ValAddressFromBech32(newVal.GetOperator())
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.DistrKeeper.SetValidatorHistoricalRewards(ctx, valAddr, 0, distrtypes.NewValidatorHistoricalRewards(sdk.DecCoins{}, 1))
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.DistrKeeper.SetValidatorCurrentRewards(ctx, valAddr, distrtypes.NewValidatorCurrentRewards(sdk.DecCoins{}, 1))
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.DistrKeeper.SetValidatorAccumulatedCommission(ctx, valAddr, distrtypes.InitialValidatorAccumulatedCommission())
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.DistrKeeper.SetValidatorOutstandingRewards(ctx, valAddr, distrtypes.ValidatorOutstandingRewards{Rewards: sdk.DecCoins{}})
	if err != nil {
		tmos.Exit(err.Error())
	}

	// SLASHING
	//

	// Set validator signing info for our new validator.
	newConsAddr := sdk.ConsAddress(newValAddr.Bytes())
	newValidatorSigningInfo := slashingtypes.ValidatorSigningInfo{
		Address:     newConsAddr.String(),
		StartHeight: app.LastBlockHeight() - 1,
		Tombstoned:  false,
	}
	err = app.SlashingKeeper.SetValidatorSigningInfo(ctx, newConsAddr, newValidatorSigningInfo)
	if err != nil {
		tmos.Exit(err.Error())
	}

	//
	// Optional Changes:
	//

	// GOV
	//

	newExpeditedVotingPeriod := time.Minute
	newVotingPeriod := time.Minute * 2

	govParams, err := app.GovKeeper.Params.Get(ctx)
	if err != nil {
		tmos.Exit(err.Error())
	}
	govParams.ExpeditedVotingPeriod = &newExpeditedVotingPeriod
	govParams.VotingPeriod = &newVotingPeriod
	govParams.MinDeposit = sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseCoinUnit, 100000000))
	govParams.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseCoinUnit, 150000000))

	err = app.GovKeeper.Params.Set(ctx, govParams)
	if err != nil {
		tmos.Exit(err.Error())
	}

	// EPOCHS
	//

	dayEpochInfo := app.EpochsKeeper.GetEpochInfo(ctx, "day")
	dayEpochInfo.Duration = time.Hour * 6
	// Prevents epochs from running back to back
	dayEpochInfo.CurrentEpochStartTime = time.Now().UTC()
	// If you want epoch to run a minute after starting the chain, uncomment the line below and comment the line above
	// dayEpochInfo.CurrentEpochStartTime = time.Now().UTC().Add(-dayEpochInfo.Duration).Add(time.Minute)
	dayEpochInfo.CurrentEpochStartHeight = app.LastBlockHeight()
	app.EpochsKeeper.DeleteEpochInfo(ctx, "day")
	err = app.EpochsKeeper.AddEpochInfo(ctx, dayEpochInfo)
	if err != nil {
		tmos.Exit(err.Error())
	}

	weekEpochInfo := app.EpochsKeeper.GetEpochInfo(ctx, "week")
	weekEpochInfo.Duration = time.Hour * 12
	// Prevents epochs from running back to back
	weekEpochInfo.CurrentEpochStartTime = time.Now().UTC()
	weekEpochInfo.CurrentEpochStartHeight = app.LastBlockHeight()
	app.EpochsKeeper.DeleteEpochInfo(ctx, "week")
	err = app.EpochsKeeper.AddEpochInfo(ctx, weekEpochInfo)
	if err != nil {
		tmos.Exit(err.Error())
	}

	// BANK
	//

	defaultCoins := sdk.NewCoins(
		sdk.NewInt64Coin("ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", 1000000000000), // DAI
		sdk.NewInt64Coin("ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4", 1000000000000), // USDC, for pool creation fee
		sdk.NewInt64Coin(appparams.BaseCoinUnit, 1000000000000),
		sdk.NewInt64Coin("uion", 1000000000))

	localOsmosisAccounts := []sdk.AccAddress{
		sdk.MustAccAddressFromBech32("osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj"),
		sdk.MustAccAddressFromBech32("osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks"),
		sdk.MustAccAddressFromBech32("osmo18s5lynnmx37hq4wlrw9gdn68sg2uxp5rgk26vv"),
		sdk.MustAccAddressFromBech32("osmo1qwexv7c6sm95lwhzn9027vyu2ccneaqad4w8ka"),
		sdk.MustAccAddressFromBech32("osmo14hcxlnwlqtq75ttaxf674vk6mafspg8xwgnn53"),
		sdk.MustAccAddressFromBech32("osmo12rr534cer5c0vj53eq4y32lcwguyy7nndt0u2t"),
		sdk.MustAccAddressFromBech32("osmo1nt33cjd5auzh36syym6azgc8tve0jlvklnq7jq"),
		sdk.MustAccAddressFromBech32("osmo10qfrpash5g2vk3hppvu45x0g860czur8ff5yx0"),
		sdk.MustAccAddressFromBech32("osmo1f4tvsdukfwh6s9swrc24gkuz23tp8pd3e9r5fa"),
		sdk.MustAccAddressFromBech32("osmo1myv43sqgnj5sm4zl98ftl45af9cfzk7nhjxjqh"),
		sdk.MustAccAddressFromBech32("osmo14gs9zqh8m49yy9kscjqu9h72exyf295afg6kgk"),
		sdk.MustAccAddressFromBech32("osmo1jllfytsz4dryxhz5tl7u73v29exsf80vz52ucc"),
	}

	// Fund localosmosis accounts
	for _, account := range localOsmosisAccounts {
		err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, defaultCoins)
		if err != nil {
			tmos.Exit(err.Error())
		}
		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, account, defaultCoins)
		if err != nil {
			tmos.Exit(err.Error())
		}
	}

	// Fund edgenet faucet
	faucetCoins := sdk.NewCoins(
		sdk.NewInt64Coin("ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", 1000000000000000), // DAI
		sdk.NewInt64Coin(appparams.BaseCoinUnit, 1000000000000000),
		sdk.NewInt64Coin("uion", 1000000000000))
	err = app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, faucetCoins)
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, sdk.MustAccAddressFromBech32("osmo1rqgf207csps822qwmd3k2n6k6k4e99w502e79t"), faucetCoins)
	if err != nil {
		tmos.Exit(err.Error())
	}

	// Mars bank account
	marsCoins := sdk.NewCoins(
		sdk.NewInt64Coin(appparams.BaseCoinUnit, 10000000000000),
		sdk.NewInt64Coin("ibc/903A61A498756EA560B85A85132D3AEE21B5DEDD41213725D22ABF276EA6945E", 400000000000),
		sdk.NewInt64Coin("ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", 3000000000000),
		sdk.NewInt64Coin("ibc/C140AFD542AE77BD7DCC83F13FDD8C5E5BB8C4929785E6EC2F4C636F98F17901", 200000000000),
		sdk.NewInt64Coin("ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", 700000000000),
		sdk.NewInt64Coin("ibc/D1542AA8762DB13087D8364F3EA6509FD6F009A34F00426AF9E4F9FA85CBBF1F", 2000000000),
		sdk.NewInt64Coin("ibc/EA1D43981D5C9A1C4AAEA9C23BB1D4FA126BA9BC7020A25E0AE4AA841EA25DC5", 3000000000000000000))
	err = app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, marsCoins)
	if err != nil {
		tmos.Exit(err.Error())
	}
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, sdk.MustAccAddressFromBech32("osmo1ev02crc36675xd8s029qh7wg3wjtfk37jr004z"), marsCoins)
	if err != nil {
		tmos.Exit(err.Error())
	}

	// UPGRADE
	//

	if upgradeToTrigger != "" {
		upgradePlan := upgradetypes.Plan{
			Name:   upgradeToTrigger,
			Height: app.LastBlockHeight() + 10,
		}
		err = app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan)
		if err != nil {
			panic(err)
		}
	}

	return app
}

// CheckTx will check the transaction with the provided checkTxHandler. We override the default
// handler so that we can verify bid transactions before they are inserted into the mempool.
// With the BlockSDK CheckTx, we can verify the bid transaction and all of the bundled transactions
// before inserting the bid transaction into the mempool.
func (app *OsmosisApp) CheckTx(req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	return app.checkTxHandler(req)
}

// SetCheckTx sets the checkTxHandler for the app.
func (app *OsmosisApp) SetCheckTx(handler checktx.CheckTx) {
	app.checkTxHandler = handler
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

// PreBlocker application updates before each begin block.
func (app *OsmosisApp) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	// Set gas meter to the free gas meter.
	// This is because there is currently non-deterministic gas usage in the
	// pre-blocker, e.g. due to hydration of in-memory data structures.
	//
	// Note that we don't need to reset the gas meter after the pre-blocker
	// because Go is pass by value.
	ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
	mm := app.ModuleManager()
	return mm.PreBlock(ctx)
}

// BeginBlocker application updates every begin block.
func (app *OsmosisApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	BeginBlockForks(ctx, app)
	return app.mm.BeginBlock(ctx)
}

// EndBlocker application updates every end block.
func (app *OsmosisApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.mm.EndBlock(ctx)
}

// Precommitter application updates before the commital of a block after all transactions have been delivered.
func (app *OsmosisApp) Precommitter(ctx sdk.Context) {
	mm := app.ModuleManager()
	if err := mm.Precommit(ctx); err != nil {
		panic(err)
	}
}

func (app *OsmosisApp) PrepareCheckStater(ctx sdk.Context) {
	mm := app.ModuleManager()
	if err := mm.PrepareCheckState(ctx); err != nil {
		panic(err)
	}
}

// InitChainer application update at chain initialization.
func (app *OsmosisApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	if err != nil {
		panic(err)
	}

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
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	module.NewBasicManagerFromManager(app.mm, nil).RegisterGRPCGatewayRoutes(
		clientCtx,
		apiSvr.GRPCGatewayRouter,
	)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

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
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

// RegisterNodeService registers the node gRPC Query service.
func (app *OsmosisApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// SimulationManager implements the SimulationApp interface
func (app *OsmosisApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// ChainID gets chainID from private fields of BaseApp
// Should be removed once SDK 0.50.x will be adopted
func (app *OsmosisApp) ChainID() string {
	field := reflect.ValueOf(app.BaseApp).Elem().FieldByName("chainID")
	return field.String()
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

func (app *OsmosisApp) customPreUpgradeHandler(upgradeInfo upgradetypes.Plan) {
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

// initOTELTracer initializes the OTEL tracer
// and wires it up with the Sentry exporter.
func initOTELTracer(ctx context.Context, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}
