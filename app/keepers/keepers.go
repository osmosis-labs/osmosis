package keepers

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icahost "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibcclient "github.com/cosmos/ibc-go/v3/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"

	// IBC Transfer: Defines the "transfer" IBC port
	transfer "github.com/cosmos/ibc-go/v3/modules/apps/transfer"

	_ "github.com/osmosis-labs/osmosis/v10/client/docs/statik"
	owasm "github.com/osmosis-labs/osmosis/v10/wasmbinding"
	epochskeeper "github.com/osmosis-labs/osmosis/v10/x/epochs/keeper"
	epochstypes "github.com/osmosis-labs/osmosis/v10/x/epochs/types"
	gammkeeper "github.com/osmosis-labs/osmosis/v10/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap"
	twaptypes "github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
	gammtypes "github.com/osmosis-labs/osmosis/v10/x/gamm/types"
	incentiveskeeper "github.com/osmosis-labs/osmosis/v10/x/incentives/keeper"
	incentivestypes "github.com/osmosis-labs/osmosis/v10/x/incentives/types"
	lockupkeeper "github.com/osmosis-labs/osmosis/v10/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v10/x/lockup/types"
	mintkeeper "github.com/osmosis-labs/osmosis/v10/x/mint/keeper"
	minttypes "github.com/osmosis-labs/osmosis/v10/x/mint/types"
	poolincentives "github.com/osmosis-labs/osmosis/v10/x/pool-incentives"
	poolincentiveskeeper "github.com/osmosis-labs/osmosis/v10/x/pool-incentives/keeper"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v10/x/pool-incentives/types"
	"github.com/osmosis-labs/osmosis/v10/x/superfluid"
	superfluidkeeper "github.com/osmosis-labs/osmosis/v10/x/superfluid/keeper"
	superfluidtypes "github.com/osmosis-labs/osmosis/v10/x/superfluid/types"
	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v10/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v10/x/tokenfactory/types"
	"github.com/osmosis-labs/osmosis/v10/x/txfees"
	txfeeskeeper "github.com/osmosis-labs/osmosis/v10/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/v10/x/txfees/types"
)

type AppKeepers struct {
	// keepers, by order of initialization
	// "Special" keepers
	ParamsKeeper     *paramskeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	CrisisKeeper     *crisiskeeper.Keeper
	UpgradeKeeper    *upgradekeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper  capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	ScopedWasmKeeper     capabilitykeeper.ScopedKeeper

	// "Normal" keepers
	AccountKeeper        *authkeeper.AccountKeeper
	BankKeeper           *bankkeeper.BaseKeeper
	AuthzKeeper          *authzkeeper.Keeper
	StakingKeeper        *stakingkeeper.Keeper
	DistrKeeper          *distrkeeper.Keeper
	SlashingKeeper       *slashingkeeper.Keeper
	IBCKeeper            *ibckeeper.Keeper
	ICAHostKeeper        *icahostkeeper.Keeper
	TransferKeeper       *ibctransferkeeper.Keeper
	EvidenceKeeper       *evidencekeeper.Keeper
	GAMMKeeper           *gammkeeper.Keeper
	TwapKeeper           *twap.Keeper
	LockupKeeper         *lockupkeeper.Keeper
	EpochsKeeper         *epochskeeper.Keeper
	IncentivesKeeper     *incentiveskeeper.Keeper
	MintKeeper           *mintkeeper.Keeper
	PoolIncentivesKeeper *poolincentiveskeeper.Keeper
	TxFeesKeeper         *txfeeskeeper.Keeper
	SuperfluidKeeper     *superfluidkeeper.Keeper
	GovKeeper            *govkeeper.Keeper
	WasmKeeper           *wasm.Keeper
	TokenFactoryKeeper   *tokenfactorykeeper.Keeper
	// IBC modules
	// transfer module
	TransferModule transfer.AppModule

	// keys to access the substores
	keys    map[string]*sdk.KVStoreKey
	tkeys   map[string]*sdk.TransientStoreKey
	memKeys map[string]*sdk.MemoryStoreKey
}

