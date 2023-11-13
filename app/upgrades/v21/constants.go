package v21

import (
	auctiontypes "github.com/skip-mev/block-sdk/x/auction/types"

	"github.com/osmosis-labs/osmosis/v20/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v21 upgrade.
const (
	UpgradeName    = "v21"
	TestingChainId = "testing-chain-id"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			// new modules
			auctiontypes.ModuleName,

			// v47 modules
			crisistypes.ModuleName,
			consensustypes.ModuleName,
		},
		Deleted: []string{},
	},
}
