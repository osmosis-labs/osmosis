package types_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	appParams "github.com/osmosis-labs/osmosis/v12/app/params"
)

func TestMsgSwapExactAmountIn(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")
	_, invalidAccErr := sdk.AccAddressFromBech32(invalidAddr.String())

	createMsg := func(after func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
		properMsg := gammtypes.MsgSwapExactAmountIn{
			Sender: addr1,
			Routes: []gammtypes.SwapAmountInRoute{{
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

	msg := createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), gammtypes.RouterKey)
	require.Equal(t, msg.Type(), "swap_exact_amount_in")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        gammtypes.MsgSwapExactAmountIn
		expectErr error
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
				// Do nothing
				return msg
			}),
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectErr: sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", invalidAccErr),
		},
		{
			name: "empty routes",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
				msg.Routes = nil
				return msg
			}),
			expectErr: gammtypes.ErrEmptyRoutes,
		},
		{
			name: "empty routes2",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
				msg.Routes = []gammtypes.SwapAmountInRoute{}
				return msg
			}),
			expectErr: gammtypes.ErrEmptyRoutes,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
				msg.Routes[1].TokenOutDenom = "1"
				return msg
			}),
			expectErr: fmt.Errorf("invalid denom: %s", "1"),
		},
		// should err be "invalid denom"?
		{
			name: "invalid denom2",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
				msg.TokenIn.Denom = "1"
				return msg
			}),
			expectErr: sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.TokenIn.String()),
		},
		{
			name: "zero amount token",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(0)
				return msg
			}),
			expectErr: sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.TokenIn.String()),
		},
		{
			name: "negative amount token",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.TokenIn.String()),
		},
		{
			name: "zero amount criteria",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
				msg.TokenOutMinAmount = sdk.NewInt(0)
				return msg
			}),
			expectErr: gammtypes.ErrNotPositiveCriteria,
		},
		{
			name: "negative amount criteria",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
				msg.TokenOutMinAmount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: gammtypes.ErrNotPositiveCriteria,
		},
	}

	for _, test := range tests {
		err := test.msg.ValidateBasic()
		if test.expectErr == nil {
			require.NoError(t, err, "test: %v", test.name)
		} else {
			require.Error(t, err, "test: %v", test.name)
			require.ErrorAs(t, test.expectErr, &err)
		}
	}
}

func TestMsgSwapExactAmountOut(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")
	_, invalidAccErr := sdk.AccAddressFromBech32(invalidAddr.String())

	createMsg := func(after func(msg gammtypes.MsgSwapExactAmountOut) gammtypes.MsgSwapExactAmountOut) gammtypes.MsgSwapExactAmountOut {
		properMsg := gammtypes.MsgSwapExactAmountOut{
			Sender: addr1,
			Routes: []gammtypes.SwapAmountOutRoute{{
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

	msg := createMsg(func(msg gammtypes.MsgSwapExactAmountOut) gammtypes.MsgSwapExactAmountOut {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), gammtypes.RouterKey)
	require.Equal(t, msg.Type(), "swap_exact_amount_out")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        gammtypes.MsgSwapExactAmountOut
		expectErr error
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountOut) gammtypes.MsgSwapExactAmountOut {
				// Do nothing
				return msg
			}),
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountOut) gammtypes.MsgSwapExactAmountOut {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectErr: sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", invalidAccErr),
		},
		{
			name: "empty routes",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountOut) gammtypes.MsgSwapExactAmountOut {
				msg.Routes = nil
				return msg
			}),
			expectErr: gammtypes.ErrEmptyRoutes,
		},
		{
			name: "empty routes2",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountOut) gammtypes.MsgSwapExactAmountOut {
				msg.Routes = []gammtypes.SwapAmountOutRoute{}
				return msg
			}),
			expectErr: gammtypes.ErrEmptyRoutes,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountOut) gammtypes.MsgSwapExactAmountOut {
				msg.Routes[1].TokenInDenom = "1"
				return msg
			}),
			expectErr: fmt.Errorf("invalid denom: %s", "1"),
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountOut) gammtypes.MsgSwapExactAmountOut {
				msg.TokenOut.Denom = "1"
				return msg
			}),
			expectErr: sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.TokenOut.String()),
		},
		{
			name: "zero amount token",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountOut) gammtypes.MsgSwapExactAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(0)
				return msg
			}),
			expectErr: sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.TokenOut.String()),
		},
		{
			name: "negative amount token",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountOut) gammtypes.MsgSwapExactAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.TokenOut.String()),
		},
		{
			name: "zero amount criteria",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountOut) gammtypes.MsgSwapExactAmountOut {
				msg.TokenInMaxAmount = sdk.NewInt(0)
				return msg
			}),
			expectErr: gammtypes.ErrNotPositiveCriteria,
		},
		{
			name: "negative amount criteria",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountOut) gammtypes.MsgSwapExactAmountOut {
				msg.TokenInMaxAmount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: gammtypes.ErrNotPositiveCriteria,
		},
	}

	for _, test := range tests {
		err := test.msg.ValidateBasic()
		if test.expectErr == nil {
			require.NoError(t, err, "test: %v", test.name)
		} else {
			require.Error(t, err, "test: %v", test.name)
			require.ErrorAs(t, test.expectErr, &err)
		}
	}
}