// InitNormalKeepers initializes all 'normal' keepers (account, app, bank, auth, staking, distribution, slashing, transfer, gamm, IBC router, pool incentives, governance, mint, txfees keepers).
func (appKeepers *AppKeepers) InitNormalKeepers(
	appCodec codec.Codec,
	bApp *baseapp.BaseApp,
	maccPerms map[string][]string,
	wasmDir string,
	wasmConfig wasm.Config,
	wasmEnabledProposals []wasm.ProposalType,
	wasmOpts []wasm.Option,
	blockedAddress map[string]bool,
) {
	// Add 'normal' keepers
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		appKeepers.keys[authtypes.StoreKey],
		appKeepers.GetSubspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount,
		maccPerms,
	)
	appKeepers.AccountKeeper = &accountKeeper
	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		appKeepers.keys[banktypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.GetSubspace(banktypes.ModuleName),
		blockedAddress,
	)
	appKeepers.BankKeeper = &bankKeeper

	authzKeeper := authzkeeper.NewKeeper(
		appKeepers.keys[authzkeeper.StoreKey],
		appCodec,
		bApp.MsgServiceRouter(),
	)
	appKeepers.AuthzKeeper = &authzKeeper

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[stakingtypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.GetSubspace(stakingtypes.ModuleName),
	)
	appKeepers.StakingKeeper = &stakingKeeper

	distrKeeper := distrkeeper.NewKeeper(
		appCodec, appKeepers.keys[distrtypes.StoreKey],
		appKeepers.GetSubspace(distrtypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		authtypes.FeeCollectorName,
		blockedAddress,
	)
	appKeepers.DistrKeeper = &distrKeeper

	slashingKeeper := slashingkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[slashingtypes.StoreKey],
		appKeepers.StakingKeeper,
		appKeepers.GetSubspace(slashingtypes.ModuleName),
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
	)

	// Create Transfer Keepers
	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibctransfertypes.StoreKey],
		appKeepers.GetSubspace(ibctransfertypes.ModuleName),
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.ScopedTransferKeeper,
	)
	appKeepers.TransferKeeper = &transferKeeper
	appKeepers.TransferModule = transfer.NewAppModule(*appKeepers.TransferKeeper)
	transferIBCModule := transfer.NewIBCModule(*appKeepers.TransferKeeper)

	icaHostKeeper := icahostkeeper.NewKeeper(
		appCodec, appKeepers.keys[icahosttypes.StoreKey],
		appKeepers.GetSubspace(icahosttypes.SubModuleName),
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		appKeepers.ScopedICAHostKeeper,
		bApp.MsgServiceRouter(),
	)
	appKeepers.ICAHostKeeper = &icaHostKeeper

	icaHostIBCModule := icahost.NewIBCModule(*appKeepers.ICAHostKeeper)
	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
		AddRoute(ibctransfertypes.ModuleName, transferIBCModule)
	// Note: the sealing is done after creating wasmd and wiring that up

	// create evidence keeper with router
	// If evidence needs to be handled for the app, set routes in router here and seal
	appKeepers.EvidenceKeeper = evidencekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[evidencetypes.StoreKey],
		appKeepers.StakingKeeper,
		appKeepers.SlashingKeeper,
	)

	gammKeeper := gammkeeper.NewKeeper(
		appCodec, appKeepers.keys[gammtypes.StoreKey],
		appKeepers.GetSubspace(gammtypes.ModuleName),
		appKeepers.AccountKeeper, appKeepers.BankKeeper, appKeepers.DistrKeeper)
	appKeepers.GAMMKeeper = &gammKeeper

	appKeepers.TwapKeeper = twap.NewKeeper(
		appKeepers.keys[twaptypes.StoreKey],
		appKeepers.tkeys[twaptypes.TransientStoreKey],
		appKeepers.GetSubspace(twaptypes.ModuleName),
		appKeepers.GAMMKeeper)

	appKeepers.LockupKeeper = lockupkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[lockuptypes.StoreKey],
		// TODO: Visit why this needs to be deref'd
		*appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.DistrKeeper)

	appKeepers.EpochsKeeper = epochskeeper.NewKeeper(appCodec, appKeepers.keys[epochstypes.StoreKey])

	appKeepers.IncentivesKeeper = incentiveskeeper.NewKeeper(
		appCodec,
		appKeepers.keys[incentivestypes.StoreKey],
		appKeepers.GetSubspace(incentivestypes.ModuleName),
		appKeepers.BankKeeper,
		appKeepers.LockupKeeper,
		appKeepers.EpochsKeeper,
	)

	appKeepers.SuperfluidKeeper = superfluidkeeper.NewKeeper(
		appCodec, appKeepers.keys[superfluidtypes.StoreKey], appKeepers.GetSubspace(superfluidtypes.ModuleName),
		*appKeepers.AccountKeeper, appKeepers.BankKeeper, appKeepers.StakingKeeper, appKeepers.DistrKeeper, appKeepers.EpochsKeeper, appKeepers.LockupKeeper, appKeepers.GAMMKeeper, appKeepers.IncentivesKeeper,
		lockupkeeper.NewMsgServerImpl(appKeepers.LockupKeeper))

	mintKeeper := mintkeeper.NewKeeper(
		appCodec,
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
		appCodec,
		appKeepers.keys[poolincentivestypes.StoreKey],
		appKeepers.GetSubspace(poolincentivestypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.IncentivesKeeper,
		appKeepers.DistrKeeper,
		distrtypes.ModuleName,
		authtypes.FeeCollectorName,
	)
	appKeepers.PoolIncentivesKeeper = &poolIncentivesKeeper

	txFeesKeeper := txfeeskeeper.NewKeeper(
		appCodec,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.EpochsKeeper,
		appKeepers.keys[txfeestypes.StoreKey],
		appKeepers.GAMMKeeper,
		appKeepers.GAMMKeeper,
		txfeestypes.FeeCollectorName,
		txfeestypes.NonNativeFeeCollectorName,
	)
	appKeepers.TxFeesKeeper = &txFeesKeeper

	tokenFactoryKeeper := tokenfactorykeeper.NewKeeper(
		appCodec,
		appKeepers.keys[tokenfactorytypes.StoreKey],
		appKeepers.GetSubspace(tokenfactorytypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper.WithMintCoinsRestriction(tokenfactorytypes.NewTokenFactoryDenomMintCoinsRestriction()),
		appKeepers.DistrKeeper,
	)
	appKeepers.TokenFactoryKeeper = &tokenFactoryKeeper

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	supportedFeatures := "iterator,staking,stargate,osmosis"

	wasmOpts = append(owasm.RegisterCustomPlugins(appKeepers.GAMMKeeper, appKeepers.BankKeeper, appKeepers.TokenFactoryKeeper), wasmOpts...)

	wasmKeeper := wasm.NewKeeper(
		appCodec,
		appKeepers.keys[wasm.StoreKey],
		appKeepers.GetSubspace(wasm.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		appKeepers.DistrKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		appKeepers.ScopedWasmKeeper,
		appKeepers.TransferKeeper,
		bApp.MsgServiceRouter(),
		bApp.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		supportedFeatures,
		wasmOpts...,
	)
	appKeepers.WasmKeeper = &wasmKeeper

	// wire up x/wasm to IBC
	ibcRouter.AddRoute(wasm.ModuleName, wasm.NewIBCHandler(appKeepers.WasmKeeper, appKeepers.IBCKeeper.ChannelKeeper))
	appKeepers.IBCKeeper.SetRouter(ibcRouter)

	// register the proposal types
	govRouter := govtypes.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(*appKeepers.ParamsKeeper)).
		AddRoute(distrtypes.RouterKey, distribution.NewCommunityPoolSpendProposalHandler(*appKeepers.DistrKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(appKeepers.IBCKeeper.ClientKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(*appKeepers.UpgradeKeeper)).
		AddRoute(ibchost.RouterKey, ibcclient.NewClientProposalHandler(appKeepers.IBCKeeper.ClientKeeper)).
		AddRoute(poolincentivestypes.RouterKey, poolincentives.NewPoolIncentivesProposalHandler(*appKeepers.PoolIncentivesKeeper)).
		AddRoute(txfeestypes.RouterKey, txfees.NewUpdateFeeTokenProposalHandler(*appKeepers.TxFeesKeeper)).
		AddRoute(superfluidtypes.RouterKey, superfluid.NewSuperfluidProposalHandler(*appKeepers.SuperfluidKeeper, *appKeepers.EpochsKeeper))

	// The gov proposal types can be individually enabled
	if len(wasmEnabledProposals) != 0 {
		govRouter.AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(appKeepers.WasmKeeper, wasmEnabledProposals))
	}

	govKeeper := govkeeper.NewKeeper(
		appCodec, appKeepers.keys[govtypes.StoreKey],
		appKeepers.GetSubspace(govtypes.ModuleName), appKeepers.AccountKeeper, appKeepers.BankKeeper,
		appKeepers.SuperfluidKeeper, govRouter)
	appKeepers.GovKeeper = &govKeeper
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
	bApp.SetParamStore(appKeepers.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))

	// add capability keeper and ScopeToModule for ibc module
	appKeepers.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, appKeepers.keys[capabilitytypes.StoreKey], appKeepers.memKeys[capabilitytypes.MemStoreKey])
	appKeepers.ScopedIBCKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	appKeepers.ScopedICAHostKeeper = appKeepers.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	appKeepers.ScopedTransferKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	appKeepers.ScopedWasmKeeper = appKeepers.CapabilityKeeper.ScopeToModule(wasm.ModuleName)
	appKeepers.CapabilityKeeper.Seal()

	// TODO: Make a SetInvCheckPeriod fn on CrisisKeeper.
	// IMO, its bad design atm that it requires this in state machine initialization
	crisisKeeper := crisiskeeper.NewKeeper(
		appKeepers.GetSubspace(crisistypes.ModuleName), invCheckPeriod, appKeepers.BankKeeper, authtypes.FeeCollectorName,
	)
	appKeepers.CrisisKeeper = &crisisKeeper

	upgradeKeeper := upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		appKeepers.keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		bApp,
	)
	appKeepers.UpgradeKeeper = &upgradeKeeper
}

