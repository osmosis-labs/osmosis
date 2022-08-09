package twap_test

import (
	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

// TestCreatePoolFlow tests that upon a pool being created,
// we have made the correct store entries.
func (s *TestSuite) TestCreateTwoAssetPoolFlow() {
	poolId := s.PrepareUni2PoolWithAssets(defaultUniV2Coins[0], defaultUniV2Coins[1])

	expectedTwap, err := types.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, poolId, denom0, denom1)
	s.Require().NoError(err)

	twap, err := s.twapkeeper.GetMostRecentRecordStoreRepresentation(s.Ctx, poolId, denom0, denom1)
	s.Require().NoError(err)
	s.Require().Equal(expectedTwap, twap)

	twap, err = s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime(), denom0, denom1)
	s.Require().NoError(err)
	s.Require().Equal(expectedTwap, twap)
}

// TODO: Write test for create pool, swap, and EndBlock in same block, we use post-swap spot price in record.

// Tests that after a swap, we are triggering internal tracking logic for a pool.
func (s *TestSuite) TestSwapTriggeringTrackPoolId() {
	poolId := s.PrepareUni2PoolWithAssets(defaultUniV2Coins[0], defaultUniV2Coins[1])
	s.BeginNewBlock(false)
	s.RunBasicSwap(poolId)

	s.Require().Equal([]uint64{poolId}, s.twapkeeper.GetChangedPools(s.Ctx))
}
