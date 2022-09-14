package types_test

import (
	fmt "fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

func TestT(t *testing.T) {
	msg := types.MsgSwapExactAmountIn{
		Sender: "test",
		Routes: []types.SwapAmountInRoute{
			{PoolId: 1,
				TokenOutDenom: "test"},
		},
		TokenIn:           sdk.Coin{},
		TokenOutMinAmount: sdk.Int{},
	}

	fmt.Println(string(msg.GetSignBytes()))
}
