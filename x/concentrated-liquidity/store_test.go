package concentrated_liquidity_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
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
		setupPositions    []position
		
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
