package v22

import (
	"github.com/osmosis-labs/osmosis/v21/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"

	authenticatortypes "github.com/osmosis-labs/osmosis/v21/x/authenticator/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v22 upgrade.
const (
	UpgradeName = "v22"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			authenticatortypes.ManagerStoreKey,
			authenticatortypes.AuthenticatorStoreKey,
		},
		Deleted: []string{},
	},
}
