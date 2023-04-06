package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	appParams "github.com/osmosis-labs/osmosis/v15/app/params"
)

func TestMsgSwapExactAmountIn(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
		properMsg := types.MsgSwapExactAmountIn{
			Sender: addr1,
			Routes: []types.SwapAmountInRoute{{
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

	msg := createMsg(func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), types.RouterKey)
	require.Equal(t, msg.Type(), "swap_exact_amount_in")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        types.MsgSwapExactAmountIn
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes",
			msg: createMsg(func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.Routes = nil
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes2",
			msg: createMsg(func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.Routes = []types.SwapAmountInRoute{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.Routes[1].TokenOutDenom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom2",
			msg: createMsg(func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.TokenIn.Denom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount token",
			msg: createMsg(func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount token",
			msg: createMsg(func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount criteria",
			msg: createMsg(func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.TokenOutMinAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount criteria",
			msg: createMsg(func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.TokenOutMinAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func TestMsgSwapExactAmountOut(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
		properMsg := types.MsgSwapExactAmountOut{
			Sender: addr1,
			Routes: []types.SwapAmountOutRoute{{
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

	msg := createMsg(func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), types.RouterKey)
	require.Equal(t, msg.Type(), "swap_exact_amount_out")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        types.MsgSwapExactAmountOut
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes",
			msg: createMsg(func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.Routes = nil
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes2",
			msg: createMsg(func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.Routes = []types.SwapAmountOutRoute{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.Routes[1].TokenInDenom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.TokenOut.Denom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount token",
			msg: createMsg(func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount token",
			msg: createMsg(func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount criteria",
			msg: createMsg(func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.TokenInMaxAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount criteria",
			msg: createMsg(func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.TokenInMaxAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

// Test authz serialize and de-serializes for poolmanager msg.
func TestAuthzMsg(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	coin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))

	testCases := []struct {
		name string
		msg  sdk.Msg
	}{
		{
			name: "MsgSwapExactAmountOut",
			msg: &types.MsgSwapExactAmountIn{
				Sender: addr1,
				Routes: []types.SwapAmountInRoute{{
					PoolId:        0,
					TokenOutDenom: "test",
				}, {
					PoolId:        1,
					TokenOutDenom: "test2",
				}},
				TokenIn:           coin,
				TokenOutMinAmount: sdk.NewInt(1),
			},
		},
		{
			name: "MsgSwapExactAmountOut",
			msg: &types.MsgSwapExactAmountOut{
				Sender: addr1,
				Routes: []types.SwapAmountOutRoute{{
					PoolId:       0,
					TokenInDenom: "test",
				}, {
					PoolId:       1,
					TokenInDenom: "test2",
				}},
				TokenOut:         coin,
				TokenInMaxAmount: sdk.NewInt(1),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apptesting.TestMessageAuthzSerialization(t, tc.msg)
		})
	}
}
