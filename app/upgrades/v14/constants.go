package v14

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	routertypes "github.com/strangelove-ventures/packet-forward-middleware/v4/router/types"

	ibchookstypes "github.com/osmosis-labs/osmosis/x/ibc-hooks/types"

	"github.com/osmosis-labs/osmosis/v13/app/upgrades"
	cltypes "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	downtimetypes "github.com/osmosis-labs/osmosis/v13/x/downtime-detector/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v13/x/protorev/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
	valsetpreftypes "github.com/osmosis-labs/osmosis/v13/x/valset-pref/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v14 upgrade.
const UpgradeName = "v14"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{valsetpreftypes.StoreKey, protorevtypes.StoreKey, swaproutertypes.StoreKey, downtimetypes.StoreKey, ibchookstypes.StoreKey, cltypes.StoreKey, routertypes.StoreKey},
		Deleted: []string{},
	},
}
