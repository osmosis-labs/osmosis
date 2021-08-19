package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/gamm/types"
)

func (suite *KeeperTestSuite) TestJoinPoolTwap() {
	suite.SetupTest()

	tests := []struct {
		fn func()
	}{
		{
			fn: func() {
				poolId := suite.preparePool()

				// JoinPool should not be causing any changes to PoolTwap
				err := suite.app.GAMMKeeper.JoinPool(suite.ctx, acc2, poolId, types.OneShare.MulRaw(50), sdk.Coins{})
				suite.Require().NoError(err)

				// try calling pool twap that does not exist
				_, err = suite.app.GAMMKeeper.GetPoolTwapHistory(suite.ctx, 100, time.Now())
				suite.Require().Error(err)
			},
		},
		{
			fn: func() {
				poolId := suite.preparePool()

				// change in time is neccessary in that using same time
				// would omit an error since GetRecentPoolTwap is current time exclusive
				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))
				poolTwap, err := suite.app.GAMMKeeper.GetPoolTwapHistory(ctx, poolId, time.Now().Add(time.Second))
				suite.Require().NoError(err)

				var foobarSpotPrice sdk.Dec
				var barbazSpotPrice sdk.Dec
				var bazfooSpotPrice sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarSpotPrice = twapPair.SpotPrice
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazSpotPrice = twapPair.SpotPrice
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooSpotPrice = twapPair.SpotPrice
					}
				}

				// no changes should exist when any sort of changes didnt happen
				suite.Require().Equal(sdk.MustNewDecFromStr("2.000000000000000000"), foobarSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("1.500000000000000000"), barbazSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.333333333333333333"), bazfooSpotPrice)
			},
		},
		{
			// test JoinSwapExternAmountIn twap
			fn: func() {
				poolId := suite.preparePool()

				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))
				foocoin := sdk.NewCoin("foo", sdk.NewInt(10))
				_, err := suite.app.GAMMKeeper.JoinSwapExternAmountIn(ctx, acc1, poolId, foocoin, sdk.ZeroInt())
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 2))
				poolTwap, err := suite.app.GAMMKeeper.GetPoolTwapHistory(ctx, poolId, time.Now().Add(time.Second*2))
				suite.Require().NoError(err)

				var barfooPriceCumulative sdk.Dec
				var bazfooPriceCumulative sdk.Dec
				var foobarPriceCumulative sdk.Dec
				var foobazPriceCumulative sdk.Dec
				var barbazPriceCumulative sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazPriceCumulative = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("0.499999000002004952"), barfooPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.333332666668003301"), bazfooPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("2.000003999999980177"), foobarPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("3.000005999999970265"), foobazPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("1.500000000000000000"), barbazPriceCumulative)

				// testing the logics for price cumulation
				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 3))
				_, err = suite.app.GAMMKeeper.JoinSwapExternAmountIn(ctx, acc1, poolId, foocoin, sdk.ZeroInt())
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 4))
				poolTwap, err = suite.app.GAMMKeeper.GetPoolTwapHistory(ctx, poolId, time.Now().Add(time.Second*4))
				suite.Require().NoError(err)

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazPriceCumulative = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("1.499995000018003236"), barfooPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.999996666678668825"), bazfooPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("6.000019999999986783"), foobarPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("9.000029999999980175"), foobazPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("4.500000000000000000"), barbazPriceCumulative)
			},
		},
		{
			// test JoinSwapShareAmountOut twap
			fn: func() {
				poolId := suite.preparePool()

				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))

				_, err := suite.app.GAMMKeeper.JoinSwapShareAmountOut(ctx, acc1, poolId, "foo", types.OneShare.MulRaw(10), sdk.NewInt(1000000000000000000))
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 2))
				poolTwap, err := suite.app.GAMMKeeper.GetPoolTwapHistory(ctx, poolId, time.Now().Add(time.Second*2))
				suite.Require().NoError(err)

				var barfooPriceCumulative sdk.Dec
				var bazfooPriceCumulative sdk.Dec
				var foobarPriceCumulative sdk.Dec
				var foobazPriceCumulative sdk.Dec
				var barbazPriceCumulative sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazPriceCumulative = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("0.282236996889973988"), barfooPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.188157997926649326"), bazfooPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("3.543121600000001200"), foobarPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("5.314682400000001800"), foobazPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("1.500000000000000000"), barbazPriceCumulative)
			},
		},
	}

	for _, test := range tests {
		test.fn()
	}
}

