package v4_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/osmosis-labs/osmosis/v10/app"
	v4 "github.com/osmosis-labs/osmosis/v10/app/upgrades/v4"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
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

const dummyUpgradeHeight = 5

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

				bal := int64(1000000000000)
				coin := sdk.NewInt64Coin("uosmo", bal)
				coins := sdk.NewCoins(coin)
				err := suite.app.BankKeeper.MintCoins(suite.ctx, "mint", coins)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, "mint", "distribution", coins)
				suite.Require().NoError(err)
				feePool := suite.app.DistrKeeper.GetFeePool(suite.ctx)
				feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinFromCoin(coin))
				suite.app.DistrKeeper.SetFeePool(suite.ctx, feePool)
			},
			func() {
				// run upgrade
				suite.ctx = suite.ctx.WithBlockHeight(dummyUpgradeHeight - 1)
				plan := upgradetypes.Plan{Name: "v4", Height: dummyUpgradeHeight}
				err := suite.app.UpgradeKeeper.ScheduleUpgrade(suite.ctx, plan)
				suite.Require().NoError(err)
				plan, exists := suite.app.UpgradeKeeper.GetUpgradePlan(suite.ctx)
				suite.Require().True(exists)

				suite.ctx = suite.ctx.WithBlockHeight(dummyUpgradeHeight)
				suite.Require().NotPanics(func() {
					beginBlockRequest := abci.RequestBeginBlock{}
					suite.app.BeginBlocker(suite.ctx, beginBlockRequest)
				})
			},
			func() {
				total := int64(0)

				// check that each account got the payment expected
				payments := v4.GetProp12Payments()
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

				// check that the total paid out was as expected
				suite.Require().Equal(total, int64(367926557424))

				expectedBal := 1000000000000 - total

				// check that distribution module account balance has been reduced correctly
				distAddr := suite.app.AccountKeeper.GetModuleAddress("distribution")
				distBal := suite.app.BankKeeper.GetBalance(suite.ctx, distAddr, "uosmo")
				suite.Require().Equal(distBal, sdk.NewInt64Coin("uosmo", expectedBal))

				// check that feepool.communitypool has been reduced correctly
				feePool := suite.app.DistrKeeper.GetFeePool(suite.ctx)
				suite.Require().Equal(feePool.GetCommunityPool(), sdk.NewDecCoins(sdk.NewInt64DecCoin("uosmo", expectedBal)))

				// Check that gamm Minimum Fee has been set correctly
				gammParams := suite.app.GAMMKeeper.GetParams(suite.ctx)
				expectedCreationFee := sdk.NewCoins(sdk.NewCoin("uosmo", sdk.OneInt()))
				suite.Require().Equal(gammParams.PoolCreationFee, expectedCreationFee)
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
