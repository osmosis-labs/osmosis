package v20

import (
	"github.com/osmosis-labs/osmosis/v20/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v20 upgrade.
const UpgradeName = "v20"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			crisistypes.ModuleName,
			consensustypes.ModuleName,
			group.ModuleName,
			nft.ModuleName,
		},
		Deleted: []string{},
	},
}
