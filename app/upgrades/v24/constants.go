package v24

import (
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"

	"github.com/osmosis-labs/osmosis/v28/app/upgrades"

	store "cosmossdk.io/store/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v24 upgrade.
const UpgradeName = "v24"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{ibcwasmtypes.StoreKey, icacontrollertypes.StoreKey},
		Deleted: []string{},
	},
}

var WhitelistedFeeTokenSetters = []string{"osmo17eqe9dpglajwd48r65lasq3mftra5q4uxce525htyvjdp0q037vqpurhve"}
