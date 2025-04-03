package keepers

import (
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icq "github.com/cosmos/ibc-apps/modules/async-icq/v8"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v8/types"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icacontroller "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	custombankkeeper "github.com/osmosis-labs/osmosis/v27/custom/bank/keeper"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	customwasmkeeper "github.com/osmosis-labs/osmosis/v27/custom/wasm/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
	downtimedetector "github.com/osmosis-labs/osmosis/v27/x/downtime-detector"
	downtimetypes "github.com/osmosis-labs/osmosis/v27/x/downtime-detector/types"
	"github.com/osmosis-labs/osmosis/v27/x/gamm"
	ibcratelimit "github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit"
	ibcratelimittypes "github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/types"
	marketkeeper "github.com/osmosis-labs/osmosis/v27/x/market/keeper"
	markettypes "github.com/osmosis-labs/osmosis/v27/x/market/types"
	oraclekeeper "github.com/osmosis-labs/osmosis/v27/x/oracle/keeper"
	oracletypes "github.com/osmosis-labs/osmosis/v27/x/oracle/types"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v27/x/protorev"
	stablestakingincentviceskeeper "github.com/osmosis-labs/osmosis/v27/x/stable-staking-incentives/keeper"
	stablestakingincentvicestypes "github.com/osmosis-labs/osmosis/v27/x/stable-staking-incentives/types"
	treasurykeeper "github.com/osmosis-labs/osmosis/v27/x/treasury/keeper"
	treasurytypes "github.com/osmosis-labs/osmosis/v27/x/treasury/types"
	ibchooks "github.com/osmosis-labs/osmosis/x/ibc-hooks"
	ibchookskeeper "github.com/osmosis-labs/osmosis/x/ibc-hooks/keeper"
	ibchookstypes "github.com/osmosis-labs/osmosis/x/ibc-hooks/types"

	icqkeeper "github.com/cosmos/ibc-apps/modules/async-icq/v8/keeper"
	ibcwasmkeeper "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/keeper"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	icahost "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	packetforward "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"

	// IBC Transfer: Defines the "transfer" IBC port
	transfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"

	"github.com/osmosis-labs/osmosis/v27/x/smart-account/authenticator"
	smartaccountkeeper "github.com/osmosis-labs/osmosis/v27/x/smart-account/keeper"
	smartaccounttypes "github.com/osmosis-labs/osmosis/v27/x/smart-account/types"

	_ "github.com/osmosis-labs/osmosis/v27/client/docs/statik"
	owasm "github.com/osmosis-labs/osmosis/v27/wasmbinding"
	concentratedliquidity "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	concentratedliquiditytypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	epochskeeper "github.com/osmosis-labs/osmosis/v27/x/epochs/keeper"
	epochstypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	gammkeeper "github.com/osmosis-labs/osmosis/v27/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	incentiveskeeper "github.com/osmosis-labs/osmosis/v27/x/incentives/keeper"
	incentivestypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockupkeeper "github.com/osmosis-labs/osmosis/v27/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	mintkeeper "github.com/osmosis-labs/osmosis/v27/x/mint/keeper"
	minttypes "github.com/osmosis-labs/osmosis/v27/x/mint/types"
	poolincentives "github.com/osmosis-labs/osmosis/v27/x/pool-incentives"
	poolincentiveskeeper "github.com/osmosis-labs/osmosis/v27/x/pool-incentives/keeper"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
	protorevkeeper "github.com/osmosis-labs/osmosis/v27/x/protorev/keeper"
	protorevtypes "github.com/osmosis-labs/osmosis/v27/x/protorev/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid"
	superfluidkeeper "github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper"
	superfluidtypes "github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v27/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"
	"github.com/osmosis-labs/osmosis/v27/x/twap"
	twaptypes "github.com/osmosis-labs/osmosis/v27/x/twap/types"
	txfeeskeeper "github.com/osmosis-labs/osmosis/v27/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/v27/x/txfees/types"
	valsetpref "github.com/osmosis-labs/osmosis/v27/x/valset-pref"
	valsetpreftypes "github.com/osmosis-labs/osmosis/v27/x/valset-pref/types"

	auctionkeeper "github.com/skip-mev/block-sdk/v2/x/auction/keeper"
	auctiontypes "github.com/skip-mev/block-sdk/v2/x/auction/types"

	storetypes "cosmossdk.io/store/types"
)

const (
	AccountAddressPrefix = "melody"
)