// initParamsKeeper init params keeper and its subspaces.
func (appKeepers *AppKeepers) initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey sdk.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypes.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(incentivestypes.ModuleName)
	paramsKeeper.Subspace(poolincentivestypes.ModuleName)
	paramsKeeper.Subspace(superfluidtypes.ModuleName)
	paramsKeeper.Subspace(gammtypes.ModuleName)
	paramsKeeper.Subspace(wasm.ModuleName)
	paramsKeeper.Subspace(tokenfactorytypes.ModuleName)

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
			appKeepers.TokenFactoryKeeper.Hooks(*appKeepers.WasmKeeper),
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
		),
	)

	appKeepers.EpochsKeeper.SetHooks(
		epochstypes.NewMultiEpochHooks(
			// insert epoch hooks receivers here
			appKeepers.TxFeesKeeper.Hooks(),
			appKeepers.SuperfluidKeeper.Hooks(),
			appKeepers.IncentivesKeeper.Hooks(),
			appKeepers.MintKeeper.Hooks(),
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
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		paramstypes.StoreKey,
		ibchost.StoreKey,
		icahosttypes.StoreKey,
		upgradetypes.StoreKey,
		evidencetypes.StoreKey,
		ibctransfertypes.StoreKey,
		capabilitytypes.StoreKey,
		gammtypes.StoreKey,
		twaptypes.StoreKey,
		lockuptypes.StoreKey,
		incentivestypes.StoreKey,
		epochstypes.StoreKey,
		poolincentivestypes.StoreKey,
		authzkeeper.StoreKey,
		txfeestypes.StoreKey,
		superfluidtypes.StoreKey,
		wasm.StoreKey,
		tokenfactorytypes.StoreKey,
	}
}
