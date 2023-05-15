package downtimedetector_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	"github.com/osmosis-labs/osmosis/v15/x/downtime-detector/types"
)

var (
	baseTime = time.Unix(1257894000, 0).UTC()
	sec      = time.Second
	min      = time.Minute
)

type blocktimes []time.Duration

func (b blocktimes) EndTime() time.Time {
	endTime := baseTime
	for _, d := range b {
		endTime = endTime.Add(d)
	}
	return endTime
}

func (suite *KeeperTestSuite) runBlocktimes(times blocktimes) {
	suite.Ctx = suite.Ctx.WithBlockTime(baseTime)
	suite.App.DowntimeKeeper.BeginBlock(suite.Ctx)
	for _, duration := range times {
		suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(duration))
		suite.App.DowntimeKeeper.BeginBlock(suite.Ctx)
	}
}

var (
	abruptRecovery5minDowntime10min   blocktimes = []time.Duration{sec, 10 * min, 5 * min}
	smootherRecovery5minDowntime10min blocktimes = []time.Duration{sec, 10 * min, min, min, min, min, min}
	fifteenMinEndtime                            = abruptRecovery5minDowntime10min.EndTime()
	tenMinEndtime                                = abruptRecovery5minDowntime10min.EndTime().Add(-5 * min)
)

func (suite *KeeperTestSuite) TestBeginBlock() {
	tests := map[string]struct {
		times     blocktimes
		downtimes []types.GenesisDowntimeEntry
	}{
		"10 min halt, then 5 min halt": {
			times: abruptRecovery5minDowntime10min,
			downtimes: []types.GenesisDowntimeEntry{
				types.NewGenesisDowntimeEntry(types.Downtime_DURATION_1M, fifteenMinEndtime),
				types.NewGenesisDowntimeEntry(types.Downtime_DURATION_3M, fifteenMinEndtime),
				types.NewGenesisDowntimeEntry(types.Downtime_DURATION_5M, fifteenMinEndtime),
				types.NewGenesisDowntimeEntry(types.Downtime_DURATION_10M, tenMinEndtime),
			},
		},
		"10 min halt, then 1 min sequence": {
			times: smootherRecovery5minDowntime10min,
			downtimes: []types.GenesisDowntimeEntry{
				types.NewGenesisDowntimeEntry(types.Downtime_DURATION_1M, fifteenMinEndtime),
				types.NewGenesisDowntimeEntry(types.Downtime_DURATION_2M, tenMinEndtime),
				types.NewGenesisDowntimeEntry(types.Downtime_DURATION_5M, tenMinEndtime),
				types.NewGenesisDowntimeEntry(types.Downtime_DURATION_10M, tenMinEndtime),
			},
		},
	}
	for name, test := range tests {
		suite.Run(name, func() {
			suite.runBlocktimes(test.times)
			suite.Require().Equal(test.times.EndTime(), suite.Ctx.BlockTime())
			for _, downtime := range test.downtimes {
				lastDowntime, err := suite.App.DowntimeKeeper.GetLastDowntimeOfLength(suite.Ctx, downtime.Duration)
				suite.Require().NoError(err)
				suite.Require().Equal(downtime.LastDowntime, lastDowntime)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestRecoveryQuery() {
	type queryTestcase struct {
		downtime        types.Downtime
		recovTime       time.Duration
		expectRecovered bool
	}

	tests := map[string]struct {
		times blocktimes
		cases []queryTestcase
	}{
		"10 min halt, then 5 min halt": {
			times: abruptRecovery5minDowntime10min,
			cases: []queryTestcase{
				{types.Downtime_DURATION_10M, 4 * min, true},
				{types.Downtime_DURATION_10M, 5 * min, true},
				{types.Downtime_DURATION_10M, 6 * min, false},
				{types.Downtime_DURATION_30S, 1 * min, false},
			},
		},
		"10 min halt, then 1 min sequence": {
			times: smootherRecovery5minDowntime10min,
			cases: []queryTestcase{
				{types.Downtime_DURATION_10M, 4 * min, true},
				{types.Downtime_DURATION_10M, 5 * min, true},
				{types.Downtime_DURATION_10M, 6 * min, false},
				{types.Downtime_DURATION_30S, 1 * min, false},
			},
		},
	}
	for name, test := range tests {
		suite.Run(name, func() {
			suite.runBlocktimes(test.times)
			suite.Require().Equal(test.times.EndTime(), suite.Ctx.BlockTime())
			for _, query := range test.cases {
				recovered, err := suite.App.DowntimeKeeper.RecoveredSinceDowntimeOfLength(
					suite.Ctx, query.downtime, query.recovTime)
				suite.Require().NoError(err)
				suite.Require().Equal(query.expectRecovered, recovered)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestRecoveryQueryErrors() {
	tests := map[string]struct {
		times     blocktimes
		downtime  types.Downtime
		recovTime time.Duration
	}{
		"invalid downtime": {
			times:     abruptRecovery5minDowntime10min,
			downtime:  types.Downtime(0x7F),
			recovTime: min,
		},
		"0 recovery time": {
			times:     abruptRecovery5minDowntime10min,
			downtime:  types.Downtime_DURATION_1H,
			recovTime: time.Duration(0),
		},
	}
	for name, test := range tests {
		suite.Run(name, func() {
			suite.runBlocktimes(test.times)
			_, err := suite.App.DowntimeKeeper.RecoveredSinceDowntimeOfLength(
				suite.Ctx, test.downtime, test.recovTime)
			suite.Require().Error(err)
		})
	}
}

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
