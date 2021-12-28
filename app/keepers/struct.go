package keepers

import (
	// TODO: Find a coherent way to organize these imports
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
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/ibc-go/v2/modules/apps/transfer"

	ibctransferkeeper "github.com/cosmos/ibc-go/v2/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	porttypes "github.com/cosmos/ibc-go/v2/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v2/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v2/modules/core/keeper"

	appparams "github.com/osmosis-labs/osmosis/app/params"
	_ "github.com/osmosis-labs/osmosis/client/docs/statik"
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
	poolincentives "github.com/osmosis-labs/osmosis/x/pool-incentives"
	poolincentiveskeeper "github.com/osmosis-labs/osmosis/x/pool-incentives/keeper"
	poolincentivestypes "github.com/osmosis-labs/osmosis/x/pool-incentives/types"
	txfeeskeeper "github.com/osmosis-labs/osmosis/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/x/txfees/types"

	"github.com/osmosis-labs/bech32-ibc/x/bech32ibc"
	bech32ibckeeper "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/keeper"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"
	bech32ics20keeper "github.com/osmosis-labs/bech32-ibc/x/bech32ics20/keeper"

	ibcclient "github.com/cosmos/ibc-go/v2/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"

	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

// AppKeepers is a struct of a pointer to every keeper in Osmosis.
type AppKeepers struct {
	// keepers, by order of initialization
	// "Special" keepers
	ParamsKeeper     *paramskeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	CrisisKeeper     *crisiskeeper.Keeper
	UpgradeKeeper    *upgradekeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper

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
	GovKeeper            *govkeeper.Keeper
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (keepers AppKeepers) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, ok := keepers.ParamsKeeper.GetSubspace(moduleName)
	if !ok {
		panic("We did not register a subspace that was expected")
	}
	return subspace
}

func InitEmptyKeepers() AppKeepers {
	return AppKeepers{}
}

func InitSpecialKeepers(
	keepers *AppKeepers,
	bApp *baseapp.BaseApp,
	encodingConfig appparams.EncodingConfig,
	legacyAmino *codec.LegacyAmino,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
) {
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	keys := AllKVStoreKeys()
	tkeys := AllTStoreKeys()
	memKeys := AllMemStoreKeys()

	paramsKeeper := initParamsKeeper(appCodec, cdc, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])
	keepers.ParamsKeeper = &paramsKeeper

	// set the BaseApp's parameter store
	bApp.SetParamStore(keepers.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))

	// add capability keeper and ScopeToModule for ibc module
	keepers.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])
	keepers.ScopedIBCKeeper = keepers.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	keepers.ScopedTransferKeeper = keepers.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	keepers.CapabilityKeeper.Seal()

	// TODO: Make a SetInvCheckPeriod fn on CrisisKeeper.
	// IMO, its bad design atm that it requires this in state machine initialization
	crisisKeeper := crisiskeeper.NewKeeper(
		keepers.GetSubspace(crisistypes.ModuleName), invCheckPeriod, keepers.BankKeeper, authtypes.FeeCollectorName,
	)
	keepers.CrisisKeeper = &crisisKeeper

	upgradeKeeper := upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		bApp,
	)
	keepers.UpgradeKeeper = &upgradeKeeper
}

