package twap_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

// TestSwapAndEndBlockTriggeringSave tests that if we:
// * create a pool in block 1
// * swap in block 2
// then after block 2 end block, we have saved records for the pool,
// for both block 1 & 2, with distinct spot prices in their records, and accumulators incremented.
func (s *TestSuite) TestSwapAndEndBlockTriggeringSave() {

	tests := []struct {
		name string
		// this pool will be created at height 1
		createPoolFunc func() uint64
		// token to be swaped in at each block height
		tokenInAtBlock map[int64]sdk.Coin
		// block height to stop at
		stopHeight int64
	}{
		{
			name: "create pool; swap",
			createPoolFunc: func() uint64 {
				return s.PrepareUni2PoolWithAssets(
					sdk.NewInt64Coin(denom0, 524912712309487),
					sdk.NewInt64Coin(denom1, 23890479387234),
				)
			},
			tokenInAtBlock: map[int64]sdk.Coin{
				2: sdk.NewInt64Coin(denom1, 120718932),
			},
			stopHeight: 2,
		},
		{
			name: "create pool and swap; swap",
			createPoolFunc: func() uint64 {
				return s.PrepareUni2PoolWithAssets(
					sdk.NewInt64Coin(denom0, 3249712309487),
					sdk.NewInt64Coin(denom1, 88290479387234),
				)
			},
			tokenInAtBlock: map[int64]sdk.Coin{
				1: sdk.NewInt64Coin(denom0, 2340984),
				2: sdk.NewInt64Coin(denom1, 53412184),
			},

			stopHeight: 2,
		},
		{
			name: "create pool; swap; no swap; swap",
			createPoolFunc: func() uint64 {
				return s.PrepareUni2PoolWithAssets(
					sdk.NewInt64Coin(denom0, 984912712309487),
					sdk.NewInt64Coin(denom1, 890888479387234),
				)
			},
			tokenInAtBlock: map[int64]sdk.Coin{
				2: sdk.NewInt64Coin(denom1, 120718932),
				4: sdk.NewInt64Coin(denom1, 9345734),
			},
			stopHeight: 4,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()

			poolId := tc.createPoolFunc()
			twapAtPoolCreation, err := types.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, poolId, denom0, denom1)
			s.Require().NoError(err)
			expectedHistoricalTwap := twapAtPoolCreation

			tokenIn, doSwap := tc.tokenInAtBlock[1]
			if doSwap {
				s.RunCustomSwap(poolId, tokenIn)
				twapAfterSwap, err := types.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, poolId, denom0, denom1)
				s.Require().NoError(err)

				s.Require().NotEqual(twapAtPoolCreation.P0LastSpotPrice, twapAfterSwap.P0LastSpotPrice)
				s.Require().NotEqual(twapAtPoolCreation.P1LastSpotPrice, twapAfterSwap.P1LastSpotPrice)

				expectedHistoricalTwap = twapAfterSwap
			}

			s.EndBlock()
			s.Commit()

			for blockHeight := int64(2); blockHeight <= tc.stopHeight; blockHeight++ {
				tokenIn, doSwap := tc.tokenInAtBlock[blockHeight]
				if !doSwap {
					s.EndBlock()
					s.Commit()

					continue
				}

				s.RunCustomSwap(poolId, tokenIn)

				expectedLatestTwapUpToAccum := s.twapkeeper.UpdateRecord(s.Ctx, expectedHistoricalTwap)
				// ensure different spot prices
				s.Require().NotEqual(expectedHistoricalTwap.P0LastSpotPrice, expectedLatestTwapUpToAccum.P0LastSpotPrice)
				s.Require().NotEqual(expectedHistoricalTwap.P1LastSpotPrice, expectedLatestTwapUpToAccum.P1LastSpotPrice)

				s.EndBlock()
				s.Commit()

				// check records
				historicalOldTwap, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime().Add(time.Second*-2), denom0, denom1)

				s.Require().NoError(err)
				s.Require().Equal(expectedHistoricalTwap, historicalOldTwap)

				expectedHistoricalTwap = expectedLatestTwapUpToAccum

				latestTwap, err := s.twapkeeper.GetMostRecentRecordStoreRepresentation(s.Ctx, poolId, denom0, denom1)
				s.Require().NoError(err)
				s.Require().Equal(latestTwap.P0LastSpotPrice, expectedLatestTwapUpToAccum.P0LastSpotPrice)
				s.Require().Equal(latestTwap.P1LastSpotPrice, expectedLatestTwapUpToAccum.P1LastSpotPrice)

				latestTwap2, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime(), denom0, denom1)
				s.Require().NoError(err)
				s.Require().Equal(latestTwap, latestTwap2)

				// check accumulators incremented - we test details of correct increment in logic

				s.Require().True(latestTwap.P0ArithmeticTwapAccumulator.GT(historicalOldTwap.P0ArithmeticTwapAccumulator))
				s.Require().True(latestTwap.P1ArithmeticTwapAccumulator.GT(historicalOldTwap.P1ArithmeticTwapAccumulator))
			}
		})
	}
}
