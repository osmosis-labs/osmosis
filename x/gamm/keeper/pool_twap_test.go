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
				_, err = suite.app.GAMMKeeper.GetRecentPoolTwap(suite.ctx, 100)
				suite.Require().Error(err)
			},
		},
		{
			fn: func() {
				poolId := suite.preparePool()

				// change in time is neccessary in that using same time
				// would omit an error since GetRecentPoolTwap is current time exclusive
				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))
				poolTwap, err := suite.app.GAMMKeeper.GetRecentPoolTwap(ctx, poolId)
				suite.Require().NoError(err)

				var foobarSpotPrice sdk.Dec
				var barbazSpotPrice sdk.Dec
				var bazfooSpotPrice sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooSpotPrice = twapPair.PriceCumulative
					}
				}

				// no changes should exist when any sort of changes didnt happen
				suite.Require().Equal(sdk.MustNewDecFromStr("2.000000000000000000"), foobarSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("1.500000000000000000"), barbazSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.333333333333333333"), bazfooSpotPrice)
			},
		},
		{
			fn: func() {
				poolId := suite.preparePool()

				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))

				foocoin := sdk.NewCoin("foo", sdk.NewInt(10))
				_, err := suite.app.GAMMKeeper.JoinSwapExternAmountIn(ctx, acc1, poolId, foocoin, sdk.ZeroInt())
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 2))
				poolTwap, err := suite.app.GAMMKeeper.GetRecentPoolTwap(ctx, poolId)
				suite.Require().NoError(err)

				var barfooSpotPrice sdk.Dec
				var bazfooSpotPrice sdk.Dec
				var foobarSpotPrice sdk.Dec
				var foobazSpotPrice sdk.Dec
				var barbazSpotPrice sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazSpotPrice = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("0.999999000002004952"), barfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.666666000001336634"), bazfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("4.000003999999980177"), foobarSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("6.000005999999970265"), foobazSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("3.000000000000000000"), barbazSpotPrice)
			},
		},
		{
			fn: func() {
				poolId := suite.preparePool()

				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))

				_, err := suite.app.GAMMKeeper.JoinSwapShareAmountOut(ctx, acc1, poolId, "foo", types.OneShare.MulRaw(10), sdk.NewInt(1000000000000000000))
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 2))
				poolTwap, err := suite.app.GAMMKeeper.GetRecentPoolTwap(ctx, poolId)
				suite.Require().NoError(err)

				var barfooSpotPrice sdk.Dec
				var bazfooSpotPrice sdk.Dec
				var foobarSpotPrice sdk.Dec
				var foobazSpotPrice sdk.Dec
				var barbazSpotPrice sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazSpotPrice = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("0.782236996889973988"), barfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.521491331259982659"), bazfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("5.543121600000001200"), foobarSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("8.314682400000001800"), foobazSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("3.000000000000000000"), barbazSpotPrice)
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
			fn: func() {
				poolId := suite.preparePool()

				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))

				_, err := suite.app.GAMMKeeper.ExitSwapShareAmountIn(ctx, acc1, poolId, "foo", types.OneShare.MulRaw(10), sdk.ZeroInt())
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 2))
				poolTwap, err := suite.app.GAMMKeeper.GetRecentPoolTwap(ctx, poolId)
				suite.Require().NoError(err)

				var barfooSpotPrice sdk.Dec
				var bazfooSpotPrice sdk.Dec
				var foobarSpotPrice sdk.Dec
				var foobazSpotPrice sdk.Dec
				var barbazSpotPrice sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazSpotPrice = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("1.440837857508977912"), barfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.960558571672651941"), bazfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("3.062882399999999507"), foobarSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("4.594323599999999260"), foobazSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("3.000000000000000000"), barbazSpotPrice)
			},
		},
		{
			fn: func() {
				poolId := suite.preparePool()

				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))
				foocoin := sdk.NewCoin("foo", sdk.NewInt(10))

				_, err := suite.app.GAMMKeeper.ExitSwapExternAmountOut(ctx, acc1, poolId, foocoin, sdk.NewInt(1000000000000000000))
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 2))
				poolTwap, err := suite.app.GAMMKeeper.GetRecentPoolTwap(ctx, poolId)
				suite.Require().NoError(err)

				var barfooSpotPrice sdk.Dec
				var bazfooSpotPrice sdk.Dec
				var foobarSpotPrice sdk.Dec
				var foobazSpotPrice sdk.Dec
				var barbazSpotPrice sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazSpotPrice = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("1.000001000002005786"), barfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.666667333334670523"), bazfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("3.999995999999976874"), foobarSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("5.999993999999965310"), foobazSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("3.000000000000000000"), barbazSpotPrice)
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
				poolTwap, err := suite.app.GAMMKeeper.GetRecentPoolTwap(ctx, poolId)
				suite.Require().NoError(err)

				var barfooSpotPrice sdk.Dec
				var bazfooSpotPrice sdk.Dec
				var foobarSpotPrice sdk.Dec
				var foobazSpotPrice sdk.Dec
				var barbazSpotPrice sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazSpotPrice = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("0.999998600002794048"), barfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.666666000001336634"), bazfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("4.000005600004503791"), foobarSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("6.000005999999970265"), foobazSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("2.999998799999967292"), barbazSpotPrice)
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
				poolTwap, err := suite.app.GAMMKeeper.GetRecentPoolTwap(ctx, poolId)
				suite.Require().NoError(err)

				var barfooSpotPrice sdk.Dec
				var bazfooSpotPrice sdk.Dec
				var foobarSpotPrice sdk.Dec
				var foobazSpotPrice sdk.Dec
				var barbazSpotPrice sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "bar" && twapPair.TokenOut == "baz" {
						barbazSpotPrice = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("0.970596008884849791"), barfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.653466672710785818"), bazfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("4.124964897959196572"), foobarSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("6.123698399999983366"), foobazSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("2.969999999999983250"), barbazSpotPrice)
			},
		},
	}

	for _, test := range tests {
		test.fn()
	}
}
