package v10

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
<<<<<<< HEAD

	"github.com/osmosis-labs/osmosis/v9/app/keepers"
=======
>>>>>>> e7993a4 (Delete irregular state change logic (#1769))
)

func CreateUpgradeHandler() upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		return fromVM, nil
	}
}
