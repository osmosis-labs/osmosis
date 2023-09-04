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
	// abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func (s *TestSuite) TestGetExtraArithmeticTwap() {
	s.SetupTest()

	client := client.Querier{K: *s.App.TwapKeeper}

	s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(200_000_000)), sdk.NewCoin("uosmo", sdk.NewInt(5_000_000_000))))

	timeNow := time.Now()

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
	s.Ctx = s.Ctx.WithBlockTime(timeNow.Add(-time.Hour * 72)).WithBlockHeight(1)
	poolId, _ := s.App.PoolManagerKeeper.CreatePool(s.Ctx, msg)
	s.EndBlock()
	s.Commit()

	// Create 2 swaps to have historical spot spice

	s.Ctx = s.App.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime(timeNow.Add(-time.Hour * 60)).WithBlockHeight(2)
	s.App.PoolManagerKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], poolId, sdk.NewCoin("uosmo", sdk.NewInt(1_000_000)), "atom", sdk.ZeroInt())
	sp0, _ := s.App.PoolManagerKeeper.RouteCalculateSpotPrice(s.Ctx, poolId, "uosmo", "atom")
	s.EndBlock()
	s.Commit()
	fmt.Println("sp060", sp0)

	s.Ctx = s.App.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime(timeNow.Add(-time.Hour * 50)).WithBlockHeight(3)
	s.App.PoolManagerKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], poolId, sdk.NewCoin("uosmo", sdk.NewInt(2_000_000)), "atom", sdk.ZeroInt())
	sp0, _ = s.App.PoolManagerKeeper.RouteCalculateSpotPrice(s.Ctx, poolId, "uosmo", "atom")
	s.EndBlock()
	s.Commit()
	fmt.Println("sp050", sp0)

	tempCtx := s.App.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime(timeNow.Add(-time.Hour * 60)).WithBlockHeight(2)
	sp0, _ = s.App.PoolManagerKeeper.RouteCalculateSpotPrice(tempCtx, poolId, "uosmo", "atom")
	fmt.Println("sp0 60", sp0)

	tempCtx = s.App.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime(timeNow.Add(-time.Hour * 50)).WithBlockHeight(3)
	sp0, _ = s.App.PoolManagerKeeper.RouteCalculateSpotPrice(tempCtx, poolId, "uosmo", "atom")
	fmt.Println("sp0 50", sp0)

	s.Ctx = s.Ctx.WithBlockTime(timeNow).WithBlockHeight(4)

	t1 := dummyTwapRecord(poolId, timeNow.Add(-time.Hour*48), "atom", "uosmo", sdk.MustNewDecFromStr("0.052278737000000000"),
		sdk.MustNewDecFromStr("2084460698.472701991000000000"),
		sdk.MustNewDecFromStr("413622017762.413858430000000000"),
		sdk.ZeroDec())

	t2 := dummyTwapRecord(poolId, timeNow.Add(-time.Hour*24), "atom", "uosmo", sdk.MustNewDecFromStr("0.052246889000000000"),
		sdk.MustNewDecFromStr("2090272238.355103396000000000"),
		sdk.MustNewDecFromStr("415728489555.130694250000000000"),
		sdk.ZeroDec())

	t3 := dummyTwapRecord(poolId, timeNow.Add(-time.Hour), "atom", "uosmo", sdk.MustNewDecFromStr("0.051870691000000000"),
		sdk.MustNewDecFromStr("2095034109.934779828000000000"),
		sdk.MustNewDecFromStr("417471046637.064223410000000000"),
		sdk.ZeroDec())

	s.App.TwapKeeper.StoreNewRecord(s.Ctx, t1)
	s.App.TwapKeeper.StoreNewRecord(s.Ctx, t2)
	s.App.TwapKeeper.StoreNewRecord(s.Ctx, t3)

	fmt.Println("calculate", (sdk.MustNewDecFromStr("413482958212.909858430000000000").Sub(sdk.MustNewDecFromStr("412788503389.549858430000000000")).QuoInt64(36000000)))

	endTimeAfterPruned := timeNow.Add(-time.Hour)
	endTimeBeforePruned := timeNow.Add(-time.Hour * 50)
	invalidEndTime := timeNow.Add(time.Hour)

	tests := map[string]struct {
		req          queryproto.ArithmeticTwapRequest
		expErr       bool
		expectedTwap sdk.Dec
	}{
		// We have accumDiff = 417471046637.064223410000000000 - 415728489555.130694250000000000 = 1742557081.933529160000000000
		// timeDelta = 23h = 82800000ms
		// twap = accumDiff / timeDelta = 21.045375385670642028
		"In range [now - period, now]": {
			req: queryproto.ArithmeticTwapRequest{
				PoolId:     poolId,
				BaseAsset:  "atom",
				QuoteAsset: "uosmo",
				StartTime:  timeNow.Add(-time.Hour * 24),
				EndTime:    &endTimeAfterPruned,
			},
			expErr: false,
			expectedTwap: sdk.MustNewDecFromStr("21.045375385670642028"),
		},
		"endTime > ctx.BlockTime": {
			req: queryproto.ArithmeticTwapRequest{
				PoolId:     poolId,
				BaseAsset:  "atom",
				QuoteAsset: "uosmo",
				StartTime:  timeNow.Add(-time.Hour * 24),
				EndTime:    &invalidEndTime,
			},
			expErr: true,
		},
		// We have endTime accum = 417471046637.064223410000000000
		// lastRecord saved (now - 48h) have accum = 413622017762.413858430000000000
		// startTime(now -50h) have spot price = 19.313826320000000000
		// So we will calculate statTime accum = 413622017762.413858430000000000 - 19.313826320000000000 * 7200000 = 413482958212.909858430000000000
		// twap = (417471046637.064223410000000000 - 413482958212.909858430000000000) / 176400000 = 22.608211021283248185
		"startTime < ctx.BlockTime - period < endTime": {
			req: queryproto.ArithmeticTwapRequest{
				PoolId:     poolId,
				BaseAsset:  "atom",
				QuoteAsset: "uosmo",
				StartTime:  timeNow.Add(-time.Hour * 50),
				EndTime:    &endTimeAfterPruned,
			},
			expErr: false,
			expectedTwap: sdk.MustNewDecFromStr("22.608211021283248185"),
		},
		// lastRecord saved (now - 48h) have accum = 413622017762.413858430000000000
		// endTime(now -50h) have spot price = 19.313826320000000000
		// So we will calculate endTime accum = 413622017762.413858430000000000 - 19.313826320000000000 * 7200000 = 413482958212.909858430000000000
		// startTime(now -60h) have spot price = 19.290411760000000000
		// So we will calculate start accum = 413482958212.909858430000000000 - 19.290411760000000000 * 36000000 = 412788503389.549858430000000000
		// twap = (413482958212.909858430000000000 - 412788503389.549858430000000000) / 36000000 = 19.290411760000000000

		"startTime < endTime < ctx.BlockTime - period": {
			req: queryproto.ArithmeticTwapRequest{
				PoolId:     poolId,
				BaseAsset:  "atom",
				QuoteAsset: "uosmo",
				StartTime:  timeNow.Add(-time.Hour * 60),
				EndTime:    &endTimeBeforePruned,
			},
			expErr: false,
			expectedTwap: sdk.MustNewDecFromStr("19.290411760000000000"),
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
				s.Require().Equal(test.expectedTwap, twap.ArithmeticTwap)
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
