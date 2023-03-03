package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	appParams "github.com/osmosis-labs/osmosis/v15/app/params"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

func TestMsgCreatePosition(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	tests := []struct {
		name       string
		msg        types.MsgCreatePosition
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: types.MsgCreatePosition{
				PoolId:          1,
				Sender:          addr1,
				LowerTick:       1,
				UpperTick:       10,
				TokenDesired0:   sdk.NewCoin("stake", sdk.OneInt()),
				TokenDesired1:   sdk.NewCoin("osmo", sdk.OneInt()),
				TokenMinAmount0: sdk.OneInt(),
				TokenMinAmount1: sdk.OneInt(),
			},
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: types.MsgCreatePosition{
				PoolId:          1,
				Sender:          invalidAddr.String(),
				LowerTick:       1,
				UpperTick:       10,
				TokenDesired0:   sdk.NewCoin("stake", sdk.OneInt()),
				TokenDesired1:   sdk.NewCoin("osmo", sdk.OneInt()),
				TokenMinAmount0: sdk.OneInt(),
				TokenMinAmount1: sdk.OneInt(),
			},
			expectPass: false,
		},
		{
			name: "invalid price range, lower tick > upper",
			msg: types.MsgCreatePosition{
				PoolId:          1,
				Sender:          addr1,
				LowerTick:       10,
				UpperTick:       1,
				TokenDesired0:   sdk.NewCoin("stake", sdk.OneInt()),
				TokenDesired1:   sdk.NewCoin("osmo", sdk.OneInt()),
				TokenMinAmount0: sdk.OneInt(),
				TokenMinAmount1: sdk.OneInt(),
			},
			expectPass: false,
		},
		{
			name: "negative token 0 desire",
			msg: types.MsgCreatePosition{
				PoolId:          1,
				Sender:          addr1,
				LowerTick:       1,
				UpperTick:       10,
				TokenDesired0:   sdk.Coin{Denom: "stake", Amount: sdk.NewInt(-10)},
				TokenDesired1:   sdk.NewCoin("osmo", sdk.OneInt()),
				TokenMinAmount0: sdk.OneInt(),
				TokenMinAmount1: sdk.OneInt(),
			},
			expectPass: false,
		},
		{
			name: "negative token 1 desire",
			msg: types.MsgCreatePosition{
				PoolId:          1,
				Sender:          addr1,
				LowerTick:       1,
				UpperTick:       10,
				TokenDesired0:   sdk.NewCoin("stake", sdk.OneInt()),
				TokenDesired1:   sdk.Coin{Denom: "osmo", Amount: sdk.NewInt(-10)},
				TokenMinAmount0: sdk.OneInt(),
				TokenMinAmount1: sdk.OneInt(),
			},
			expectPass: false,
		},
		{
			name: "zero desire",
			msg: types.MsgCreatePosition{
				PoolId:          1,
				Sender:          addr1,
				LowerTick:       1,
				UpperTick:       10,
				TokenDesired0:   sdk.NewCoin("stake", sdk.ZeroInt()),
				TokenDesired1:   sdk.NewCoin("osmo", sdk.ZeroInt()),
				TokenMinAmount0: sdk.OneInt(),
				TokenMinAmount1: sdk.OneInt(),
			},
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: types.MsgCreatePosition{
				PoolId:          1,
				Sender:          addr1,
				LowerTick:       1,
				UpperTick:       10,
				TokenDesired0:   sdk.NewCoin("stake", sdk.OneInt()),
				TokenDesired1:   sdk.NewCoin("osmo", sdk.OneInt()),
				TokenMinAmount0: sdk.NewInt(-1),
				TokenMinAmount1: sdk.NewInt(-1),
			},
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: types.MsgCreatePosition{
				PoolId:          1,
				Sender:          addr1,
				LowerTick:       1,
				UpperTick:       10,
				TokenDesired0:   sdk.NewCoin("stake", sdk.OneInt()),
				TokenDesired1:   sdk.NewCoin("osmo", sdk.OneInt()),
				TokenMinAmount0: sdk.ZeroInt(),
				TokenMinAmount1: sdk.ZeroInt(),
			},
			expectPass: true,
		},
	}

	for _, test := range tests {
		msg := test.msg

		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
			require.Equal(t, msg.Route(), types.RouterKey)
			require.Equal(t, msg.Type(), "create-position")
			signers := msg.GetSigners()
			require.Equal(t, len(signers), 1)
			require.Equal(t, signers[0].String(), addr1)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func TestMsgWithdrawPosition(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	tests := []struct {
		name       string
		msg        types.MsgWithdrawPosition
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: types.MsgWithdrawPosition{
				PoolId:          1,
				Sender:          addr1,
				LowerTick:       1,
				UpperTick:       10,
				LiquidityAmount: sdk.OneDec(),
			},
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: types.MsgWithdrawPosition{
				PoolId:          1,
				Sender:          invalidAddr.String(),
				LowerTick:       1,
				UpperTick:       10,
				LiquidityAmount: sdk.OneDec(),
			},
			expectPass: false,
		},
		{
			name: "invalid price range, lower tick > upper",
			msg: types.MsgWithdrawPosition{
				PoolId:          1,
				Sender:          addr1,
				LowerTick:       10,
				UpperTick:       1,
				LiquidityAmount: sdk.OneDec(),
			},
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: types.MsgWithdrawPosition{
				PoolId:          1,
				Sender:          addr1,
				LowerTick:       1,
				UpperTick:       10,
				LiquidityAmount: sdk.NewDec(-10),
			},
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: types.MsgWithdrawPosition{
				PoolId:          1,
				Sender:          addr1,
				LowerTick:       1,
				UpperTick:       10,
				LiquidityAmount: sdk.ZeroDec(),
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		msg := test.msg

		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
			require.Equal(t, msg.Route(), types.RouterKey)
			require.Equal(t, msg.Type(), "withdraw-position")
			signers := msg.GetSigners()
			require.Equal(t, len(signers), 1)
			require.Equal(t, signers[0].String(), addr1)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func TestConcentratedLiquiditySerialization(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	defaultPoolId := uint64(1)

	testCases := []struct {
		name  string
		clMsg sdk.Msg
	}{
		{
			name: "MsgCreateConcentratedPool",
			clMsg: &model.MsgCreateConcentratedPool{
				Sender:      addr1,
				Denom0:      "foo",
				Denom1:      "bar",
				TickSpacing: uint64(1),
			},
		},
		{
			name: "MsgWithdrawPosition",
			clMsg: &types.MsgWithdrawPosition{
				PoolId:          defaultPoolId,
				Sender:          addr1,
				LowerTick:       int64(10000),
				UpperTick:       int64(20000),
				LiquidityAmount: sdk.NewDec(100),
			},
		},
		{
			name: "MsgCreatePosition",
			clMsg: &types.MsgCreatePosition{
				PoolId:          defaultPoolId,
				Sender:          addr1,
				LowerTick:       int64(10000),
				UpperTick:       int64(20000),
				TokenDesired0:   sdk.NewCoin("foo", sdk.NewInt(1000)),
				TokenDesired1:   sdk.NewCoin("bar", sdk.NewInt(1000)),
				TokenMinAmount0: sdk.OneInt(),
				TokenMinAmount1: sdk.OneInt(),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apptesting.TestMessageAuthzSerialization(t, tc.clMsg)
		})
	}
}