type AppKeepers struct {
	// keepers, by order of initialization
	// "Special" keepers
	ParamsKeeper          *paramskeeper.Keeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ConsensusParamsKeeper *consensusparamkeeper.Keeper
	BankKeeper            *custombankkeeper.CustomKeeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper
	ScopedWasmKeeper          capabilitykeeper.ScopedKeeper
	ScopedICQKeeper           capabilitykeeper.ScopedKeeper

	// "Normal" keepers
	AccountKeeper                 *authkeeper.AccountKeeper
	AuthzKeeper                   *authzkeeper.Keeper
	StakingKeeper                 *stakingkeeper.Keeper
	DistrKeeper                   *distrkeeper.Keeper
	DowntimeKeeper                *downtimedetector.Keeper
	SlashingKeeper                *slashingkeeper.Keeper
	IBCKeeper                     *ibckeeper.Keeper
	IBCHooksKeeper                *ibchookskeeper.Keeper
	ICAHostKeeper                 *icahostkeeper.Keeper
	ICAControllerKeeper           *icacontrollerkeeper.Keeper
	ICQKeeper                     *icqkeeper.Keeper
	TransferKeeper                *ibctransferkeeper.Keeper
	IBCWasmClientKeeper           *ibcwasmkeeper.Keeper
	EvidenceKeeper                *evidencekeeper.Keeper
	GAMMKeeper                    *gammkeeper.Keeper
	TwapKeeper                    *twap.Keeper
	LockupKeeper                  *lockupkeeper.Keeper
	EpochsKeeper                  *epochskeeper.Keeper
	IncentivesKeeper              *incentiveskeeper.Keeper
	ProtoRevKeeper                *protorevkeeper.Keeper
	MintKeeper                    *mintkeeper.Keeper
	PoolIncentivesKeeper          *poolincentiveskeeper.Keeper
	TxFeesKeeper                  *txfeeskeeper.Keeper
	SuperfluidKeeper              *superfluidkeeper.Keeper
	GovKeeper                     *govkeeper.Keeper
	WasmKeeper                    *wasmkeeper.Keeper
	ContractKeeper                *wasmkeeper.PermissionedKeeper
	TokenFactoryKeeper            *tokenfactorykeeper.Keeper
	PoolManagerKeeper             *poolmanager.Keeper
	OracleKeeper                  *oraclekeeper.Keeper
	MarketKeeper                  *marketkeeper.Keeper
	TreasuryKeeper                *treasurykeeper.Keeper
	StableStakingIncentivesKeeper *stablestakingincentviceskeeper.Keeper
	ValidatorSetPreferenceKeeper  *valsetpref.Keeper
	ConcentratedLiquidityKeeper   *concentratedliquidity.Keeper
	CosmwasmPoolKeeper            *cosmwasmpool.Keeper
	SmartAccountKeeper            *smartaccountkeeper.Keeper
	AuthenticatorManager          *authenticator.AuthenticatorManager

	// IBC modules
	// transfer module
	RawIcs20TransferAppModule transfer.AppModule
	RateLimitingICS4Wrapper   *ibcratelimit.ICS4Wrapper
	TransferStack             *ibchooks.IBCMiddleware
	Ics20WasmHooks            *ibchooks.WasmHooks
	HooksICS4Wrapper          ibchooks.ICS4Middleware
	PacketForwardKeeper       *packetforwardkeeper.Keeper

	// BlockSDK
	AuctionKeeper *auctionkeeper.Keeper

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey
}