func InitNormalKeepers(
	keepers *AppKeepers,
	bApp *baseapp.BaseApp,
	encodingConfig appparams.EncodingConfig,
	legacyAmino *codec.LegacyAmino,
	// TODO: We should move BlockedAddrs to be a method on the bank keeper that can get configured independently
	blockedAddrs map[string]bool) {

	appCodec := encodingConfig.Marshaler
	keys := AllKVStoreKeys()
	// tkeys := AllTStoreKeys()
	// memKeys := AllMemStoreKeys()

	// Add 'normal' keepers
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		keepers.GetSubspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount,
		MaccPerms,
	)
	keepers.AccountKeeper = &accountKeeper
	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		keepers.AccountKeeper,
		keepers.GetSubspace(banktypes.ModuleName),
		blockedAddrs,
	)
	keepers.BankKeeper = &bankKeeper

	authzKeeper := authzkeeper.NewKeeper(
		keys[authzkeeper.StoreKey],
		appCodec,
		bApp.MsgServiceRouter(),
	)
	keepers.AuthzKeeper = &authzKeeper

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		keepers.AccountKeeper,
		keepers.BankKeeper,
		keepers.GetSubspace(stakingtypes.ModuleName),
	)
	keepers.StakingKeeper = &stakingKeeper

	distrKeeper := distrkeeper.NewKeeper(
		appCodec, keys[distrtypes.StoreKey],
		keepers.GetSubspace(distrtypes.ModuleName), keepers.AccountKeeper, keepers.BankKeeper,
		&stakingKeeper, authtypes.FeeCollectorName, blockedAddrs,
	)
	keepers.DistrKeeper = &distrKeeper

	slashingKeeper := slashingkeeper.NewKeeper(
		appCodec, keys[slashingtypes.StoreKey], &stakingKeeper, keepers.GetSubspace(slashingtypes.ModuleName),
	)
	keepers.SlashingKeeper = &slashingKeeper

	// Create IBC Keeper
	keepers.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibchost.StoreKey],
		keepers.GetSubspace(ibchost.ModuleName),
		&stakingKeeper,
		keepers.UpgradeKeeper,
		keepers.ScopedIBCKeeper)

	// Create Transfer Keepers
	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec, keys[ibctransfertypes.StoreKey], keepers.GetSubspace(ibctransfertypes.ModuleName),
		keepers.IBCKeeper.ChannelKeeper, &keepers.IBCKeeper.PortKeeper,
		keepers.AccountKeeper, keepers.BankKeeper, keepers.ScopedTransferKeeper,
	)
	keepers.TransferKeeper = &transferKeeper
	transferModule := transfer.NewAppModule(*keepers.TransferKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferModule)
	keepers.IBCKeeper.SetRouter(ibcRouter)

	keepers.Bech32IBCKeeper = bech32ibckeeper.NewKeeper(
		keepers.IBCKeeper.ChannelKeeper, appCodec, keys[bech32ibctypes.StoreKey],
		keepers.TransferKeeper,
	)

	// TODO: Should we be passing this instead of bank in many places?
	// Where do we want send coins to be cross-chain?
	keepers.Bech32ICS20Keeper = bech32ics20keeper.NewKeeper(
		keepers.IBCKeeper.ChannelKeeper,
		keepers.BankKeeper, keepers.TransferKeeper,
		keepers.Bech32IBCKeeper,
		keepers.TransferKeeper,
		appCodec,
	)

	// create evidence keeper with router
	// If evidence needs to be handled for the app, set routes in router here and seal
	keepers.EvidenceKeeper = evidencekeeper.NewKeeper(
		appCodec, keys[evidencetypes.StoreKey], keepers.StakingKeeper, keepers.SlashingKeeper,
	)

	keepers.ClaimKeeper = claimkeeper.NewKeeper(
		appCodec,
		keys[claimtypes.StoreKey],
		keepers.AccountKeeper,
		keepers.BankKeeper, keepers.StakingKeeper, keepers.DistrKeeper)

	gammKeeper := gammkeeper.NewKeeper(
		appCodec, keys[gammtypes.StoreKey],
		keepers.GetSubspace(gammtypes.ModuleName),
		keepers.AccountKeeper, keepers.BankKeeper, keepers.DistrKeeper)
	keepers.GAMMKeeper = &gammKeeper

	keepers.LockupKeeper = lockupkeeper.NewKeeper(
		appCodec, keys[lockuptypes.StoreKey],
		// TODO: Visit why this needs to be deref'd
		*keepers.AccountKeeper,
		keepers.BankKeeper)

	keepers.EpochsKeeper = epochskeeper.NewKeeper(appCodec, keys[epochstypes.StoreKey])

	keepers.IncentivesKeeper = incentiveskeeper.NewKeeper(
		appCodec, keys[incentivestypes.StoreKey],
		keepers.GetSubspace(incentivestypes.ModuleName),
		*keepers.AccountKeeper,
		keepers.BankKeeper, keepers.LockupKeeper, keepers.EpochsKeeper)

	mintKeeper := mintkeeper.NewKeeper(
		appCodec, keys[minttypes.StoreKey],
		keepers.GetSubspace(minttypes.ModuleName),
		keepers.AccountKeeper, keepers.BankKeeper, keepers.DistrKeeper, keepers.EpochsKeeper,
		authtypes.FeeCollectorName,
	)
	keepers.MintKeeper = &mintKeeper

	poolIncentivesKeeper := poolincentiveskeeper.NewKeeper(
		appCodec,
		keys[poolincentivestypes.StoreKey],
		keepers.GetSubspace(poolincentivestypes.ModuleName),
		keepers.AccountKeeper,
		keepers.BankKeeper,
		keepers.IncentivesKeeper,
		keepers.DistrKeeper,
		distrtypes.ModuleName,
		authtypes.FeeCollectorName,
	)
	keepers.PoolIncentivesKeeper = &poolIncentivesKeeper

	txFeesKeeper := txfeeskeeper.NewKeeper(
		appCodec,
		keys[txfeestypes.StoreKey],
		keepers.GAMMKeeper,
	)
	keepers.TxFeesKeeper = &txFeesKeeper

	// register the proposal types
	// TODO: This appears to be missing tx fees proposal type
	govRouter := govtypes.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(*keepers.ParamsKeeper)).
		AddRoute(distrtypes.RouterKey, distribution.NewCommunityPoolSpendProposalHandler(*keepers.DistrKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(keepers.IBCKeeper.ClientKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(*keepers.UpgradeKeeper)).
		AddRoute(ibchost.RouterKey, ibcclient.NewClientProposalHandler(keepers.IBCKeeper.ClientKeeper)).
		AddRoute(poolincentivestypes.RouterKey, poolincentives.NewPoolIncentivesProposalHandler(*keepers.PoolIncentivesKeeper)).
		AddRoute(bech32ibctypes.RouterKey, bech32ibc.NewBech32IBCProposalHandler(*keepers.Bech32IBCKeeper))

	govKeeper := govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey],
		keepers.GetSubspace(govtypes.ModuleName), keepers.AccountKeeper, keepers.BankKeeper,
		keepers.StakingKeeper, govRouter)
	keepers.GovKeeper = &govKeeper
}

