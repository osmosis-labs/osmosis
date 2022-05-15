package app

import (
	// Utilities from the Cosmos-SDK other than Cosmos modules
	"github.com/cosmos/cosmos-sdk/types/module"

	// Cosmos-SDK Modules
	// https://github.com/cosmos/cosmos-sdk/tree/master/x
	// NB: Osmosis uses a fork of the cosmos-sdk which can be found at: https://github.com/osmosis-labs/cosmos-sdk

	// Auth: Authentication of accounts and transactions for Cosmos SDK applications.
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	// Vesting: Allows the lock and periodic release of tokens
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	// Authz: Authorization for accounts to perform actions on behalf of other accounts.
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"

	// Bank: allows users to transfer tokens
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	// Capability: allows developers to atomically define what a module can and cannot do
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	// Crisis: Halting the blockchain under certain circumstances (e.g. if an invariant is broken).
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"

	// Distribution: Fee distribution, and staking token provision distribution.
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	// Evidence handling for double signing, misbehaviour, etc.
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"

	// Genesis Utilities: Used for evertything to do with the very first block of a chain
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	// Governance: Allows stakeholders to make decisions concering a Cosmos-SDK blockchain's economy and development
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	// Params: Parameters that are always available
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	// Slashing:
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	// Staking: Allows the Tendermint validator set to be chosen based on bonded stake.
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// Upgrade:  Software upgrades handling and coordination.
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	// IBC Transfer: Defines the "transfer" IBC port
	transfer "github.com/cosmos/ibc-go/v2/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"

	// IBC: Inter-blockchain communication
	ibc "github.com/cosmos/ibc-go/v2/modules/core"
	ibcclientclient "github.com/cosmos/ibc-go/v2/modules/core/02-client/client"
	ibchost "github.com/cosmos/ibc-go/v2/modules/core/24-host"

	// Osmosis application prarmeters
	appparams "github.com/osmosis-labs/osmosis/v7/app/params"

	// Upgrades from earlier versions of Osmosis
	_ "github.com/osmosis-labs/osmosis/v7/client/docs/statik"

	// Modules that live in the Osmosis repository and are specific to Osmosis
	"github.com/osmosis-labs/osmosis/v7/x/claim"
	claimtypes "github.com/osmosis-labs/osmosis/v7/x/claim/types"

	// Epochs: gives Osmosis a sense of "clock time" so that events can be based on days instead of "number of blocks"
	"github.com/osmosis-labs/osmosis/v7/x/epochs"
	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"

	// Generalized Automated Market Maker
	"github.com/osmosis-labs/osmosis/v7/x/gamm"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"

	// Incentives: Allows Osmosis and foriegn chain communities to incentivize users to provide liquidity
	"github.com/osmosis-labs/osmosis/v7/x/incentives"
	incentivestypes "github.com/osmosis-labs/osmosis/v7/x/incentives/types"

	// Lockup: allows tokens to be locked (made non-transferrable)
	"github.com/osmosis-labs/osmosis/v7/x/lockup"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	// Mint: Our modified version of github.com/cosmos/cosmos-sdk/x/mint
	"github.com/osmosis-labs/osmosis/v7/x/mint"
	minttypes "github.com/osmosis-labs/osmosis/v7/x/mint/types"

	// Pool incentives:
	poolincentives "github.com/osmosis-labs/osmosis/v7/x/pool-incentives"
	poolincentivesclient "github.com/osmosis-labs/osmosis/v7/x/pool-incentives/client"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v7/x/pool-incentives/types"

	// Superfluid: Allows users to stake gamm (bonded liquidity)
	superfluid "github.com/osmosis-labs/osmosis/v7/x/superfluid"
	superfluidclient "github.com/osmosis-labs/osmosis/v7/x/superfluid/client"
	superfluidtypes "github.com/osmosis-labs/osmosis/v7/x/superfluid/types"

	// txfees: Allows Osmosis to charge transaction fees without harming IBC user experience
	"github.com/osmosis-labs/osmosis/v7/x/txfees"
	txfeestypes "github.com/osmosis-labs/osmosis/v7/x/txfees/types"

	// Wasm: Allows Osmosis to interact with web assembly smart contracts
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmclient "github.com/CosmWasm/wasmd/x/wasm/client"

	// Modules related to bech32-ibc, which allows new ibc funcationality based on the bech32 prefix of addresses
	"github.com/osmosis-labs/bech32-ibc/x/bech32ibc"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"
	"github.com/osmosis-labs/bech32-ibc/x/bech32ics20"
)