func TestMsgJoinPool(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg gammtypes.MsgJoinPool) gammtypes.MsgJoinPool) gammtypes.MsgJoinPool {
		properMsg := gammtypes.MsgJoinPool{
			Sender:         addr1,
			PoolId:         1,
			ShareOutAmount: sdk.NewInt(10),
			TokenInMaxs:    sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(10)), sdk.NewCoin("test2", sdk.NewInt(20))),
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg gammtypes.MsgJoinPool) gammtypes.MsgJoinPool {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), gammtypes.RouterKey)
	require.Equal(t, msg.Type(), "join_pool")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        gammtypes.MsgJoinPool
		expectErr error
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg gammtypes.MsgJoinPool) gammtypes.MsgJoinPool {
				// Do nothing
				return msg
			}),
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg gammtypes.MsgJoinPool) gammtypes.MsgJoinPool {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "negative requirement",
			msg: createMsg(func(msg gammtypes.MsgJoinPool) gammtypes.MsgJoinPool {
				msg.ShareOutAmount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveRequireAmount, msg.ShareOutAmount.String()),
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg gammtypes.MsgJoinPool) gammtypes.MsgJoinPool {
				msg.TokenInMaxs[1].Amount = sdk.NewInt(0)
				return msg
			}),
			expectErr: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg gammtypes.MsgJoinPool) gammtypes.MsgJoinPool {
				msg.TokenInMaxs[1].Amount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "'empty token max in' can pass",
			msg: createMsg(func(msg gammtypes.MsgJoinPool) gammtypes.MsgJoinPool {
				msg.TokenInMaxs = nil
				return msg
			}),
		},
		{
			name: "'empty token max in' can pass 2",
			msg: createMsg(func(msg gammtypes.MsgJoinPool) gammtypes.MsgJoinPool {
				msg.TokenInMaxs = sdk.Coins{}
				return msg
			}),
		},
	}

	for _, test := range tests {
		err := test.msg.ValidateBasic()
		if test.expectErr == nil {
			require.NoError(t, err, "test: %v", test.name)
		} else {
			require.Error(t, err, "test: %v", test.name)
			require.ErrorAs(t, test.expectErr, &err)
		}
	}
}

func TestMsgExitPool(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg gammtypes.MsgExitPool) gammtypes.MsgExitPool) gammtypes.MsgExitPool {
		properMsg := gammtypes.MsgExitPool{
			Sender:        addr1,
			PoolId:        1,
			ShareInAmount: sdk.NewInt(10),
			TokenOutMins:  sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(10)), sdk.NewCoin("test2", sdk.NewInt(20))),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg gammtypes.MsgExitPool) gammtypes.MsgExitPool {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), gammtypes.RouterKey)
	require.Equal(t, msg.Type(), "exit_pool")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        gammtypes.MsgExitPool
		expectErr error
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg gammtypes.MsgExitPool) gammtypes.MsgExitPool {
				// Do nothing
				return msg
			}),
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg gammtypes.MsgExitPool) gammtypes.MsgExitPool {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "negative requirement",
			msg: createMsg(func(msg gammtypes.MsgExitPool) gammtypes.MsgExitPool {
				msg.ShareInAmount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveRequireAmount, msg.ShareInAmount.String()),
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg gammtypes.MsgExitPool) gammtypes.MsgExitPool {
				msg.TokenOutMins[1].Amount = sdk.NewInt(0)
				return msg
			}),
			expectErr: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg gammtypes.MsgExitPool) gammtypes.MsgExitPool {
				msg.TokenOutMins[1].Amount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "'empty token min out' can pass",
			msg: createMsg(func(msg gammtypes.MsgExitPool) gammtypes.MsgExitPool {
				msg.TokenOutMins = nil
				return msg
			}),
		},
		{
			name: "'empty token min out' can pass 2",
			msg: createMsg(func(msg gammtypes.MsgExitPool) gammtypes.MsgExitPool {
				msg.TokenOutMins = sdk.Coins{}
				return msg
			}),
		},
	}

	for _, test := range tests {
		err := test.msg.ValidateBasic()
		if test.expectErr == nil {
			require.NoError(t, err, "test: %v", test.name)
		} else {
			require.Error(t, err, "test: %v", test.name)
			require.ErrorAs(t, test.expectErr, &err)
		}
	}
}

