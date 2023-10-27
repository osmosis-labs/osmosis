package v21

import (
	buildertypes "github.com/skip-mev/pob/x/builder/types"

	"github.com/osmosis-labs/osmosis/v20/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v21 upgrade.
const UpgradeName = "v21"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			// new modules
			buildertypes.ModuleName,

			// v47 modules
			crisistypes.ModuleName,
			consensustypes.ModuleName,
			feegranttypes.ModuleName,
			group.ModuleName,
			nft.ModuleName,
		},
		Deleted: []string{},
	},
}
