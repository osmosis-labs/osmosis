package upgrades

import (
	store "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/v27/app/keepers"
)

// BaseAppParamManager defines an interrace that BaseApp is expected to fulfill
// that allows upgrade handlers to modify BaseApp parameters.
type BaseAppParamManager interface {
	GetConsensusParams(ctx sdk.Context) tmproto.ConsensusParams
	StoreConsensusParams(ctx sdk.Context, cp tmproto.ConsensusParams) error
}

// Upgrade defines a struct containing necessary fields that a SoftwareUpgradeProposal
// must have written, in order for the state migration to go smoothly.
// An upgrade must implement this struct, and then set it in the app.go.
// The app.go will then define the handler.
type Upgrade struct {
	// Upgrade version name, for the upgrade handler, e.g. `v7`
	UpgradeName string

	// CreateUpgradeHandler defines the function that creates an upgrade handler
	CreateUpgradeHandler func(*module.Manager, module.Configurator, BaseAppParamManager, *keepers.AppKeepers) upgradetypes.UpgradeHandler

	// Store upgrades, should be used for any new modules introduced, new modules deleted, or store names renamed.
	StoreUpgrades store.StoreUpgrades
}

// Fork defines a struct containing the requisite fields for a non-software upgrade proposal
// Hard Fork at a given height to implement.
// There is one time code that can be added for the start of the Fork, in `BeginForkLogic`.
// Any other change in the code should be height-gated, if the goal is to have old and new binaries
// to be compatible prior to the upgrade height.
type Fork struct {
	// Upgrade version name, for the upgrade handler, e.g. `v7`
	UpgradeName string
	// height the upgrade occurs at
	UpgradeHeight int64

	// Function that runs some custom state transition code at the beginning of a fork.
	BeginForkLogic func(ctx sdk.Context, keepers *keepers.AppKeepers)
}
