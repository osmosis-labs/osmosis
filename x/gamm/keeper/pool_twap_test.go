package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/gamm/types"
)

func (suite *KeeperTestSuite) TestCreatePoolTwap() {
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

				// this case should omit error since ctx time has not changed
				// get pool twap gets pool twap with current time exclusive
				_, err = suite.app.GAMMKeeper.GetPoolTwap(suite.ctx, poolId)
				suite.Require().Error(err)

				// try calling pool twap that does not exist
				_, err = suite.app.GAMMKeeper.GetPoolTwap(suite.ctx, 100)
				suite.Require().Error(err)
			},
		},
		{
			fn: func() {
				poolId := suite.preparePool()

				// change in time is neccessary in that using same time
				// would omit an error since GetPoolTwap is current time exclusive
				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))
				poolTwap, err := suite.app.GAMMKeeper.GetPoolTwap(ctx, poolId)
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

				suite.Equal(sdk.NewDec(2), foobarSpotPrice)
				suite.Equal(sdk.NewDecWithPrec(15, 1), barbazSpotPrice)
				suite.Equal(sdk.NewDec(1).Quo(sdk.NewDec(3)), bazfooSpotPrice)
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
				poolTwap, err := suite.app.GAMMKeeper.GetPoolTwap(ctx, poolId)
				suite.Require().NoError(err)

				var barfooSpotPrice sdk.Dec
				var bazfooSpotPrice sdk.Dec
				var foobarSpotPrice sdk.Dec
				var foobazSpotPrice sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazSpotPrice = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("0.499999000002004952"), barfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.333332666668003301"), bazfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("2.000003999999980177"), foobarSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("3.000005999999970265"), foobazSpotPrice)
			},
		},
		{
			fn: func() {
				poolId := suite.preparePool()

				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))

				_, err := suite.app.GAMMKeeper.JoinSwapShareAmountOut(ctx, acc1, poolId, "foo", types.OneShare.MulRaw(10), sdk.NewInt(1000000000000000000))
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 2))
				poolTwap, err := suite.app.GAMMKeeper.GetPoolTwap(ctx, poolId)
				suite.Require().NoError(err)

				var barfooSpotPrice sdk.Dec
				var bazfooSpotPrice sdk.Dec
				var foobarSpotPrice sdk.Dec
				var foobazSpotPrice sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazSpotPrice = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("0.282236996889973988"), barfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.188157997926649326"), bazfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("3.543121600000001200"), foobarSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("5.314682400000001800"), foobazSpotPrice)
			},
		},
		{
			fn: func() {
				poolId := suite.preparePool()

				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))

				_, err := suite.app.GAMMKeeper.JoinSwapShareAmountOut(ctx, acc1, poolId, "foo", types.OneShare.MulRaw(10), sdk.NewInt(1000000000000000000))
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 2))
				poolTwap, err := suite.app.GAMMKeeper.GetPoolTwap(ctx, poolId)
				suite.Require().NoError(err)

				var barfooSpotPrice sdk.Dec
				var bazfooSpotPrice sdk.Dec
				var foobarSpotPrice sdk.Dec
				var foobazSpotPrice sdk.Dec

				for _, twapPair := range poolTwap.TwapPairs {
					if twapPair.TokenIn == "bar" && twapPair.TokenOut == "foo" {
						barfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "baz" && twapPair.TokenOut == "foo" {
						bazfooSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "bar" {
						foobarSpotPrice = twapPair.PriceCumulative
					} else if twapPair.TokenIn == "foo" && twapPair.TokenOut == "baz" {
						foobazSpotPrice = twapPair.PriceCumulative
					}
				}

				suite.Require().Equal(sdk.MustNewDecFromStr("0.282236996889973988"), barfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("0.188157997926649326"), bazfooSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("3.543121600000001200"), foobarSpotPrice)
				suite.Require().Equal(sdk.MustNewDecFromStr("5.314682400000001800"), foobazSpotPrice)
			},
		},
	}

	for _, test := range tests {
		test.fn()
	}
}
