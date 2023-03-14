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
)

func (s *KeeperTestSuite) TestGetAllPositionsWithVaryingFreezeTimes() {
	s.Setup()
	defaultAddress := s.TestAccs[0]
	secondAddress := s.TestAccs[1]
	defaultJoinTime := s.Ctx.BlockTime()

	type position struct {
		poolId         uint64
		acc            sdk.AccAddress
		coin0          sdk.Coin
		coin1          sdk.Coin
		lowerTick      int64
		upperTick      int64
		joinTime       time.Time
		freezeDuration time.Duration
	}

	tests := map[string]struct {
		setupPositions []position
	}{
		"no positions": {
			setupPositions: []position{},
		},
		"one position": {
			setupPositions: []position{
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration},
			},
		},
		"multiple positions": {
			setupPositions: []position{
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration + time.Hour},
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration + time.Hour*2},
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration + time.Hour*3},
			},
		},
		"multiple positions, some different owner": {
			setupPositions: []position{
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration + time.Hour},
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration + time.Hour*2},
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration + time.Hour*3},
				{1, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration + time.Hour},
				{1, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration + time.Hour*2},
				{1, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration + time.Hour*3},
			},
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			// Setup.
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
			s.PrepareConcentratedPool()
			expectedPositions := []sdk.Dec{}
			for _, pos := range tc.setupPositions {
				positionLiquidity := s.SetupPosition(pos.poolId, pos.acc, pos.coin0, pos.coin1, pos.lowerTick, pos.upperTick, pos.joinTime, pos.freezeDuration)
				if pos.acc.Equals(defaultAddress) {
					expectedPositions = append(expectedPositions, positionLiquidity)
				}
			}

			// System under test.
			actualPositions, err := s.App.ConcentratedLiquidityKeeper.GetAllPositionsWithVaryingFreezeTimes(s.Ctx, 1, defaultAddress, DefaultLowerTick, DefaultUpperTick)
			s.NoError(err)

			// Assertions.
			s.Equal(expectedPositions, actualPositions)
		})
	}
}

func (s *KeeperTestSuite) TestParseFullPositionFromBytes() {
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
			key:          types.KeyFullPosition(defaultPoolId, defaultAddress, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration),
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
			key:          types.KeyFullPosition(defaultPoolId, defaultAddress, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration),
			val:          []byte{1, 2, 3, 4, 5, 6, 7},
			expectingErr: true,
		},
		"Sufficient test case": {
			key:          types.KeyFullPosition(defaultPoolId, defaultAddress, DefaultLowerTick, DefaultUpperTick, defaultJoinTime, DefaultFreezeDuration),
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
