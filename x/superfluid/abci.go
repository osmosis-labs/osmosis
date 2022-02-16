package superfluid

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
)

// BeginBlocker is called on every block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
}

// Called every block to automatically unlock matured locks
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
