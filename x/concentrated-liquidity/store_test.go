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
	defaultFrozenUntil := s.Ctx.BlockTime().Add(DefaultFreezeDuration)
	defaultAddress := s.TestAccs[0]
	secondAddress := s.TestAccs[1]
	type position struct {
		poolId      uint64
		acc         sdk.AccAddress
		coin0       sdk.Coin
		coin1       sdk.Coin
		lowerTick   int64
		upperTick   int64
		frozenUntil time.Time
	}

	tests := map[string]struct {
		setupPositions []position
	}{
		"no positions": {
			setupPositions: []position{},
		},
		"one position": {
			setupPositions: []position{
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil},
			},
		},
		"multiple positions": {
			setupPositions: []position{
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil.Add(time.Hour)},
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil.Add(time.Hour * 2)},
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil.Add(time.Hour * 3)},
			},
		},
		"multiple positions, some different owner": {
			setupPositions: []position{
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil.Add(time.Hour)},
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil.Add(time.Hour * 2)},
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil.Add(time.Hour * 3)},
				{1, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil.Add(time.Hour)},
				{1, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil.Add(time.Hour * 2)},
				{1, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil.Add(time.Hour * 3)},
			},
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			// Setup.
			s.SetupTest()
			ctx := s.Ctx
			s.PrepareConcentratedPool()
			expectedPositions := []model.Position{}
			for _, pos := range tc.setupPositions {
				position := s.SetupPosition(pos.poolId, pos.acc, pos.coin0, pos.coin1, pos.lowerTick, pos.upperTick, pos.frozenUntil)
				if pos.acc.Equals(defaultAddress) {
					expectedPositions = append(expectedPositions, position)
				}
			}

			// System under test.
			actualPositions, err := s.App.ConcentratedLiquidityKeeper.GetAllPositionsWithVaryingFreezeTimes(ctx, 1, defaultAddress, DefaultLowerTick, DefaultUpperTick)
			s.NoError(err)

			// Assertions.
			s.Equal(expectedPositions, actualPositions)
		})
	}
}

