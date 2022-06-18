package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appParams "github.com/osmosis-labs/osmosis/v7/app/params"
)

func TestSybilResistantFeeMethods(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	//invalidAddr := sdk.AccAddress("invalid")

	createSwapInMsg := func(after func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn) MsgSwapExactAmountIn {
		properMsg := MsgSwapExactAmountIn{
			Sender: addr1,
			Routes: []SwapAmountInRoute{{
				PoolId:        0,
				TokenOutDenom: "test",
			}, {
				PoolId:        1,
				TokenOutDenom: "test2",
			}},
			TokenIn:           sdk.NewCoin("test", sdk.NewInt(100)),
			TokenOutMinAmount: sdk.NewInt(200),
		}

		return after(properMsg)
	}

	msgSwapIn := createSwapInMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
		// do nothing
		return msg
	})

	require.Equal(t, msgSwapIn.Route(), RouterKey)
	require.Equal(t, msgSwapIn.Type(), "swap_exact_amount_in")
	signersIn := msgSwapIn.GetSigners()
	require.Equal(t, len(signersIn), 1)
	require.Equal(t, signersIn[0].String(), addr1)

	createSwapInMultihop := func(after func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn) MsgSwapExactAmountIn {
		properMsg := MsgSwapExactAmountIn{
			Sender: addr1,
			Routes: []SwapAmountInRoute{{
				PoolId:        0,
				TokenOutDenom: "test",
			}, {
				PoolId:        1,
				TokenOutDenom: "test1",
			}, {
				PoolId:        2,
				TokenOutDenom: "test2",
			}},
			TokenIn:           sdk.NewCoin("test", sdk.NewInt(100)),
			TokenOutMinAmount: sdk.NewInt(200),
		}

		return after(properMsg)
	}

	msgMultSwapIn := createSwapInMultihop(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
		// do nothing
		return msg
	})

	require.Equal(t, msgSwapIn.Route(), RouterKey)
	require.Equal(t, msgSwapIn.Type(), "swap_exact_amount_in")
	signersMultIn := msgMultSwapIn.GetSigners()
	require.Equal(t, len(signersMultIn), 1)
	require.Equal(t, signersMultIn[0].String(), addr1)

	createMsgJoinPool := func(after func(msg MsgJoinPool) MsgJoinPool) MsgJoinPool {
		properMsg := MsgJoinPool{
			Sender:         addr1,
			PoolId:         1,
			ShareOutAmount: sdk.NewInt(10),
			TokenInMaxs:    sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(10)), sdk.NewCoin("test2", sdk.NewInt(20))),
		}

		return after(properMsg)
	}

	msgJoinPool := createMsgJoinPool(func(msg MsgJoinPool) MsgJoinPool {
		// Do nothing
		return msg
	})

	require.Equal(t, msgJoinPool.Route(), RouterKey)
	require.Equal(t, msgJoinPool.Type(), "join_pool")
	signersJoinPool := msgJoinPool.GetSigners()
	require.Equal(t, len(signersJoinPool), 1)
	require.Equal(t, signersJoinPool[0].String(), addr1)

	createSwapOutMsg := func(after func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut) MsgSwapExactAmountOut {
		properMsg := MsgSwapExactAmountOut{
			Sender: addr1,
			Routes: []SwapAmountOutRoute{{
				PoolId:       0,
				TokenInDenom: "test",
			}, {
				PoolId:       1,
				TokenInDenom: "test2",
			}},
			TokenOut:         sdk.NewCoin("test", sdk.NewInt(100)),
			TokenInMaxAmount: sdk.NewInt(200),
		}

		return after(properMsg)
	}

	msgSwapOut := createSwapOutMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
		// do nothing
		return msg
	})

	require.Equal(t, msgSwapOut.Route(), RouterKey)
	require.Equal(t, msgSwapOut.Type(), "swap_exact_amount_out")
	signersOut := msgSwapOut.GetSigners()
	require.Equal(t, len(signersOut), 1)
	require.Equal(t, signersOut[0].String(), addr1)

	createSwapOutMultihop := func(after func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut) MsgSwapExactAmountOut {
		properMsg := MsgSwapExactAmountOut{
			Sender: addr1,
			Routes: []SwapAmountOutRoute{{
				PoolId:       2,
				TokenInDenom: "test2",
			}, {
				PoolId:       1,
				TokenInDenom: "test1",
			}, {
				PoolId:       0,
				TokenInDenom: "test",
			}},
			TokenOut:         sdk.NewCoin("test", sdk.NewInt(100)),
			TokenInMaxAmount: sdk.NewInt(200),
		}

		return after(properMsg)
	}

	msgMultSwapOut := createSwapOutMultihop(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
		// do nothing
		return msg
	})

	require.Equal(t, msgSwapOut.Route(), RouterKey)
	require.Equal(t, msgSwapOut.Type(), "swap_exact_amount_out")
	signersMultOut := msgMultSwapOut.GetSigners()
	require.Equal(t, len(signersMultOut), 1)
	require.Equal(t, signersMultOut[0].String(), addr1)

	tests := []struct {
		name             string
		msg              SybilResistantFee
		expectDenomPath  []string
		expectPoolIdPath []uint64
		expectToken      sdk.Coin
	}{
		{
			name: "proper swap in",
			msg: createSwapInMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				// do nothing
				return msg
			}),
			expectDenomPath:  []string{"test", "test", "test2"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
		},
		{
			name: "proper swap in multihop",
			msg: createSwapInMultihop(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				// do nothing
				return msg
			}),
			expectDenomPath:  []string{"test", "test", "test1", "test2"},
			expectPoolIdPath: []uint64{0, 1, 2},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
		},
		{
			name: "proper swap out",
			msg: createSwapOutMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				// do nothing
				return msg
			}),
			expectDenomPath:  []string{"test", "test2", "test"},
			expectPoolIdPath: []uint64{0, 1},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
		},
		{
			name: "proper swap out multihop",
			msg: createSwapOutMultihop(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				// do nothing
				return msg
			}),
			expectDenomPath:  []string{"test", "test2", "test1", "test"},
			expectPoolIdPath: []uint64{2, 1, 0},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
		},
		{
			name: "empty routes swap in",
			msg: createSwapInMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.Routes = nil
				return msg
			}),
			expectDenomPath:  []string{},
			expectPoolIdPath: []uint64{},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
		},
		{
			name: "empty routes2 swap in",
			msg: createSwapInMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.Routes = []SwapAmountInRoute{}
				return msg
			}),
			expectDenomPath:  []string{},
			expectPoolIdPath: []uint64{},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
		},
		{
			name: "empty routes swap out",
			msg: createSwapOutMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.Routes = nil
				return msg
			}),
			expectDenomPath:  []string{},
			expectPoolIdPath: []uint64{},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
		},
		{
			name: "empty routes2 swap out",
			msg: createSwapOutMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.Routes = []SwapAmountOutRoute{}
				return msg
			}),
			expectDenomPath:  []string{},
			expectPoolIdPath: []uint64{},
			expectToken:      sdk.NewCoin("test", sdk.NewInt(100)),
		},
	}

	for _, test := range tests {
		denomPath := test.msg.GetTokenDenomsOnPath()
		poolIdPath := test.msg.GetPoolIdOnPath()
		token := test.msg.GetTokenToFee()

		require.Equal(t, test.expectToken, token)
		require.ElementsMatch(t, test.expectDenomPath, denomPath)
		require.ElementsMatch(t, test.expectPoolIdPath, poolIdPath)

		// this is a bit convoluted but works
		msgIn, ok := test.msg.(MsgSwapExactAmountIn)
		if !ok {
			msgOut, ok := test.msg.(MsgSwapExactAmountOut)
			require.True(t, ok)
			require.NoError(t, msgOut.ValidateBasic(), "test %v", test.name)
			return
		}

		require.True(t, ok)
		require.NoError(t, msgIn.ValidateBasic(), "test %v", test.name)

	}

}