func TestMsgJoinSwapExternAmountIn(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg gammtypes.MsgJoinSwapExternAmountIn) gammtypes.MsgJoinSwapExternAmountIn) gammtypes.MsgJoinSwapExternAmountIn {
		properMsg := gammtypes.MsgJoinSwapExternAmountIn{
			Sender:            addr1,
			PoolId:            1,
			TokenIn:           sdk.NewCoin("test", sdk.NewInt(100)),
			ShareOutMinAmount: sdk.NewInt(100),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg gammtypes.MsgJoinSwapExternAmountIn) gammtypes.MsgJoinSwapExternAmountIn {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), gammtypes.RouterKey)
	require.Equal(t, msg.Type(), "join_swap_extern_amount_in")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        gammtypes.MsgJoinSwapExternAmountIn
		expectErr error
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg gammtypes.MsgJoinSwapExternAmountIn) gammtypes.MsgJoinSwapExternAmountIn {
				// Do nothing
				return msg
			}),
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg gammtypes.MsgJoinSwapExternAmountIn) gammtypes.MsgJoinSwapExternAmountIn {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg gammtypes.MsgJoinSwapExternAmountIn) gammtypes.MsgJoinSwapExternAmountIn {
				msg.TokenIn.Denom = "1"
				return msg
			}),
			expectErr: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg gammtypes.MsgJoinSwapExternAmountIn) gammtypes.MsgJoinSwapExternAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(0)
				return msg
			}),
			expectErr: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg gammtypes.MsgJoinSwapExternAmountIn) gammtypes.MsgJoinSwapExternAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "zero criteria",
			msg: createMsg(func(msg gammtypes.MsgJoinSwapExternAmountIn) gammtypes.MsgJoinSwapExternAmountIn {
				msg.ShareOutMinAmount = sdk.NewInt(0)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveCriteria, msg.ShareOutMinAmount.String()),
		},
		{
			name: "negative criteria",
			msg: createMsg(func(msg gammtypes.MsgJoinSwapExternAmountIn) gammtypes.MsgJoinSwapExternAmountIn {
				msg.ShareOutMinAmount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveCriteria, msg.ShareOutMinAmount.String()),
		},
	}

	for _, test := range tests {
		err := test.msg.ValidateBasic()
		if test.expectErr == nil {
			require.NoError(t, err, "test: %v", test.name)
		} else {
			require.Error(t, err, "test: %v", test.name)
			require.ErrorAs(t, test.expectErr, &err)
		}
	}
}

func TestMsgJoinSwapShareAmountOut(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg gammtypes.MsgJoinSwapShareAmountOut) gammtypes.MsgJoinSwapShareAmountOut) gammtypes.MsgJoinSwapShareAmountOut {
		properMsg := gammtypes.MsgJoinSwapShareAmountOut{
			Sender:           addr1,
			PoolId:           1,
			TokenInDenom:     "test",
			ShareOutAmount:   sdk.NewInt(100),
			TokenInMaxAmount: sdk.NewInt(100),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg gammtypes.MsgJoinSwapShareAmountOut) gammtypes.MsgJoinSwapShareAmountOut {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), gammtypes.RouterKey)
	require.Equal(t, msg.Type(), "join_swap_share_amount_out")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        gammtypes.MsgJoinSwapShareAmountOut
		expectErr error
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg gammtypes.MsgJoinSwapShareAmountOut) gammtypes.MsgJoinSwapShareAmountOut {
				// Do nothing
				return msg
			}),
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg gammtypes.MsgJoinSwapShareAmountOut) gammtypes.MsgJoinSwapShareAmountOut {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg gammtypes.MsgJoinSwapShareAmountOut) gammtypes.MsgJoinSwapShareAmountOut {
				msg.TokenInDenom = "1"
				return msg
			}),
			expectErr: fmt.Errorf("invalid denom: %s", msg.TokenInDenom),
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg gammtypes.MsgJoinSwapShareAmountOut) gammtypes.MsgJoinSwapShareAmountOut {
				msg.ShareOutAmount = sdk.NewInt(0)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveRequireAmount, msg.ShareOutAmount.String()),
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg gammtypes.MsgJoinSwapShareAmountOut) gammtypes.MsgJoinSwapShareAmountOut {
				msg.ShareOutAmount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveRequireAmount, msg.ShareOutAmount.String()),
		},
		{
			name: "zero criteria",
			msg: createMsg(func(msg gammtypes.MsgJoinSwapShareAmountOut) gammtypes.MsgJoinSwapShareAmountOut {
				msg.TokenInMaxAmount = sdk.NewInt(0)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveCriteria, msg.TokenInMaxAmount.String()),
		},
		{
			name: "negative criteria",
			msg: createMsg(func(msg gammtypes.MsgJoinSwapShareAmountOut) gammtypes.MsgJoinSwapShareAmountOut {
				msg.TokenInMaxAmount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveCriteria, msg.TokenInMaxAmount.String()),
		},
	}

	for _, test := range tests {
		err := test.msg.ValidateBasic()
		if test.expectErr == nil {
			require.NoError(t, err, "test: %v", test.name)
		} else {
			require.Error(t, err, "test: %v", test.name)
			require.ErrorAs(t, test.expectErr, &err)
		}
	}
}

