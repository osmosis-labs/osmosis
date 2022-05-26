package app

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
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
	claimkeeper "github.com/osmosis-labs/osmosis/x/claim/keeper"
	claimtypes "github.com/osmosis-labs/osmosis/x/claim/types"
	epochskeeper "github.com/osmosis-labs/osmosis/x/epochs/keeper"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	gammkeeper "github.com/osmosis-labs/osmosis/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
	incentiveskeeper "github.com/osmosis-labs/osmosis/x/incentives/keeper"
	incentivestypes "github.com/osmosis-labs/osmosis/x/incentives/types"
	lockupkeeper "github.com/osmosis-labs/osmosis/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	mintkeeper "github.com/osmosis-labs/osmosis/x/mint/keeper"
	minttypes "github.com/osmosis-labs/osmosis/x/mint/types"
	"github.com/osmosis-labs/osmosis/x/osmolbp"
	osmolbpkeeper "github.com/osmosis-labs/osmosis/x/osmolbp/keeper"
	poolincentives "github.com/osmosis-labs/osmosis/x/pool-incentives"
	poolincentiveskeeper "github.com/osmosis-labs/osmosis/x/pool-incentives/keeper"
	poolincentivestypes "github.com/osmosis-labs/osmosis/x/pool-incentives/types"
	superfluidkeeper "github.com/osmosis-labs/osmosis/x/superfluid/keeper"
	superfluidtypes "github.com/osmosis-labs/osmosis/x/superfluid/types"
	"github.com/osmosis-labs/osmosis/x/txfees"
	txfeeskeeper "github.com/osmosis-labs/osmosis/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/x/txfees/types"
)

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

	paramsKeeper := initParamsKeeper(appCodec, cdc, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])
	app.ParamsKeeper = &paramsKeeper

	// set the BaseApp's parameter store
	bApp.SetParamStore(app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])
	app.ScopedIBCKeeper = app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	app.ScopedTransferKeeper = app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
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

func (app *OsmosisApp) InitNormalKeepers() {
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
	app.IBCKeeper.SetRouter(ibcRouter)

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

	app.SuperfluidKeeper = *superfluidkeeper.NewKeeper(
		appCodec, keys[superfluidtypes.StoreKey], app.GetSubspace(superfluidtypes.ModuleName),
		*app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.DistrKeeper, app.EpochsKeeper, app.LockupKeeper, gammKeeper, app.IncentivesKeeper)

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

	app.OsmolbpKeeper = osmolbpkeeper.NewKeeper(keys[osmolbpkeeper.StoreKey], appCodec, bankKeeper, app.GetSubspace(osmolbp.ModuleName))

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
		AddRoute(txfeestypes.RouterKey, txfees.NewUpdateFeeTokenProposalHandler(*app.TxFeesKeeper))

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
