package lockup

import (
	"encoding/binary"
	"fmt"

	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"
)

// BeginBlocker is called on every block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
}

// Called every block to automatically unlock matured locks
func EndBlocker(ctx sdk.Context, k keeper.Keeper, keys map[string]*store.KVStoreKey) []abci.ValidatorUpdate {
	AppVersion := uint64(1)
	test := ctx.ConsensusParams().Version.GetAppVersion()
	fmt.Println("HI BOLD LETTERS HERE IN THE LOG-- app version is", test)
	store := ctx.KVStore(keys[types.StoreKey])
	versionBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(versionBytes, AppVersion)
	store.Set([]byte{types.ProtocolVersionByte}, versionBytes)

	second := ctx.ConsensusParams().Version.GetAppVersion()
	fmt.Println("HI BOLD LETTERS HERE IN THE LOG-- app version is now", second)

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