func TestMsgExitSwapExternAmountOut(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg gammtypes.MsgExitSwapExternAmountOut) gammtypes.MsgExitSwapExternAmountOut) gammtypes.MsgExitSwapExternAmountOut {
		properMsg := gammtypes.MsgExitSwapExternAmountOut{
			Sender:           addr1,
			PoolId:           1,
			TokenOut:         sdk.NewCoin("test", sdk.NewInt(100)),
			ShareInMaxAmount: sdk.NewInt(100),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg gammtypes.MsgExitSwapExternAmountOut) gammtypes.MsgExitSwapExternAmountOut {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), gammtypes.RouterKey)
	require.Equal(t, msg.Type(), "exit_swap_extern_amount_out")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        gammtypes.MsgExitSwapExternAmountOut
		expectErr error
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg gammtypes.MsgExitSwapExternAmountOut) gammtypes.MsgExitSwapExternAmountOut {
				// Do nothing
				return msg
			}),
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg gammtypes.MsgExitSwapExternAmountOut) gammtypes.MsgExitSwapExternAmountOut {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg gammtypes.MsgExitSwapExternAmountOut) gammtypes.MsgExitSwapExternAmountOut {
				msg.TokenOut.Denom = "1"
				return msg
			}),
			expectErr: fmt.Errorf("invalid denom: %s", msg.TokenOut.Denom),
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg gammtypes.MsgExitSwapExternAmountOut) gammtypes.MsgExitSwapExternAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(0)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveRequireAmount, msg.TokenOut.Amount.String()),
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg gammtypes.MsgExitSwapExternAmountOut) gammtypes.MsgExitSwapExternAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveRequireAmount, msg.TokenOut.Amount.String()),
		},
		{
			name: "zero criteria",
			msg: createMsg(func(msg gammtypes.MsgExitSwapExternAmountOut) gammtypes.MsgExitSwapExternAmountOut {
				msg.ShareInMaxAmount = sdk.NewInt(0)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveCriteria, msg.ShareInMaxAmount.String()),
		},
		{
			name: "negative criteria",
			msg: createMsg(func(msg gammtypes.MsgExitSwapExternAmountOut) gammtypes.MsgExitSwapExternAmountOut {
				msg.ShareInMaxAmount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveCriteria, msg.ShareInMaxAmount.String()),
		},
	}

	for _, test := range tests {
		err := test.msg.ValidateBasic()
		if test.expectErr == nil {
			require.NoError(t, err, "test: %v", test.name)
		} else {
			require.Error(t, err, "test: %v", test.name)
			require.ErrorAs(t, test.expectErr, &err)
		}
	}
}

