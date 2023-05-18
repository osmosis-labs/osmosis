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
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("stake", sdk.OneInt()), sdk.NewCoin("osmo", sdk.OneInt())),
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
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("stake", sdk.OneInt()), sdk.NewCoin("osmo", sdk.OneInt())),
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
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("stake", sdk.OneInt()), sdk.NewCoin("osmo", sdk.OneInt())),
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
				TokensProvided:  sdk.Coins{sdk.Coin{Denom: "stake", Amount: sdk.NewInt(-10)}, sdk.NewCoin("osmo", sdk.OneInt())},
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
				TokensProvided:  sdk.Coins{sdk.NewCoin("stake", sdk.OneInt()), sdk.Coin{Denom: "osmo", Amount: sdk.NewInt(-10)}},
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
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("stake", sdk.ZeroInt()), sdk.NewCoin("osmo", sdk.ZeroInt())),
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
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("stake", sdk.OneInt()), sdk.NewCoin("osmo", sdk.OneInt())),
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
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("stake", sdk.OneInt()), sdk.NewCoin("osmo", sdk.OneInt())),
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
			require.Equal(t, msg.Type(), types.TypeMsgCreatePosition)
			signers := msg.GetSigners()
			require.Equal(t, len(signers), 1)
			require.Equal(t, signers[0].String(), addr1)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func TestMsgFungifyChargedPositions(t *testing.T) {
	appParams.SetAddressPrefixes()
	var (
		pk1              = ed25519.GenPrivKey().PubKey()
		addr1            = sdk.AccAddress(pk1.Address()).String()
		invalidAddr      = sdk.AccAddress("invalid")
		validPositionIds = []uint64{1, 2}
	)

	tests := []struct {
		name       string
		msg        types.MsgFungifyChargedPositions
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: types.MsgFungifyChargedPositions{
				Sender:      addr1,
				PositionIds: validPositionIds,
			},
			expectPass: true,
		},
		{
			name: "error: invalid sender",
			msg: types.MsgFungifyChargedPositions{
				Sender:      invalidAddr.String(),
				PositionIds: validPositionIds,
			},
			expectPass: false,
		},
		{
			name: "error: only one id given, must have at least 2",
			msg: types.MsgFungifyChargedPositions{
				Sender:      addr1,
				PositionIds: []uint64{1},
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		msg := test.msg

		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
			require.Equal(t, msg.Route(), types.RouterKey)
			require.Equal(t, msg.Type(), types.TypeMsgFungifyChargedPositions)
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
				PositionId:      1,
				Sender:          addr1,
				LiquidityAmount: sdk.OneDec(),
			},
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: types.MsgWithdrawPosition{
				PositionId:      1,
				Sender:          invalidAddr.String(),
				LiquidityAmount: sdk.OneDec(),
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		msg := test.msg

		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
			require.Equal(t, msg.Route(), types.RouterKey)
			require.Equal(t, msg.Type(), types.TypeMsgWithdrawPosition)
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
				PositionId:      1,
				Sender:          addr1,
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
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000)), sdk.NewCoin("bar", sdk.NewInt(1000))),
				TokenMinAmount0: sdk.OneInt(),
				TokenMinAmount1: sdk.OneInt(),
			},
		},
		{
			name: "MsgFungifyChargedPositions",
			clMsg: &types.MsgFungifyChargedPositions{
				Sender:      addr1,
				PositionIds: []uint64{1, 2},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apptesting.TestMessageAuthzSerialization(t, tc.clMsg)
		})
	}
}
