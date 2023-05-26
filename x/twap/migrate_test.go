package twap_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/v15/x/twap/types"
)

func (s *TestSuite) TestMigrateExistingPools() {
	// create two pools before migration
	s.PrepareBalancerPoolWithCoins(defaultTwoAssetCoins[0], defaultTwoAssetCoins[1])
	s.PrepareBalancerPool()

	// suppose upgrade happened and increment block height and block time
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second * 10))

	// run migration logic
	latestPoolId := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx) - 1
	err := s.twapkeeper.MigrateExistingPools(s.Ctx, latestPoolId)
	s.Require().NoError(err)

	// check transient store
	changedPools := s.twapkeeper.GetChangedPools(s.Ctx)
	s.Require().Equal(int(latestPoolId), len(changedPools))

	upgradeTime := s.Ctx.BlockTime()

	// iterate through all pools, check that all state entries have been correctly updated
	for poolId := 1; poolId <= int(latestPoolId); poolId++ {
		recentTwapRecords, err := s.twapkeeper.GetAllMostRecentRecordsForPool(s.Ctx, uint64(poolId))
		s.Require().NoError(err)
		poolDenoms, err := s.App.GAMMKeeper.GetPoolDenoms(s.Ctx, uint64(poolId))
		s.Require().NoError(err)
		denomPairs := types.GetAllUniqueDenomPairs(poolDenoms)
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

func (s *TestSuite) TestMigrateExistingPoolsError() {
	// create two pools before migration
	s.PrepareBalancerPoolWithCoins(defaultTwoAssetCoins[0], defaultTwoAssetCoins[1])
	s.PrepareBalancerPool()

	// suppose upgrade happened and increment block height and block time
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second * 10))

	// run migration logic
	// should error when we try to migrate with pool ID that does not exist
	latestPoolIdPlusOne := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx)
	err := s.twapkeeper.MigrateExistingPools(s.Ctx, latestPoolIdPlusOne)
	s.Require().Error(err)
}

// TestTwapRecord_GeometricTwap_MarshalUnmarshal this test proves that migrations
// to initialize geometric twap accumulators are not required.
// This is because proto marshalling will initialize the field to the zero value.
// Zero value is the expected initialization value for the geometric twap accumulator.
func (suite *TestSuite) TestTwapRecord_GeometricTwap_MarshalUnmarshal() {
	originalRecord := types.TwapRecord{
		Asset0Denom: "uatom",
		Asset1Denom: "uusd",
	}

	suite.Require().True(originalRecord.GeometricTwapAccumulator.IsNil())

	bz, err := proto.Marshal(&originalRecord)
	suite.Require().NoError(err)

	var deserialized types.TwapRecord
	err = proto.Unmarshal(bz, &deserialized)
	suite.Require().NoError(err)

	suite.Require().Equal(originalRecord, deserialized)
	suite.Require().Equal(originalRecord.String(), deserialized.String())

	suite.Require().False(originalRecord.GeometricTwapAccumulator.IsNil())
	suite.Require().Equal(sdk.ZeroDec(), originalRecord.GeometricTwapAccumulator)
}
