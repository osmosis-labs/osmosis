package client_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/twap/client"
	"github.com/osmosis-labs/osmosis/v27/x/twap/client/queryproto"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}

func (suite *QueryTestSuite) SetupTest() {
	suite.Setup()
}

func (suite *QueryTestSuite) TestQueryTwap() {
	suite.SetupTest()

	var (
		coins = sdk.NewCoins(
			sdk.NewInt64Coin("tokenA", 1000),
			sdk.NewInt64Coin("tokenB", 2000),
			sdk.NewInt64Coin("tokenC", 3000),
			sdk.NewInt64Coin("tokenD", 4000),
			sdk.NewInt64Coin("tokenE", 4000), // 4000 intentional
		)
		poolID          = suite.PrepareBalancerPoolWithCoins(coins...)
		validStartTime  = suite.Ctx.BlockTime()
		newBlockTime    = validStartTime.Add(time.Hour)
		startTimeTooOld = validStartTime.Add(-time.Hour)

		// Set current block time one hour from initial.
		ctx = suite.Ctx.WithBlockTime(newBlockTime)
	)

	testCases := []struct {
		name               string
		poolId             uint64
		baseAssetDenom     string
		quoteAssetDenom    string
		startTimeOverwrite *time.Time
		endTime            *time.Time
		expectErr          bool
		result             string
	}{
		{
			name:            "non-existent pool",
			poolId:          0,
			baseAssetDenom:  "tokenA",
			quoteAssetDenom: "tokenB",
			expectErr:       true,
		},
		{
			name:   "missing asset denoms",
			poolId: poolID,

			expectErr: true,
		},
		{
			name:           "missing pool ID and quote denom",
			baseAssetDenom: "tokenA",

			expectErr: true,
		},
		{
			name:            "missing pool ID and base denom",
			quoteAssetDenom: "tokenB",

			expectErr: true,
		},
		{
			name:            "tokenA in terms of tokenB",
			poolId:          poolID,
			baseAssetDenom:  "tokenA",
			quoteAssetDenom: "tokenB",
			endTime:         &newBlockTime,

			result: osmomath.NewDec(2).String(),
		},
		{
			name:            "tokenB in terms of tokenA",
			poolId:          poolID,
			baseAssetDenom:  "tokenB",
			quoteAssetDenom: "tokenA",
			endTime:         &newBlockTime,

			result: osmomath.NewDecWithPrec(5, 1).String(),
		},
		{
			name:            "tokenC in terms of tokenD (rounded decimal of 4/3)",
			poolId:          poolID,
			baseAssetDenom:  "tokenC",
			quoteAssetDenom: "tokenD",
			endTime:         &newBlockTime,

			result: osmomath.MustNewDecFromStr("1.333333330000000000").String(),
		},
		{
			name:            "tokenB in terms of tokenC (rounded decimal of 2/3)",
			poolId:          poolID,
			baseAssetDenom:  "tokenC",
			quoteAssetDenom: "tokenB",
			endTime:         &newBlockTime,
			result:          osmomath.MustNewDecFromStr("0.666666670000000000").String(),
		},
		{
			name:            "tokenD in terms of tokenE (1)",
			poolId:          poolID,
			baseAssetDenom:  "tokenD",
			quoteAssetDenom: "tokenE",
			endTime:         &newBlockTime,

			result: osmomath.OneDec().String(),
		},
		{
			name:            "tokenA in terms of tokenB - no end time",
			poolId:          poolID,
			baseAssetDenom:  "tokenA",
			quoteAssetDenom: "tokenB",
			endTime:         nil,

			result: osmomath.NewDec(2).String(),
		},
		{
			name:            "tokenA in terms of tokenB - end time is empty",
			poolId:          poolID,
			baseAssetDenom:  "tokenA",
			quoteAssetDenom: "tokenB",
			endTime:         &time.Time{},

			result: osmomath.NewDec(2).String(),
		},
		{
			name:               "tokenA in terms of tokenB - start time too old",
			poolId:             poolID,
			baseAssetDenom:     "tokenA",
			quoteAssetDenom:    "tokenB",
			startTimeOverwrite: &startTimeTooOld,

			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name+" arithmetic", func() {
			client := client.Querier{K: *suite.App.TwapKeeper}

			startTime := validStartTime
			if tc.startTimeOverwrite != nil {
				startTime = *tc.startTimeOverwrite
			}

			result, err := client.ArithmeticTwap(ctx, queryproto.ArithmeticTwapRequest{
				PoolId:     tc.poolId,
				BaseAsset:  tc.baseAssetDenom,
				QuoteAsset: tc.quoteAssetDenom,
				StartTime:  startTime,
				EndTime:    tc.endTime,
			})

			if tc.expectErr {
				suite.Require().Error(err, "expected error - ArithmeticTwap")
			} else {
				suite.Require().NoError(err, "unexpected error - ArithmeticTwap")
				suite.Require().Equal(tc.result, result.ArithmeticTwap.String())
			}

			resultToNow, err := client.ArithmeticTwapToNow(ctx, queryproto.ArithmeticTwapToNowRequest{
				PoolId:     tc.poolId,
				BaseAsset:  tc.baseAssetDenom,
				QuoteAsset: tc.quoteAssetDenom,
				StartTime:  startTime,
			})

			if tc.expectErr {
				suite.Require().Error(err, "expected error - ArithmeticTwapToNow")
			} else {
				suite.Require().NoError(err, "unexpected error - ArithmeticTwapToNow")
				suite.Require().Equal(tc.result, resultToNow.ArithmeticTwap.String())
			}
		})

		suite.Run(tc.name+" geometric", func() {
			client := client.Querier{K: *suite.App.TwapKeeper}

			startTime := validStartTime
			if tc.startTimeOverwrite != nil {
				startTime = *tc.startTimeOverwrite
			}

			result, err := client.GeometricTwap(ctx, queryproto.GeometricTwapRequest{
				PoolId:     tc.poolId,
				BaseAsset:  tc.baseAssetDenom,
				QuoteAsset: tc.quoteAssetDenom,
				StartTime:  startTime,
				EndTime:    tc.endTime,
			})

			if tc.expectErr {
				suite.Require().Error(err, "expected error - GeometricTwap")
			} else {
				suite.Require().NoError(err, "unexpected error - GeometricTwap")
				suite.Require().Equal(result.GeometricTwap.String(), result.GeometricTwap.String())
			}

			resultToNow, err := client.GeometricTwapToNow(ctx, queryproto.GeometricTwapToNowRequest{
				PoolId:     tc.poolId,
				BaseAsset:  tc.baseAssetDenom,
				QuoteAsset: tc.quoteAssetDenom,
				StartTime:  startTime,
			})

			if tc.expectErr {
				suite.Require().Error(err, "expected error - GeometricTwapToNow")
			} else {
				suite.Require().NoError(err, "unexpected error - GeometricTwapToNow")
				suite.Require().Equal(result.GeometricTwap.String(), resultToNow.GeometricTwap.String())
			}
		})
	}
}
