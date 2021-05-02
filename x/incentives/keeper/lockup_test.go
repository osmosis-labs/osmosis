package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func (suite *KeeperTestSuite) proceedBlock() {
	header := suite.ctx.BlockHeader()
	suite.ctx = suite.ctx.WithBlockHeader(tmproto.Header{
		Height:  header.Height + 1,
		ChainID: header.ChainID,
		Time:    header.Time.Add(time.Second * 5),
	})
}

func (suite *KeeperTestSuite) TestTotalLocked() {
	suite.SetupTest()

	coins := suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// locked: 10
	// unlocking: 0
	addr, _, _, _ := suite.SetupLockAndPot(false)
	denom := "lptoken"

	amount := suite.app.IncentivesKeeper.GetTotalLockedDenom(suite.ctx, denom)
	suite.Require().Equal(amount, sdk.NewInt(10))

	// locked: 10
	// unlocking: 0
	suite.UnlockAllUnlockableCoins(addr)
	amount = suite.app.IncentivesKeeper.GetTotalLockedDenom(suite.ctx, denom)
	suite.Require().Equal(amount, sdk.NewInt(10))

	// locked: 0
	// unlocking: 10
	suite.BeginUnlockAllNotUnlockings(addr)
	amount = suite.app.IncentivesKeeper.GetTotalLockedDenom(suite.ctx, denom)
	suite.Require().Equal(amount, sdk.NewInt(10))

	// locked: 0
	// unlocking: 0
	suite.proceedBlock()
	suite.UnlockAllUnlockableCoins(addr)
	amount = suite.app.IncentivesKeeper.GetTotalLockedDenom(suite.ctx, denom)
	suite.Require().Equal(amount, sdk.ZeroInt())

	// locked: 30
	// unlocking: 0
	suite.LockTokens(addr, sdk.Coins{sdk.NewInt64Coin(denom, 10)}, time.Second*8)
	suite.LockTokens(addr, sdk.Coins{sdk.NewInt64Coin(denom, 20)}, time.Second*13)
	amount = suite.app.IncentivesKeeper.GetTotalLockedDenom(suite.ctx, denom)
	suite.Require().Equal(amount, sdk.NewInt(30))

	// locked: 0
	// unlocking: 30
	suite.BeginUnlockAllNotUnlockings(addr)
	suite.proceedBlock()
	suite.UnlockAllUnlockableCoins(addr)
	amount = suite.app.IncentivesKeeper.GetTotalLockedDenom(suite.ctx, denom)
	suite.Require().Equal(amount, sdk.NewInt(30))

	// locked: 0
	// unlocking: 20
	suite.proceedBlock()
	suite.UnlockAllUnlockableCoins(addr)
	amount = suite.app.IncentivesKeeper.GetTotalLockedDenom(suite.ctx, denom)
	suite.Require().Equal(amount, sdk.NewInt(20))

	// locked: 0
	// unlocking: 0
	suite.proceedBlock()
	suite.UnlockAllUnlockableCoins(addr)
	amount = suite.app.IncentivesKeeper.GetTotalLockedDenom(suite.ctx, denom)
	suite.Require().Equal(amount, sdk.ZeroInt())
}
