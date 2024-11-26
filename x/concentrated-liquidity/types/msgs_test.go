package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appParams "github.com/osmosis-labs/osmosis/v27/app/params"

	osmosisapp "github.com/osmosis-labs/osmosis/v27/app"
	clmod "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/clmodule"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

type extMsg interface {
	sdk.Msg
	Route() string
	Type() string
	ValidateBasic() error
}

var (
	addr1       string
	addr2       string
	invalidAddr sdk.AccAddress
)

func init() {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 = sdk.AccAddress(pk1.Address()).String()
	pk2 := ed25519.GenPrivKey().PubKey()
	addr2 = sdk.AccAddress(pk2.Address()).String()
	invalidAddr = sdk.AccAddress("invalid")
}

func runValidateBasicTest(t *testing.T, name string, msg extMsg, expectPass bool, expType string) {
	if expectPass {
		require.NoError(t, msg.ValidateBasic(), "test: %v", name)
		require.Equal(t, msg.Route(), types.RouterKey)
		require.Equal(t, msg.Type(), expType)
		encCfg := osmosisapp.GetEncodingConfig().Marshaler
		signers, _, err := encCfg.GetMsgV1Signers(msg)
		require.NoError(t, err)
		require.Equal(t, len(signers), 1)
		require.Equal(t, sdk.AccAddress(signers[0]).String(), addr1)
	} else {
		require.Error(t, msg.ValidateBasic(), "test: %v", name)
	}
}

func TestMsgCreatePosition(t *testing.T) {
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
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("stake", osmomath.OneInt()), sdk.NewCoin("osmo", osmomath.OneInt())),
				TokenMinAmount0: osmomath.OneInt(),
				TokenMinAmount1: osmomath.OneInt(),
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
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("stake", osmomath.OneInt()), sdk.NewCoin("osmo", osmomath.OneInt())),
				TokenMinAmount0: osmomath.OneInt(),
				TokenMinAmount1: osmomath.OneInt(),
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
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("stake", osmomath.OneInt()), sdk.NewCoin("osmo", osmomath.OneInt())),
				TokenMinAmount0: osmomath.OneInt(),
				TokenMinAmount1: osmomath.OneInt(),
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
				TokensProvided:  sdk.Coins{sdk.Coin{Denom: "stake", Amount: osmomath.NewInt(-10)}, sdk.NewCoin("osmo", osmomath.OneInt())},
				TokenMinAmount0: osmomath.OneInt(),
				TokenMinAmount1: osmomath.OneInt(),
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
				TokensProvided:  sdk.Coins{sdk.NewCoin("stake", osmomath.OneInt()), sdk.Coin{Denom: "osmo", Amount: osmomath.NewInt(-10)}},
				TokenMinAmount0: osmomath.OneInt(),
				TokenMinAmount1: osmomath.OneInt(),
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
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("stake", osmomath.ZeroInt()), sdk.NewCoin("osmo", osmomath.ZeroInt())),
				TokenMinAmount0: osmomath.OneInt(),
				TokenMinAmount1: osmomath.OneInt(),
			},
			expectPass: false,
		},
		{
			name: "upper tick is same as lower tick",
			msg: types.MsgCreatePosition{
				PoolId:          1,
				Sender:          addr1,
				LowerTick:       10,
				UpperTick:       10,
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("stake", osmomath.ZeroInt()), sdk.NewCoin("osmo", osmomath.ZeroInt())),
				TokenMinAmount0: osmomath.OneInt(),
				TokenMinAmount1: osmomath.OneInt(),
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
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("stake", osmomath.OneInt()), sdk.NewCoin("osmo", osmomath.OneInt())),
				TokenMinAmount0: osmomath.NewInt(-1),
				TokenMinAmount1: osmomath.NewInt(-1),
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
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("stake", osmomath.OneInt()), sdk.NewCoin("osmo", osmomath.OneInt())),
				TokenMinAmount0: osmomath.ZeroInt(),
				TokenMinAmount1: osmomath.ZeroInt(),
			},
			expectPass: true,
		},
	}

	for _, test := range tests {
		runValidateBasicTest(t, test.name, &test.msg, test.expectPass, types.TypeMsgCreatePosition)
	}
}

