package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appParams "github.com/osmosis-labs/osmosis/v27/app/params"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/module"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

var (
	pk1         = ed25519.GenPrivKey().PubKey()
	addr1       = sdk.AccAddress(pk1.Address()).String()
	invalidAddr = sdk.AccAddress("invalid")

	validSwapRoutePoolThreeAmountIn = types.SwapAmountInRoute{
		PoolId:        3,
		TokenOutDenom: "uatom",
	}

	validSwapExactAmountInRoutes = []types.SwapAmountInRoute{{
		PoolId:        1,
		TokenOutDenom: appparams.BaseCoinUnit,
	}, {
		PoolId:        2,
		TokenOutDenom: "uatom",
	}}

	validSwapRoutePoolThreeAmountOut = types.SwapAmountOutRoute{
		PoolId:       3,
		TokenInDenom: "uatom",
	}

	validSwapExactAmountOutRoutes = []types.SwapAmountOutRoute{{
		PoolId:       1,
		TokenInDenom: "uatom",
	}, {
		PoolId:       2,
		TokenInDenom: appparams.BaseCoinUnit,
	}}
)

func createMsg[T any](properMsg T, after func(msg T) T) T {
	return after(properMsg)
}

func TestMsgSwapExactAmountIn(t *testing.T) {
	properMsg := types.MsgSwapExactAmountIn{
		Sender:            addr1,
		Routes:            validSwapExactAmountInRoutes,
		TokenIn:           sdk.NewCoin("test", osmomath.NewInt(100)),
		TokenOutMinAmount: osmomath.NewInt(200),
	}

	msg := createMsg(properMsg, func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
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
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.Routes = nil
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes2",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.Routes = []types.SwapAmountInRoute{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				// Create a deep copy.
				routesCopy := msg.Routes
				msg.Routes = make([]types.SwapAmountInRoute, 2)
				msg.Routes[0] = routesCopy[0]
				msg.Routes[1].TokenOutDenom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom2",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.TokenIn.Denom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount token",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.TokenIn.Amount = osmomath.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount token",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.TokenIn.Amount = osmomath.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount criteria",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.TokenOutMinAmount = osmomath.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount criteria",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountIn) types.MsgSwapExactAmountIn {
				msg.TokenOutMinAmount = osmomath.NewInt(-10)
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

	properMsg := types.MsgSwapExactAmountOut{
		Sender: addr1,
		Routes: []types.SwapAmountOutRoute{{
			PoolId:       0,
			TokenInDenom: "test",
		}, {
			PoolId:       1,
			TokenInDenom: "test2",
		}},
		TokenOut:         sdk.NewCoin("test", osmomath.NewInt(100)),
		TokenInMaxAmount: osmomath.NewInt(200),
	}

	msg := createMsg(properMsg, func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
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
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.Routes = nil
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes2",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.Routes = []types.SwapAmountOutRoute{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				// Create a deep copy.
				routesCopy := msg.Routes
				msg.Routes = make([]types.SwapAmountOutRoute, 2)
				msg.Routes[1] = routesCopy[1]
				msg.Routes[1].TokenInDenom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.TokenOut.Denom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount token",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.TokenOut.Amount = osmomath.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount token",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.TokenOut.Amount = osmomath.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount criteria",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.TokenInMaxAmount = osmomath.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount criteria",
			msg: createMsg(properMsg, func(msg types.MsgSwapExactAmountOut) types.MsgSwapExactAmountOut {
				msg.TokenInMaxAmount = osmomath.NewInt(-10)
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
	coin := sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(1))

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
				TokenOutMinAmount: osmomath.NewInt(1),
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
				TokenInMaxAmount: osmomath.NewInt(1),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apptesting.TestMessageAuthzSerialization(t, tc.msg, module.AppModuleBasic{})
		})
	}
}

