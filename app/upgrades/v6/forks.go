package v6

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RunForkLogic(ctx sdk.Context) {
	// All the height gated fork logic is actually in our ibc-go fork.
	// See: https://github.com/osmosis-labs/ibc-go/releases/tag/v2.0.2-osmo
	ctx.Logger().Info("Applying emergency hard fork for v6, allows IBC to create new channels.")
}
