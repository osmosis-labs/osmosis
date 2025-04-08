package v27

import (
	store "cosmossdk.io/store/types"
	"github.com/osmosis-labs/osmosis/v27/app/upgrades"
)

const UpgradeName = "v27"
const DistributionContractAddress = "symphony16jzpxp0e8550c9aht6q9svcux30vtyyyyxv5w2l2djjra46580wsq5hxxq"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
