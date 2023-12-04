package keeper

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v15/x/txfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the txfees module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	recipientAcc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	if recipientAcc == nil {
		panic(fmt.Sprintf("module account %s does not exist", types.ModuleName))
	}

	err := k.SetBaseDenom(ctx, genState.Basedenom)
	if err != nil {
		panic(err)
	}
	err = k.SetFeeTokens(ctx, genState.Feetokens)
	if err != nil {
		panic(err)
	}

	info := k.epochKeeper.GetEpochInfo(ctx, types.EpochIdentifier)
	if info.Identifier == "" {
		panic(fmt.Sprintf("epoch info for identifier %s does not exist", types.EpochIdentifier))
	}
}

// ExportGenesis returns the txfees module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Basedenom, _ = k.GetBaseDenom(ctx)
	genesis.Feetokens = k.GetFeeTokens(ctx)
	return genesis
}
