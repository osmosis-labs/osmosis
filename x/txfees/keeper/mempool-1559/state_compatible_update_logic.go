package mempool1559

import sdk "github.com/cosmos/cosmos-sdk/types"

func DeliverTxCode(ctx sdk.Context, tx sdk.FeeTx) {
	CurEipState.deliverTxCode(ctx, tx)
}

func BeginBlockCode(ctx sdk.Context) {
	CurEipState.startBlock(ctx.BlockHeight())
}

func EndBlockCode(ctx sdk.Context) {
	CurEipState.updateBaseFee(ctx.BlockHeight())
}
