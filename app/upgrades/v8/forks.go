package v8

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RunForkLogic(ctx sdk.Context) {
	ctx.Logger().Info("Applying emergency hard fork for v8.")
	// Three different parts to this upgrade
	// 1) Remove superfluid staking from Osmo/UST and Osmo/Luna
}
