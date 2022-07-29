package types

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/authz"

	appParams "github.com/osmosis-labs/osmosis/v10/app/params"
)

func TestMsgSwapExactAmountIn(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn) MsgSwapExactAmountIn {
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

	msg := createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "swap_exact_amount_in")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgSwapExactAmountIn
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.Routes = nil
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes2",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.Routes = []SwapAmountInRoute{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.Routes[1].TokenOutDenom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom2",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.TokenIn.Denom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount token",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount token",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount criteria",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.TokenOutMinAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount criteria",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
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

	createMsg := func(after func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut) MsgSwapExactAmountOut {
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

	msg := createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "swap_exact_amount_out")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgSwapExactAmountOut
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.Routes = nil
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes2",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.Routes = []SwapAmountOutRoute{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.Routes[1].TokenInDenom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.TokenOut.Denom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount token",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount token",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount criteria",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.TokenInMaxAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount criteria",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
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

func TestMsgJoinPool(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgJoinPool) MsgJoinPool) MsgJoinPool {
		properMsg := MsgJoinPool{
			Sender:         addr1,
			PoolId:         1,
			ShareOutAmount: sdk.NewInt(10),
			TokenInMaxs:    sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(10)), sdk.NewCoin("test2", sdk.NewInt(20))),
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg MsgJoinPool) MsgJoinPool {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "join_pool")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgJoinPool
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative requirement",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.ShareOutAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.TokenInMaxs[1].Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.TokenInMaxs[1].Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "'empty token max in' can pass",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.TokenInMaxs = nil
				return msg
			}),
			expectPass: true,
		},
		{
			name: "'empty token max in' can pass 2",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.TokenInMaxs = sdk.Coins{}
				return msg
			}),
			expectPass: true,
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

func TestMsgExitPool(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgExitPool) MsgExitPool) MsgExitPool {
		properMsg := MsgExitPool{
			Sender:        addr1,
			PoolId:        1,
			ShareInAmount: sdk.NewInt(10),
			TokenOutMins:  sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(10)), sdk.NewCoin("test2", sdk.NewInt(20))),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg MsgExitPool) MsgExitPool {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "exit_pool")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgExitPool
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative requirement",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.ShareInAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.TokenOutMins[1].Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.TokenOutMins[1].Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "'empty token min out' can pass",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.TokenOutMins = nil
				return msg
			}),
			expectPass: true,
		},
		{
			name: "'empty token min out' can pass 2",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.TokenOutMins = sdk.Coins{}
				return msg
			}),
			expectPass: true,
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