func (suite *KeeperTestSuite) TestExitPoolTwap() {
	suite.SetupTest()

	tests := []struct {
		fn func()
	}{
		{
			// test ExitSwapShareAmountIn twap
			fn: func() {
				poolId := suite.preparePool()

				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))

				_, err := suite.app.GAMMKeeper.ExitSwapShareAmountIn(ctx, acc1, poolId, "foo", types.OneShare.MulRaw(10), sdk.ZeroInt())
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 2))
				poolTwap, err := suite.app.GAMMKeeper.GetPoolTwapHistory(ctx, poolId, time.Now().Add(time.Second*2))
				suite.Require().NoError(err)

				var barfooPriceCumulative sdk.Dec
				var bazfooPriceCumulative sdk.Dec
				var foobarPriceCumulative sdk.Dec
				var foobazPriceCumulative sdk.Dec
				var barbazPriceCumulative sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazPriceCumulative = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("0.940837857508977912"), barfooPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.627225238339318608"), bazfooPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("1.062882399999999507"), foobarPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("1.594323599999999260"), foobazPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("1.500000000000000000"), barbazPriceCumulative)
			},
		},
		{
			// test ExitSwapExternAmountOut twap
			fn: func() {
				poolId := suite.preparePool()

				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))
				foocoin := sdk.NewCoin("foo", sdk.NewInt(10))

				_, err := suite.app.GAMMKeeper.ExitSwapExternAmountOut(ctx, acc1, poolId, foocoin, sdk.NewInt(1000000000000000000))
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 2))
				poolTwap, err := suite.app.GAMMKeeper.GetPoolTwapHistory(ctx, poolId, time.Now().Add(time.Second*2))
				suite.Require().NoError(err)

				var barfooPriceCumulative sdk.Dec
				var bazfooPriceCumulative sdk.Dec
				var foobarPriceCumulative sdk.Dec
				var foobazPriceCumulative sdk.Dec
				var barbazPriceCumulative sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazPriceCumulative = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("0.500001000002005786"), barfooPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.333334000001337190"), bazfooPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("1.999995999999976874"), foobarPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("2.999993999999965310"), foobazPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("1.500000000000000000"), barbazPriceCumulative)
			},
		},
	}

	for _, test := range tests {
		test.fn()
	}
}

func (suite *KeeperTestSuite) TestSwapPoolTwap() {
	suite.SetupTest()

	tests := []struct {
		fn func()
	}{
		{
			fn: func() {
				poolId := suite.preparePool()

				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))
				foocoin := sdk.NewCoin("foo", sdk.NewInt(10))

				_, _, err := suite.app.GAMMKeeper.SwapExactAmountIn(ctx, acc1, poolId, foocoin, "bar", sdk.ZeroInt())
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 2))
				poolTwap, err := suite.app.GAMMKeeper.GetPoolTwapHistory(ctx, poolId, time.Now().Add(time.Second*2))
				suite.Require().NoError(err)

				var barfooPriceCumulative sdk.Dec
				var bazfooPriceCumulative sdk.Dec
				var foobarPriceCumulative sdk.Dec
				var foobazPriceCumulative sdk.Dec
				var barbazPriceCumulative sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazPriceCumulative = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("0.499998600002794048"), barfooPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.333332666668003301"), bazfooPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("2.000005600004503791"), foobarPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("3.000005999999970265"), foobazPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("1.499998799999967292"), barbazPriceCumulative)
			},
		},
		{
			fn: func() {
				poolId := suite.preparePool()

				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))
				barcoin := sdk.NewCoin("bar", sdk.NewInt(100000))

				_, _, err := suite.app.GAMMKeeper.SwapExactAmountOut(ctx, acc1, poolId, "foo", sdk.NewInt(900000000), barcoin)
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 2))
				poolTwap, err := suite.app.GAMMKeeper.GetPoolTwapHistory(ctx, poolId, time.Now().Add(time.Second*2))
				suite.Require().NoError(err)

				var barfooPriceCumulative sdk.Dec
				var bazfooPriceCumulative sdk.Dec
				var foobarPriceCumulative sdk.Dec
				var foobazPriceCumulative sdk.Dec
				var barbazPriceCumulative sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazPriceCumulative = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazPriceCumulative = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("0.470596008884849791"), barfooPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.320133339377452485"), bazfooPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("2.124964897959196572"), foobarPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("3.123698399999983366"), foobazPriceCumulative)
				suite.Require().Equal(sdk.MustNewDecFromStr("1.469999999999983250"), barbazPriceCumulative)
			},
		},
	}

	for _, test := range tests {
		test.fn()
	}
}

func (suite *KeeperTestSuite) TestPoolTwapSpotPrice() {
	suite.SetupTest()

	tests := []struct {
		fn func()
	}{
		{
			fn: func() {
				poolId := suite.preparePool()

				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))
				spotPrice, err := suite.app.GAMMKeeper.GetRecentPoolTwapSpotPrice(ctx, poolId, "foo", "bar", 10)
				suite.Require().NoError(err)
				suite.Require().Equal(sdk.MustNewDecFromStr("2.0"), spotPrice)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 2))
				foocoin := sdk.NewCoin("foo", sdk.NewInt(10))
				_, err = suite.app.GAMMKeeper.JoinSwapExternAmountIn(ctx, acc1, poolId, foocoin, sdk.ZeroInt())
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 3))
				spotPrice, err = suite.app.GAMMKeeper.GetRecentPoolTwapSpotPrice(ctx, poolId, "foo", "bar", 10)
				suite.Require().NoError(err)
				suite.Require().Equal(sdk.MustNewDecFromStr("2.000003999999980177"), spotPrice)
			},
		},
	}
	for _, test := range tests {
		test.fn()
	}
}
