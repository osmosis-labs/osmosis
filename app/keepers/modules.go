package keepers

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmclient "github.com/CosmWasm/wasmd/x/wasm/client"
	transfer "github.com/cosmos/ibc-go/v3/modules/apps/transfer"
	ibc "github.com/cosmos/ibc-go/v3/modules/core"
	ibcclientclient "github.com/cosmos/ibc-go/v3/modules/core/02-client/client"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	ica "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts"

	_ "github.com/osmosis-labs/osmosis/v13/client/docs/statik"
	"github.com/osmosis-labs/osmosis/v13/x/epochs"
	"github.com/osmosis-labs/osmosis/v13/x/gamm"
	ibc_hooks "github.com/osmosis-labs/osmosis/v13/x/ibc-hooks"
	ibc_rate_limit "github.com/osmosis-labs/osmosis/v13/x/ibc-rate-limit"
	"github.com/osmosis-labs/osmosis/v13/x/incentives"
	"github.com/osmosis-labs/osmosis/v13/x/lockup"
	"github.com/osmosis-labs/osmosis/v13/x/mint"
	poolincentives "github.com/osmosis-labs/osmosis/v13/x/pool-incentives"
	poolincentivesclient "github.com/osmosis-labs/osmosis/v13/x/pool-incentives/client"
	"github.com/osmosis-labs/osmosis/v13/x/protorev"
	superfluid "github.com/osmosis-labs/osmosis/v13/x/superfluid"
	superfluidclient "github.com/osmosis-labs/osmosis/v13/x/superfluid/client"
	swaprouter "github.com/osmosis-labs/osmosis/v13/x/swaprouter/module"
	"github.com/osmosis-labs/osmosis/v13/x/tokenfactory"
	"github.com/osmosis-labs/osmosis/v13/x/twap/twapmodule"
	"github.com/osmosis-labs/osmosis/v13/x/txfees"
	valsetprefmodule "github.com/osmosis-labs/osmosis/v13/x/valset-pref/valpref-module"
)

// AppModuleBasics returns ModuleBasics for the module BasicManager.
var AppModuleBasics = []module.AppModuleBasic{
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
			paramsclient.ProposalHandler,
			distrclient.ProposalHandler,
			upgradeclient.ProposalHandler,
			upgradeclient.CancelProposalHandler,
			poolincentivesclient.UpdatePoolIncentivesHandler,
			poolincentivesclient.ReplacePoolIncentivesHandler,
			ibcclientclient.UpdateClientProposalHandler,
			ibcclientclient.UpgradeProposalHandler,
			superfluidclient.SetSuperfluidAssetsProposalHandler,
			superfluidclient.RemoveSuperfluidAssetsProposalHandler,
			superfluidclient.UpdateUnpoolWhitelistProposalHandler,
		)...,
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
	swaprouter.AppModuleBasic{},
	twapmodule.AppModuleBasic{},
	protorev.AppModuleBasic{},
	txfees.AppModuleBasic{},
	incentives.AppModuleBasic{},
	lockup.AppModuleBasic{},
	poolincentives.AppModuleBasic{},
	epochs.AppModuleBasic{},
	superfluid.AppModuleBasic{},
	tokenfactory.AppModuleBasic{},
	valsetprefmodule.AppModuleBasic{},
	wasm.AppModuleBasic{},
	ica.AppModuleBasic{},
	ibc_hooks.AppModuleBasic{},
	ibc_rate_limit.AppModuleBasic{},
}