func TestMsgExitSwapShareAmountIn(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg gammtypes.MsgExitSwapShareAmountIn) gammtypes.MsgExitSwapShareAmountIn) gammtypes.MsgExitSwapShareAmountIn {
		properMsg := gammtypes.MsgExitSwapShareAmountIn{
			Sender:            addr1,
			PoolId:            1,
			TokenOutDenom:     "test",
			ShareInAmount:     sdk.NewInt(100),
			TokenOutMinAmount: sdk.NewInt(100),
		}
		return after(properMsg)
	}

	msg := createMsg(func(msg gammtypes.MsgExitSwapShareAmountIn) gammtypes.MsgExitSwapShareAmountIn {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), gammtypes.RouterKey)
	require.Equal(t, msg.Type(), "exit_swap_share_amount_in")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        gammtypes.MsgExitSwapShareAmountIn
		expectErr error
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg gammtypes.MsgExitSwapShareAmountIn) gammtypes.MsgExitSwapShareAmountIn {
				// Do nothing
				return msg
			}),
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg gammtypes.MsgExitSwapShareAmountIn) gammtypes.MsgExitSwapShareAmountIn {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg gammtypes.MsgExitSwapShareAmountIn) gammtypes.MsgExitSwapShareAmountIn {
				msg.TokenOutDenom = "1"
				return msg
			}),
			expectErr: fmt.Errorf("invalid denom: %s", msg.TokenOutDenom),
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg gammtypes.MsgExitSwapShareAmountIn) gammtypes.MsgExitSwapShareAmountIn {
				msg.ShareInAmount = sdk.NewInt(0)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveRequireAmount, msg.ShareInAmount.String()),
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg gammtypes.MsgExitSwapShareAmountIn) gammtypes.MsgExitSwapShareAmountIn {
				msg.ShareInAmount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveRequireAmount, msg.ShareInAmount.String()),
		},
		{
			name: "zero criteria",
			msg: createMsg(func(msg gammtypes.MsgExitSwapShareAmountIn) gammtypes.MsgExitSwapShareAmountIn {
				msg.TokenOutMinAmount = sdk.NewInt(0)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveCriteria, msg.TokenOutMinAmount.String()),
		},
		{
			name: "negative criteria",
			msg: createMsg(func(msg gammtypes.MsgExitSwapShareAmountIn) gammtypes.MsgExitSwapShareAmountIn {
				msg.TokenOutMinAmount = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.Wrap(gammtypes.ErrNotPositiveCriteria, msg.TokenOutMinAmount.String()),
		},
	}

	for _, test := range tests {
		err := test.msg.ValidateBasic()
		if test.expectErr == nil {
			require.NoError(t, err, "test: %v", test.name)
		} else {
			require.Error(t, err, "test: %v", test.name)
			require.ErrorAs(t, test.expectErr, &err)
		}
	}
}

// Test authz serialize and de-serializes for gamm msg.
func TestAuthzMsg(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	coin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))

	testCases := []struct {
		name    string
		gammMsg sdk.Msg
	}{
		{
			name: "MsgExitSwapExternAmountOut",
			gammMsg: &gammtypes.MsgExitSwapShareAmountIn{
				Sender:            addr1,
				PoolId:            1,
				TokenOutDenom:     "test",
				ShareInAmount:     sdk.NewInt(100),
				TokenOutMinAmount: sdk.NewInt(100),
			},
		},
		{
			name: `MsgExitSwapExternAmountOut`,
			gammMsg: &gammtypes.MsgExitSwapExternAmountOut{
				Sender:           addr1,
				PoolId:           1,
				TokenOut:         coin,
				ShareInMaxAmount: sdk.NewInt(1),
			},
		},
		{
			name: "MsgExitPool",
			gammMsg: &gammtypes.MsgExitPool{
				Sender:        addr1,
				PoolId:        1,
				ShareInAmount: sdk.NewInt(100),
				TokenOutMins:  sdk.NewCoins(coin),
			},
		},
		{
			name: "MsgJoinPool",
			gammMsg: &gammtypes.MsgJoinPool{
				Sender:         addr1,
				PoolId:         1,
				ShareOutAmount: sdk.NewInt(1),
				TokenInMaxs:    sdk.NewCoins(coin),
			},
		},
		{
			name: "MsgJoinSwapExternAmountIn",
			gammMsg: &gammtypes.MsgJoinSwapExternAmountIn{
				Sender:            addr1,
				PoolId:            1,
				TokenIn:           coin,
				ShareOutMinAmount: sdk.NewInt(1),
			},
		},
		{
			name: "MsgJoinSwapShareAmountOut",
			gammMsg: &gammtypes.MsgSwapExactAmountIn{
				Sender: addr1,
				Routes: []gammtypes.SwapAmountInRoute{{
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
			gammMsg: &gammtypes.MsgSwapExactAmountOut{
				Sender: addr1,
				Routes: []gammtypes.SwapAmountOutRoute{{
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
			apptesting.TestMessageAuthzSerialization(t, tc.gammMsg)
		})
	}
}
