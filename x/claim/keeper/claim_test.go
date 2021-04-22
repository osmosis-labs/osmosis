package keeper_test

import (
	"github.com/c-osmosis/osmosis/x/claim/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (suite *KeeperTestSuite) TestAirdropFlow() {
	suite.SetupTest()

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	balances := []banktypes.Balance{
		{
			Address: addr1.String(),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)),
		},
		{
			Address: addr2.String(),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 20)),
		},
	}
	err := suite.app.ClaimKeeper.SetClaimables(suite.ctx, balances)
	suite.Require().NoError(err)

	coins1, err := suite.app.ClaimKeeper.GetClaimable(suite.ctx, addr1.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins1, balances[0].Coins)

	coins2, err := suite.app.ClaimKeeper.GetClaimable(suite.ctx, addr2.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins2, balances[1].Coins)

	coins3, err := suite.app.ClaimKeeper.GetClaimable(suite.ctx, addr3.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins3, sdk.Coins{})

	// get withdrawable amount before activity
	coins4, err := suite.app.ClaimKeeper.GetWithdrawableByActivity(suite.ctx, addr1.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins4.String(), sdk.NewCoins().String())

	// try to claim before any activity
	coins1, err = suite.app.ClaimKeeper.ClaimCoins(suite.ctx, addr1.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins1.String(), sdk.NewCoins().String())

	// get withdrawn activities
	actions := suite.app.ClaimKeeper.GetWithdrawnActions(suite.ctx, addr1)
	suite.Require().Len(actions, 0)

	// do half of actions
	suite.app.ClaimKeeper.SetUserAction(suite.ctx, addr1, types.ActionAddLiquidity)
	suite.app.ClaimKeeper.SetUserAction(suite.ctx, addr1, types.ActionSwap)

	// get withdrawable amount after activity
	coins5, err := suite.app.ClaimKeeper.GetWithdrawableByActivity(suite.ctx, addr1.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins5.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)).String())

	// claim after activity
	coins1, err = suite.app.ClaimKeeper.ClaimCoins(suite.ctx, addr1.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)).String())

	// get withdrawn activities
	actions = suite.app.ClaimKeeper.GetWithdrawnActions(suite.ctx, addr1)
	suite.Require().Len(actions, 2)

	// do rest of actions
	suite.app.ClaimKeeper.SetUserAction(suite.ctx, addr1, types.ActionVote)
	suite.app.ClaimKeeper.SetUserAction(suite.ctx, addr1, types.ActionDelegateStake)

	// get withdrawable amount after rest actions done
	coins6, err := suite.app.ClaimKeeper.GetWithdrawableByActivity(suite.ctx, addr1.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins6.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)).String())

	// try to claim after rest actions done
	coins1, err = suite.app.ClaimKeeper.ClaimCoins(suite.ctx, addr1.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)).String())

	// get withdrawn activities
	actions = suite.app.ClaimKeeper.GetWithdrawnActions(suite.ctx, addr1)
	suite.Require().Len(actions, 4)

	// get claimable after withdrawing all
	coins1, err = suite.app.ClaimKeeper.GetClaimable(suite.ctx, addr1.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins1, balances[0].Coins)

	// get withdrawable after withdrawing all
	coins1, err = suite.app.ClaimKeeper.GetWithdrawableByActivity(suite.ctx, addr1.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins1, sdk.Coins{})

	coins3, err = suite.app.ClaimKeeper.ClaimCoins(suite.ctx, addr3.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins1, sdk.Coins{})

	err = suite.app.ClaimKeeper.FundRemainingsToCommunity(suite.ctx)
	suite.Require().NoError(err)
	moduleAccAddr := suite.app.AccountKeeper.GetModuleAddress(types.ModuleName)
	coins := suite.app.BankKeeper.GetBalance(suite.ctx, moduleAccAddr, sdk.DefaultBondDenom)
	suite.Require().Equal(coins, sdk.NewInt64Coin(sdk.DefaultBondDenom, 0))

	suite.app.ClaimKeeper.ClearClaimables(suite.ctx)
	coins2, err = suite.app.ClaimKeeper.GetClaimable(suite.ctx, addr2.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins2, sdk.Coins{})
}
