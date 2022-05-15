package app

import (
	// Utilities from the Cosmos-SDK other than Cosmos modules
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// Cosmos-SDK Modules
	// https://github.com/cosmos/cosmos-sdk/tree/master/x
	// NB: Osmosis uses a fork of the cosmos-sdk which can be found at: https://github.com/osmosis-labs/cosmos-sdk

	// Auth: Authentication of accounts and transactions for Cosmos SDK applications.
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	// Authz: Authorization for accounts to perform actions on behalf of other accounts.
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"

	// Bank: allows users to transfer tokens
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	// Capability: allows developers to atomically define what a module can and cannot do
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	// Crisis: Halting the blockchain under certain circumstances (e.g. if an invariant is broken).
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"

	// Distribution: Fee distribution, and staking token provision distribution.
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	// Evidence handling for double signing, misbehaviour, etc.
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"

	// Governance: Allows stakeholders to make decisions concering a Cosmos-SDK blockchain's economy and development
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	// Params: Parameters that are always available
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	// Slashing:
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	// Staking: Allows the Tendermint validator set to be chosen based on bonded stake.
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// Upgrade:  Software upgrades handling and coordination.
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	// IBC Transfer: Defines the "transfer" IBC port
	transfer "github.com/cosmos/ibc-go/v2/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v2/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v2/modules/core/05-port/types"

	// IBC: Inter-blockchain communication
	ibcclient "github.com/cosmos/ibc-go/v2/modules/core/02-client"
	ibchost "github.com/cosmos/ibc-go/v2/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v2/modules/core/keeper"

	// Upgrades from earlier versions of Osmosis
	_ "github.com/osmosis-labs/osmosis/v7/client/docs/statik"

	// Modules that live in the Osmosis repository and are specific to Osmosis
	claimkeeper "github.com/osmosis-labs/osmosis/v7/x/claim/keeper"
	claimtypes "github.com/osmosis-labs/osmosis/v7/x/claim/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid"
	"github.com/osmosis-labs/osmosis/v7/x/txfees"

	// Epochs: gives Osmosis a sense of "clock time" so that events can be based on days instead of "number of blocks"
	epochskeeper "github.com/osmosis-labs/osmosis/v7/x/epochs/keeper"
	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"

	// Generalized Automated Market Maker
	gammkeeper "github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"

	// Incentives: Allows Osmosis and foriegn chain communities to incentivize users to provide liquidity
	incentiveskeeper "github.com/osmosis-labs/osmosis/v7/x/incentives/keeper"
	incentivestypes "github.com/osmosis-labs/osmosis/v7/x/incentives/types"

	// Lockup: allows tokens to be locked (made non-transferrable)
	lockupkeeper "github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	// Mint: Our modified version of github.com/cosmos/cosmos-sdk/x/mint
	mintkeeper "github.com/osmosis-labs/osmosis/v7/x/mint/keeper"
	minttypes "github.com/osmosis-labs/osmosis/v7/x/mint/types"

	// Pool incentives:
	poolincentives "github.com/osmosis-labs/osmosis/v7/x/pool-incentives"
	poolincentiveskeeper "github.com/osmosis-labs/osmosis/v7/x/pool-incentives/keeper"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v7/x/pool-incentives/types"

	// Superfluid: Allows users to stake gamm (bonded liquidity)
	superfluidkeeper "github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
	superfluidtypes "github.com/osmosis-labs/osmosis/v7/x/superfluid/types"

	// Txfees: Allows Osmosis to charge transaction fees without harming IBC user experience
	txfeeskeeper "github.com/osmosis-labs/osmosis/v7/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/v7/x/txfees/types"

	// Wasm: Allows Osmosis to interact with web assembly smart contracts
	"github.com/CosmWasm/wasmd/x/wasm"

	// Modules related to bech32-ibc, which allows new ibc funcationality based on the bech32 prefix of addresses
	"github.com/osmosis-labs/bech32-ibc/x/bech32ibc"
	bech32ibckeeper "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/keeper"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"
	bech32ics20keeper "github.com/osmosis-labs/bech32-ibc/x/bech32ics20/keeper"
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

// Note: I put x/wasm here as I need to write it up to these other ones
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
		app.GetSubspace(distrtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, authtypes.FeeCollectorName, app.BlockedAddrs(),
	)
	app.DistrKeeper = &distrKeeper

	slashingKeeper := slashingkeeper.NewKeeper(
		appCodec, keys[slashingtypes.StoreKey], app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName),
	)
	app.SlashingKeeper = &slashingKeeper

	// Create IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibchost.StoreKey],
		app.GetSubspace(ibchost.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		app.ScopedIBCKeeper)

	// Create Transfer Keepers
	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec, keys[ibctransfertypes.StoreKey], app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
		app.AccountKeeper, app.BankKeeper, app.ScopedTransferKeeper,
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
		app.BankKeeper, app.TransferKeeper,
		app.Bech32IBCKeeper,
		app.TransferKeeper,
		appCodec,
	)

	// create evidence keeper with router
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = evidencekeeper.NewKeeper(
		appCodec, keys[evidencetypes.StoreKey], app.StakingKeeper, app.SlashingKeeper,
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
		appCodec, keys[lockuptypes.StoreKey],
		// TODO: Visit why this needs to be deref'd
		*app.AccountKeeper,
		app.BankKeeper,
		app.DistrKeeper)

	app.EpochsKeeper = epochskeeper.NewKeeper(appCodec, keys[epochstypes.StoreKey])

	app.IncentivesKeeper = incentiveskeeper.NewKeeper(
		appCodec, keys[incentivestypes.StoreKey],
		app.GetSubspace(incentivestypes.ModuleName),
		app.BankKeeper, app.LockupKeeper, app.EpochsKeeper)

	app.SuperfluidKeeper = superfluidkeeper.NewKeeper(
		appCodec, keys[superfluidtypes.StoreKey], app.GetSubspace(superfluidtypes.ModuleName),
		*app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.DistrKeeper, app.EpochsKeeper, app.LockupKeeper, app.GAMMKeeper, app.IncentivesKeeper,
		lockupkeeper.NewMsgServerImpl(app.LockupKeeper))

	mintKeeper := mintkeeper.NewKeeper(
		appCodec, keys[minttypes.StoreKey],
		app.GetSubspace(minttypes.ModuleName),
		app.AccountKeeper, app.BankKeeper, app.DistrKeeper, app.EpochsKeeper,
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

	txFeesKeeper := txfeeskeeper.NewKeeper(
		appCodec,
		keys[txfeestypes.StoreKey],
		app.GAMMKeeper,
	)
	app.TxFeesKeeper = &txFeesKeeper

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	supportedFeatures := "iterator,staking,stargate"
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

// initParamsKeeper init params keeper and its subspaces
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
