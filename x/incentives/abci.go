package incentives

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/c-osmosis/osmosis/x/incentives/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker is called on every block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
}

// Called every block to distribute coins
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	// TODO: should do this on every epoch not every block
	pots := k.GetActivePots(ctx)
	for _, pot := range pots {
		k.Distribute(ctx, pot)
		if pot.NumEpochs <= pot.FilledEpochs {
			k.FinishDistribution(ctx, pot)
		}
	}
	return []abci.ValidatorUpdate{}
}
