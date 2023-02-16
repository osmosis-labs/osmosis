package v15

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	packetforwardtypes "github.com/strangelove-ventures/packet-forward-middleware/v4/router/types"

	"github.com/osmosis-labs/osmosis/v14/app/upgrades"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v14/x/protorev/types"
	valsetpreftypes "github.com/osmosis-labs/osmosis/v14/x/valset-pref/types"
	icqtypes "github.com/strangelove-ventures/async-icq/v4/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v15 upgrade.
const UpgradeName = "v15"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{poolmanagertypes.StoreKey, valsetpreftypes.StoreKey, protorevtypes.StoreKey, icqtypes.StoreKey, packetforwardtypes.StoreKey},
		Deleted: []string{},
	},
}
