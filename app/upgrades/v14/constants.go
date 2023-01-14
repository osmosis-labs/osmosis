package v14

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	ibchookstypes "github.com/osmosis-labs/osmosis/x/ibc-hooks/types"

	"github.com/osmosis-labs/osmosis/v14/app/upgrades"
	cltypes "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
	downtimetypes "github.com/osmosis-labs/osmosis/v14/x/downtime-detector/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v14/x/protorev/types"
	valsetpreftypes "github.com/osmosis-labs/osmosis/v14/x/valset-pref/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v14 upgrade.
const UpgradeName = "v14"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{valsetpreftypes.StoreKey, protorevtypes.StoreKey, poolmanagertypes.StoreKey, downtimetypes.StoreKey, ibchookstypes.StoreKey, cltypes.StoreKey},
		Deleted: []string{},
	},
}
