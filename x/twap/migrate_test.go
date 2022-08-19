package twap_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

func (s *TestSuite) TestMigrateExistingPools() {
	// create two pools before migration
	s.PrepareUni2PoolWithAssets(defaultUniV2Coins[0], defaultUniV2Coins[1])
	s.PrepareBalancerPool()

	// suppose upgrade happened and increment block height and block time
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second * 10))

	// run migration logic
	latestPoolId := s.App.GAMMKeeper.GetNextPoolId(s.Ctx) - 1
	err := s.twapkeeper.MigrateExistingPools(s.Ctx, latestPoolId)
	s.Require().NoError(err)

	upgradeTime := s.Ctx.BlockTime()

	// iterate through all pools, check that all state entries have been correctly updated
	for poolId := 1; poolId <= int(latestPoolId); poolId++ {
		recentTwapRecords, err := s.twapkeeper.GetAllMostRecentRecordsForPool(s.Ctx, uint64(poolId))
		poolDenoms, _ := s.App.GAMMKeeper.GetPoolDenoms(s.Ctx, uint64(poolId))
		denomPairs, _ := types.GetAllUniqueDenomPairs(poolDenoms)
		s.Require().NoError(err)
		s.Require().Equal(len(denomPairs), len(recentTwapRecords))

		// ensure that the migrate logic has been triggered by checking that
		// the twap record time has been updated to the current ctx block time
		s.Require().Equal(upgradeTime, recentTwapRecords[0].Time)

		twapRecord, err := s.twapkeeper.GetMostRecentRecordStoreRepresentation(s.Ctx, uint64(poolId), recentTwapRecords[0].Asset0Denom, recentTwapRecords[0].Asset1Denom)
		s.Require().NoError(err)
		s.Require().Equal(upgradeTime, twapRecord.Time)

		twapRecordBeforeTime, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, uint64(poolId), s.Ctx.BlockTime(), twapRecord.Asset0Denom, twapRecord.Asset1Denom)
		s.Require().NoError(err)
		s.Require().Equal(upgradeTime, twapRecordBeforeTime.Time)
	}
}
