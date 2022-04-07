package app

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

	// Staking: Allows the Tendermint validator set to be chosen based on bonded stake.
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	transfer "github.com/cosmos/ibc-go/v2/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v2/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	ibcclient "github.com/cosmos/ibc-go/v2/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v2/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v2/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v2/modules/core/keeper"
	"github.com/osmosis-labs/bech32-ibc/x/bech32ibc"
	bech32ibckeeper "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/keeper"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"
	bech32ics20keeper "github.com/osmosis-labs/bech32-ibc/x/bech32ics20/keeper"

	owasm "github.com/osmosis-labs/osmosis/v7/app/wasm"
	_ "github.com/osmosis-labs/osmosis/v7/client/docs/statik"
	claimkeeper "github.com/osmosis-labs/osmosis/v7/x/claim/keeper"
	claimtypes "github.com/osmosis-labs/osmosis/v7/x/claim/types"
	epochskeeper "github.com/osmosis-labs/osmosis/v7/x/epochs/keeper"
	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	gammkeeper "github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	incentiveskeeper "github.com/osmosis-labs/osmosis/v7/x/incentives/keeper"
	incentivestypes "github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	lockupkeeper "github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	mintkeeper "github.com/osmosis-labs/osmosis/v7/x/mint/keeper"
	minttypes "github.com/osmosis-labs/osmosis/v7/x/mint/types"
	poolincentives "github.com/osmosis-labs/osmosis/v7/x/pool-incentives"
	poolincentiveskeeper "github.com/osmosis-labs/osmosis/v7/x/pool-incentives/keeper"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v7/x/pool-incentives/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid"
	superfluidkeeper "github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
	superfluidtypes "github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
	"github.com/osmosis-labs/osmosis/v7/x/txfees"
	txfeeskeeper "github.com/osmosis-labs/osmosis/v7/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

type appKeepers struct {
	// keepers, by order of initialization
	// "Special" keepers
	ParamsKeeper     *paramskeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	CrisisKeeper     *crisiskeeper.Keeper
	UpgradeKeeper    *upgradekeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
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
	TransferKeeper       *ibctransferkeeper.Keeper
	Bech32IBCKeeper      *bech32ibckeeper.Keeper
	Bech32ICS20Keeper    *bech32ics20keeper.Keeper
	EvidenceKeeper       *evidencekeeper.Keeper
	ClaimKeeper          *claimkeeper.Keeper
	GAMMKeeper           *gammkeeper.Keeper
	LockupKeeper         *lockupkeeper.Keeper
	EpochsKeeper         *epochskeeper.Keeper
	IncentivesKeeper     *incentiveskeeper.Keeper
	MintKeeper           *mintkeeper.Keeper
	PoolIncentivesKeeper *poolincentiveskeeper.Keeper
	TxFeesKeeper         *txfeeskeeper.Keeper
	FeeGrantKeeper       *feegrantkeeper.Keeper
	SuperfluidKeeper     *superfluidkeeper.Keeper
	GovKeeper            *govkeeper.Keeper
	WasmKeeper           *wasm.Keeper
}

func (app *OsmosisApp) InitSpecialKeepers(
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
) {
	appCodec := app.appCodec
	bApp := app.BaseApp
	cdc := app.cdc
	keys := app.keys
	tkeys := app.tkeys
	memKeys := app.memKeys

	paramsKeeper := app.initParamsKeeper(appCodec, cdc, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])
	app.ParamsKeeper = &paramsKeeper

	// set the BaseApp's parameter store
	bApp.SetParamStore(app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])
	app.ScopedIBCKeeper = app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	app.ScopedTransferKeeper = app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	app.ScopedWasmKeeper = app.CapabilityKeeper.ScopeToModule(wasm.ModuleName)
	app.CapabilityKeeper.Seal()

	// TODO: Make a SetInvCheckPeriod fn on CrisisKeeper.
	// IMO, its bad design atm that it requires this in state machine initialization
	crisisKeeper := crisiskeeper.NewKeeper(
		app.GetSubspace(crisistypes.ModuleName), invCheckPeriod, app.BankKeeper, authtypes.FeeCollectorName,
	)
	app.CrisisKeeper = &crisisKeeper

	upgradeKeeper := upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		bApp,
	)
	app.UpgradeKeeper = &upgradeKeeper
}

