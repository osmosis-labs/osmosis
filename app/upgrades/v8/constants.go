package v8

import (
	"github.com/osmosis-labs/osmosis/v7/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v7 upgrade.
const UpgradeName = "v8"

// The historic name of the claims module, which is removed in this release.
// Cross-check against https://github.com/osmosis-labs/osmosis/blob/v7.2.0/x/claim/types/keys.go#L5
const ClaimsModuleName = "claim"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Deleted: []string{ClaimsModuleName},
	},
}