// InitNormalKeepers initializes all 'normal' keepers (account, app, bank, auth, staking, distribution, slashing, transfer, gamm, IBC router, pool incentives, governance, mint, txfees keepers).
func (appKeepers *AppKeepers) InitNormalKeepers(
	appCodec codec.Codec,
	encodingConfig appparams.EncodingConfig,
	bApp *baseapp.BaseApp,
	maccPerms map[string][]string,
	dataDir string,
	wasmDir string,
	wasmConfig wasmtypes.WasmConfig,
	wasmOpts []wasmkeeper.Option,
	blockedAddress map[string]bool,
	ibcWasmConfig ibcwasmtypes.WasmConfig,
) {
	legacyAmino := encodingConfig.Amino
	// Add 'normal' keepers
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		AccountAddressPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.AccountKeeper = &accountKeeper
	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[banktypes.StoreKey]),
		appKeepers.AccountKeeper,
		blockedAddress,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		bApp.Logger(),
	)
	customBankKeeper := custombankkeeper.NewCustomKeeper(&bankKeeper, appKeepers.AccountKeeper)
	appKeepers.BankKeeper = &customBankKeeper

	// Initialize authenticators
	appKeepers.AuthenticatorManager = authenticator.NewAuthenticatorManager()
	appKeepers.AuthenticatorManager.InitializeAuthenticators([]authenticator.Authenticator{
		authenticator.NewSignatureVerification(appKeepers.AccountKeeper),
		authenticator.NewMessageFilter(encodingConfig),
		authenticator.NewAllOf(appKeepers.AuthenticatorManager),
		authenticator.NewAnyOf(appKeepers.AuthenticatorManager),
		authenticator.NewPartitionedAnyOf(appKeepers.AuthenticatorManager),
		authenticator.NewPartitionedAllOf(appKeepers.AuthenticatorManager),
	})
	govModuleAddr := appKeepers.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	smartAccountKeeper := smartaccountkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[smartaccounttypes.StoreKey],
		govModuleAddr,
		appKeepers.GetSubspace(smartaccounttypes.ModuleName),
		appKeepers.AuthenticatorManager,
	)
	appKeepers.SmartAccountKeeper = &smartAccountKeeper

	authzKeeper := authzkeeper.NewKeeper(
		runtime.NewKVStoreService(appKeepers.keys[authzkeeper.StoreKey]),
		appCodec,
		bApp.MsgServiceRouter(),
		appKeepers.AccountKeeper,
	)
	appKeepers.AuthzKeeper = &authzKeeper

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[stakingtypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)
	appKeepers.StakingKeeper = stakingKeeper

	distrKeeper := distrkeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(appKeepers.keys[distrtypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.DistrKeeper = &distrKeeper

	appKeepers.DowntimeKeeper = downtimedetector.NewKeeper(
		appKeepers.keys[downtimetypes.StoreKey],
	)

	slashingKeeper := slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(appKeepers.keys[slashingtypes.StoreKey]),
		appKeepers.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.SlashingKeeper = &slashingKeeper

	// Create IBC Keeper
	appKeepers.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibchost.StoreKey],
		appKeepers.GetSubspace(ibchost.ModuleName),
		appKeepers.StakingKeeper,
		appKeepers.UpgradeKeeper,
		appKeepers.ScopedIBCKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Configure the hooks keeper
	hooksKeeper := ibchookskeeper.NewKeeper(
		appKeepers.keys[ibchookstypes.StoreKey],
		appKeepers.GetSubspace(ibchookstypes.ModuleName),
		appKeepers.IBCKeeper.ChannelKeeper,
		nil,
	)
	appKeepers.IBCHooksKeeper = hooksKeeper

	// We are using a separate VM here
	ibcWasmClientKeeper := ibcwasmkeeper.NewKeeperWithConfig(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[ibcwasmtypes.StoreKey]),
		appKeepers.IBCKeeper.ClientKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		ibcWasmConfig,
		bApp.GRPCQueryRouter(),
	)

	appKeepers.IBCWasmClientKeeper = &ibcWasmClientKeeper

	appKeepers.WireICS20PreWasmKeeper(appCodec, bApp, appKeepers.IBCHooksKeeper)

	icaHostKeeper := icahostkeeper.NewKeeper(
		appCodec, appKeepers.keys[icahosttypes.StoreKey],
		appKeepers.GetSubspace(icahosttypes.SubModuleName),
		appKeepers.RateLimitingICS4Wrapper,
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		appKeepers.ScopedICAHostKeeper,
		bApp.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	icaHostKeeper.WithQueryRouter(bApp.GRPCQueryRouter())
	appKeepers.ICAHostKeeper = &icaHostKeeper

	icaControllerKeeper := icacontrollerkeeper.NewKeeper(
		appCodec, appKeepers.keys[icacontrollertypes.StoreKey],
		appKeepers.GetSubspace(icacontrollertypes.SubModuleName),
		appKeepers.RateLimitingICS4Wrapper,
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.ScopedICAControllerKeeper,
		bApp.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.ICAControllerKeeper = &icaControllerKeeper

	// initialize ICA module with mock module as the authentication module on the controller side
	var icaControllerStack porttypes.IBCModule
	icaControllerStack = icacontroller.NewIBCMiddleware(icaControllerStack, *appKeepers.ICAControllerKeeper)

	// RecvPacket, message that originates from core IBC and goes down to app, the flow is:
	// channel.RecvPacket -> fee.OnRecvPacket -> icaHost.OnRecvPacket
	icaHostStack := icahost.NewIBCModule(*appKeepers.ICAHostKeeper)

	// ICQ Keeper
	icqKeeper := icqkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[icqtypes.StoreKey],
		appKeepers.IBCKeeper.ChannelKeeper, // may be replaced with middleware
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.ScopedICQKeeper,
		bApp.GRPCQueryRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.ICQKeeper = &icqKeeper

	// Create Async ICQ module
	icqModule := icq.NewIBCModule(*appKeepers.ICQKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(icahosttypes.SubModuleName, icaHostStack).
		// The transferIBC module is replaced by rateLimitingTransferModule
		AddRoute(ibctransfertypes.ModuleName, appKeepers.TransferStack).
		// Add icq modules to IBC router
		AddRoute(icqtypes.ModuleName, icqModule)
	// Note: the sealing is done after creating wasmd and wiring that up

	// create evidence keeper with router
	// If evidence needs to be handled for the app, set routes in router here and seal
	appKeepers.EvidenceKeeper = evidencekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[evidencetypes.StoreKey]),
		appKeepers.StakingKeeper,
		appKeepers.SlashingKeeper,
		addresscodec.NewBech32Codec(sdk.Bech32PrefixAccAddr),
		runtime.ProvideCometInfoService(),
	)

	appKeepers.LockupKeeper = lockupkeeper.NewKeeper(
		appKeepers.keys[lockuptypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.DistrKeeper, appKeepers.GetSubspace(lockuptypes.ModuleName))

	appKeepers.ConcentratedLiquidityKeeper = concentratedliquidity.NewKeeper(
		appCodec,
		appKeepers.keys[concentratedliquiditytypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.GAMMKeeper,
		appKeepers.PoolIncentivesKeeper,
		appKeepers.IncentivesKeeper,
		appKeepers.LockupKeeper,
		appKeepers.DistrKeeper,
		appKeepers.ContractKeeper,
		appKeepers.GetSubspace(concentratedliquiditytypes.ModuleName),
	)

	gammKeeper := gammkeeper.NewKeeper(
		appCodec, appKeepers.keys[gammtypes.StoreKey],
		appKeepers.GetSubspace(gammtypes.ModuleName),
		appKeepers.AccountKeeper,
		// TODO: Add a mintcoins restriction
		appKeepers.BankKeeper, appKeepers.DistrKeeper,
		appKeepers.ConcentratedLiquidityKeeper,
		appKeepers.PoolIncentivesKeeper,
		appKeepers.IncentivesKeeper)
	appKeepers.GAMMKeeper = &gammKeeper
	appKeepers.ConcentratedLiquidityKeeper.SetGammKeeper(appKeepers.GAMMKeeper)

	appKeepers.CosmwasmPoolKeeper = cosmwasmpool.NewKeeper(appCodec, appKeepers.keys[cosmwasmpooltypes.StoreKey], appKeepers.GetSubspace(cosmwasmpooltypes.ModuleName), appKeepers.AccountKeeper, appKeepers.BankKeeper)

	appKeepers.PoolManagerKeeper = poolmanager.NewKeeper(
		appKeepers.keys[poolmanagertypes.StoreKey],
		appKeepers.GetSubspace(poolmanagertypes.ModuleName),
		appKeepers.GAMMKeeper,
		appKeepers.ConcentratedLiquidityKeeper,
		appKeepers.CosmwasmPoolKeeper,
		appKeepers.BankKeeper,
		appKeepers.AccountKeeper,
		appKeepers.DistrKeeper,
		appKeepers.StakingKeeper,
		appKeepers.ProtoRevKeeper,
		appKeepers.WasmKeeper,
	)
	appKeepers.PoolManagerKeeper.SetStakingKeeper(appKeepers.StakingKeeper)
	appKeepers.GAMMKeeper.SetPoolManager(appKeepers.PoolManagerKeeper)
	appKeepers.ConcentratedLiquidityKeeper.SetPoolManagerKeeper(appKeepers.PoolManagerKeeper)
	appKeepers.CosmwasmPoolKeeper.SetPoolManagerKeeper(appKeepers.PoolManagerKeeper)

	appKeepers.TwapKeeper = twap.NewKeeper(
		appKeepers.keys[twaptypes.StoreKey],
		appKeepers.tkeys[twaptypes.TransientStoreKey],
		appKeepers.GetSubspace(twaptypes.ModuleName),
		appKeepers.PoolManagerKeeper)

	appKeepers.EpochsKeeper = epochskeeper.NewKeeper(appKeepers.keys[epochstypes.StoreKey])

	protorevKeeper := protorevkeeper.NewKeeper(
		appCodec, appKeepers.keys[protorevtypes.StoreKey],
		appKeepers.tkeys[protorevtypes.TransientStoreKey],
		appKeepers.GetSubspace(protorevtypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.GAMMKeeper,
		appKeepers.EpochsKeeper,
		appKeepers.PoolManagerKeeper,
		appKeepers.ConcentratedLiquidityKeeper,
		appKeepers.DistrKeeper,
	)
	appKeepers.ProtoRevKeeper = &protorevKeeper
	appKeepers.PoolManagerKeeper.SetProtorevKeeper(appKeepers.ProtoRevKeeper)

	appKeepers.IncentivesKeeper = incentiveskeeper.NewKeeper(
		appKeepers.keys[incentivestypes.StoreKey],
		appKeepers.GetSubspace(incentivestypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.LockupKeeper,
		appKeepers.EpochsKeeper,
		appKeepers.DistrKeeper,
		appKeepers.TxFeesKeeper,
		appKeepers.ConcentratedLiquidityKeeper,
		appKeepers.PoolManagerKeeper,
		appKeepers.PoolIncentivesKeeper,
		appKeepers.ProtoRevKeeper,
	)
	appKeepers.ConcentratedLiquidityKeeper.SetIncentivesKeeper(appKeepers.IncentivesKeeper)
	appKeepers.GAMMKeeper.SetIncentivesKeeper(appKeepers.IncentivesKeeper)

	mintKeeper := mintkeeper.NewKeeper(
		appKeepers.keys[minttypes.StoreKey],
		appKeepers.GetSubspace(minttypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.DistrKeeper,
		appKeepers.EpochsKeeper,
		authtypes.FeeCollectorName,
	)
	appKeepers.MintKeeper = &mintKeeper

	poolIncentivesKeeper := poolincentiveskeeper.NewKeeper(
		appKeepers.keys[poolincentivestypes.StoreKey],
		appKeepers.GetSubspace(poolincentivestypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.IncentivesKeeper,
		appKeepers.DistrKeeper,
		appKeepers.PoolManagerKeeper,
		appKeepers.EpochsKeeper,
		appKeepers.GAMMKeeper,
	)
	appKeepers.PoolIncentivesKeeper = &poolIncentivesKeeper
	appKeepers.PoolManagerKeeper.SetPoolIncentivesKeeper(appKeepers.PoolIncentivesKeeper)
	appKeepers.IncentivesKeeper.SetPoolIncentivesKeeper(appKeepers.PoolIncentivesKeeper)
	appKeepers.ConcentratedLiquidityKeeper.SetPoolIncentivesKeeper(appKeepers.PoolIncentivesKeeper)
	appKeepers.GAMMKeeper.SetPoolIncentivesKeeper(appKeepers.PoolIncentivesKeeper)

	tokenFactoryKeeper := tokenfactorykeeper.NewKeeper(
		appKeepers.keys[tokenfactorytypes.StoreKey],
		appKeepers.GetSubspace(tokenfactorytypes.ModuleName),
		maccPerms,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.DistrKeeper,
	)
	appKeepers.TokenFactoryKeeper = &tokenFactoryKeeper

	validatorSetPreferenceKeeper := valsetpref.NewKeeper(
		appKeepers.keys[valsetpreftypes.StoreKey],
		appKeepers.GetSubspace(valsetpreftypes.ModuleName),
		appKeepers.StakingKeeper,
		appKeepers.DistrKeeper,
		appKeepers.LockupKeeper,
	)

	// initialize the auction keeper
	auctionKeeper := auctionkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[auctiontypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.DistrKeeper,
		appKeepers.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.AuctionKeeper = &auctionKeeper

	appKeepers.ValidatorSetPreferenceKeeper = &validatorSetPreferenceKeeper

	appKeepers.SuperfluidKeeper = superfluidkeeper.NewKeeper(
		appKeepers.keys[superfluidtypes.StoreKey], appKeepers.GetSubspace(superfluidtypes.ModuleName),
		*appKeepers.AccountKeeper, appKeepers.BankKeeper, appKeepers.StakingKeeper, appKeepers.DistrKeeper, appKeepers.EpochsKeeper, appKeepers.LockupKeeper, appKeepers.GAMMKeeper, appKeepers.IncentivesKeeper,
		lockupkeeper.NewMsgServerImpl(appKeepers.LockupKeeper), appKeepers.ConcentratedLiquidityKeeper, appKeepers.PoolManagerKeeper, appKeepers.ValidatorSetPreferenceKeeper)

	oracleKeeper := oraclekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[oracletypes.StoreKey],
		appKeepers.GetSubspace(oracletypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.DistrKeeper,
		appKeepers.StakingKeeper,
		appKeepers.EpochsKeeper,
		distrtypes.ModuleName,
	)
	appKeepers.OracleKeeper = &oracleKeeper

	marketKeeper := marketkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[markettypes.StoreKey],
		appKeepers.GetSubspace(markettypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.OracleKeeper,
	)
	appKeepers.MarketKeeper = &marketKeeper

	treasuryKeeper := treasurykeeper.NewKeeper(
		appCodec,
		appKeepers.keys[treasurytypes.StoreKey],
		appKeepers.GetSubspace(treasurytypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.MarketKeeper,
		appKeepers.OracleKeeper)
	appKeepers.TreasuryKeeper = &treasuryKeeper

	stableStakingIncetivesKeeper := stablestakingincentviceskeeper.NewKeeper(
		appKeepers.keys[stablestakingincentvicestypes.StoreKey],
		appKeepers.GetSubspace(stablestakingincentvicestypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper)
	appKeepers.StableStakingIncentivesKeeper = &stableStakingIncetivesKeeper

	txFeesKeeper := txfeeskeeper.NewKeeper(
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.keys[txfeestypes.StoreKey],
		appKeepers.MarketKeeper,
		appKeepers.OracleKeeper,
		appKeepers.DistrKeeper,
		appKeepers.ConsensusParamsKeeper,
		dataDir,
		appKeepers.GetSubspace(txfeestypes.ModuleName),
	)
	appKeepers.TxFeesKeeper = &txFeesKeeper

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	supportedFeatures := []string{
		"iterator",
		"staking",
		"stargate",
		"osmosis",
		"cosmwasm_1_1",
		"cosmwasm_1_2",
		"cosmwasm_1_4",
		"cosmwasm_2_0",
		"cosmwasm_2_1",
	}

	wasmMsgHandler := customwasmkeeper.NewMessageHandler(
		bApp.MsgServiceRouter(),
		appKeepers.HooksICS4Wrapper,
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.ScopedIBCKeeper,
		appKeepers.WasmKeeper,
		appKeepers.BankKeeper,
		appKeepers.TreasuryKeeper,
		*appKeepers.AccountKeeper,
		appCodec,
		appKeepers.TransferKeeper,
	)
	// the first slice will replace all default msh handler with custom one
	wasmOpts = append([]wasmkeeper.Option{wasmkeeper.WithMessageHandler(wasmMsgHandler)}, wasmOpts...)
	wasmOpts = append(owasm.RegisterCustomPlugins(appKeepers.BankKeeper, appKeepers.TokenFactoryKeeper), wasmOpts...)
	wasmOpts = append(owasm.RegisterStargateQueries(*bApp.GRPCQueryRouter(), appCodec), wasmOpts...)

	wasmKeeper := wasmkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[wasmtypes.StoreKey]),
		*appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		*appKeepers.StakingKeeper,
		distrkeeper.NewQuerier(*appKeepers.DistrKeeper),
		appKeepers.RateLimitingICS4Wrapper,
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.ScopedWasmKeeper,
		appKeepers.TransferKeeper,
		bApp.MsgServiceRouter(),
		bApp.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		supportedFeatures,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)
	appKeepers.WasmKeeper = &wasmKeeper
	appKeepers.CosmwasmPoolKeeper.SetWasmKeeper(appKeepers.WasmKeeper)
	appKeepers.PoolManagerKeeper.SetWasmKeeper(appKeepers.WasmKeeper)

	// Pass the contract keeper to all the structs (generally ICS4Wrappers for ibc middlewares) that need it
	appKeepers.ContractKeeper = wasmkeeper.NewDefaultPermissionKeeper(appKeepers.WasmKeeper)
	appKeepers.RateLimitingICS4Wrapper.ContractKeeper = appKeepers.ContractKeeper
	appKeepers.Ics20WasmHooks.ContractKeeper = appKeepers.WasmKeeper
	appKeepers.CosmwasmPoolKeeper.SetContractKeeper(appKeepers.ContractKeeper)
	appKeepers.IBCHooksKeeper.ContractKeeper = appKeepers.ContractKeeper
	appKeepers.ConcentratedLiquidityKeeper.SetContractKeeper(appKeepers.ContractKeeper)
	appKeepers.StableStakingIncentivesKeeper.SetContractKeeper(appKeepers.ContractKeeper)

	// register CosmWasm authenticator
	appKeepers.AuthenticatorManager.RegisterAuthenticator(
		authenticator.NewCosmwasmAuthenticator(appKeepers.ContractKeeper, appKeepers.AccountKeeper, appCodec))

	// set token factory contract keeper
	appKeepers.TokenFactoryKeeper.SetContractKeeper(appKeepers.ContractKeeper)

	// wire up x/wasm to IBC
	ibcRouter.AddRoute(wasmtypes.ModuleName, wasm.NewIBCHandler(appKeepers.WasmKeeper, appKeepers.IBCKeeper.ChannelKeeper, appKeepers.IBCKeeper.ChannelKeeper))

	// Seal the router
	appKeepers.IBCKeeper.SetRouter(ibcRouter)

	// register the proposal types
	govRouter := govtypesv1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypesv1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(*appKeepers.ParamsKeeper)).
		AddRoute(poolincentivestypes.RouterKey, poolincentives.NewPoolIncentivesProposalHandler(*appKeepers.PoolIncentivesKeeper)).
		AddRoute(superfluidtypes.RouterKey, superfluid.NewSuperfluidProposalHandler(*appKeepers.SuperfluidKeeper, *appKeepers.EpochsKeeper, *appKeepers.GAMMKeeper)).
		AddRoute(protorevtypes.RouterKey, protorev.NewProtoRevProposalHandler(*appKeepers.ProtoRevKeeper)).
		AddRoute(gammtypes.RouterKey, gamm.NewGammProposalHandler(*appKeepers.GAMMKeeper)).
		AddRoute(concentratedliquiditytypes.RouterKey, concentratedliquidity.NewConcentratedLiquidityProposalHandler(*appKeepers.ConcentratedLiquidityKeeper)).
		AddRoute(cosmwasmpooltypes.RouterKey, cosmwasmpool.NewCosmWasmPoolProposalHandler(*appKeepers.CosmwasmPoolKeeper)).
		AddRoute(poolmanagertypes.RouterKey, poolmanager.NewPoolManagerProposalHandler(*appKeepers.PoolManagerKeeper)).
		AddRoute(incentivestypes.RouterKey, incentiveskeeper.NewIncentivesProposalHandler(*appKeepers.IncentivesKeeper))

	govConfig := govtypes.DefaultConfig()
	// Set the maximum metadata length for government-related configurations to 10,200, deviating from the default value of 256.
	govConfig.MaxMetadataLen = 10200
	govKeeper := govkeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(appKeepers.keys[govtypes.StoreKey]),
		appKeepers.AccountKeeper, appKeepers.BankKeeper, appKeepers.SuperfluidKeeper, appKeepers.DistrKeeper, bApp.MsgServiceRouter(),
		govConfig, authtypes.NewModuleAddress(govtypes.ModuleName).String())
	appKeepers.GovKeeper = govKeeper
	appKeepers.GovKeeper.SetLegacyRouter(govRouter)
}

// WireICS20PreWasmKeeper Create the IBC Transfer Stack from bottom to top:
//
// * SendPacket. Originates from the transferKeeper and goes up the stack:
// transferKeeper.SendPacket -> ibc_rate_limit.SendPacket -> ibc_hooks.SendPacket -> channel.SendPacket
// * RecvPacket, message that originates from core IBC and goes down to app, the flow is the other way
// channel.RecvPacket -> ibc_hooks.OnRecvPacket -> ibc_rate_limit.OnRecvPacket -> forward.OnRecvPacket -> transfer.OnRecvPacket
//
// Note that the forward middleware is only integrated on the "receive" direction. It can be safely skipped when sending.
// Note also that the forward middleware is called "router", but we are using the name "forward" for clarity
// This may later be renamed upstream: https://github.com/ibc-apps/middleware/packet-forward-middleware/issues/10
//
// After this, the wasm keeper is required to be set on both
// appkeepers.WasmHooks AND appKeepers.RateLimitingICS4Wrapper
func (appKeepers *AppKeepers) WireICS20PreWasmKeeper(
	appCodec codec.Codec,
	bApp *baseapp.BaseApp,
	hooksKeeper *ibchookskeeper.Keeper,
) {
	// Setup the ICS4Wrapper used by the hooks middleware
	melodyPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	wasmHooks := ibchooks.NewWasmHooks(hooksKeeper, nil, melodyPrefix) // The contract keeper needs to be set later
	appKeepers.Ics20WasmHooks = &wasmHooks
	appKeepers.HooksICS4Wrapper = ibchooks.NewICS4Middleware(
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.Ics20WasmHooks,
	)

	// ChannelKeeper wrapper for rate limiting SendPacket(). The wasmKeeper needs to be added after it's created
	rateLimitingICS4Wrapper := ibcratelimit.NewICS4Middleware(
		appKeepers.HooksICS4Wrapper,
		appKeepers.AccountKeeper,
		// wasm keeper we set later.
		nil,
		appKeepers.BankKeeper,
		appKeepers.GetSubspace(ibcratelimittypes.ModuleName),
	)
	appKeepers.RateLimitingICS4Wrapper = &rateLimitingICS4Wrapper

	// Create Transfer Keepers
	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibctransfertypes.StoreKey],
		appKeepers.GetSubspace(ibctransfertypes.ModuleName),
		// The ICS4Wrapper is replaced by the rateLimitingICS4Wrapper instead of the channel
		appKeepers.RateLimitingICS4Wrapper,
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.ScopedTransferKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.TransferKeeper = &transferKeeper
	appKeepers.RawIcs20TransferAppModule = transfer.NewAppModule(*appKeepers.TransferKeeper)

	// Packet Forward Middleware
	// Initialize packet forward middleware router
	appKeepers.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[packetforwardtypes.StoreKey],
		appKeepers.TransferKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.DistrKeeper,
		appKeepers.BankKeeper,
		// The ICS4Wrapper is replaced by the HooksICS4Wrapper instead of the channel so that sending can be overridden by the middleware
		appKeepers.HooksICS4Wrapper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	packetForwardMiddleware := packetforward.NewIBCMiddleware(
		transfer.NewIBCModule(*appKeepers.TransferKeeper),
		appKeepers.PacketForwardKeeper,
		0,
		packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
		packetforwardkeeper.DefaultRefundTransferPacketTimeoutTimestamp,
	)

	// RateLimiting IBC Middleware
	rateLimitingTransferModule := ibcratelimit.NewIBCModule(packetForwardMiddleware, appKeepers.RateLimitingICS4Wrapper)

	// Hooks Middleware
	hooksTransferModule := ibchooks.NewIBCMiddleware(&rateLimitingTransferModule, &appKeepers.HooksICS4Wrapper)
	appKeepers.TransferStack = &hooksTransferModule
}

// InitSpecialKeepers initiates special keepers (crisis appkeeper, upgradekeeper, params keeper)
func (appKeepers *AppKeepers) InitSpecialKeepers(
	appCodec codec.Codec,
	bApp *baseapp.BaseApp,
	wasmDir string,
	cdc *codec.LegacyAmino,
	invCheckPeriod uint,
	skipUpgradeHeights map[int64]bool,
	homePath string,
) {
	appKeepers.GenerateKeys()
	paramsKeeper := appKeepers.initParamsKeeper(appCodec, cdc, appKeepers.keys[paramstypes.StoreKey], appKeepers.tkeys[paramstypes.TStoreKey])
	appKeepers.ParamsKeeper = &paramsKeeper

	// set the BaseApp's parameter store
	consensusParamsKeeper := consensusparamkeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(appKeepers.keys[consensusparamtypes.StoreKey]), authtypes.NewModuleAddress(govtypes.ModuleName).String(), runtime.EventService{})
	appKeepers.ConsensusParamsKeeper = &consensusParamsKeeper
	bApp.SetParamStore(appKeepers.ConsensusParamsKeeper.ParamsStore)

	// add capability keeper and ScopeToModule for ibc module
	appKeepers.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, appKeepers.keys[capabilitytypes.StoreKey], appKeepers.memKeys[capabilitytypes.MemStoreKey])
	appKeepers.ScopedIBCKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	appKeepers.ScopedICAHostKeeper = appKeepers.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	appKeepers.ScopedICAControllerKeeper = appKeepers.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	appKeepers.ScopedTransferKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	appKeepers.ScopedWasmKeeper = appKeepers.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)
	appKeepers.ScopedICQKeeper = appKeepers.CapabilityKeeper.ScopeToModule(icqtypes.ModuleName)
	appKeepers.CapabilityKeeper.Seal()

	// TODO: Make a SetInvCheckPeriod fn on CrisisKeeper.
	// IMO, its bad design atm that it requires this in state machine initialization
	crisisKeeper := crisiskeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(appKeepers.keys[crisistypes.StoreKey]), invCheckPeriod, appKeepers.BankKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String(), addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()))
	appKeepers.CrisisKeeper = crisisKeeper

	upgradeKeeper := upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(appKeepers.keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		bApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.UpgradeKeeper = upgradeKeeper
}