func (s *KeeperTestSuite) TestParseFullPositionFromBytes() {
	defaultFrozenUntil := s.Ctx.BlockTime().Add(DefaultFreezeDuration)
	defaultAddress := s.TestAccs[0]
	cdc := s.App.AppCodec()

	frozenFormat := osmoutils.FormatTimeString
	addrFormat := address.MustLengthPrefix

	tests := map[string]struct {
		key          []byte
		val          []byte
		expectingErr bool
	}{
		"Empty val": {
			key:          types.KeyFullPosition(defaultPoolId, defaultAddress, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil),
			val:          []byte{},
			expectingErr: true,
		},
		"Empty key": {
			key:          []byte{},
			val:          cdc.MustMarshal(&model.Position{Liquidity: DefaultLiquidityAmt, FrozenUntil: defaultFrozenUntil}),
			expectingErr: true,
		},
		"Random key": {
			key:          []byte{112, 12, 14, 4, 5},
			val:          cdc.MustMarshal(&model.Position{Liquidity: DefaultLiquidityAmt, FrozenUntil: defaultFrozenUntil}),
			expectingErr: true,
		},
		"Using not full key (wrong key)": {
			key:          types.KeyPosition(defaultPoolId, defaultAddress, DefaultLowerTick, DefaultUpperTick),
			val:          cdc.MustMarshal(&model.Position{Liquidity: DefaultLiquidityAmt, FrozenUntil: defaultFrozenUntil}),
			expectingErr: true,
		},
		"One key separator missing in key": {
			key:          []byte(fmt.Sprintf("%s%s%s%d%s%d%s%d%s%s", types.PositionPrefix, addrFormat(defaultAddress.Bytes()), "|", defaultPoolId, "|", DefaultLowerTick, "|", DefaultUpperTick, "|", frozenFormat(defaultFrozenUntil))),
			val:          cdc.MustMarshal(&model.Position{Liquidity: DefaultLiquidityAmt, FrozenUntil: defaultFrozenUntil}),
			expectingErr: true,
		},
		"Wrong position prefix": {
			key:          []byte(fmt.Sprintf("%s%s%s%s%d%s%d%s%d%s%s", []byte{0x01}, "|", addrFormat(defaultAddress), "|", defaultPoolId, "|", DefaultLowerTick, "|", DefaultUpperTick, "|", frozenFormat(defaultFrozenUntil))),
			val:          cdc.MustMarshal(&model.Position{Liquidity: DefaultLiquidityAmt, FrozenUntil: defaultFrozenUntil}),
			expectingErr: true,
		},
		"Wrong poolid": {
			key:          []byte(fmt.Sprintf("%s%s%s%s%d%s%d%s%d%s%s", types.PositionPrefix, "|", addrFormat(defaultAddress), "|", -1, "|", DefaultLowerTick, "|", DefaultUpperTick, "|", frozenFormat(defaultFrozenUntil))),
			val:          cdc.MustMarshal(&model.Position{Liquidity: DefaultLiquidityAmt, FrozenUntil: defaultFrozenUntil}),
			expectingErr: true,
		},
		"Wrong lower tick": {
			key:          []byte(fmt.Sprintf("%s%s%s%s%d%s%s%s%d%s%s", types.PositionPrefix, "|", addrFormat(defaultAddress), "|", defaultPoolId, "|", "WrongLowerTick", "|", DefaultUpperTick, "|", frozenFormat(defaultFrozenUntil))),
			val:          cdc.MustMarshal(&model.Position{Liquidity: DefaultLiquidityAmt, FrozenUntil: defaultFrozenUntil}),
			expectingErr: true,
		},
		"Wrong upper tick": {
			key:          []byte(fmt.Sprintf("%s%s%s%s%d%s%d%s%s%s%s", types.PositionPrefix, "|", addrFormat(defaultAddress), "|", defaultPoolId, "|", DefaultLowerTick, "|", "WrongUpperTick", "|", frozenFormat(defaultFrozenUntil))),
			val:          cdc.MustMarshal(&model.Position{Liquidity: DefaultLiquidityAmt, FrozenUntil: defaultFrozenUntil}),
			expectingErr: true,
		},
		"Wrong frozen until": {
			key:          []byte(fmt.Sprintf("%s%s%s%s%d%s%d%s%d%s%s", types.PositionPrefix, "|", addrFormat(defaultAddress), "|", defaultPoolId, "|", DefaultLowerTick, "|", DefaultUpperTick, "|", defaultFrozenUntil)),
			val:          cdc.MustMarshal(&model.Position{Liquidity: DefaultLiquidityAmt, FrozenUntil: defaultFrozenUntil}),
			expectingErr: true,
		},
		"Invalid val bytes": {
			key:          types.KeyFullPosition(defaultPoolId, defaultAddress, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil),
			val:          []byte{1, 2, 3, 4, 5, 6, 7},
			expectingErr: true,
		},
		"Sufficient test case": {
			key:          types.KeyFullPosition(defaultPoolId, defaultAddress, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil),
			val:          cdc.MustMarshal(&model.Position{Liquidity: DefaultLiquidityAmt, FrozenUntil: defaultFrozenUntil}),
			expectingErr: false,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			fullPosition, err := cl.ParseFullPositionFromBytes(tc.key, tc.val)
			if tc.expectingErr {
				s.Require().Error(err)
				s.Require().Equal(fullPosition, types.FullPositionByOwnerResult{})
			} else {
				s.Require().NoError(err)

				// check result
				s.Require().Equal(defaultPoolId, fullPosition.PoolId)
				s.Require().Equal(DefaultLowerTick, fullPosition.LowerTick)
				s.Require().Equal(DefaultUpperTick, fullPosition.UpperTick)
				s.Require().Equal(defaultFrozenUntil, fullPosition.FrozenUntil)
				s.Require().Equal(DefaultLiquidityAmt, fullPosition.Liquidity)

			}
		})
	}
}