func SetupHooks(keepers *AppKeepers) {
	// For every module that has hooks set on it,
	// you must check InitNormalKeepers to ensure that its not passed by de-reference
	// e.g. *keepers.StakingKeeper doesn't appear

	// Recall that SetHooks is a mutative call.
	keepers.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			keepers.DistrKeeper.Hooks(),
			keepers.SlashingKeeper.Hooks(),
			keepers.ClaimKeeper.Hooks()),
	)

	keepers.GAMMKeeper.SetHooks(
		gammtypes.NewMultiGammHooks(
			// insert gamm hooks receivers here
			keepers.PoolIncentivesKeeper.Hooks(),
			keepers.ClaimKeeper.Hooks(),
		),
	)

	keepers.LockupKeeper.SetHooks(
		lockuptypes.NewMultiLockupHooks(
		// insert lockup hooks receivers here
		),
	)

	keepers.IncentivesKeeper.SetHooks(
		incentivestypes.NewMultiIncentiveHooks(
		// insert incentive hooks receivers here
		),
	)

	keepers.MintKeeper.SetHooks(
		minttypes.NewMultiMintHooks(
			// insert mint hooks receivers here
			keepers.PoolIncentivesKeeper.Hooks(),
		),
	)

	keepers.EpochsKeeper.SetHooks(
		epochstypes.NewMultiEpochHooks(
			// insert epoch hooks receivers here
			keepers.IncentivesKeeper.Hooks(),
			keepers.MintKeeper.Hooks(),
		),
	)

	keepers.GovKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
			// insert governance hooks receivers here
			keepers.ClaimKeeper.Hooks(),
		),
	)
}

func NewBlankKeepers() AppKeepers {
	return AppKeepers{}
}

// initParamsKeeper init params keeper and its subspaces
// TODO: Figure out why this doesn't include every module.
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey sdk.StoreKey) paramskeeper.Keeper {
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
	paramsKeeper.Subspace(gammtypes.ModuleName)

	return paramsKeeper
}
