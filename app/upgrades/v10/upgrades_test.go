package v10_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	v10 "github.com/osmosis-labs/osmosis/v7/app/upgrades/v10"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.Setup()
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
			"Test that upgrade does a token transfer",
			func() {
				for i, addr := range v10.TransferFromAddresses {
					suite.FundAcc(addr.Addr, sdk.NewCoins(
						sdk.NewInt64Coin(fmt.Sprintf("coin-%d", i), 1),
						sdk.NewInt64Coin("uosmo", 1)))
				}
				balances := suite.App.AppKeepers.BankKeeper.GetAllBalances(suite.Ctx, v10.RecoveryAddress)
				suite.Require().True(balances.Empty())
			},
			func() {
				// run upgrade
				// First run block N-1, begin new block takes ctx height + 1
				suite.Ctx = suite.Ctx.WithBlockHeight(v10.ForkHeight - 2)
				suite.BeginNewBlock(false)
				balances := suite.App.AppKeepers.BankKeeper.GetAllBalances(suite.Ctx, v10.RecoveryAddress)
				suite.Require().True(balances.Empty())

				// run upgrade height
				suite.Require().NotPanics(func() {
					suite.BeginNewBlock(false)
				})
			},
			func() {
				expectedBalance := sdk.NewCoins()
				for i, addr := range v10.TransferFromAddresses {
					expectedBalance = expectedBalance.Add(
						sdk.NewInt64Coin(fmt.Sprintf("coin-%d", i), 1),
						sdk.NewInt64Coin("uosmo", 1))
					balances := suite.App.AppKeepers.BankKeeper.GetAllBalances(suite.Ctx, addr.Addr)
					suite.Require().True(balances.Empty())
				}
				balances := suite.App.AppKeepers.BankKeeper.GetAllBalances(suite.Ctx, v10.RecoveryAddress)
				suite.Require().Equal(expectedBalance, balances)
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
