package apptesting

import (
	"sort"

	"github.com/osmosis-labs/osmosis/osmomath"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v32/x/pool-incentives/types"
)

// validates that Group and group Gauge exist
func (s *KeeperTestHelper) ValidateGroupExists(gaugeID uint64) {
	_, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)

	_, err = s.App.IncentivesKeeper.GetGroupByGaugeID(s.Ctx, gaugeID)
	s.Require().NoError(err)
}

// ReplacePoolIncentivesDistrRecords replaces the pool-incentives distribution records with the given records.
func (s *KeeperTestHelper) ReplacePoolIncentivesDistrRecords(records ...poolincentivestypes.DistrRecord) {
	totalWeight := osmomath.ZeroInt()
	for _, record := range records {
		totalWeight = totalWeight.Add(record.Weight)
	}
	s.App.PoolIncentivesKeeper.SetDistrInfo(s.Ctx, poolincentivestypes.DistrInfo{TotalWeight: totalWeight, Records: records})
}

// UpdatePoolIncentivesDistrRecords merges the given records into the pool-incentives distribution records by
// gauge ID, dropping the records with zero weight.
func (s *KeeperTestHelper) UpdatePoolIncentivesDistrRecords(records ...poolincentivestypes.DistrRecord) {
	recordsByGaugeID := make(map[uint64]poolincentivestypes.DistrRecord)
	for _, record := range s.App.PoolIncentivesKeeper.GetDistrInfo(s.Ctx).Records {
		recordsByGaugeID[record.GaugeId] = record
	}
	for _, record := range records {
		recordsByGaugeID[record.GaugeId] = record
	}

	merged := make([]poolincentivestypes.DistrRecord, 0, len(recordsByGaugeID))
	for _, record := range recordsByGaugeID {
		if !record.Weight.IsZero() {
			merged = append(merged, record)
		}
	}
	sort.Slice(merged, func(i, j int) bool { return merged[i].GaugeId < merged[j].GaugeId })
	s.ReplacePoolIncentivesDistrRecords(merged...)
}
