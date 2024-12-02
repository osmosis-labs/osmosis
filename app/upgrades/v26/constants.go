package v26

import (
	"github.com/osmosis-labs/osmosis/v28/app/upgrades"

	storetypes "cosmossdk.io/store/types"
	circuittypes "cosmossdk.io/x/circuit/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v26 upgrade.
const (
	UpgradeName = "v26"

	// MaximumUnauthenticatedGas for smart account transactions to verify the fee payer
	MaximumUnauthenticatedGas = uint64(250_000)

	// BlockMaxBytes is the max bytes for a block, 3mb
	BlockMaxBytes = int64(3000000)

	// BlockMaxGas is the max gas allowed in a block
	BlockMaxGas = int64(300000000)

	// CostPerByte is the gas cost per byte
	CostPerByte = uint64(30)
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			// Add circuittypes as per 0.47 to 0.50 upgrade handler
			// https://github.com/cosmos/cosmos-sdk/blob/b7d9d4c8a9b6b8b61716d2023982d29bdc9839a6/simapp/upgrades.go#L21
			circuittypes.ModuleName,
		},
		Deleted: []string{},
	},
}
