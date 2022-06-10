package v10

import (
	"github.com/osmosis-labs/osmosis/v7/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Last executed block on the v9 code was 4713064.
// Last committed block is assumed to be 4713064, as we have block proposals that were not precommitted upon
// for 4713065.
const ForkHeight = 4713065

// UpgradeName defines the on-chain upgrade name for the Osmosis v9-fork for recovery.
// This is not called v10, due to this bug that would require a state migration to fix:
const UpgradeName = "v10"

// RecoveryAddress that the irregular state change transfers to.
// TODO: Derive it, currently set to 'distribution' module account.
var RecoveryAddress, recoveryAddressErr = sdk.AccAddressFromBech32("osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld")

func init() {
	if recoveryAddressErr != nil {
		panic("recovery address decoding failure")
	}
}

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}

var Fork = upgrades.Fork{
	UpgradeName:    UpgradeName,
	UpgradeHeight:  ForkHeight,
	BeginForkLogic: RunForkLogic,
}
