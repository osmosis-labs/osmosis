package mempool1559

import sdk "github.com/cosmos/cosmos-sdk/types"

// DeliverTxCode is run on every transaction and will collect
// the gas for every transaction for use calculating gas
func DeliverTxCode(ctx sdk.Context, tx sdk.FeeTx) {
	CurEipState.deliverTxCode(ctx, tx)
}

// BeginBlockCode runs at the start of every block and it
// reset the CurEipStates lastBlockHeight and totalGasWantedThisBlock
func BeginBlockCode(ctx sdk.Context) {
	CurEipState.startBlock(ctx.BlockHeight())
}

// EndBlockCode runs at the end of every block and it
// updates the base fee based on the block attributes
func EndBlockCode(ctx sdk.Context) {
	CurEipState.updateBaseFee(ctx.BlockHeight())
}
