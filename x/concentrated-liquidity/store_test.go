package concentrated_liquidity_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/osmosis-labs/osmosis/osmoutils"
	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/genesis"
)

const (
	defaultTickIndex = 1
)

var (
	defaultTickInfo = model.TickInfo{
		LiquidityGross:   DefaultLiquidityAmt,
		LiquidityNet:     DefaultLiquidityAmt,
		FeeGrowthOutside: DefaultFeeAccumCoins,
		UptimeTrackers:   wrapUptimeTrackers(getExpectedUptimes().hundredTokensMultiDenom),
	}

	defaultTick = genesis.FullTick{
		PoolId:    defaultPoolId,
		TickIndex: defaultTickIndex,
		Info:      defaultTickInfo,
	}
)

func (s *KeeperTestSuite) TestParseFullPositionFromBytes() {
	s.Setup()
	defaultAddress := s.TestAccs[0]
	cdc := s.App.AppCodec()
	joinTimeFormat := osmoutils.FormatTimeString
	addrFormat := address.MustLengthPrefix
	defaultJoinTime := time.Unix(0, 0).UTC()

	tests := map[string]struct {
		key          []byte
		val          []byte
		expectingErr bool
	}{
		"Empty val": {
			key:          types.KeyFullPosition(defaultPoolId, defaultAddress, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration, 1),
			val:          []byte{},
			expectingErr: true,
		},
		"Empty key": {
			key:          []byte{},
			val:          cdc.MustMarshal(&sdk.DecProto{Dec: DefaultLiquidityAmt}),
			expectingErr: true,
		},
		"Random key": {
			key:          []byte{112, 12, 14, 4, 5},
			val:          cdc.MustMarshal(&sdk.DecProto{Dec: DefaultLiquidityAmt}),
			expectingErr: true,
		},
		"Using not full key (wrong key)": {
			key:          types.KeyPosition(defaultPoolId, defaultAddress, DefaultLowerTick, DefaultUpperTick),
			val:          cdc.MustMarshal(&sdk.DecProto{Dec: DefaultLiquidityAmt}),
			expectingErr: true,
		},
		"One key separator missing in key": {
			key:          []byte(fmt.Sprintf("%s%s%s%d%s%d%s%d%s%s%s%s", types.PositionPrefix, addrFormat(defaultAddress.Bytes()), "|", defaultPoolId, "|", DefaultLowerTick, "|", DefaultUpperTick, "|", joinTimeFormat(defaultJoinTime), "|", DefaultFreezeDuration.String())),
			val:          cdc.MustMarshal(&sdk.DecProto{Dec: DefaultLiquidityAmt}),
			expectingErr: true,
		},
		"Wrong position prefix": {
			key:          []byte(fmt.Sprintf("%s%s%s%s%d%s%d%s%d%s%s%s%s", []byte{0x01}, "|", addrFormat(defaultAddress), "|", defaultPoolId, "|", DefaultLowerTick, "|", DefaultUpperTick, "|", joinTimeFormat(defaultJoinTime), "|", DefaultFreezeDuration.String())),
			val:          cdc.MustMarshal(&sdk.DecProto{Dec: DefaultLiquidityAmt}),
			expectingErr: true,
		},
		"Wrong poolid": {
			key:          []byte(fmt.Sprintf("%s%s%s%s%d%s%d%s%d%s%s%s%s", types.PositionPrefix, "|", addrFormat(defaultAddress), "|", -1, "|", DefaultLowerTick, "|", DefaultUpperTick, "|", joinTimeFormat(defaultJoinTime), "|", DefaultFreezeDuration.String())),
			val:          cdc.MustMarshal(&sdk.DecProto{Dec: DefaultLiquidityAmt}),
			expectingErr: true,
		},
		"Wrong lower tick": {
			key:          []byte(fmt.Sprintf("%s%s%s%s%d%s%s%s%d%s%s%s%s", types.PositionPrefix, "|", addrFormat(defaultAddress), "|", defaultPoolId, "|", "WrongLowerTick", "|", DefaultUpperTick, "|", joinTimeFormat(defaultJoinTime), "|", DefaultFreezeDuration.String())),
			val:          cdc.MustMarshal(&sdk.DecProto{Dec: DefaultLiquidityAmt}),
			expectingErr: true,
		},
		"Wrong upper tick": {
			key:          []byte(fmt.Sprintf("%s%s%s%s%d%s%d%s%s%s%s%s%s", types.PositionPrefix, "|", addrFormat(defaultAddress), "|", defaultPoolId, "|", DefaultLowerTick, "|", "WrongUpperTick", "|", joinTimeFormat(defaultJoinTime), "|", DefaultFreezeDuration.String())),
			val:          cdc.MustMarshal(&sdk.DecProto{Dec: DefaultLiquidityAmt}),
			expectingErr: true,
		},
		"Wrong join time": {
			key:          []byte(fmt.Sprintf("%s%s%s%s%d%s%d%s%d%s%s%s%s", types.PositionPrefix, "|", addrFormat(defaultAddress), "|", defaultPoolId, "|", DefaultLowerTick, "|", DefaultUpperTick, "|", defaultJoinTime, "|", DefaultFreezeDuration.String())),
			val:          cdc.MustMarshal(&sdk.DecProto{Dec: DefaultLiquidityAmt}),
			expectingErr: true,
		},
		"Wrong freeze duration": {
			key:          []byte(fmt.Sprintf("%s%s%s%s%d%s%d%s%d%s%s%s%s", types.PositionPrefix, "|", addrFormat(defaultAddress), "|", defaultPoolId, "|", DefaultLowerTick, "|", DefaultUpperTick, "|", defaultJoinTime, "|", DefaultFreezeDuration)),
			val:          cdc.MustMarshal(&sdk.DecProto{Dec: DefaultLiquidityAmt}),
			expectingErr: true,
		},
		"Invalid val bytes": {
			key:          types.KeyFullPosition(defaultPoolId, defaultAddress, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration, 1),
			val:          []byte{1, 2, 3, 4, 5, 6, 7},
			expectingErr: true,
		},
		"Sufficient test case": {
			key:          types.KeyFullPosition(defaultPoolId, defaultAddress, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration, 1),
			val:          cdc.MustMarshal(&sdk.DecProto{Dec: DefaultLiquidityAmt}),
			expectingErr: false,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {

			fullPosition, err := cl.ParseFullPositionFromBytes(tc.key, tc.val)
			if tc.expectingErr {
				s.Require().Error(err)
				s.Require().Equal(fullPosition, model.Position{})
			} else {
				s.Require().NoError(err)

				// check result
				s.Require().Equal(defaultPoolId, fullPosition.PoolId)
				s.Require().Equal(DefaultLowerTick, fullPosition.LowerTick)
				s.Require().Equal(DefaultUpperTick, fullPosition.UpperTick)
				s.Require().Equal(defaultJoinTime, fullPosition.JoinTime)
				s.Require().Equal(DefaultFreezeDuration, fullPosition.FreezeDuration)
				s.Require().Equal(DefaultLiquidityAmt, fullPosition.Liquidity)
			}
		})
	}
}

