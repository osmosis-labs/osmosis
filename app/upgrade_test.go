package app_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/osmosis-labs/osmosis/app"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.OsmosisApp
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) TestUpgradePayments() {
	testCases := []struct {
		msg         string
		pre_update  func()
		update      func()
		post_update func()
		expPass     bool
	}{
		{
			"Test community pool payouts for Prop 12",
			func() {
				// mint coins to distribution module / feepool.communitypool

				var bal = int64(1000000000000)
				coin := sdk.NewInt64Coin("uosmo", bal)
				coins := sdk.NewCoins(coin)
				suite.app.BankKeeper.MintCoins(suite.ctx, "mint", coins)
				suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, "mint", "distribution", coins)
				feePool := suite.app.DistrKeeper.GetFeePool(suite.ctx)
				feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinFromCoin(coin))
				suite.app.DistrKeeper.SetFeePool(suite.ctx, feePool)

			},
			func() {
				// run upgrade

				plan := upgradetypes.Plan{Name: "v4", Height: 5}
				suite.app.UpgradeKeeper.ScheduleUpgrade(suite.ctx, plan)
				plan, exists := suite.app.UpgradeKeeper.GetUpgradePlan(suite.ctx)
				suite.Require().True(exists)
				suite.Require().NotPanics(func() {
					suite.app.UpgradeKeeper.ApplyUpgrade(suite.ctx.WithBlockHeight(5), plan)
				})
			},
			func() {
				var total = int64(0)

				// check that each account got the payment expected
				payments := app.GetProp12Payments()
				for _, payment := range payments {
					addr, err := sdk.AccAddressFromBech32(payment[0])
					suite.Require().NoError(err)
					amount, err := strconv.ParseInt(strings.TrimSpace(payment[1]), 10, 64)
					suite.Require().NoError(err)
					coin := sdk.NewInt64Coin("uosmo", amount)

					accBal := suite.app.BankKeeper.GetBalance(suite.ctx, addr, "uosmo")
					suite.Require().Equal(coin, accBal)

					total += amount
				}

				//check that the total paid out was as expected
				suite.Require().Equal(total, int64(367926557424))

				expectedBal := 1000000000000 - total

				// check that distribution module account balance has been reduced correctly
				distAddr := suite.app.AccountKeeper.GetModuleAddress("distribution")
				distBal := suite.app.BankKeeper.GetBalance(suite.ctx, distAddr, "uosmo")
				suite.Require().Equal(distBal, sdk.NewInt64Coin("uosmo", expectedBal))

				// check that feepool.communitypool has been reduced correctly
				feePool := suite.app.DistrKeeper.GetFeePool(suite.ctx)
				suite.Require().Equal(feePool.GetCommunityPool(), sdk.NewDecCoins(sdk.NewInt64DecCoin("uosmo", expectedBal)))

			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.pre_update()
			tc.update()
			tc.post_update()

		})
	}
}
