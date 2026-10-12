package v32

import (
	"github.com/osmosis-labs/osmosis/v31/app/upgrades"

	store "cosmossdk.io/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v32 upgrade.
const UpgradeName = "v32"

// LiquiditySubDAOAddress is the DAO DAO core contract of the liquidity sub-DAO.
// Non-base-denom balances stuck in the protorev module account are sent here so the
// sub-DAO can swap them to OSMO and burn it.
// https://daodao.zone/dao/osmo1rvq5cq2j35k7sqqz49e5e8zezl45fcywcawazh46qnc0g96d0d6sasqsgc
const LiquiditySubDAOAddress = "osmo1rvq5cq2j35k7sqqz49e5e8zezl45fcywcawazh46qnc0g96d0d6sasqsgc"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
	},
}
