package v31

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v30/app/upgrades"

	store "cosmossdk.io/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v31 upgrade.
const UpgradeName = "v31"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
	},
}

var (
	// All superfluid delegation will not increase there is no asset allowed to perform superfluid delegation
	// So it is ok to set the check for total undelegated amount to the current total amount
	// https://lcd.osmosis.zone/osmosis/superfluid/v1beta1/all_superfluid_delegations
	totalSuperfluidDelegationAmount = osmomath.NewInt(385298452)
)