func TestMsgJoinSwapExternAmountIn(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
		properMsg := MsgJoinSwapExternAmountIn{
			Sender:            addr1,
			PoolId:            1,
			TokenIn:           sdk.NewCoin("test", sdk.NewInt(100)),
			ShareOutMinAmount: sdk.NewInt(100),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "join_swap_extern_amount_in")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgJoinSwapExternAmountIn
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
				msg.TokenIn.Denom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero criteria",
			msg: createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
				msg.ShareOutMinAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative criteria",
			msg: createMsg(func(msg MsgJoinSwapExternAmountIn) MsgJoinSwapExternAmountIn {
				msg.ShareOutMinAmount = sdk.NewInt(-10)
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

func TestMsgJoinSwapShareAmountOut(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
		properMsg := MsgJoinSwapShareAmountOut{
			Sender:           addr1,
			PoolId:           1,
			TokenInDenom:     "test",
			ShareOutAmount:   sdk.NewInt(100),
			TokenInMaxAmount: sdk.NewInt(100),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "join_swap_share_amount_out")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgJoinSwapShareAmountOut
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
				msg.TokenInDenom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
				msg.ShareOutAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
				msg.ShareOutAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero criteria",
			msg: createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
				msg.TokenInMaxAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative criteria",
			msg: createMsg(func(msg MsgJoinSwapShareAmountOut) MsgJoinSwapShareAmountOut {
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

func TestMsgExitSwapExternAmountOut(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
		properMsg := MsgExitSwapExternAmountOut{
			Sender:           addr1,
			PoolId:           1,
			TokenOut:         sdk.NewCoin("test", sdk.NewInt(100)),
			ShareInMaxAmount: sdk.NewInt(100),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "exit_swap_extern_amount_out")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgExitSwapExternAmountOut
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
				msg.TokenOut.Denom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero criteria",
			msg: createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
				msg.ShareInMaxAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative criteria",
			msg: createMsg(func(msg MsgExitSwapExternAmountOut) MsgExitSwapExternAmountOut {
				msg.ShareInMaxAmount = sdk.NewInt(-10)
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

func TestMsgExitSwapShareAmountIn(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
		properMsg := MsgExitSwapShareAmountIn{
			Sender:            addr1,
			PoolId:            1,
			TokenOutDenom:     "test",
			ShareInAmount:     sdk.NewInt(100),
			TokenOutMinAmount: sdk.NewInt(100),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "exit_swap_share_amount_in")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgExitSwapShareAmountIn
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
				msg.TokenOutDenom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
				msg.ShareInAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
				msg.ShareInAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero criteria",
			msg: createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
				msg.TokenOutMinAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative criteria",
			msg: createMsg(func(msg MsgExitSwapShareAmountIn) MsgExitSwapShareAmountIn {
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

// Test authz serialize and de-serializes for gamm msg.
func TestAuthzMsg(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	coin := sdk.NewCoin("stake", sdk.NewInt(1))
	someDate := time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC)

	const (
		mockGranter string = "cosmos1abc"
		mockGrantee string = "cosmos1xyz"
	)

	testCases := []struct {
		name                       string
		expectedGrantSignByteMsg   string
		expectedRevokeSignByteMsg  string
		expectedExecStrSignByteMsg string
		gammMsg                    sdk.Msg
	}{
		{
			name: "MsgExitSwapExternAmountOut",
			expectedGrantSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgGrant",
					  "value":{
						 "grant":{
							"authorization":{
							   "type":"cosmos-sdk/GenericAuthorization",
							   "value":{
								  "msg":"/osmosis.gamm.v1beta1.MsgExitSwapShareAmountIn"
							   }
							},
							"expiration":"0001-01-01T02:01:01.000000001Z"
						 },
						 "grantee":"%s",
						 "granter":"%s"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedRevokeSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgRevoke",
					  "value":{
						 "grantee":"%s",
						 "granter":"%s",
						 "msg_type_url":"/osmosis.gamm.v1beta1.MsgExitSwapShareAmountIn"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedExecStrSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgExec",
					  "value":{
						 "grantee":"%s",
						 "msgs":[
							{
							   "type":"osmosis/gamm/exit-swap-share-amount-in",
							   "value":{
								  "pool_id":"1",
								  "sender":"%s",
								  "share_in_amount":"100",
								  "token_out_denom":"test",
								  "token_out_min_amount":"100"
							   }
							}
						 ]
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, addr1),
			gammMsg: &MsgExitSwapShareAmountIn{
				Sender:            addr1,
				PoolId:            1,
				TokenOutDenom:     "test",
				ShareInAmount:     sdk.NewInt(100),
				TokenOutMinAmount: sdk.NewInt(100),
			},
		},
		{
			name: `MsgExitSwapExternAmountOut`,
			expectedGrantSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgGrant",
					  "value":{
						 "grant":{
							"authorization":{
							   "type":"cosmos-sdk/GenericAuthorization",
							   "value":{
								  "msg":"/osmosis.gamm.v1beta1.MsgExitSwapExternAmountOut"
							   }
							},
							"expiration":"0001-01-01T02:01:01.000000001Z"
						 },
						 "grantee":"%s",
						 "granter":"%s"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedRevokeSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgRevoke",
					  "value":{
						 "grantee":"%s",
						 "granter":"%s",
						 "msg_type_url":"/osmosis.gamm.v1beta1.MsgExitSwapExternAmountOut"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedExecStrSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgExec",
					  "value":{
						 "grantee":"%s",
						 "msgs":[
							{
							   "type":"osmosis/gamm/exit-swap-extern-amount-out",
							   "value":{
								  "pool_id":"1",
								  "sender":"%s",
								  "share_in_max_amount":"1",
								  "token_out":{
									 "amount":"1",
									 "denom":"stake"
								  }
							   }
							}
						 ]
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, addr1),
			gammMsg: &MsgExitSwapExternAmountOut{
				Sender:           addr1,
				PoolId:           1,
				TokenOut:         coin,
				ShareInMaxAmount: sdk.NewInt(1),
			},
		},
		{
			name: "MsgExitPool",
			expectedGrantSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgGrant",
					  "value":{
						 "grant":{
							"authorization":{
							   "type":"cosmos-sdk/GenericAuthorization",
							   "value":{
								  "msg":"/osmosis.gamm.v1beta1.MsgExitPool"
							   }
							},
							"expiration":"0001-01-01T02:01:01.000000001Z"
						 },
						 "grantee":"%s",
						 "granter":"%s"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedRevokeSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgRevoke",
					  "value":{
						 "grantee":"%s",
						 "granter":"%s",
						 "msg_type_url":"/osmosis.gamm.v1beta1.MsgExitPool"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedExecStrSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgExec",
					  "value":{
						 "grantee":"%s",
						 "msgs":[
							{
							   "type":"osmosis/gamm/exit-pool",
							   "value":{
								  "pool_id":"1",
								  "sender":"%s",
								  "share_in_amount":"100",
								  "token_out_mins":[
									 {
										"amount":"1",
										"denom":"stake"
									 }
								  ]
							   }
							}
						 ]
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, addr1),
			gammMsg: &MsgExitPool{
				Sender:        addr1,
				PoolId:        1,
				ShareInAmount: sdk.NewInt(100),
				TokenOutMins:  sdk.NewCoins(coin),
			},
		},
		{
			name: "MsgJoinPool",
			expectedGrantSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgGrant",
					  "value":{
						 "grant":{
							"authorization":{
							   "type":"cosmos-sdk/GenericAuthorization",
							   "value":{
								  "msg":"/osmosis.gamm.v1beta1.MsgJoinPool"
							   }
							},
							"expiration":"0001-01-01T02:01:01.000000001Z"
						 },
						 "grantee":"%s",
						 "granter":"%s"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedRevokeSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgRevoke",
					  "value":{
						 "grantee":"%s",
						 "granter":"%s",
						 "msg_type_url":"/osmosis.gamm.v1beta1.MsgJoinPool"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedExecStrSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgExec",
					  "value":{
						 "grantee":"%s",
						 "msgs":[
							{
							   "type":"osmosis/gamm/join-pool",
							   "value":{
								  "pool_id":"1",
								  "sender":"%s",
								  "share_out_amount":"1",
								  "token_in_maxs":[
									 {
										"amount":"1",
										"denom":"stake"
									 }
								  ]
							   }
							}
						 ]
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, addr1),
			gammMsg: &MsgJoinPool{
				Sender:         addr1,
				PoolId:         1,
				ShareOutAmount: sdk.NewInt(1),
				TokenInMaxs:    sdk.NewCoins(coin),
			},
		},
		{
			name: "MsgJoinSwapExternAmountIn",
			expectedGrantSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgGrant",
					  "value":{
						 "grant":{
							"authorization":{
							   "type":"cosmos-sdk/GenericAuthorization",
							   "value":{
								  "msg":"/osmosis.gamm.v1beta1.MsgJoinSwapExternAmountIn"
							   }
							},
							"expiration":"0001-01-01T02:01:01.000000001Z"
						 },
						 "grantee":"%s",
						 "granter":"%s"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedRevokeSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgRevoke",
					  "value":{
						 "grantee":"%s",
						 "granter":"%s",
						 "msg_type_url":"/osmosis.gamm.v1beta1.MsgJoinSwapExternAmountIn"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedExecStrSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgExec",
					  "value":{
						 "grantee":"%s",
						 "msgs":[
							{
							   "type":"osmosis/gamm/join-swap-extern-amount-in",
							   "value":{
								  "pool_id":"1",
								  "sender":"%s",
								  "share_out_min_amount":"1",
								  "token_in":{
									 "amount":"1",
									 "denom":"stake"
								  }
							   }
							}
						 ]
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, addr1),
			gammMsg: &MsgJoinSwapExternAmountIn{
				Sender:            addr1,
				PoolId:            1,
				TokenIn:           coin,
				ShareOutMinAmount: sdk.NewInt(1),
			},
		},
		{
			name: "MsgJoinSwapShareAmountOut",
			expectedGrantSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgGrant",
					  "value":{
						 "grant":{
							"authorization":{
							   "type":"cosmos-sdk/GenericAuthorization",
							   "value":{
								  "msg":"/osmosis.gamm.v1beta1.MsgJoinSwapShareAmountOut"
							   }
							},
							"expiration":"0001-01-01T02:01:01.000000001Z"
						 },
						 "grantee":"%s",
						 "granter":"%s"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedRevokeSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgRevoke",
					  "value":{
						 "grantee":"%s",
						 "granter":"%s",
						 "msg_type_url":"/osmosis.gamm.v1beta1.MsgJoinSwapShareAmountOut"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedExecStrSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgExec",
					  "value":{
						 "grantee":"%s",
						 "msgs":[
							{
							   "type":"osmosis/gamm/join-swap-share-amount-out",
							   "value":{
								  "pool_id":"1",
								  "sender":"%s",
								  "share_out_amount":"1",
								  "token_in_denom":"denom",
								  "token_in_max_amount":"1"
							   }
							}
						 ]
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, addr1),
			gammMsg: &MsgJoinSwapShareAmountOut{
				Sender:           addr1,
				PoolId:           1,
				TokenInDenom:     "denom",
				ShareOutAmount:   sdk.NewInt(1),
				TokenInMaxAmount: sdk.NewInt(1),
			},
		},
		{
			name: "MsgSwapExactAmountIn",
			expectedGrantSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgGrant",
					  "value":{
						 "grant":{
							"authorization":{
							   "type":"cosmos-sdk/GenericAuthorization",
							   "value":{
								  "msg":"/osmosis.gamm.v1beta1.MsgSwapExactAmountIn"
							   }
							},
							"expiration":"0001-01-01T02:01:01.000000001Z"
						 },
						 "grantee":"%s",
						 "granter":"%s"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedRevokeSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgRevoke",
					  "value":{
						 "grantee":"%s",
						 "granter":"%s",
						 "msg_type_url":"/osmosis.gamm.v1beta1.MsgSwapExactAmountIn"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedExecStrSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgExec",
					  "value":{
						 "grantee":"%s",
						 "msgs":[
							{
							   "type":"osmosis/gamm/swap-exact-amount-in",
							   "value":{
								  "routes":[
									 {
										"token_out_denom":"test"
									 },
									 {
										"pool_id":"1",
										"token_out_denom":"test2"
									 }
								  ],
								  "sender":"%s",
								  "token_in":{
									 "amount":"1",
									 "denom":"stake"
								  },
								  "token_out_min_amount":"1"
							   }
							}
						 ]
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, addr1),
			gammMsg: &MsgSwapExactAmountIn{
				Sender: addr1,
				Routes: []SwapAmountInRoute{{
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
			expectedGrantSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgGrant",
					  "value":{
						 "grant":{
							"authorization":{
							   "type":"cosmos-sdk/GenericAuthorization",
							   "value":{
								  "msg":"/osmosis.gamm.v1beta1.MsgSwapExactAmountOut"
							   }
							},
							"expiration":"0001-01-01T02:01:01.000000001Z"
						 },
						 "grantee":"%s",
						 "granter":"%s"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedRevokeSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgRevoke",
					  "value":{
						 "grantee":"%s",
						 "granter":"%s",
						 "msg_type_url":"/osmosis.gamm.v1beta1.MsgSwapExactAmountOut"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedExecStrSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgExec",
					  "value":{
						 "grantee":"%s",
						 "msgs":[
							{
							   "type":"osmosis/gamm/swap-exact-amount-out",
							   "value":{
								  "routes":[
									 {
										"token_in_denom":"test"
									 },
									 {
										"pool_id":"1",
										"token_in_denom":"test2"
									 }
								  ],
								  "sender":"%s",
								  "token_in_max_amount":"1",
								  "token_out":{
									 "amount":"1",
									 "denom":"stake"
								  }
							   }
							}
						 ]
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, addr1),
			gammMsg: &MsgSwapExactAmountOut{
				Sender: addr1,
				Routes: []SwapAmountOutRoute{{
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
			// Authz: Grant Msg
			typeURL := sdk.MsgTypeURL(tc.gammMsg)
			grant, err := authz.NewGrant(someDate, authz.NewGenericAuthorization(typeURL), someDate.Add(time.Hour))
			require.NoError(t, err)
			msgGrant := &authz.MsgGrant{Granter: mockGranter, Grantee: mockGrantee, Grant: grant}

			require.Equal(t,
				formatJsonStr(tc.expectedGrantSignByteMsg),
				string(legacytx.StdSignBytes("foo", 1, 1, 1, legacytx.StdFee{}, []sdk.Msg{msgGrant}, "memo")),
			)

			// Authz: Revoke Msg
			msgRevoke := &authz.MsgRevoke{Granter: mockGranter, Grantee: mockGrantee, MsgTypeUrl: typeURL}

			require.Equal(t,
				formatJsonStr(tc.expectedRevokeSignByteMsg),
				string(legacytx.StdSignBytes("foo", 1, 1, 1, legacytx.StdFee{}, []sdk.Msg{msgRevoke}, "memo")),
			)

			// Authz: Exec Msg
			msgAny, _ := cdctypes.NewAnyWithValue(tc.gammMsg)
			msgExec := &authz.MsgExec{Grantee: mockGrantee, Msgs: []*cdctypes.Any{msgAny}}

			require.Equal(t,
				formatJsonStr(tc.expectedExecStrSignByteMsg),
				string(legacytx.StdSignBytes("foo", 1, 1, 1, legacytx.StdFee{}, []sdk.Msg{msgExec}, "memo")),
			)
		})
	}
}

func formatJsonStr(jsonStrMsg string) string {
	ans := strings.ReplaceAll(jsonStrMsg, "\n", "")
	ans = strings.ReplaceAll(ans, "\t", "")
	ans = strings.ReplaceAll(ans, " ", "")

	return ans
}