// initParamsKeeper init params keeper and its subspaces.
func (appKeepers *AppKeepers) initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)

	// register the key tables for legacy param subspaces
	keyTable := ibcclienttypes.ParamKeyTable()
	keyTable.RegisterParamSet(&ibcconnectiontypes.Params{})
	paramsKeeper.Subspace(ibchost.ModuleName).WithKeyTable(keyTable)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName).WithKeyTable(ibctransfertypes.ParamKeyTable())
	paramsKeeper.Subspace(icahosttypes.SubModuleName).WithKeyTable(icahosttypes.ParamKeyTable())
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName).WithKeyTable(icacontrollertypes.ParamKeyTable())
	paramsKeeper.Subspace(incentivestypes.ModuleName)
	paramsKeeper.Subspace(lockuptypes.ModuleName)
	paramsKeeper.Subspace(poolincentivestypes.ModuleName)
	paramsKeeper.Subspace(stablestakingincentvicestypes.ModuleName)
	paramsKeeper.Subspace(protorevtypes.ModuleName)
	paramsKeeper.Subspace(superfluidtypes.ModuleName)
	paramsKeeper.Subspace(poolmanagertypes.ModuleName)
	paramsKeeper.Subspace(oracletypes.ModuleName)
	paramsKeeper.Subspace(markettypes.ModuleName)
	paramsKeeper.Subspace(treasurytypes.ModuleName)
	paramsKeeper.Subspace(gammtypes.ModuleName)
	paramsKeeper.Subspace(wasmtypes.ModuleName)
	paramsKeeper.Subspace(tokenfactorytypes.ModuleName)
	paramsKeeper.Subspace(twaptypes.ModuleName)
	paramsKeeper.Subspace(ibcratelimittypes.ModuleName)
	paramsKeeper.Subspace(concentratedliquiditytypes.ModuleName)
	paramsKeeper.Subspace(icqtypes.ModuleName)
	paramsKeeper.Subspace(packetforwardtypes.ModuleName).WithKeyTable(packetforwardtypes.ParamKeyTable())
	paramsKeeper.Subspace(cosmwasmpooltypes.ModuleName)
	paramsKeeper.Subspace(ibchookstypes.ModuleName)
	paramsKeeper.Subspace(smartaccounttypes.ModuleName).WithKeyTable(smartaccounttypes.ParamKeyTable())
	paramsKeeper.Subspace(txfeestypes.ModuleName)
	paramsKeeper.Subspace(auctiontypes.ModuleName)

	return paramsKeeper
}