func (s *KeeperTestSuite) TestParseFullTickFromBytes() {
	const (
		emptyKeySeparator   = ""
		invalidKeySeparator = "-"
	)

	var (
		cdc = s.App.AppCodec()

		formatFullKey = func(tickPrefix []byte, poolIdBytes []byte, tickIndexBytes []byte) []byte {
			key := make([]byte, 0)
			key = append(key, tickPrefix...)
			key = append(key, poolIdBytes...)
			key = append(key, tickIndexBytes...)
			return key
		}
	)

	tests := map[string]struct {
		key           []byte
		val           []byte
		expectedValue genesis.FullTick
		expectedErr   error
	}{
		"valid positive tick": {
			key:           types.KeyTick(defaultPoolId, defaultTickIndex),
			val:           cdc.MustMarshal(&defaultTickInfo),
			expectedValue: defaultTick,
		},
		"valid zero tick": {
			key:           types.KeyTick(defaultPoolId, 0),
			val:           cdc.MustMarshal(&defaultTickInfo),
			expectedValue: withTickIndex(defaultTick, 0),
		},
		"valid negative tick": {
			key:           types.KeyTick(defaultPoolId, -1),
			val:           cdc.MustMarshal(&defaultTickInfo),
			expectedValue: withTickIndex(defaultTick, -1),
		},
		"valid negative tick large": {
			key:           types.KeyTick(defaultPoolId, -200),
			val:           cdc.MustMarshal(&defaultTickInfo),
			expectedValue: withTickIndex(defaultTick, -200),
		},
		"empty key": {
			key:         []byte{},
			val:         cdc.MustMarshal(&defaultTickInfo),
			expectedErr: types.ErrKeyNotFound,
		},
		"random key": {
			key: []byte{112, 12, 14, 4, 5},
			val: cdc.MustMarshal(&defaultTickInfo),
			expectedErr: types.InvalidTickKeyByteLengthError{
				Length: 5,
			},
		},
		"using not full key (wrong key)": {
			key: types.KeyTickPrefixByPoolId(defaultPoolId),
			val: cdc.MustMarshal(&defaultTickInfo),
			expectedErr: types.InvalidTickKeyByteLengthError{
				Length: len(types.TickPrefix) + cl.Uint64Bytes,
			},
		},
		"invalid prefix key": {
			key:         formatFullKey(types.PositionPrefix, sdk.Uint64ToBigEndian(defaultPoolId), types.TickIndexToBytes(defaultTickIndex)),
			val:         cdc.MustMarshal(&defaultTickInfo),
			expectedErr: types.InvalidPrefixError{Actual: string(types.PositionPrefix), Expected: string(types.TickPrefix)},
		},
		"invalid value": {
			key:         types.KeyTick(defaultPoolId, defaultTickIndex),
			val:         cdc.MustMarshal(&defaultTick), // should be tick info
			expectedErr: types.ErrValueParse,
		},
		"invalid tick index encoding": {
			// must use types.TickIndexToBytes() on tick index for correct encoding.
			key: formatFullKey(types.TickPrefix, sdk.Uint64ToBigEndian(defaultPoolId), sdk.Uint64ToBigEndian(defaultTickIndex)),
			val: cdc.MustMarshal(&defaultTickInfo),
			expectedErr: types.InvalidTickKeyByteLengthError{
				Length: len(types.TickPrefix) + cl.Uint64Bytes + cl.Uint64Bytes,
			},
		},
		"invalid pool id encoding": {
			// format 1 byte.
			key: formatFullKey(types.TickPrefix, []byte(fmt.Sprintf("%x", defaultPoolId)), types.TickIndexToBytes(defaultTickIndex)),
			val: cdc.MustMarshal(&defaultTickInfo),
			expectedErr: types.InvalidTickKeyByteLengthError{
				Length: len(types.TickPrefix) + 2 + cl.Uint64Bytes,
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {

			fullTick, err := cl.ParseFullTickFromBytes(tc.key, tc.val)
			if tc.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectedErr)
				s.Require().Equal(fullTick, genesis.FullTick{})
			} else {
				s.Require().NoError(err)

				// check result
				s.Require().Equal(tc.expectedValue, fullTick)
			}
		})
	}
}