func (app *OsmosisApp) InitNormalKeepers(
	wasmDir string,
	wasmConfig wasm.Config,
	wasmEnabledProposals []wasm.ProposalType,
	wasmOpts []wasm.Option,
) {
	appCodec := app.appCodec
	bApp := app.BaseApp
	keys := app.keys

	// Add 'normal' keepers
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		app.GetSubspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount,
		maccPerms,
	)
	app.AccountKeeper = &accountKeeper
	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		app.GetSubspace(banktypes.ModuleName),
		app.BlockedAddrs(),
	)
	app.BankKeeper = &bankKeeper

	authzKeeper := authzkeeper.NewKeeper(
		keys[authzkeeper.StoreKey],
		appCodec,
		bApp.MsgServiceRouter(),
	)
	app.AuthzKeeper = &authzKeeper

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.GetSubspace(stakingtypes.ModuleName),
	)
	app.StakingKeeper = &stakingKeeper

	distrKeeper := distrkeeper.NewKeeper(
		appCodec, keys[distrtypes.StoreKey],
		app.GetSubspace(distrtypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authtypes.FeeCollectorName,
		app.BlockedAddrs(),
	)
	app.DistrKeeper = &distrKeeper

	slashingKeeper := slashingkeeper.NewKeeper(
		appCodec,
		keys[slashingtypes.StoreKey],
		app.StakingKeeper,
		app.GetSubspace(slashingtypes.ModuleName),
	)
	app.SlashingKeeper = &slashingKeeper

	// Create IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibchost.StoreKey],
		app.GetSubspace(ibchost.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		app.ScopedIBCKeeper,
	)

	// Create Transfer Keepers
	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.ScopedTransferKeeper,
	)
	app.TransferKeeper = &transferKeeper
	app.transferModule = transfer.NewAppModule(*app.TransferKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, app.transferModule)
	// Note: the sealing is done after creating wasmd and wiring that up

	app.Bech32IBCKeeper = bech32ibckeeper.NewKeeper(
		app.IBCKeeper.ChannelKeeper, appCodec, keys[bech32ibctypes.StoreKey],
		app.TransferKeeper,
	)

	// TODO: Should we be passing this instead of bank in many places?
	// Where do we want send coins to be cross-chain?
	app.Bech32ICS20Keeper = bech32ics20keeper.NewKeeper(
		app.IBCKeeper.ChannelKeeper,
		app.BankKeeper,
		app.TransferKeeper,
		app.Bech32IBCKeeper,
		app.TransferKeeper,
		appCodec,
	)

	// create evidence keeper with router
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = evidencekeeper.NewKeeper(
		appCodec,
		keys[evidencetypes.StoreKey],
		app.StakingKeeper,
		app.SlashingKeeper,
	)

	app.ClaimKeeper = claimkeeper.NewKeeper(
		appCodec,
		keys[claimtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper, app.StakingKeeper, app.DistrKeeper)

	gammKeeper := gammkeeper.NewKeeper(
		appCodec, keys[gammtypes.StoreKey],
		app.GetSubspace(gammtypes.ModuleName),
		app.AccountKeeper, app.BankKeeper, app.DistrKeeper)
	app.GAMMKeeper = &gammKeeper

	app.LockupKeeper = lockupkeeper.NewKeeper(
		appCodec,
		keys[lockuptypes.StoreKey],
		// TODO: Visit why this needs to be deref'd
		*app.AccountKeeper,
		app.BankKeeper,
		app.DistrKeeper)

	app.EpochsKeeper = epochskeeper.NewKeeper(appCodec, keys[epochstypes.StoreKey])

	app.IncentivesKeeper = incentiveskeeper.NewKeeper(
		appCodec,
		keys[incentivestypes.StoreKey],
		app.GetSubspace(incentivestypes.ModuleName),
		app.BankKeeper,
		app.LockupKeeper,
		app.EpochsKeeper,
	)

	app.SuperfluidKeeper = superfluidkeeper.NewKeeper(
		appCodec,
		keys[superfluidtypes.StoreKey],
		app.GetSubspace(superfluidtypes.ModuleName),
		*app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistrKeeper,
		app.EpochsKeeper,
		app.LockupKeeper,
		gammKeeper,
		app.IncentivesKeeper,
		lockupkeeper.NewMsgServerImpl(app.LockupKeeper),
	)

	mintKeeper := mintkeeper.NewKeeper(
		appCodec,
		keys[minttypes.StoreKey],
		app.GetSubspace(minttypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.DistrKeeper,
		app.EpochsKeeper,
		authtypes.FeeCollectorName,
	)
	app.MintKeeper = &mintKeeper

	poolIncentivesKeeper := poolincentiveskeeper.NewKeeper(
		appCodec,
		keys[poolincentivestypes.StoreKey],
		app.GetSubspace(poolincentivestypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.IncentivesKeeper,
		app.DistrKeeper,
		distrtypes.ModuleName,
		authtypes.FeeCollectorName,
	)
	app.PoolIncentivesKeeper = &poolIncentivesKeeper

	// Note: gammKeeper is expected to satisfy the SpotPriceCalculator interface parameter
	txFeesKeeper := txfeeskeeper.NewKeeper(
		appCodec,
		app.AccountKeeper,
		app.BankKeeper,
		app.EpochsKeeper,
		keys[txfeestypes.StoreKey],
		app.GAMMKeeper,
		app.GAMMKeeper,
		txfeestypes.FeeCollectorName,
		txfeestypes.NonNativeFeeCollectorName,
	)
	app.TxFeesKeeper = &txFeesKeeper

	feeGrantKeeper := feegrantkeeper.NewKeeper(
		appCodec,
		keys[txfeestypes.StoreKey],
		app.AccountKeeper,
	)
	app.FeeGrantKeeper = &feeGrantKeeper

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	supportedFeatures := "iterator,staking,stargate,osmosis"

	wasmOpts = append(owasm.RegisterCustomPlugins(app.GAMMKeeper, app.BankKeeper), wasmOpts...)

	wasmKeeper := wasm.NewKeeper(
		appCodec,
		keys[wasm.StoreKey],
		app.GetSubspace(wasm.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistrKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.ScopedWasmKeeper,
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		supportedFeatures,
		wasmOpts...,
	)
	app.WasmKeeper = &wasmKeeper

	// wire up x/wasm to IBC
	ibcRouter.AddRoute(wasm.ModuleName, wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper))
	app.IBCKeeper.SetRouter(ibcRouter)

	// register the proposal types
	// TODO: This appears to be missing tx fees proposal type
	govRouter := govtypes.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(*app.ParamsKeeper)).
		AddRoute(distrtypes.RouterKey, distribution.NewCommunityPoolSpendProposalHandler(*app.DistrKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(*app.UpgradeKeeper)).
		AddRoute(ibchost.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(poolincentivestypes.RouterKey, poolincentives.NewPoolIncentivesProposalHandler(*app.PoolIncentivesKeeper)).
		AddRoute(bech32ibctypes.RouterKey, bech32ibc.NewBech32IBCProposalHandler(*app.Bech32IBCKeeper)).
		AddRoute(txfeestypes.RouterKey, txfees.NewUpdateFeeTokenProposalHandler(*app.TxFeesKeeper)).
		AddRoute(superfluidtypes.RouterKey, superfluid.NewSuperfluidProposalHandler(*app.SuperfluidKeeper, *app.EpochsKeeper))

	// The gov proposal types can be individually enabled
	if len(wasmEnabledProposals) != 0 {
		govRouter.AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(app.WasmKeeper, wasmEnabledProposals))
	}

	govKeeper := govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey],
		app.GetSubspace(govtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, govRouter)
	app.GovKeeper = &govKeeper
}

func (app *OsmosisApp) SetupHooks() {
	// For every module that has hooks set on it,
	// you must check InitNormalKeepers to ensure that its not passed by de-reference
	// e.g. *app.StakingKeeper doesn't appear

	// Recall that SetHooks is a mutative call.
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			app.DistrKeeper.Hooks(),
			app.SlashingKeeper.Hooks(),
			app.ClaimKeeper.Hooks(),
			app.SuperfluidKeeper.Hooks(),
		),
	)

	app.GAMMKeeper.SetHooks(
		gammtypes.NewMultiGammHooks(
			// insert gamm hooks receivers here
			app.PoolIncentivesKeeper.Hooks(),
			app.ClaimKeeper.Hooks(),
		),
	)

	app.LockupKeeper.SetHooks(
		lockuptypes.NewMultiLockupHooks(
			// insert lockup hooks receivers here
			app.SuperfluidKeeper.Hooks(),
		),
	)

	app.IncentivesKeeper.SetHooks(
		incentivestypes.NewMultiIncentiveHooks(
		// insert incentive hooks receivers here
		),
	)

	app.MintKeeper.SetHooks(
		minttypes.NewMultiMintHooks(
			// insert mint hooks receivers here
			app.PoolIncentivesKeeper.Hooks(),
		),
	)

	app.EpochsKeeper.SetHooks(
		epochstypes.NewMultiEpochHooks(
			// insert epoch hooks receivers here
			app.TxFeesKeeper.Hooks(),
			app.SuperfluidKeeper.Hooks(),
			app.IncentivesKeeper.Hooks(),
			app.MintKeeper.Hooks(),
		),
	)

	app.GovKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
			// insert governance hooks receivers here
			app.ClaimKeeper.Hooks(),
		),
	)
}

// initParamsKeeper init params keeper and its subspaces.
func (app *OsmosisApp) initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey sdk.StoreKey) paramskeeper.Keeper {
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
	paramsKeeper.Subspace(incentivestypes.ModuleName)
	paramsKeeper.Subspace(poolincentivestypes.ModuleName)
	paramsKeeper.Subspace(superfluidtypes.ModuleName)
	paramsKeeper.Subspace(gammtypes.ModuleName)
	paramsKeeper.Subspace(wasm.ModuleName)

	return paramsKeeper
}

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
		upgradetypes.StoreKey,
		evidencetypes.StoreKey,
		ibctransfertypes.StoreKey,
		capabilitytypes.StoreKey,
		gammtypes.StoreKey,
		lockuptypes.StoreKey,
		claimtypes.StoreKey,
		incentivestypes.StoreKey,
		epochstypes.StoreKey,
		poolincentivestypes.StoreKey,
		authzkeeper.StoreKey,
		txfeestypes.StoreKey,
		superfluidtypes.StoreKey,
		bech32ibctypes.StoreKey,
		wasm.StoreKey,
	}
}
