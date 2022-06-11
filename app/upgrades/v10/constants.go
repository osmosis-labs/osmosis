package v10

import (
<<<<<<< HEAD
	"github.com/osmosis-labs/osmosis/v9/app/upgrades"

	sdk "github.com/cosmos/cosmos-sdk/types"
=======
	"github.com/osmosis-labs/osmosis/v7/app/upgrades"
>>>>>>> e7993a4 (Delete irregular state change logic (#1769))
)

// Last executed block on the v9 code was 4713064.
// Last committed block is assumed to be 4713064, as we have block proposals that were not precommitted upon
// for 4713065.
const ForkHeight = 4713065

// UpgradeName defines the on-chain upgrade name for the Osmosis v9-fork for recovery.
// This is not called v10, due to this bug that would require a state migration to fix:
const UpgradeName = "v10"

// Created synthetically via fork
// var Upgrade = upgrades.Upgrade{
// 	UpgradeName:          UpgradeName,
// 	CreateUpgradeHandler: CreateUpgradeHandler,
// 	StoreUpgrades:        store.StoreUpgrades{},
// }

var Fork = upgrades.Fork{
	UpgradeName:    UpgradeName,
	UpgradeHeight:  ForkHeight,
	BeginForkLogic: RunForkLogic,
}
