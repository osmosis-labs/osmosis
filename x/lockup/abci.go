package lockup

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"
)

// BeginBlocker is called on every block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
}

// Called every block to automatically unlock matured locks
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	test := ctx.ConsensusParams().Version.GetAppVersion()
	fmt.Println("HI BOLD LETTERS HERE IN THE LOG-- app version is", test)
	// disable automatic withdraw before specific block height
	// it is actually for testing with legacy
	MinBlockHeightToBeginAutoWithdrawing := int64(6)
	if ctx.BlockHeight() < MinBlockHeightToBeginAutoWithdrawing {
		return []abci.ValidatorUpdate{}
	}

	// delete synthetic locks matured before lockup deletion
	k.DeleteAllMaturedSyntheticLocks(ctx)

	// withdraw and delete locks
	k.WithdrawAllMaturedLocks(ctx)
	return []abci.ValidatorUpdate{}
}

// TODO: add invariant that no native lockup existent synthetic lockup exists by calling GetAllSyntheticLockups
// TODO: if superfluid does not delete synthetic lockup before native lockup deletion, it won't be able to be deleted
