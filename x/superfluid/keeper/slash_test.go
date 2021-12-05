package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestSlashLockupsForSlashedOnDelegation() {
	valAddr, lock := suite.SetupSuperfluidDelegate()

	expAcc := types.SuperfluidIntermediaryAccount{
		Denom:   lock.Coins[0].Denom,
		ValAddr: valAddr.String(),
	}

	// check delegation from intermediary account to validator
	validator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
	suite.Require().True(found)
	delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAddress(), valAddr)
	suite.Require().True(found)
	suite.Require().Equal(delegation.Shares, sdk.NewDec(1900000)) // 95% x 2 x 1000000
	delegatedTokens := validator.TokensFromShares(delegation.Shares).TruncateInt()
	suite.Require().Equal(delegatedTokens, sdk.NewInt(1900000))

	// slash validator
	suite.ctx = suite.ctx.WithBlockHeight(100)
	consAddr, err := validator.GetConsAddr()
	suite.Require().NoError(err)
	suite.app.StakingKeeper.Slash(suite.ctx, consAddr, 80, 1, sdk.NewDecWithPrec(5, 2))

	// check delegation changes
	validator, found = suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
	suite.Require().True(found)
	delegation, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAddress(), valAddr)
	suite.Require().True(found)
	suite.Require().Equal(delegation.Shares, sdk.NewDec(1900000)) // 95% x 2 x 1000000
	delegatedTokens = validator.TokensFromShares(delegation.Shares).TruncateInt()
	suite.Require().True(delegatedTokens.LT(sdk.NewInt(1900000)))

	// refresh intermediary account delegations
	suite.NotPanics(func() {
		suite.app.SuperfluidKeeper.SlashLockupsForSlashedOnDelegation(suite.ctx)
	})

	// check lock changes after slash
	gotLock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lock.ID)
	suite.Require().NoError(err)
	suite.Require().True(gotLock.Coins.AmountOf("lptoken").LT(sdk.NewInt(1000000)))
}
