package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	query "github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	pooltypes "github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
)

var _ = suite.TestingSuite(nil)

// TestGRPCGaugeByID tests querying gauges via gRPC returns the correct response.
func (s *KeeperTestSuite) TestGRPCGaugeByID() {
	s.SetupTest()

	// create a gauge
	gaugeID, _, coins, startTime := s.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// ensure that querying for a gauge with an ID that doesn't exist returns an error.
	res, err := s.querier.GaugeByID(s.Ctx, &types.GaugeByIDRequest{Id: 1000})
	s.Require().Error(err)
	s.Require().Equal(res, (*types.GaugeByIDResponse)(nil))

	// check that querying a gauge with an ID that exists returns the gauge.
	res, err = s.querier.GaugeByID(s.Ctx, &types.GaugeByIDRequest{Id: gaugeID})
	s.Require().NoError(err)
	s.Require().NotEqual(res.Gauge, nil)
	expectedGauge := types.Gauge{
		Id:          gaugeID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	s.Require().Equal(res.Gauge.String(), expectedGauge.String())
}

// TestGRPCGauges tests querying upcoming and active gauges via gRPC returns the correct response.
func (s *KeeperTestSuite) TestGRPCGauges() {
	s.SetupTest()

	// ensure initially querying gauges returns no gauges
	res, err := s.querier.Gauges(s.Ctx, &types.GaugesRequest{})
	s.Require().NoError(err)
	s.Require().Len(res.Data, 0)

	// create a gauge
	gaugeID, _, coins, startTime := s.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// query gauges again, but this time expect the gauge created earlier in the response
	res, err = s.querier.Gauges(s.Ctx, &types.GaugesRequest{})
	s.Require().NoError(err)
	s.Require().Len(res.Data, 1)
	expectedGauge := types.Gauge{
		Id:          gaugeID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	s.Require().Equal(res.Data[0].String(), expectedGauge.String())

	// create 10 more gauges
	for i := 0; i < 10; i++ {
		s.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)})
		s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Second))
	}

	// check that setting page request limit to 10 will only return 10 out of the 11 gauges
	filter := query.PageRequest{Limit: 10}
	res, err = s.querier.Gauges(s.Ctx, &types.GaugesRequest{Pagination: &filter})
	s.Require().NoError(err)
	s.Require().Len(res.Data, 10)
}

// TestGRPCActiveGauges tests querying active gauges via gRPC returns the correct response.
func (s *KeeperTestSuite) TestGRPCActiveGauges() {
	s.SetupTest()

	// ensure initially querying active gauges returns no gauges
	res, err := s.querier.ActiveGauges(s.Ctx, &types.ActiveGaugesRequest{})
	s.Require().NoError(err)
	s.Require().Len(res.Data, 0)

	// create a gauge and move it from upcoming to active
	gaugeID, gauge, coins, startTime := s.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Second))
	err = s.querier.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge)
	s.Require().NoError(err)

	// query active gauges again, but this time expect the gauge created earlier in the response
	res, err = s.querier.ActiveGauges(s.Ctx, &types.ActiveGaugesRequest{})
	s.Require().NoError(err)
	s.Require().Len(res.Data, 1)
	expectedGauge := types.Gauge{
		Id:          gaugeID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	s.Require().Equal(res.Data[0].String(), expectedGauge.String())

	// create 20 more gauges
	for i := 0; i < 20; i++ {
		_, gauge, _, _ := s.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)})
		s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Second))

		// move the first 9 gauges from upcoming to active (now 10 active gauges, 30 total gauges)
		if i < 9 {
			err = s.querier.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge)
			s.Require().NoError(err)
		}
	}

	// set page request limit to 5, expect only 5 active gauge responses
	res, err = s.querier.ActiveGauges(s.Ctx, &types.ActiveGaugesRequest{Pagination: &query.PageRequest{Limit: 5}})
	s.Require().NoError(err)
	s.Require().Len(res.Data, 5)

	// set page request limit to 15, expect only 10 active gauge responses
	res, err = s.querier.ActiveGauges(s.Ctx, &types.ActiveGaugesRequest{Pagination: &query.PageRequest{Limit: 15}})
	s.Require().NoError(err)
	s.Require().Len(res.Data, 10)
}

