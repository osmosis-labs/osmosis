package twap_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v12/x/twap/client/queryproto"
	"github.com/osmosis-labs/osmosis/v12/x/twap/types"
)

func (s *TestSuite) TestGetArithmeticTwap_Query() {
	tests := map[string]struct {
		recordsToSet []types.TwapRecord
		ctxTime      time.Time
		request 	 queryproto.ArithmeticTwapRequest
		expResponse  queryproto.ArithmeticTwapResponse
		expErr		 bool
	}{
		"Start and end point to same record, no err": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			request: queryproto.ArithmeticTwapRequest{
				PoolId: baseRecord.PoolId,
				StartTime: baseTime,
				EndTime: &tPlusOne,
				BaseAsset: denom1,
				QuoteAsset: denom0,
			},
			expResponse: queryproto.ArithmeticTwapResponse{ArithmeticTwap: sdk.NewDec(10)},
		},
		"Spot price error exactly at end time, err return": {
			recordsToSet: []types.TwapRecord{withLastErrTime(baseRecord, tPlusOne)},
			ctxTime:      tPlusOneMin,
			request: queryproto.ArithmeticTwapRequest{
				PoolId: baseRecord.PoolId,
				StartTime: baseTime,
				EndTime: &tPlusOne,
				BaseAsset: denom1,
				QuoteAsset: denom0,
			},
			expResponse: queryproto.ArithmeticTwapResponse{ArithmeticTwap: sdk.NewDec(10), IsResponseUnstable: true},
			expErr: true,
		},
		
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.preSetRecords(test.recordsToSet)
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)
			res, err := s.querier.ArithmeticTwap(sdk.WrapSDKContext(s.Ctx), &test.request)
			if test.expErr {
				s.Require().Error(err)
			}
			s.Require().Equal(&test.expResponse, res)
		})
	}
}

func (s *TestSuite) TestGetArithmeticTwapToNow_Query() {
	tests := map[string]struct {
		recordsToSet []types.TwapRecord
		ctxTime      time.Time
		request 	 queryproto.ArithmeticTwapToNowRequest
		expResponse  queryproto.ArithmeticTwapToNowResponse
		expErr		 bool
	}{
		"Start time = record time, no err": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			request: queryproto.ArithmeticTwapToNowRequest{
				PoolId: baseRecord.PoolId,
				StartTime: baseTime,
				BaseAsset: denom1,
				QuoteAsset: denom0,
			},
			expResponse: queryproto.ArithmeticTwapToNowResponse{ArithmeticTwap: sdk.NewDec(10)},
		},
		"Spot price error": {
			recordsToSet: []types.TwapRecord{withLastErrTime(baseRecord, tPlusOne)},
			ctxTime:      tPlusOneMin,
			request: queryproto.ArithmeticTwapToNowRequest{
				PoolId: baseRecord.PoolId,
				StartTime: baseTime,
				BaseAsset: denom1,
				QuoteAsset: denom0,
			},
			expResponse: queryproto.ArithmeticTwapToNowResponse{ArithmeticTwap: sdk.NewDec(10), IsResponseUnstable: true},
			expErr: true,
		},
		
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.preSetRecords(test.recordsToSet)
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)
			res, err := s.querier.ArithmeticTwapToNow(sdk.WrapSDKContext(s.Ctx), &test.request)
			if test.expErr {
				s.Require().Error(err)
			}
			s.Require().Equal(&test.expResponse, res)
		})
	}
}

