package downtimedetector_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/types"
)

func (s *KeeperTestSuite) TestImportExport() {
	tests := map[string]struct {
		Downtimes     []types.GenesisDowntimeEntry
		LastBlockTime time.Time
	}{
		"no downtimes": {
			LastBlockTime: baseTime,
		},
		"some downtimes": {
			LastBlockTime: baseTime,
			Downtimes: []types.GenesisDowntimeEntry{
				{Duration: types.Downtime_DURATION_10M, LastDowntime: baseTime.Add(-time.Hour)},
				{Duration: types.Downtime_DURATION_30M, LastDowntime: baseTime.Add(-time.Hour)},
			},
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.Ctx = s.Ctx.WithBlockTime(test.LastBlockTime.Add(time.Hour))
			genState := &types.GenesisState{Downtimes: test.Downtimes, LastBlockTime: test.LastBlockTime}
			s.App.DowntimeKeeper.InitGenesis(s.Ctx, genState)
			exportedState := s.App.DowntimeKeeper.ExportGenesis(s.Ctx)
			s.Require().Equal(test.LastBlockTime, exportedState.LastBlockTime)
			// O(N^2) method of checking downtimes, not concerned with run-time as its bounded.
			for _, downtime := range test.Downtimes {
				found := false
				for _, exportedDowntime := range exportedState.Downtimes {
					if exportedDowntime.Duration == downtime.Duration {
						s.Require().Equal(downtime.LastDowntime, exportedDowntime.LastDowntime)
						found = true
						break
					}
				}
				s.Require().True(found, "downtime %s not found in exported state", downtime.Duration.String())
			}
		})
	}
}
