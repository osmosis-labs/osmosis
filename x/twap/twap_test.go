package twap_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v19/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v19/x/twap/types"

	// "github.com/osmosis-labs/osmosis/v19/app/apptesting"
	"github.com/osmosis-labs/osmosis/v19/x/twap/client"
	"github.com/osmosis-labs/osmosis/v19/x/twap/client/queryproto"
)

func (s *TestSuite) TestGetExtraArithmeticTwap() {
	s.SetupTest()

	client := client.Querier{K: *s.App.TwapKeeper}

	s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(200_000_000)), sdk.NewCoin("uosmo", sdk.NewInt(5_000_000_000))))

	msg := balancer.NewMsgCreateBalancerPool(s.TestAccs[0], balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}, []balancer.PoolAsset{
		{
			Token:  sdk.NewCoin("atom", sdk.NewInt(170082003)),
			Weight: sdk.NewInt(1000000),
		},
		{
			Token:  sdk.NewCoin("uosmo", sdk.NewInt(3278961579)),
			Weight: sdk.NewInt(1000000),
		},
	}, "")

	// Create a pool 3 days ago
	tempCtx := s.Ctx.WithBlockTime(time.Now().Add(-time.Hour * 72))
	poolId, err := s.App.PoolManagerKeeper.CreatePool(tempCtx, msg)
	sp0, err := s.App.PoolManagerKeeper.RouteCalculateSpotPrice(tempCtx, poolId, "atom", "uosmo")
	fmt.Println(poolId, sp0, err)

	// Create 2 swaps to have historical spot spice

	tempCtx = s.Ctx.WithBlockTime(time.Now().Add(-time.Hour * 60))
	_, err = s.App.PoolManagerKeeper.SwapExactAmountIn(tempCtx, s.TestAccs[0], poolId, sdk.NewCoin("uosmo", sdk.NewInt(1_000_000)), "atom", sdk.ZeroInt())
	sp0, err = s.App.PoolManagerKeeper.RouteCalculateSpotPrice(tempCtx, poolId, "atom", "uosmo")

	tempCtx = s.Ctx.WithBlockTime(time.Now().Add(-time.Hour * 50))
	_, err = s.App.PoolManagerKeeper.SwapExactAmountIn(tempCtx, s.TestAccs[0], poolId, sdk.NewCoin("uosmo", sdk.NewInt(2_000_000)), "atom", sdk.ZeroInt())
	sp0, err = s.App.PoolManagerKeeper.RouteCalculateSpotPrice(tempCtx, poolId, "atom", "uosmo")

	s.Ctx = s.Ctx.WithBlockTime(time.Now())

	t1 := dummyTwapRecord(poolId, time.Now().Add(-time.Hour*48), "atom", "uosmo", sdk.MustNewDecFromStr("0.052278737000000000"),
		sdk.MustNewDecFromStr("2084460698.472701991000000000"),
		sdk.MustNewDecFromStr("413622017762.413858430000000000"),
		sdk.ZeroDec())

	t2 := dummyTwapRecord(poolId, time.Now().Add(-time.Hour*24), "atom", "uosmo", sdk.MustNewDecFromStr("0.052246889000000000"),
		sdk.MustNewDecFromStr("2090272238.355103396000000000"),
		sdk.MustNewDecFromStr("415728489555.130694250000000000"),
		sdk.ZeroDec())

	t3 := dummyTwapRecord(poolId, time.Now(), "atom", "uosmo", sdk.MustNewDecFromStr("0.051870691000000000"),
		sdk.MustNewDecFromStr("2095034109.934779828000000000"),
		sdk.MustNewDecFromStr("417471046637.064223410000000000"),
		sdk.ZeroDec())

	s.App.TwapKeeper.StoreNewRecord(s.Ctx, t1)
	s.App.TwapKeeper.StoreNewRecord(s.Ctx, t2)
	s.App.TwapKeeper.StoreNewRecord(s.Ctx, t3)

	endTimeAfterPruned := time.Now().Add(-time.Hour)
	endTimeBeforePruned := time.Now().Add(-time.Hour * 50)
	invalidEndTime := time.Now().Add(time.Hour)

	tests := map[string]struct {
		req          queryproto.ArithmeticTwapRequest
		expErr       bool
		expectedTwap sdk.Dec
	}{
		"In range [now - period, now]": {
			req: queryproto.ArithmeticTwapRequest{
				PoolId:     poolId,
				BaseAsset:  "atom",
				QuoteAsset: "uosmo",
				StartTime:  time.Now().Add(-time.Hour * 24),
				EndTime:    &endTimeAfterPruned,
			},
			expErr: false,
			expectedTwap: sdk.MustNewDecFromStr("19.139895583065242411"),
		},
		"endTime > ctx.BlockTime": {
			req: queryproto.ArithmeticTwapRequest{
				PoolId:     poolId,
				BaseAsset:  "atom",
				QuoteAsset: "uosmo",
				StartTime:  time.Now().Add(-time.Hour * 24),
				EndTime:    &invalidEndTime,
			},
			expErr: true,
		},
		"startTime < ctx.BlockTime - period < endTime": {
			req: queryproto.ArithmeticTwapRequest{
				PoolId:     poolId,
				BaseAsset:  "atom",
				QuoteAsset: "uosmo",
				StartTime:  time.Now().Add(-time.Hour * 50),
				EndTime:    &endTimeAfterPruned,
			},
			expErr: false,
		},
		"startTime < endTime < ctx.BlockTime - period": {
			req: queryproto.ArithmeticTwapRequest{
				PoolId:     poolId,
				BaseAsset:  "atom",
				QuoteAsset: "uosmo",
				StartTime:  time.Now().Add(-time.Hour * 60),
				EndTime:    &endTimeBeforePruned,
			},
			expErr: false,
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			twap, err := client.ExtraArithmeticTwap(s.Ctx, test.req)
			fmt.Println(twap, err)
			if test.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func dummyTwapRecord(poolId uint64, t time.Time, asset0 string, asset1 string, sp0, accum0, accum1, geomAccum sdk.Dec) types.TwapRecord {
	return types.TwapRecord{
		PoolId:      poolId,
		Time:        t,
		Asset0Denom: asset0,
		Asset1Denom: asset1,

		P0LastSpotPrice:             sp0,
		P1LastSpotPrice:             sdk.OneDec().Quo(sp0),
		P0ArithmeticTwapAccumulator: accum0,
		P1ArithmeticTwapAccumulator: accum1,
		GeometricTwapAccumulator:    geomAccum,
	}
}