// TestGRPCActiveGaugesPerDenom tests querying active gauges by denom via gRPC returns the correct response.
func (s *KeeperTestSuite) TestGRPCActiveGaugesPerDenom() {
	s.SetupTest()

	// ensure initially querying gauges by denom returns no gauges
	res, err := s.querier.ActiveGaugesPerDenom(s.Ctx, &types.ActiveGaugesPerDenomRequest{})
	s.Require().NoError(err)
	s.Require().Len(res.Data, 0)

	// create a gauge
	gaugeID, gauge, coins, startTime := s.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Second))
	err = s.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge)
	s.Require().NoError(err)

	// query gauges by denom again, but this time expect the gauge created earlier in the response
	res, err = s.querier.ActiveGaugesPerDenom(s.Ctx, &types.ActiveGaugesPerDenomRequest{Denom: "lptoken", Pagination: nil})
	s.Require().NoError(err)
	s.Require().Len(res.Data, 1)
	expectedGauge := types.Gauge{
		Id:          gaugeID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	s.Require().Equal(res.Data[0].String(), expectedGauge.String())

	// setup 20 more gauges with the pool denom
	for i := 0; i < 20; i++ {
		_, gauge, _, _ := s.SetupNewGaugeWithDenom(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)}, "pool")
		s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Second))

		// move the first 10 of 20 gauges to an active status
		if i < 10 {
			err = s.querier.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge)
			s.Require().NoError(err)
		}
	}

	// query active gauges by lptoken denom with a page request of 5 should only return one gauge
	res, err = s.querier.ActiveGaugesPerDenom(s.Ctx, &types.ActiveGaugesPerDenomRequest{Denom: "lptoken", Pagination: &query.PageRequest{Limit: 5}})
	s.Require().Len(res.Data, 1)
	s.Require().NoError(err)

	// query active gauges by pool denom with a page request of 5 should return 5 gauges
	res, err = s.querier.ActiveGaugesPerDenom(s.Ctx, &types.ActiveGaugesPerDenomRequest{Denom: "pool", Pagination: &query.PageRequest{Limit: 5}})
	s.Require().Len(res.Data, 5)
	s.Require().NoError(err)

	// query active gauges by pool denom with a page request of 15 should return 10 gauges
	res, err = s.querier.ActiveGaugesPerDenom(s.Ctx, &types.ActiveGaugesPerDenomRequest{Denom: "pool", Pagination: &query.PageRequest{Limit: 15}})
	s.Require().Len(res.Data, 10)
	s.Require().NoError(err)
}

// TestGRPCParams tests querying params via gRPC returns the correct response.
func (s *KeeperTestSuite) TestGRPCParams() {
	s.SetupTest()

	expectedParams := s.App.IncentivesKeeper.GetParams(s.Ctx)
	res, err := s.querier.Params(s.Ctx, &types.ParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(expectedParams, res.Params)
}

// TestGRPCUpcomingGauges tests querying upcoming gauges via gRPC returns the correct response.
func (s *KeeperTestSuite) TestGRPCUpcomingGauges() {
	s.SetupTest()

	// ensure initially querying upcoming gauges returns no gauges
	res, err := s.querier.UpcomingGauges(s.Ctx, &types.UpcomingGaugesRequest{})
	s.Require().NoError(err)
	s.Require().Len(res.Data, 0)

	// create a gauge
	gaugeID, _, coins, startTime := s.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// query upcoming gauges again, but this time expect the gauge created earlier in the response
	res, err = s.querier.UpcomingGauges(s.Ctx, &types.UpcomingGaugesRequest{})
	s.Require().NoError(err)
	s.Require().Len(res.Data, 1)
	expectedGauge := types.Gauge{
		Id:          gaugeID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	s.Require().Equal(res.Data[0].String(), expectedGauge.String())

	// setup 20 more upcoming gauges
	for i := 0; i < 20; i++ {
		_, gauge, _, _ := s.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)})
		s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Second))

		// move the first 9 created gauges to an active status
		// 1 + (20 -9) = 12 upcoming gauges
		if i < 9 {
			err = s.querier.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge)
			s.Require().NoError(err)
		}
	}

	// query upcoming gauges with a page request of 5 should return 5 gauges
	res, err = s.querier.UpcomingGauges(s.Ctx, &types.UpcomingGaugesRequest{Pagination: &query.PageRequest{Limit: 5}})
	s.Require().NoError(err)
	s.Require().Len(res.Data, 5)

	// query upcoming gauges with a page request of 15 should return 12 gauges
	res, err = s.querier.UpcomingGauges(s.Ctx, &types.UpcomingGaugesRequest{Pagination: &query.PageRequest{Limit: 15}})
	s.Require().NoError(err)
	s.Require().Len(res.Data, 12)
}

