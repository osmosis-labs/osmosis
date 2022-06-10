package v10

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v7/app/keepers"
)

func RunForkLogic(ctx sdk.Context, appKeepers *keepers.AppKeepers) {
	if ctx.BlockHeight() != ForkHeight {
		panic(fmt.Sprintf("current height %d, expected it to be the fork height %d", ctx.BlockHeight(), ForkHeight))
	}
	plan := upgradetypes.Plan{
		Name:   UpgradeName,
		Height: ForkHeight,
		Info:   "",
	}
	err := appKeepers.UpgradeKeeper.ScheduleUpgrade(ctx, plan)
	if err != nil {
		panic(err)
	}
	// appKeepers.UpgradeKeeper.
}