func TestMsgSplitRouteSwapExactAmountIn(t *testing.T) {
	var (
		validMultihopRouteOne = types.SwapAmountInSplitRoute{
			Pools:         validSwapExactAmountInRoutes,
			TokenInAmount: osmomath.OneInt(),
		}
		validMultihopRouteTwo = types.SwapAmountInSplitRoute{
			Pools: []types.SwapAmountInRoute{
				validSwapRoutePoolThreeAmountIn,
			},
			TokenInAmount: osmomath.OneInt(),
		}

		defaultValidMsg = types.MsgSplitRouteSwapExactAmountIn{
			Sender: addr1,
			Routes: []types.SwapAmountInSplitRoute{
				validMultihopRouteOne,
				validMultihopRouteTwo,
			},
			TokenInDenom:      "udai",
			TokenOutMinAmount: osmomath.OneInt(),
		}
	)
	msg := createMsg(defaultValidMsg, func(msg types.MsgSplitRouteSwapExactAmountIn) types.MsgSplitRouteSwapExactAmountIn {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), types.RouterKey)
	require.Equal(t, msg.Type(), types.TypeMsgSplitRouteSwapExactAmountIn)
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := map[string]struct {
		msg         types.MsgSplitRouteSwapExactAmountIn
		expectError bool
	}{
		"valid": {
			msg: defaultValidMsg,
		},
		"invalid sender": {
			msg: createMsg(defaultValidMsg, func(msg types.MsgSplitRouteSwapExactAmountIn) types.MsgSplitRouteSwapExactAmountIn {
				msg.Sender = ""
				return msg
			}),
			expectError: true,
		},
		"duplicate multihop routes": {
			msg: createMsg(defaultValidMsg, func(msg types.MsgSplitRouteSwapExactAmountIn) types.MsgSplitRouteSwapExactAmountIn {
				msg.Routes = []types.SwapAmountInSplitRoute{
					validMultihopRouteOne,
					validMultihopRouteOne,
				}
				return msg
			}),
			expectError: true,
		},
		"different final token out": {
			msg: createMsg(defaultValidMsg, func(msg types.MsgSplitRouteSwapExactAmountIn) types.MsgSplitRouteSwapExactAmountIn {
				differentFinalTokenOut := validMultihopRouteOne

				// Initialize new slice for deep copy
				differentFinalTokenOut.Pools = make([]types.SwapAmountInRoute, len(validMultihopRouteOne.Pools))
				copy(differentFinalTokenOut.Pools, validMultihopRouteOne.Pools)

				// change last token out denom
				differentFinalTokenOut.Pools[len(differentFinalTokenOut.Pools)-1].TokenOutDenom = "other"

				msg.Routes = []types.SwapAmountInSplitRoute{
					differentFinalTokenOut,
					validMultihopRouteOne,
				}
				return msg
			}),
			expectError: true,
		},
		"invalid token in denom": {
			msg: createMsg(defaultValidMsg, func(msg types.MsgSplitRouteSwapExactAmountIn) types.MsgSplitRouteSwapExactAmountIn {
				msg.TokenInDenom = ""
				return msg
			}),
			expectError: true,
		},
		"invalid token out min amount": {
			msg: createMsg(defaultValidMsg, func(msg types.MsgSplitRouteSwapExactAmountIn) types.MsgSplitRouteSwapExactAmountIn {
				msg.TokenOutMinAmount = osmomath.ZeroInt()
				return msg
			}),
			expectError: true,
		},
		"empty routes": {
			msg: createMsg(defaultValidMsg, func(msg types.MsgSplitRouteSwapExactAmountIn) types.MsgSplitRouteSwapExactAmountIn {
				msg.Routes = []types.SwapAmountInSplitRoute{}
				return msg
			}),
			expectError: true,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()

			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgSplitRouteSwapExactAmountOut(t *testing.T) {
	var (
		validMultihopRouteOne = types.SwapAmountOutSplitRoute{
			Pools:          validSwapExactAmountOutRoutes,
			TokenOutAmount: osmomath.OneInt(),
		}
		validMultihopRouteTwo = types.SwapAmountOutSplitRoute{
			Pools: []types.SwapAmountOutRoute{
				validSwapRoutePoolThreeAmountOut,
			},
			TokenOutAmount: osmomath.OneInt(),
		}

		defaultValidMsg = types.MsgSplitRouteSwapExactAmountOut{
			Sender: addr1,
			Routes: []types.SwapAmountOutSplitRoute{
				validMultihopRouteOne,
				validMultihopRouteTwo,
			},
			TokenOutDenom:    "udai",
			TokenInMaxAmount: osmomath.OneInt(),
		}
	)
	msg := createMsg(defaultValidMsg, func(msg types.MsgSplitRouteSwapExactAmountOut) types.MsgSplitRouteSwapExactAmountOut {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), types.RouterKey)
	require.Equal(t, msg.Type(), types.TypeMsgSplitRouteSwapExactAmountOut)
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := map[string]struct {
		msg         types.MsgSplitRouteSwapExactAmountOut
		expectError bool
	}{
		"valid": {
			msg: defaultValidMsg,
		},
		"invalid sender": {
			msg: createMsg(defaultValidMsg, func(msg types.MsgSplitRouteSwapExactAmountOut) types.MsgSplitRouteSwapExactAmountOut {
				msg.Sender = ""
				return msg
			}),
			expectError: true,
		},
		"duplicate multihop routes": {
			msg: createMsg(defaultValidMsg, func(msg types.MsgSplitRouteSwapExactAmountOut) types.MsgSplitRouteSwapExactAmountOut {
				msg.Routes = []types.SwapAmountOutSplitRoute{
					validMultihopRouteOne,
					validMultihopRouteOne,
				}
				return msg
			}),
			expectError: true,
		},
		"different first token in": {
			msg: createMsg(defaultValidMsg, func(msg types.MsgSplitRouteSwapExactAmountOut) types.MsgSplitRouteSwapExactAmountOut {
				differentFirstTokenIn := validMultihopRouteOne

				// Initialize new slice for deep copy
				differentFirstTokenIn.Pools = make([]types.SwapAmountOutRoute, len(validMultihopRouteOne.Pools))
				copy(differentFirstTokenIn.Pools, validMultihopRouteOne.Pools)

				// change last token out denom
				differentFirstTokenIn.Pools[0].TokenInDenom = "other"

				msg.Routes = []types.SwapAmountOutSplitRoute{
					differentFirstTokenIn,
					validMultihopRouteOne,
				}
				return msg
			}),
			expectError: true,
		},
		"invalid token out denom": {
			msg: createMsg(defaultValidMsg, func(msg types.MsgSplitRouteSwapExactAmountOut) types.MsgSplitRouteSwapExactAmountOut {
				msg.TokenOutDenom = ""
				return msg
			}),
			expectError: true,
		},
		"invalid token in max amount": {
			msg: createMsg(defaultValidMsg, func(msg types.MsgSplitRouteSwapExactAmountOut) types.MsgSplitRouteSwapExactAmountOut {
				msg.TokenInMaxAmount = osmomath.ZeroInt()
				return msg
			}),
			expectError: true,
		},
		"empty routes": {
			msg: createMsg(defaultValidMsg, func(msg types.MsgSplitRouteSwapExactAmountOut) types.MsgSplitRouteSwapExactAmountOut {
				msg.Routes = []types.SwapAmountOutSplitRoute{}
				return msg
			}),
			expectError: true,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()

			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgSetDenomPairTakerFee(t *testing.T) {
	createMsg := func(after func(msg types.MsgSetDenomPairTakerFee) types.MsgSetDenomPairTakerFee) types.MsgSetDenomPairTakerFee {
		properMsg := types.MsgSetDenomPairTakerFee{
			Sender: addr1,
			DenomPairTakerFee: []types.DenomPairTakerFee{
				{
					TokenInDenom:  appparams.BaseCoinUnit,
					TokenOutDenom: "uatom",
					TakerFee:      osmomath.MustNewDecFromStr("0.003"),
				},
				{
					TokenInDenom:  appparams.BaseCoinUnit,
					TokenOutDenom: "uion",
					TakerFee:      osmomath.MustNewDecFromStr("0.006"),
				},
			},
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg types.MsgSetDenomPairTakerFee) types.MsgSetDenomPairTakerFee {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), types.RouterKey)
	require.Equal(t, msg.Type(), types.TypeMsgSetDenomPairTakerFee)
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := map[string]struct {
		msg         types.MsgSetDenomPairTakerFee
		expectError bool
	}{
		"valid": {
			msg: createMsg(func(msg types.MsgSetDenomPairTakerFee) types.MsgSetDenomPairTakerFee {
				// Do nothing
				return msg
			}),
		},
		"invalid sender": {
			msg: createMsg(func(msg types.MsgSetDenomPairTakerFee) types.MsgSetDenomPairTakerFee {
				msg.Sender = ""
				return msg
			}),
			expectError: true,
		},
		"invalid denom0": {
			msg: createMsg(func(msg types.MsgSetDenomPairTakerFee) types.MsgSetDenomPairTakerFee {
				msg.DenomPairTakerFee[0].TokenInDenom = ""
				return msg
			}),
			expectError: true,
		},
		"invalid denom1": {
			msg: createMsg(func(msg types.MsgSetDenomPairTakerFee) types.MsgSetDenomPairTakerFee {
				msg.DenomPairTakerFee[0].TokenOutDenom = ""
				return msg
			}),
			expectError: true,
		},
		"invalid denom0 = denom1": {
			msg: createMsg(func(msg types.MsgSetDenomPairTakerFee) types.MsgSetDenomPairTakerFee {
				msg.DenomPairTakerFee[0].TokenInDenom = msg.DenomPairTakerFee[0].TokenOutDenom
				return msg
			}),
			expectError: true,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()

			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgSetTakerFeeShareAgreementForDenom(t *testing.T) {
	createMsg := func(after func(msg types.MsgSetTakerFeeShareAgreementForDenom) types.MsgSetTakerFeeShareAgreementForDenom) types.MsgSetTakerFeeShareAgreementForDenom {
		properMsg := types.MsgSetTakerFeeShareAgreementForDenom{
			Sender:      addr1,
			Denom:       "uatom",
			SkimPercent: osmomath.MustNewDecFromStr("0.01"),
			SkimAddress: addr1,
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg types.MsgSetTakerFeeShareAgreementForDenom) types.MsgSetTakerFeeShareAgreementForDenom {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), types.RouterKey)
	require.Equal(t, msg.Type(), types.TypeMsgSetTakerFeeShareAgreementForDenomPair)
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := map[string]struct {
		msg         types.MsgSetTakerFeeShareAgreementForDenom
		expectError bool
	}{
		"valid": {
			msg: createMsg(func(msg types.MsgSetTakerFeeShareAgreementForDenom) types.MsgSetTakerFeeShareAgreementForDenom {
				// Do nothing
				return msg
			}),
		},
		"invalid sender": {
			msg: createMsg(func(msg types.MsgSetTakerFeeShareAgreementForDenom) types.MsgSetTakerFeeShareAgreementForDenom {
				msg.Sender = ""
				return msg
			}),
			expectError: true,
		},
		"invalid skim address": {
			msg: createMsg(func(msg types.MsgSetTakerFeeShareAgreementForDenom) types.MsgSetTakerFeeShareAgreementForDenom {
				msg.SkimAddress = ""
				return msg
			}),
			expectError: true,
		},
		"invalid skim percent (zero or less)": {
			msg: createMsg(func(msg types.MsgSetTakerFeeShareAgreementForDenom) types.MsgSetTakerFeeShareAgreementForDenom {
				msg.SkimPercent = osmomath.MustNewDecFromStr("-0.01")
				return msg
			}),
			expectError: true,
		},
		"invalid skim percent (greater than one)": {
			msg: createMsg(func(msg types.MsgSetTakerFeeShareAgreementForDenom) types.MsgSetTakerFeeShareAgreementForDenom {
				msg.SkimPercent = osmomath.MustNewDecFromStr("1.01")
				return msg
			}),
			expectError: true,
		},
		"invalid denom": {
			msg: createMsg(func(msg types.MsgSetTakerFeeShareAgreementForDenom) types.MsgSetTakerFeeShareAgreementForDenom {
				msg.Denom = ""
				return msg
			}),
			expectError: true,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()

			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgSetRegisteredAlloyedPool(t *testing.T) {
	createMsg := func(after func(msg types.MsgSetRegisteredAlloyedPool) types.MsgSetRegisteredAlloyedPool) types.MsgSetRegisteredAlloyedPool {
		properMsg := types.MsgSetRegisteredAlloyedPool{
			Sender: addr1,
			PoolId: 1,
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg types.MsgSetRegisteredAlloyedPool) types.MsgSetRegisteredAlloyedPool {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), types.RouterKey)
	require.Equal(t, msg.Type(), types.TypeMsgSetRegisteredAlloyedPool)
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := map[string]struct {
		msg         types.MsgSetRegisteredAlloyedPool
		expectError bool
	}{
		"valid": {
			msg: createMsg(func(msg types.MsgSetRegisteredAlloyedPool) types.MsgSetRegisteredAlloyedPool {
				// Do nothing
				return msg
			}),
		},
		"invalid sender": {
			msg: createMsg(func(msg types.MsgSetRegisteredAlloyedPool) types.MsgSetRegisteredAlloyedPool {
				msg.Sender = ""
				return msg
			}),
			expectError: true,
		},
		"invalid pool id": {
			msg: createMsg(func(msg types.MsgSetRegisteredAlloyedPool) types.MsgSetRegisteredAlloyedPool {
				msg.PoolId = 0
				return msg
			}),
			expectError: true,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()

			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
