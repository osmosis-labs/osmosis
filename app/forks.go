package app

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlockForks is intended to be ran in.
func BeginBlockForks(ctx sdk.Context, app *OsmosisApp) {
	for _, fork := range Forks {
		if !strings.Contains(ctx.ChainID(), "e2e") {
			if ctx.BlockHeight() == fork.UpgradeHeight {
				fork.BeginForkLogic(ctx, &app.AppKeepers)
				return
			}
		} else {
			if ctx.BlockHeight() == TestFork.UpgradeHeight {
				TestFork.BeginForkLogic(ctx, &app.AppKeepers)
				return
			}
		}
	}
}
