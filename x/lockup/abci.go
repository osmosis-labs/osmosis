package lockup

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/lockup/keeper"
)

// BeginBlocker is called on every block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
}

// Called every block to automatically unlock matured locks
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	// disable automatic withdraw before specific block height
	// it is actually for testing with legacy
	MinBlockHeightToBeginAutoWithdrawing := int64(6)
	if ctx.BlockHeight() < MinBlockHeightToBeginAutoWithdrawing {
		return []abci.ValidatorUpdate{}
	}
	k.WithdrawAllMaturedLocks(ctx)
	return []abci.ValidatorUpdate{}
}
