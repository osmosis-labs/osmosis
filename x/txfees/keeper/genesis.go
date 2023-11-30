package keeper

import (
	"github.com/osmosis-labs/osmosis/v21/x/txfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the txfees module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	err := k.SetBaseDenom(ctx, genState.Basedenom)
	if err != nil {
		panic(err)
	}
	err = k.SetFeeTokens(ctx, genState.Feetokens)
	if err != nil {
		panic(err)
	}

	// We track the txfees generated in the KVStore.
	// If the values were exported, we set them here.
	// If the values were not exported, we initialize them to zero as well as use the current block height.
	if genState.TxFeesTracker != nil {
		k.SetTxFeesTrackerValue(ctx, genState.TxFeesTracker.TxFees)
		k.SetTxFeesTrackerStartHeight(ctx, genState.TxFeesTracker.HeightAccountingStartsFrom)
	} else {
		k.SetTxFeesTrackerValue(ctx, sdk.NewCoins())
		k.SetTxFeesTrackerStartHeight(ctx, ctx.BlockHeight())
	}
}

// ExportGenesis returns the txfees module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// Export KVStore values to the genesis state so they can be imported in init genesis.
	txFeesTracker := types.TxFeesTracker{
		TxFees:                     k.GetTxFeesTrackerValue(ctx),
		HeightAccountingStartsFrom: k.GetTxFeesTrackerStartHeight(ctx),
	}

	genesis := types.DefaultGenesis()
	genesis.Basedenom, _ = k.GetBaseDenom(ctx)
	genesis.Feetokens = k.GetFeeTokens(ctx)
	genesis.TxFeesTracker = &txFeesTracker
	return genesis
}