// TestGRPCUpcomingGaugesPerDenom tests querying upcoming gauges by denom via gRPC returns the correct response.
func (s *KeeperTestSuite) TestGRPCUpcomingGaugesPerDenom() {
	s.SetupTest()

	// ensure initially querying upcoming gauges by denom returns no gauges
	upcomingGaugeRequest := types.UpcomingGaugesPerDenomRequest{Denom: "lptoken", Pagination: nil}
	res, err := s.querier.UpcomingGaugesPerDenom(s.Ctx, &upcomingGaugeRequest)
	s.Require().NoError(err)
	s.Require().Len(res.UpcomingGauges, 0)

	// create a gauge, and check upcoming gauge is working
	gaugeID, gauge, coins, startTime := s.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// query upcoming gauges by denom again, but this time expect the gauge created earlier in the response
	res, err = s.querier.UpcomingGaugesPerDenom(s.Ctx, &upcomingGaugeRequest)
	s.Require().NoError(err)
	s.Require().Len(res.UpcomingGauges, 1)
	expectedGauge := types.Gauge{
		Id:          gaugeID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	s.Require().Equal(res.UpcomingGauges[0].String(), expectedGauge.String())

	// move gauge from upcoming to active
	// ensure the query no longer returns a response
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Second))
	err = s.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge)
	s.Require().NoError(err)
	res, err = s.querier.UpcomingGaugesPerDenom(s.Ctx, &upcomingGaugeRequest)
	s.Require().NoError(err)
	s.Require().Len(res.UpcomingGauges, 0)

	// setup 20 more upcoming gauges with pool denom
	for i := 0; i < 20; i++ {
		_, gauge, _, _ := s.SetupNewGaugeWithDenom(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)}, "pool")
		s.Ctx = s.Ctx.WithBlockTime(startTime.Add(time.Second))

		// move the first 10 created gauges from upcoming to active
		// this leaves 10 upcoming gauges
		if i < 10 {
			err = s.querier.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge)
			s.Require().NoError(err)
		}
	}

	// query upcoming gauges by lptoken denom with a page request of 5 should return 0 gauges
	res, err = s.querier.UpcomingGaugesPerDenom(s.Ctx, &types.UpcomingGaugesPerDenomRequest{Denom: "lptoken", Pagination: &query.PageRequest{Limit: 5}})
	s.Require().NoError(err)
	s.Require().Len(res.UpcomingGauges, 0)

	// query upcoming gauges by pool denom with a page request of 5 should return 5 gauges
	res, err = s.querier.UpcomingGaugesPerDenom(s.Ctx, &types.UpcomingGaugesPerDenomRequest{Denom: "pool", Pagination: &query.PageRequest{Limit: 5}})
	s.Require().NoError(err)
	s.Require().Len(res.UpcomingGauges, 5)

	// query upcoming gauges by pool denom with a page request of 15 should return 10 gauges
	res, err = s.querier.UpcomingGaugesPerDenom(s.Ctx, &types.UpcomingGaugesPerDenomRequest{Denom: "pool", Pagination: &query.PageRequest{Limit: 15}})
	s.Require().NoError(err)
	s.Require().Len(res.UpcomingGauges, 10)
}

// TestGRPCRewardsEst tests querying rewards estimation at a future specific time (by epoch) via gRPC returns the correct response.
func (s *KeeperTestSuite) TestGRPCRewardsEst() {
	s.SetupTest()

	// create an address with no locks
	// ensure rewards estimation returns a nil coins struct
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	res, err := s.querier.RewardsEst(s.Ctx, &types.RewardsEstRequest{
		Owner: lockOwner.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})

	// setup a lock and gauge for a new address
	lockOwner, _, coins, _ := s.SetupLockAndGauge(false)

	// query the rewards of the new address after 100 epochs
	// since it is the only address the gauge is paying out to, the future rewards should equal the entirety of the gauge
	res, err = s.querier.RewardsEst(s.Ctx, &types.RewardsEstRequest{
		Owner:    lockOwner.String(),
		EndEpoch: 100,
	})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, coins)
}

// TestRewardsEstWithPoolIncentives tests querying rewards estimation at a future specific time (by epoch) via gRPC returns the correct response.
// Also changes distribution records for the pool incentives to distribute to the respective lock owner.
func (s *KeeperTestSuite) TestRewardsEstWithPoolIncentives() {
	s.SetupTest()

	// create an address with no locks
	// ensure rewards estimation returns a nil coins struct
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	res, err := s.querier.RewardsEst(s.Ctx, &types.RewardsEstRequest{
		Owner: lockOwner.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})

	// setup a lock and gauge for a new address
	lockOwner, gaugeID, coins, _ := s.SetupLockAndGauge(true)

	// take newly created gauge and modify its pool incentives distribution weight to 100
	distrRecord := pooltypes.DistrRecord{
		GaugeId: gaugeID,
		Weight:  osmomath.NewInt(100),
	}
	err = s.App.PoolIncentivesKeeper.ReplaceDistrRecords(s.Ctx, distrRecord)
	s.Require().NoError(err)

	// query the rewards of the new address after the 10th epoch
	// since it is the only address the gauge is paying out to, the future rewards should equal the entirety of the gauge
	res, err = s.querier.RewardsEst(s.Ctx, &types.RewardsEstRequest{
		Owner:    lockOwner.String(),
		EndEpoch: 10,
	})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, coins)

	// after the current epoch ends, mint more coins that matches the lock coin demon created earlier
	epochIdentifier := s.App.MintKeeper.GetParams(s.Ctx).EpochIdentifier
	curEpochNumber := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, epochIdentifier).CurrentEpoch
	s.App.EpochsKeeper.AfterEpochEnd(s.Ctx, epochIdentifier, curEpochNumber)
	// TODO: Figure out what this number should be
	// TODO: Respond to this
	mintCoins := sdk.NewCoin(coins[0].Denom, osmomath.NewInt(1500000))

	// query the rewards of the new address after the 10th epoch
	// since it is the only address the gauge is paying out to, the future rewards should equal the entirety of the gauge plus the newly minted coins
	res, err = s.querier.RewardsEst(s.Ctx, &types.RewardsEstRequest{
		Owner:    lockOwner.String(),
		EndEpoch: 10,
	})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, coins.Add(mintCoins))
}

