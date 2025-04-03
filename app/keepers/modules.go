package keepers

// UNFORKING v2 TODO: Eventually should get rid of this in favor of NewBasicManagerFromManager
// Right now is strictly used for default genesis creation and registering codecs prior to app init
// Unclear to me how to use NewBasicManagerFromManager for this purpose though prior to app init
import (
	"github.com/CosmWasm/wasmd/x/wasm"
	packetforward "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	transfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	tendermint "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	stablestakingincentives "github.com/osmosis-labs/osmosis/v27/x/stable-staking-incentives"

	"cosmossdk.io/x/evidence"
	"cosmossdk.io/x/upgrade"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	icq "github.com/cosmos/ibc-apps/modules/async-icq/v8"
	"github.com/cosmos/ibc-go/modules/capability"
	ibcwasm "github.com/cosmos/ibc-go/modules/light-clients/08-wasm"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"

	"github.com/cosmos/cosmos-sdk/x/consensus"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	smartaccount "github.com/osmosis-labs/osmosis/v27/x/smart-account"

	"github.com/skip-mev/block-sdk/v2/x/auction"

	_ "github.com/osmosis-labs/osmosis/v27/client/docs/statik"
	clclient "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/client"
	concentratedliquidity "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/clmodule"
	cwpoolclient "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/client"
	cosmwasmpoolmodule "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/module"
	downtimemodule "github.com/osmosis-labs/osmosis/v27/x/downtime-detector/module"
	"github.com/osmosis-labs/osmosis/v27/x/epochs"
	"github.com/osmosis-labs/osmosis/v27/x/gamm"
	gammclient "github.com/osmosis-labs/osmosis/v27/x/gamm/client"
	"github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/ibcratelimitmodule"
	"github.com/osmosis-labs/osmosis/v27/x/incentives"
	incentivesclient "github.com/osmosis-labs/osmosis/v27/x/incentives/client"
	"github.com/osmosis-labs/osmosis/v27/x/lockup"
	"github.com/osmosis-labs/osmosis/v27/x/market"
	"github.com/osmosis-labs/osmosis/v27/x/mint"
	"github.com/osmosis-labs/osmosis/v27/x/oracle"
	poolincentives "github.com/osmosis-labs/osmosis/v27/x/pool-incentives"
	poolincentivesclient "github.com/osmosis-labs/osmosis/v27/x/pool-incentives/client"
	poolmanagerclient "github.com/osmosis-labs/osmosis/v27/x/poolmanager/client"
	poolmanager "github.com/osmosis-labs/osmosis/v27/x/poolmanager/module"
	"github.com/osmosis-labs/osmosis/v27/x/protorev"
	superfluid "github.com/osmosis-labs/osmosis/v27/x/superfluid"
	superfluidclient "github.com/osmosis-labs/osmosis/v27/x/superfluid/client"
	"github.com/osmosis-labs/osmosis/v27/x/tokenfactory"
	"github.com/osmosis-labs/osmosis/v27/x/treasury"
	"github.com/osmosis-labs/osmosis/v27/x/twap/twapmodule"
	"github.com/osmosis-labs/osmosis/v27/x/txfees"
	valsetprefmodule "github.com/osmosis-labs/osmosis/v27/x/valset-pref/valpref-module"
	ibc_hooks "github.com/osmosis-labs/osmosis/x/ibc-hooks"
)

// AppModuleBasics returns ModuleBasics for the module BasicManager.
var AppModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
	bank.AppModuleBasic{},
	capability.AppModuleBasic{},
	staking.AppModuleBasic{},
	mint.AppModuleBasic{},
	downtimemodule.AppModuleBasic{},
	distr.AppModuleBasic{},
	gov.NewAppModuleBasic(
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
			poolmanagerclient.DenomPairTakerFeeProposalHandler,
			incentivesclient.HandleCreateGroupsProposal,
		},
	),
	params.AppModuleBasic{},
	crisis.AppModuleBasic{},
	slashing.AppModuleBasic{},
	authzmodule.AppModuleBasic{},
	consensus.AppModuleBasic{},
	ibc.AppModuleBasic{},
	upgrade.AppModuleBasic{},
	evidence.AppModuleBasic{},
	transfer.AppModuleBasic{},
	vesting.AppModuleBasic{},
	gamm.AppModuleBasic{},
	poolmanager.AppModuleBasic{},
	oracle.AppModuleBasic{},
	market.AppModuleBasic{},
	treasury.AppModuleBasic{},
	twapmodule.AppModuleBasic{},
	concentratedliquidity.AppModuleBasic{},
	protorev.AppModuleBasic{},
	txfees.AppModuleBasic{},
	incentives.AppModuleBasic{},
	lockup.AppModuleBasic{},
	poolincentives.AppModuleBasic{},
	stablestakingincentives.AppModuleBasic{},
	epochs.AppModuleBasic{},
	superfluid.AppModuleBasic{},
	tokenfactory.AppModuleBasic{},
	valsetprefmodule.AppModuleBasic{},
	wasm.AppModuleBasic{},
	icq.AppModuleBasic{},
	ica.AppModuleBasic{},
	ibc_hooks.AppModuleBasic{},
	ibcratelimitmodule.AppModuleBasic{},
	ibcwasm.AppModuleBasic{},
	packetforward.AppModuleBasic{},
	cosmwasmpoolmodule.AppModuleBasic{},
	tendermint.AppModuleBasic{},
	auction.AppModuleBasic{},
	smartaccount.AppModuleBasic{},
)