// SetupHooks sets up hooks for modules.
func (appKeepers *AppKeepers) SetupHooks() {
	// For every module that has hooks set on it,
	// you must check InitNormalKeepers to ensure that its not passed by de-reference
	// e.g. *app.StakingKeeper doesn't appear

	// Recall that SetHooks is a mutative call.
	appKeepers.BankKeeper.SetHooks(
		banktypes.NewMultiBankHooks(
			appKeepers.TokenFactoryKeeper.Hooks(),
		),
	)

	appKeepers.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			appKeepers.DistrKeeper.Hooks(),
			appKeepers.SlashingKeeper.Hooks(),
			appKeepers.SuperfluidKeeper.Hooks(),
		),
	)

	appKeepers.GAMMKeeper.SetHooks(
		gammtypes.NewMultiGammHooks(
			// insert gamm hooks receivers here
			appKeepers.PoolIncentivesKeeper.Hooks(),
			appKeepers.TwapKeeper.GammHooks(),
			appKeepers.ProtoRevKeeper.Hooks(),
		),
	)

	appKeepers.ConcentratedLiquidityKeeper.SetListeners(
		concentratedliquiditytypes.NewConcentratedLiquidityListeners(
			appKeepers.TwapKeeper.ConcentratedLiquidityListener(),
			appKeepers.PoolIncentivesKeeper.Hooks(),
			appKeepers.ProtoRevKeeper.Hooks(),
		),
	)

	appKeepers.LockupKeeper.SetHooks(
		lockuptypes.NewMultiLockupHooks(
			// insert lockup hooks receivers here
			appKeepers.SuperfluidKeeper.Hooks(),
		),
	)

	appKeepers.IncentivesKeeper.SetHooks(
		incentivestypes.NewMultiIncentiveHooks(
		// insert incentive hooks receivers here
		),
	)

	appKeepers.MintKeeper.SetHooks(
		minttypes.NewMultiMintHooks(
			// insert mint hooks receivers here
			appKeepers.PoolIncentivesKeeper.Hooks(),
			appKeepers.StableStakingIncentivesKeeper.Hooks(),
		),
	)

	appKeepers.EpochsKeeper.SetHooks(
		epochstypes.NewMultiEpochHooks(
			// insert epoch hooks receivers here
			appKeepers.TxFeesKeeper.Hooks(),
			appKeepers.TwapKeeper.EpochHooks(),
			appKeepers.SuperfluidKeeper.Hooks(),
			appKeepers.IncentivesKeeper.Hooks(),
			appKeepers.MintKeeper.Hooks(),
			appKeepers.ProtoRevKeeper.EpochHooks(),
			appKeepers.OracleKeeper.Hooks(),
			appKeepers.TreasuryKeeper.Hooks(),
		),
	)

	appKeepers.GovKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// insert governance hooks receivers here
		),
	)
}

