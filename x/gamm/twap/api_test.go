package twap_test

import (
	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

func (s *TestSuite) TestGetBeginBlockAccumulatorRecord() {
	poolId := s.PrepareUni2PoolWithAssets(defaultUniV2Coins[0], defaultUniV2Coins[1])
	denomA := defaultUniV2Coins[0].Denom
	denomB := defaultUniV2Coins[1].Denom

	tests := map[string]struct {
		setRecords []types.TwapRecord
		expRecord  types.TwapRecord
		poolId     uint64
		quoteDenom string
		baseDenom  string
		expError   bool
	}{
		"no record": {[]types.TwapRecord{}, types.TwapRecord{}, 4, denomA, denomB, true},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			// setup records
			for _, r := range tc.setRecords {
				// set pool id if not not provided
				if r.PoolId == 0 {
					r.PoolId = poolId
				}
				s.twapkeeper.StoreNewRecord(s.Ctx, r)
			}
			actualRecord, err := s.twapkeeper.GetBeginBlockAccumulatorRecord(s.Ctx, tc.poolId, tc.baseDenom, tc.quoteDenom)
			if tc.expError {
				s.Require().Error(err)
				return
			} else {
				s.Require().NoError(err)
			}
			s.Require().Equal(tc.expRecord, actualRecord)
		})
	}
}