// appModuleBasics returns ModuleBasics for the module BasicManager.
var appModuleBasics = []module.AppModuleBasic{
	auth.AppModuleBasic{},
	genutil.AppModuleBasic{},
	bank.AppModuleBasic{},
	capability.AppModuleBasic{},
	staking.AppModuleBasic{},
	mint.AppModuleBasic{},
	distr.AppModuleBasic{},
	gov.NewAppModuleBasic(
		append(
			wasmclient.ProposalHandlers,
			paramsclient.ProposalHandler, distrclient.ProposalHandler, upgradeclient.ProposalHandler, upgradeclient.CancelProposalHandler,
			poolincentivesclient.UpdatePoolIncentivesHandler,
			ibcclientclient.UpdateClientProposalHandler, ibcclientclient.UpgradeProposalHandler,
			superfluidclient.SetSuperfluidAssetsProposalHandler, superfluidclient.RemoveSuperfluidAssetsProposalHandler)...,
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
	superfluid.AppModuleBasic{},
	bech32ibc.AppModuleBasic{},
	wasm.AppModuleBasic{},
}

// module account permissions
var moduleAaccountPermissions = map[string][]string{
	authtypes.FeeCollectorName:               nil,
	distrtypes.ModuleName:                    nil,
	minttypes.ModuleName:                     {authtypes.Minter, authtypes.Burner},
	minttypes.DeveloperVestingModuleAcctName: nil,
	stakingtypes.BondedPoolName:              {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName:           {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:                      {authtypes.Burner},
	ibctransfertypes.ModuleName:              {authtypes.Minter, authtypes.Burner},
	claimtypes.ModuleName:                    {authtypes.Minter, authtypes.Burner},
	gammtypes.ModuleName:                     {authtypes.Minter, authtypes.Burner},
	incentivestypes.ModuleName:               {authtypes.Minter, authtypes.Burner},
	lockuptypes.ModuleName:                   {authtypes.Minter, authtypes.Burner},
	poolincentivestypes.ModuleName:           nil,
	superfluidtypes.ModuleName:               {authtypes.Minter, authtypes.Burner},
	txfeestypes.ModuleName:                   nil,
	wasm.ModuleName:                          {authtypes.Burner},
}

// appModules return modules to initlize module manager
func appModules(app *OsmosisApp, encodingConfig appparams.EncodingConfig, skipGenesisInvariants bool) []module.AppModule {
	appCodec := encodingConfig.Marshaler

	return []module.AppModule{
		genutil.NewAppModule(
			app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, *app.AccountKeeper, nil),
		vesting.NewAppModule(*app.AccountKeeper, app.BankKeeper),
		bech32ics20.NewAppModule(appCodec, *app.Bech32ICS20Keeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants),
		gov.NewAppModule(appCodec, *app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, *app.MintKeeper, app.AccountKeeper, app.BankKeeper),
		slashing.NewAppModule(appCodec, *app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper),
		distr.NewAppModule(appCodec, *app.DistrKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper),
		staking.NewAppModule(appCodec, *app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		upgrade.NewAppModule(*app.UpgradeKeeper),
		wasm.NewAppModule(appCodec, app.WasmKeeper, app.StakingKeeper),
		evidence.NewAppModule(*app.EvidenceKeeper),
		authzmodule.NewAppModule(appCodec, *app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(*app.ParamsKeeper),
		app.transferModule,
		claim.NewAppModule(appCodec, *app.ClaimKeeper),
		gamm.NewAppModule(appCodec, *app.GAMMKeeper, app.AccountKeeper, app.BankKeeper),
		txfees.NewAppModule(appCodec, *app.TxFeesKeeper),
		incentives.NewAppModule(appCodec, *app.IncentivesKeeper, app.AccountKeeper, app.BankKeeper, app.EpochsKeeper),
		lockup.NewAppModule(appCodec, *app.LockupKeeper, app.AccountKeeper, app.BankKeeper, app.keys),
		poolincentives.NewAppModule(appCodec, *app.PoolIncentivesKeeper),
		epochs.NewAppModule(appCodec, *app.EpochsKeeper),
		superfluid.NewAppModule(appCodec, *app.SuperfluidKeeper, app.AccountKeeper, app.BankKeeper,
			app.StakingKeeper, app.LockupKeeper, app.GAMMKeeper, app.EpochsKeeper),
		bech32ibc.NewAppModule(appCodec, *app.Bech32IBCKeeper),
	}
}

// orderBeginBlockers Tell the app's module manager how to set the order of BeginBlockers, which are run at the beginning of every block.
func orderBeginBlockers() []string {
	return []string{
		// Upgrades should be run _very_ first
		upgradetypes.ModuleName,
		// Note: epochs' begin should be "real" start of epochs, we keep epochs beginblock at the beginning
		epochstypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		poolincentivestypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		paramstypes.ModuleName, vestingtypes.ModuleName,
		gammtypes.ModuleName, incentivestypes.ModuleName, lockuptypes.ModuleName, claimtypes.ModuleName,
		poolincentivestypes.ModuleName,
		// superfluid must come after distribution and epochs
		superfluidtypes.ModuleName,
		bech32ibctypes.ModuleName, txfeestypes.ModuleName,
		wasm.ModuleName,
	}
}

// orderEndBlockers Tell the app's module manager how to set the order of EndBlockers, which are run at the end of every block.
var orderEndBlockers = []string{
	lockuptypes.ModuleName,
	crisistypes.ModuleName, govtypes.ModuleName, stakingtypes.ModuleName, claimtypes.ModuleName,
	capabilitytypes.ModuleName, authtypes.ModuleName, banktypes.ModuleName, distrtypes.ModuleName,
	slashingtypes.ModuleName, minttypes.ModuleName,
	genutiltypes.ModuleName, evidencetypes.ModuleName, authz.ModuleName,
	paramstypes.ModuleName, upgradetypes.ModuleName, vestingtypes.ModuleName,
	ibchost.ModuleName, ibctransfertypes.ModuleName,
	gammtypes.ModuleName, incentivestypes.ModuleName, lockuptypes.ModuleName,
	poolincentivestypes.ModuleName, superfluidtypes.ModuleName, bech32ibctypes.ModuleName, txfeestypes.ModuleName,
	// Note: epochs' endblock should be "real" end of epochs, we keep epochs endblock at the end
	epochstypes.ModuleName,
	wasm.ModuleName,
}

// modulesOrderInitGenesis returns module names in order for init genesis calls
var modulesOrderInitGenesis = []string{
	capabilitytypes.ModuleName,
	authtypes.ModuleName,
	banktypes.ModuleName,
	distrtypes.ModuleName,
	stakingtypes.ModuleName,
	slashingtypes.ModuleName,
	govtypes.ModuleName,
	minttypes.ModuleName,
	crisistypes.ModuleName,
	ibchost.ModuleName,
	gammtypes.ModuleName,
	txfeestypes.ModuleName,
	genutiltypes.ModuleName,
	evidencetypes.ModuleName,
	paramstypes.ModuleName,
	upgradetypes.ModuleName,
	vestingtypes.ModuleName,
	ibctransfertypes.ModuleName,
	bech32ibctypes.ModuleName, // comes after ibctransfertypes
	poolincentivestypes.ModuleName,
	superfluidtypes.ModuleName,
	claimtypes.ModuleName,
	incentivestypes.ModuleName,
	epochstypes.ModuleName,
	lockuptypes.ModuleName,
	authz.ModuleName,
	// wasm after ibc transfer
	wasm.ModuleName,
}

// simulationModules returns modules for simulation manager
func simulationModules(app *OsmosisApp, encodingConfig appparams.EncodingConfig, skipGenesisInvariants bool) []module.AppModuleSimulation {
	appCodec := encodingConfig.Marshaler

	return []module.AppModuleSimulation{
		auth.NewAppModule(appCodec, *app.AccountKeeper, authsims.RandomGenesisAccounts),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		authzmodule.NewAppModule(appCodec, *app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		gamm.NewAppModule(appCodec, *app.GAMMKeeper, app.AccountKeeper, app.BankKeeper),
		txfees.NewAppModule(appCodec, *app.TxFeesKeeper),
		gov.NewAppModule(appCodec, *app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, *app.MintKeeper, app.AccountKeeper, app.BankKeeper),
		slashing.NewAppModule(appCodec, *app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper),
		distr.NewAppModule(appCodec, *app.DistrKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper),
		staking.NewAppModule(appCodec, *app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		params.NewAppModule(*app.ParamsKeeper),
		evidence.NewAppModule(*app.EvidenceKeeper),
		wasm.NewAppModule(appCodec, app.WasmKeeper, app.StakingKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		incentives.NewAppModule(appCodec, *app.IncentivesKeeper, app.AccountKeeper, app.BankKeeper, app.EpochsKeeper),
		//		lockup.NewAppModule(appCodec, *app.LockupKeeper, app.AccountKeeper, app.BankKeeper, app.keys),
		poolincentives.NewAppModule(appCodec, *app.PoolIncentivesKeeper),
		epochs.NewAppModule(appCodec, *app.EpochsKeeper),
		superfluid.NewAppModule(appCodec, *app.SuperfluidKeeper, app.AccountKeeper, app.BankKeeper,
			app.StakingKeeper, app.LockupKeeper, app.GAMMKeeper, app.EpochsKeeper),
		app.transferModule,
	}
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *OsmosisApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}