// TestGRPCToDistributeCoins tests querying coins that are going to be distributed via gRPC returns the correct response.
func (s *KeeperTestSuite) TestGRPCToDistributeCoins() {
	s.SetupTest()

	// ensure initially querying to distribute coins returns no coins
	res, err := s.querier.ModuleToDistributeCoins(s.Ctx, &types.ModuleToDistributeCoinsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})

	// create two locks with different durations
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	s.LockTokens(addr1, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)
	s.LockTokens(addr2, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, 2*time.Second)

	// setup a non perpetual gauge
	gaugeID, _, coins, startTime := s.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	gauge, err := s.querier.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)
	s.Require().NotNil(gauge)
	gauges := []types.Gauge{*gauge}

	// check to distribute coins after gauge creation
	// ensure this equals the coins within the previously created non perpetual gauge
	res, err = s.querier.ModuleToDistributeCoins(s.Ctx, &types.ModuleToDistributeCoinsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, coins)

	// distribute coins to stakers
	distrCoins, err := s.querier.Distribute(s.Ctx, gauges)
	s.Require().NoError(err)
	s.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})

	// check gauge changes after distribution
	// ensure the gauge's filled epochs have been increased by 1
	// ensure we have distributed 4 out of the 10 stake tokens
	gauge, err = s.querier.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)
	s.Require().NotNil(gauge)
	s.Require().Equal(gauge.FilledEpochs, uint64(1))
	s.Require().Equal(gauge.DistributedCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})
	gauges = []types.Gauge{*gauge}

	// move gauge from an upcoming to an active status
	s.Ctx = s.Ctx.WithBlockTime(startTime)
	err = s.querier.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge)
	s.Require().NoError(err)

	// check that the to distribute coins is equal to the initial gauge coin balance minus what has been distributed already (10-4=6)
	res, err = s.querier.ModuleToDistributeCoins(s.Ctx, &types.ModuleToDistributeCoinsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, coins.Sub(distrCoins...))

	// distribute second round to stakers
	distrCoins, err = s.querier.Distribute(s.Ctx, gauges)
	s.Require().NoError(err)
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 6)}, distrCoins)

	// now that all coins have been distributed (4 in first found 6 in the second round)
	// to distribute coins should be null
	res, err = s.querier.ModuleToDistributeCoins(s.Ctx, &types.ModuleToDistributeCoinsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.Coins, sdk.Coins{})
}

// TestGRPCDistributedCoins tests querying coins that have been distributed via gRPC returns the correct response.
func (s *KeeperTestSuite) TestGRPCDistributedCoins() {
	s.SetupTest()

	// create two locks with different durations
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	s.LockTokens(addr1, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)
	s.LockTokens(addr2, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, 2*time.Second)

	// setup a non perpetual gauge
	gaugeID, _, _, startTime := s.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	gauge, err := s.querier.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)
	s.Require().NotNil(gauge)
	gauges := []types.Gauge{*gauge}

	// move gauge from upcoming to active
	s.Ctx = s.Ctx.WithBlockTime(startTime)
	err = s.querier.MoveUpcomingGaugeToActiveGauge(s.Ctx, *gauge)
	s.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := s.querier.Distribute(s.Ctx, gauges)
	s.Require().NoError(err)
	s.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})

	// check gauge changes after distribution
	// ensure the gauge's filled epochs have been increased by 1
	// ensure we have distributed 4 out of the 10 stake tokens
	gauge, err = s.querier.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)
	s.Require().NotNil(gauge)
	s.Require().Equal(gauge.FilledEpochs, uint64(1))
	s.Require().Equal(gauge.DistributedCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})
	gauges = []types.Gauge{*gauge}

	// distribute second round to stakers
	distrCoins, err = s.querier.Distribute(s.Ctx, gauges)
	s.Require().NoError(err)
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 6)}, distrCoins)
}