func TestMsgAddToPosition(t *testing.T) {
	baseMsg := types.MsgAddToPosition{
		PositionId:      1,
		Sender:          addr1,
		Amount0:         osmomath.OneInt(),
		Amount1:         osmomath.OneInt(),
		TokenMinAmount0: osmomath.OneInt(),
		TokenMinAmount1: osmomath.OneInt(),
	}

	tests := []struct {
		name       string
		msgFn      func() types.MsgAddToPosition
		expectPass bool
	}{
		{
			name:       "proper msg",
			msgFn:      func() types.MsgAddToPosition { return baseMsg },
			expectPass: true,
		},
		{
			name:       "proper msg",
			msgFn:      func() types.MsgAddToPosition { copy := baseMsg; copy.Sender = invalidAddr.String(); return copy },
			expectPass: false,
		},
		{
			name:       "position id zero",
			msgFn:      func() types.MsgAddToPosition { copy := baseMsg; copy.PositionId = 0; return copy },
			expectPass: false,
		},
		{
			name:       "amount0 is negative",
			msgFn:      func() types.MsgAddToPosition { copy := baseMsg; copy.Amount0 = osmomath.OneInt().Neg(); return copy },
			expectPass: false,
		},
		{
			name:       "amount1 is negative",
			msgFn:      func() types.MsgAddToPosition { copy := baseMsg; copy.Amount1 = osmomath.OneInt().Neg(); return copy },
			expectPass: false,
		},
		{
			name: "token min amount0 is negative",
			msgFn: func() types.MsgAddToPosition {
				copy := baseMsg
				copy.TokenMinAmount0 = osmomath.OneInt().Neg()
				return copy
			},
			expectPass: false,
		},
		{
			name: "token min amount1 is negative",
			msgFn: func() types.MsgAddToPosition {
				copy := baseMsg
				copy.TokenMinAmount1 = osmomath.OneInt().Neg()
				return copy
			},
			expectPass: false,
		},
		{
			name: "proper msg",
			// sanity check that above edits weren't mutative
			msgFn:      func() types.MsgAddToPosition { return baseMsg },
			expectPass: true,
		},
	}

	for _, test := range tests {
		msg := test.msgFn()
		runValidateBasicTest(t, test.name, &msg, test.expectPass, types.TypeAddToPosition)
	}
}

func TestMsgFungifyChargedPositions(t *testing.T) {
	var validPositionIds = []uint64{1, 2}

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
		runValidateBasicTest(t, test.name, &test.msg, test.expectPass, types.TypeMsgFungifyChargedPositions)
	}
}

func TestMsgWithdrawPosition(t *testing.T) {
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
				LiquidityAmount: osmomath.OneDec(),
			},
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: types.MsgWithdrawPosition{
				PositionId:      1,
				Sender:          invalidAddr.String(),
				LiquidityAmount: osmomath.OneDec(),
			},
			expectPass: false,
		},
	}
	for _, test := range tests {
		runValidateBasicTest(t, test.name, &test.msg, test.expectPass, types.TypeMsgWithdrawPosition)
	}
}

func TestConcentratedLiquiditySerialization(t *testing.T) {
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
				LiquidityAmount: osmomath.NewDec(100),
			},
		},
		{
			name: "MsgCreatePosition",
			clMsg: &types.MsgCreatePosition{
				PoolId:          defaultPoolId,
				Sender:          addr1,
				LowerTick:       int64(10000),
				UpperTick:       int64(20000),
				TokensProvided:  sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(1000)), sdk.NewCoin("bar", osmomath.NewInt(1000))),
				TokenMinAmount0: osmomath.OneInt(),
				TokenMinAmount1: osmomath.OneInt(),
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
			apptesting.TestMessageAuthzSerialization(t, tc.clMsg, clmod.AppModuleBasic{})
		})
	}
}

func TestMsgTransferPositions(t *testing.T) {
	tests := []struct {
		name       string
		msg        types.MsgTransferPositions
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: types.MsgTransferPositions{
				PositionIds: []uint64{1, 2, 5, 9, 20},
				Sender:      addr1,
				NewOwner:    addr2,
			},
			expectPass: true,
		},
		{
			name: "position ids are not unique",
			msg: types.MsgTransferPositions{
				PositionIds: []uint64{1, 2, 5, 9, 1},
				Sender:      addr1,
				NewOwner:    addr2,
			},
			expectPass: false,
		},
		{
			name: "no position ids",
			msg: types.MsgTransferPositions{
				Sender:   addr1,
				NewOwner: addr2,
			},
			expectPass: false,
		},
		{
			name: "invalid sender",
			msg: types.MsgTransferPositions{
				PositionIds: []uint64{1, 2, 5, 9, 20},
				Sender:      invalidAddr.String(),
				NewOwner:    addr2,
			},
			expectPass: false,
		},
		{
			name: "invalid new owner",
			msg: types.MsgTransferPositions{
				PositionIds: []uint64{1, 2, 5, 9, 20},
				Sender:      addr1,
				NewOwner:    invalidAddr.String(),
			},
			expectPass: false,
		},
		{
			name: "sender and new owner are the same",
			msg: types.MsgTransferPositions{
				PositionIds: []uint64{1, 2, 5, 9, 20},
				Sender:      addr1,
				NewOwner:    addr1,
			},
			expectPass: false,
		},
	}
	for _, test := range tests {
		runValidateBasicTest(t, test.name, &test.msg, test.expectPass, types.TypeMsgTransferPositions)
	}
}
