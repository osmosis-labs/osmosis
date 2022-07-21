package twap_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

func (s *TestSuite) TestGetBeginBlockAccumulatorRecord() {
	poolId := s.PrepareUni2PoolWithAssets(defaultUniV2Coins[0], defaultUniV2Coins[1])
	denomA, denomB := defaultUniV2Coins[1].Denom, defaultUniV2Coins[0].Denom
	initStartRecord := newRecord(s.Ctx.BlockTime(), sdk.OneDec(), sdk.ZeroDec(), sdk.ZeroDec())
	initStartRecord.PoolId, initStartRecord.Height = poolId, s.Ctx.BlockHeight()
	initStartRecord.Asset0Denom, initStartRecord.Asset1Denom = denomA, denomB

	zeroAccumTenPoint1Record := recordWithUpdatedSpotPrice(initStartRecord, sdk.NewDec(10), sdk.NewDecWithPrec(1, 1))

	blankRecord := types.TwapRecord{}
	defaultTime := s.Ctx.BlockTime()

	tPlusOneSec := defaultTime.Add(time.Second)

	tests := map[string]struct {
		// if start record is blank, don't do any sets
		startRecord types.TwapRecord
		// We set it to have the updated time
		expRecord  types.TwapRecord
		time       time.Time
		poolId     uint64
		quoteDenom string
		baseDenom  string
		expError   bool
	}{
		"no record (wrong pool ID)": {blankRecord, blankRecord, defaultTime, 4, denomA, denomB, true},
		"default record":            {blankRecord, initStartRecord, defaultTime, 1, denomA, denomB, false},
		"one second later record":   {blankRecord, recordWithUpdatedAccum(initStartRecord, OneSec, OneSec), tPlusOneSec, 1, denomA, denomB, false},
		"idempotent overwrite":      {initStartRecord, initStartRecord, defaultTime, 1, denomA, denomB, false},
		"idempotent overwrite2":     {initStartRecord, recordWithUpdatedAccum(initStartRecord, OneSec, OneSec), tPlusOneSec, 1, denomA, denomB, false},
		"diff spot price": {zeroAccumTenPoint1Record,
			recordWithUpdatedAccum(zeroAccumTenPoint1Record, OneSec.MulInt64(10), OneSec.QuoInt64(10)),
			tPlusOneSec, 1, denomA, denomB, false},
		// TODO: Overflow
	}
	for name, tc := range tests {
		s.Run(name, func() {
			// setup time
			s.Ctx = s.Ctx.WithBlockTime(tc.time)
			tc.expRecord.Time = tc.time

			// setup record
			initSetRecord := tc.startRecord
			if (tc.startRecord == types.TwapRecord{}) {
				initSetRecord = initStartRecord
			}
			s.twapkeeper.StoreNewRecord(s.Ctx, initSetRecord)

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