// TODO: We need to automate this, by bundling with a module struct...
func KVStoreKeys() []string {
	return []string{
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		minttypes.StoreKey,
		distrtypes.StoreKey,
		downtimetypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		crisistypes.StoreKey,
		paramstypes.StoreKey,
		consensusparamtypes.StoreKey,
		ibchost.StoreKey,
		icahosttypes.StoreKey,
		icacontrollertypes.StoreKey,
		upgradetypes.StoreKey,
		evidencetypes.StoreKey,
		ibctransfertypes.StoreKey,
		ibcwasmtypes.StoreKey,
		capabilitytypes.StoreKey,
		gammtypes.StoreKey,
		twaptypes.StoreKey,
		lockuptypes.StoreKey,
		incentivestypes.StoreKey,
		epochstypes.StoreKey,
		poolincentivestypes.StoreKey,
		concentratedliquiditytypes.StoreKey,
		poolmanagertypes.StoreKey,
		oracletypes.StoreKey,
		markettypes.StoreKey,
		treasurytypes.StoreKey,
		authzkeeper.StoreKey,
		txfeestypes.StoreKey,
		superfluidtypes.StoreKey,
		wasmtypes.StoreKey,
		tokenfactorytypes.StoreKey,
		valsetpreftypes.StoreKey,
		protorevtypes.StoreKey,
		ibchookstypes.StoreKey,
		icqtypes.StoreKey,
		packetforwardtypes.StoreKey,
		cosmwasmpooltypes.StoreKey,
		auctiontypes.StoreKey,
		smartaccounttypes.StoreKey,
	}
}
