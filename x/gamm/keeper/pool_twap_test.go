package keeper_test

import (
	"fmt"
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
				//  would omit an error since GetPoolTwap is current time exclusive
				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))
				poolTwap, err := suite.app.GAMMKeeper.GetPoolTwap(ctx, poolId)
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

				suite.Equal(sdk.NewDec(2), foobarSpotPrice)
				suite.Equal(sdk.NewDecWithPrec(15, 1), barbazSpotPrice)
				suite.Equal(sdk.NewDec(1).Quo(sdk.NewDec(3)), bazfooSpotPrice)
			},
		},
		{
			fn: func() {
				poolId := suite.preparePool()

				fmt.Println("\n===============POOL ID===========")
				fmt.Print(poolId)

				ctx := suite.ctx.WithBlockTime(time.Now().Add(time.Second))

				b, _ := suite.app.GAMMKeeper.GetPoolTwap(ctx, poolId)

				fmt.Println("\n==============POOL TWAP BEFORE CHANGE===========")
				fmt.Print(b.String())
				foocoin := sdk.NewCoin("foo", sdk.NewInt(10))
				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 2))

				_, err := suite.app.GAMMKeeper.JoinSwapExternAmountIn(ctx, acc1, poolId, foocoin, sdk.ZeroInt())
				suite.Require().NoError(err)

				ctx = suite.ctx.WithBlockTime(time.Now().Add(time.Second * 3))
				a, err := suite.app.GAMMKeeper.GetPoolTwap(ctx, poolId)
				suite.Require().NoError(err)

				fmt.Println("\n==============POOL TWAP===========")
				fmt.Print(a.String())

				suite.Require().Equal(poolId, 3)
			},
		},
	}

	for _, test := range tests {
		test.fn()
	}
}
