package v4_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v21/app"
	v4 "github.com/osmosis-labs/osmosis/v21/app/upgrades/v4"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.OsmosisApp
}

func (s *UpgradeTestSuite) SetupTest() {
	s.app = app.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const dummyUpgradeHeight = 5

func (s *UpgradeTestSuite) TestUpgradePayments() {
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
				err := s.app.BankKeeper.MintCoins(s.ctx, "mint", coins)
				s.Require().NoError(err)
				err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, "mint", "distribution", coins)
				s.Require().NoError(err)
				feePool := s.app.DistrKeeper.GetFeePool(s.ctx)
				feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinFromCoin(coin))
				s.app.DistrKeeper.SetFeePool(s.ctx, feePool)
			},
			func() {
				// run upgrade
				s.ctx = s.ctx.WithBlockHeight(dummyUpgradeHeight - 1)
				plan := upgradetypes.Plan{Name: "v4", Height: dummyUpgradeHeight}
				err := s.app.UpgradeKeeper.ScheduleUpgrade(s.ctx, plan)
				s.Require().NoError(err)
				_, exists := s.app.UpgradeKeeper.GetUpgradePlan(s.ctx)
				s.Require().True(exists)

				s.ctx = s.ctx.WithBlockHeight(dummyUpgradeHeight)
				s.Require().NotPanics(func() {
					beginBlockRequest := abci.RequestBeginBlock{}
					s.app.BeginBlocker(s.ctx, beginBlockRequest)
				})
			},
			func() {
				total := int64(0)

				// check that each account got the payment expected
				payments := v4.GetProp12Payments()
				for _, payment := range payments {
					addr, err := sdk.AccAddressFromBech32(payment[0])
					s.Require().NoError(err)
					amount, err := strconv.ParseInt(strings.TrimSpace(payment[1]), 10, 64)
					s.Require().NoError(err)
					coin := sdk.NewInt64Coin("uosmo", amount)

					accBal := s.app.BankKeeper.GetBalance(s.ctx, addr, "uosmo")
					s.Require().Equal(coin, accBal)

					total += amount
				}

				// check that the total paid out was as expected
				s.Require().Equal(total, int64(367926557424))

				expectedBal := 1000000000000 - total

				// check that distribution module account balance has been reduced correctly
				distAddr := s.app.AccountKeeper.GetModuleAddress("distribution")
				distBal := s.app.BankKeeper.GetBalance(s.ctx, distAddr, "uosmo")
				s.Require().Equal(distBal, sdk.NewInt64Coin("uosmo", expectedBal))

				// check that feepool.communitypool has been reduced correctly
				feePool := s.app.DistrKeeper.GetFeePool(s.ctx)
				s.Require().Equal(feePool.GetCommunityPool(), sdk.NewDecCoins(sdk.NewInt64DecCoin("uosmo", expectedBal)))

				// Check that gamm Minimum Fee has been set correctly

				// Kept as comments for recordkeeping. Since SetParams is now private, the changes being tested for can no longer be made:
				//  	gammParams := s.app.GAMMKeeper.GetParams(suite.ctx)
				//  	expectedCreationFee := sdk.NewCoins(sdk.NewCoin("uosmo", osmomath.OneInt()))
				//  	s.Require().Equal(gammParams.PoolCreationFee, expectedCreationFee)
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			s.SetupTest() // reset

			tc.pre_update()
			tc.update()
			tc.post_update()
		})
	}
}
