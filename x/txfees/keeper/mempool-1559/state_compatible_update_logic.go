package mempool1559

import sdk "github.com/cosmos/cosmos-sdk/types"

func DeliverTxCode(ctx sdk.Context, tx sdk.Tx) {

}

func EndBlockCode(ctx sdk.Context) {
	curEipState.updateBaseFee(ctx.BlockHeight())
}
