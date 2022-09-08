package v12

import (
	"github.com/osmosis-labs/osmosis/v11/app/upgrades"
)

// Last executed block on the v11 code was TODO.
// TODO: UPDATE ME
const ForkHeight = 100

// UpgradeName defines the on-chain upgrade name for the Osmosis v12-fork for recovery.
const UpgradeName = "v12"

var Fork = upgrades.Fork{
	UpgradeName:    UpgradeName,
	UpgradeHeight:  ForkHeight,
	BeginForkLogic: RunForkLogic,
}
