package downtimedetector_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/types"
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

func (s *KeeperTestSuite) runBlocktimes(times blocktimes) {
	s.Ctx = s.Ctx.WithBlockTime(baseTime)
	s.App.DowntimeKeeper.BeginBlock(s.Ctx)
	for _, duration := range times {
		s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(duration))
		s.App.DowntimeKeeper.BeginBlock(s.Ctx)
	}
}

var (
	abruptRecovery5minDowntime10min   blocktimes = []time.Duration{sec, 10 * min, 5 * min}
	smootherRecovery5minDowntime10min blocktimes = []time.Duration{sec, 10 * min, min, min, min, min, min}
	fifteenMinEndtime                            = abruptRecovery5minDowntime10min.EndTime()
	tenMinEndtime                                = abruptRecovery5minDowntime10min.EndTime().Add(-5 * min)
)

func (s *KeeperTestSuite) TestBeginBlock() {
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
		s.Run(name, func() {
			s.runBlocktimes(test.times)
			s.Require().Equal(test.times.EndTime(), s.Ctx.BlockTime())
			for _, downtime := range test.downtimes {
				lastDowntime, err := s.App.DowntimeKeeper.GetLastDowntimeOfLength(s.Ctx, downtime.Duration)
				s.Require().NoError(err)
				s.Require().Equal(downtime.LastDowntime, lastDowntime)
			}
		})
	}
}

func (s *KeeperTestSuite) TestRecoveryQuery() {
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
		s.Run(name, func() {
			s.runBlocktimes(test.times)
			s.Require().Equal(test.times.EndTime(), s.Ctx.BlockTime())
			for _, query := range test.cases {
				recovered, err := s.App.DowntimeKeeper.RecoveredSinceDowntimeOfLength(
					s.Ctx, query.downtime, query.recovTime)
				s.Require().NoError(err)
				s.Require().Equal(query.expectRecovered, recovered)
			}
		})
	}
}

func (s *KeeperTestSuite) TestRecoveryQueryErrors() {
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
		s.Run(name, func() {
			s.runBlocktimes(test.times)
			_, err := s.App.DowntimeKeeper.RecoveredSinceDowntimeOfLength(
				s.Ctx, test.downtime, test.recovTime)
			s.Require().Error(err)
		})
	}
}

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
