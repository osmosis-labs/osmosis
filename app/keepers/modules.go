package keepers

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmclient "github.com/CosmWasm/wasmd/x/wasm/client"
	transfer "github.com/cosmos/ibc-go/v4/modules/apps/transfer"
	ibc "github.com/cosmos/ibc-go/v4/modules/core"
	ibcclientclient "github.com/cosmos/ibc-go/v4/modules/core/02-client/client"
	"github.com/strangelove-ventures/packet-forward-middleware/v4/router"

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
	ica "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts"
	icq "github.com/strangelove-ventures/async-icq/v4"

	_ "github.com/osmosis-labs/osmosis/v15/client/docs/statik"
	concentratedliquidity "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/clmodule"
	downtimemodule "github.com/osmosis-labs/osmosis/v15/x/downtime-detector/module"
	"github.com/osmosis-labs/osmosis/v15/x/epochs"
	"github.com/osmosis-labs/osmosis/v15/x/gamm"
	"github.com/osmosis-labs/osmosis/v15/x/ibc-rate-limit/ibcratelimitmodule"
	"github.com/osmosis-labs/osmosis/v15/x/incentives"
	"github.com/osmosis-labs/osmosis/v15/x/lockup"
	"github.com/osmosis-labs/osmosis/v15/x/mint"
	poolincentives "github.com/osmosis-labs/osmosis/v15/x/pool-incentives"
	poolincentivesclient "github.com/osmosis-labs/osmosis/v15/x/pool-incentives/client"
	poolmanager "github.com/osmosis-labs/osmosis/v15/x/poolmanager/module"
	"github.com/osmosis-labs/osmosis/v15/x/protorev"
	superfluid "github.com/osmosis-labs/osmosis/v15/x/superfluid"
	superfluidclient "github.com/osmosis-labs/osmosis/v15/x/superfluid/client"
	"github.com/osmosis-labs/osmosis/v15/x/tokenfactory"
	"github.com/osmosis-labs/osmosis/v15/x/twap/twapmodule"
	"github.com/osmosis-labs/osmosis/v15/x/txfees"
	valsetprefmodule "github.com/osmosis-labs/osmosis/v15/x/valset-pref/valpref-module"
	ibc_hooks "github.com/osmosis-labs/osmosis/x/ibc-hooks"
)

// AppModuleBasics returns ModuleBasics for the module BasicManager.
var AppModuleBasics = []module.AppModuleBasic{
	auth.AppModuleBasic{},
	genutil.AppModuleBasic{},
	bank.AppModuleBasic{},
	capability.AppModuleBasic{},
	staking.AppModuleBasic{},
	mint.AppModuleBasic{},
	downtimemodule.AppModuleBasic{},
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
	poolmanager.AppModuleBasic{},
	twapmodule.AppModuleBasic{},
	concentratedliquidity.AppModuleBasic{},
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
	icq.AppModuleBasic{},
	ica.AppModuleBasic{},
	ibc_hooks.AppModuleBasic{},
	ibcratelimitmodule.AppModuleBasic{},
	router.AppModuleBasic{},
}
