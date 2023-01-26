package concentrated_liquidity_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
)

func (s *KeeperTestSuite) TestGetAllPositionsWithVaryingFreezeTimes() {
	defaultFrozenUntil := s.Ctx.BlockTime().Add(DefaultFreezeDuration)
	address := s.TestAccs[0]
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
		positions []position
	}{
		// "no positions": {
		// 	positions: []position{},
		// },
		"one position": {
			positions: []position{
				{1, address, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil},
				{1, address, DefaultCoin0, DefaultCoin1, DefaultLowerTick - 1, DefaultUpperTick + 1, defaultFrozenUntil},
			},
		},
		// "multiple positions": {
		// 	positions: []position{
		// 		{1, s.TestAccs[0], DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil.Add(time.Hour)},
		// 		{1, s.TestAccs[0], DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil.Add(time.Hour * 2)},
		// 		{1, s.TestAccs[0], DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, defaultFrozenUntil.Add(time.Hour * 3)},
		// 	},
		// },
	}
	for name, tc := range tests {
		s.Run(name, func() {
			fmt.Println("heeeeerreee")
			// Setup.
			s.SetupTest()
			ctx := s.Ctx
			s.PrepareConcentratedPool()
			expectedPositions := []model.Position{}
			for _, pos := range tc.positions {
				expectedPositions = append(expectedPositions, s.SetupPosition(pos.poolId, pos.acc, pos.coin0, pos.coin1, pos.lowerTick, pos.upperTick, pos.frozenUntil))
			}

			// System under test.
			actualPositions, err := s.App.ConcentratedLiquidityKeeper.GetAllPositionsWithVaryingFreezeTimes(ctx, 1, address, DefaultLowerTick, DefaultUpperTick)
			s.NoError(err)

			// Assertions.
			s.Equal(expectedPositions, actualPositions)
		})
	}
}
