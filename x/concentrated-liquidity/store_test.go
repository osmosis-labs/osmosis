package concentrated_liquidity_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types/genesis"
)

const (
	defaultTickIndex = 1
)

var (
	defaultTickInfo = model.TickInfo{
		LiquidityGross: DefaultLiquidityAmt,
		LiquidityNet:   DefaultLiquidityAmt,
		SpreadRewardGrowthOppositeDirectionOfLastTraversal: DefaultSpreadRewardAccumCoins,
		UptimeTrackers: model.UptimeTrackers{List: wrapUptimeTrackers(getExpectedUptimes().hundredTokensMultiDenom)},
	}

	defaultTick = genesis.FullTick{
		PoolId:    defaultPoolId,
		TickIndex: defaultTickIndex,
		Info:      defaultTickInfo,
	}
)

func (s *KeeperTestSuite) TestParseFullTickFromBytes() {
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
