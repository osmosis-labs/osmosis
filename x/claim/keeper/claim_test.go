package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/x/claim/types"
)

func (suite *KeeperTestSuite) TestDelegationAutoWithdrawAndDelegateMore() {
	// Can you add a test to make sure that delegating can use the claimable amount as part of the delegation?
	suite.SetupTest()

	pub1 := secp256k1.GenPrivKey().PubKey()
	pub2 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())
	addr2 := sdk.AccAddress(pub2.Address())

	balances := []banktypes.Balance{
		{
			Address: addr1.String(),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
		},
		{
			Address: addr2.String(),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
		},
	}
	err := suite.app.ClaimKeeper.SetInitialClaimables(suite.ctx, balances)
	suite.Require().NoError(err)

	coins1, err := suite.app.ClaimKeeper.GetClaimable(suite.ctx, addr1.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins1, balances[0].Coins)
	coins2, err := suite.app.ClaimKeeper.GetClaimable(suite.ctx, addr2.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins2, balances[1].Coins)

	validator, err := stakingtypes.NewValidator(sdk.ValAddress(addr1), pub1, stakingtypes.Description{})
	suite.Require().NoError(err)
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())

	validator, _ = validator.AddTokensFromDel(sdk.TokensFromConsensusPower(1))
	delAmount := sdk.TokensFromConsensusPower(1)
	suite.app.BankKeeper.SetBalance(suite.ctx, addr2, sdk.NewCoin(sdk.DefaultBondDenom, delAmount))
	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, addr2, delAmount, stakingtypes.Unbonded, validator, true)
	suite.NoError(err)

	// delegation should automatically call claim and withdraw balance
	claimedCoins := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr2)
	suite.Require().Equal(claimedCoins.AmountOf(sdk.DefaultBondDenom), balances[1].Coins.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)))

	_, err = suite.app.StakingKeeper.Delegate(suite.ctx, addr2, claimedCoins.AmountOf(sdk.DefaultBondDenom), stakingtypes.Unbonded, validator, true)
	suite.NoError(err)
}

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
	err := suite.app.ClaimKeeper.SetInitialClaimables(suite.ctx, balances)
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

	// get rewards amount per action
	coins4, err := suite.app.ClaimKeeper.GetWithdrawableByActivity(suite.ctx, addr1.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins4.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 2)).String())

	// get completed activities
	actions := suite.app.ClaimKeeper.GetUserActions(suite.ctx, addr1)
	suite.Require().Len(actions, 0)

	// do half of actions
	suite.app.ClaimKeeper.AfterAddLiquidity(suite.ctx, addr1)
	suite.app.ClaimKeeper.AfterSwap(suite.ctx, addr1)

	// get balance after 2 actions done
	coins1 = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 4)).String())

	// do rest of actions
	suite.app.ClaimKeeper.AfterProposalVote(suite.ctx, 1, addr1)
	suite.app.ClaimKeeper.BeforeDelegationCreated(suite.ctx, addr1, sdk.ValAddress(addr1))

	// get balance after rest actions done
	coins1 = suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1)
	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 8)).String())

	// get completed activities
	actions = suite.app.ClaimKeeper.GetUserActions(suite.ctx, addr1)
	suite.Require().Len(actions, 4)

	// get claimable after withdrawing all
	coins1, err = suite.app.ClaimKeeper.GetClaimable(suite.ctx, addr1.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins1, balances[0].Coins)

	err = suite.app.ClaimKeeper.EndAirdrop(suite.ctx)
	suite.Require().NoError(err)

	moduleAccAddr := suite.app.AccountKeeper.GetModuleAddress(types.ModuleName)
	coins := suite.app.BankKeeper.GetBalance(suite.ctx, moduleAccAddr, sdk.DefaultBondDenom)
	suite.Require().Equal(coins, sdk.NewInt64Coin(sdk.DefaultBondDenom, 0))

	coins2, err = suite.app.ClaimKeeper.GetClaimable(suite.ctx, addr2.String())
	suite.Require().NoError(err)
	suite.Require().Equal(coins2, sdk.Coins{})
}
