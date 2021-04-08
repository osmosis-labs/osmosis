package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestAllocateAssetsFromAccountToFarm() {
	suite.prepareAccounts()

	suite.Run("allocate to the non-existing farm", func() {
		// Can't allocate the assets to the non-existing farm.
		err := suite.app.FarmKeeper.AllocateAssetsFromAccountToFarm(suite.ctx, 1, allocatorAcc, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000))))
		suite.Error(err)
	})

	suite.Run("allocate the negative asset", func() {
		farm, err := suite.app.FarmKeeper.NewFarm(suite.ctx)
		suite.NoError(err)
		// Can't allocate the invalid assets.
		err = suite.app.FarmKeeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.Coins{
			sdk.Coin{
				Denom:  "foo",
				Amount: sdk.NewInt(-1000),
			},
		})
		suite.Error(err)
	})

	suite.Run("allocate the 0 balance asset", func() {
		farm, err := suite.app.FarmKeeper.NewFarm(suite.ctx)
		suite.NoError(err)
		// Can't allocate the invalid assets.
		err = suite.app.FarmKeeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(0))))
		suite.Error(err)
	})

	suite.Run("allocate the empty assets", func() {
		farm, err := suite.app.FarmKeeper.NewFarm(suite.ctx)
		suite.NoError(err)
		// Can't allocate the invalid assets.
		err = suite.app.FarmKeeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.NewCoins())
		suite.Error(err)
	})

	suite.Run("allocate the duplicated assets", func() {
		farm, err := suite.app.FarmKeeper.NewFarm(suite.ctx)
		suite.NoError(err)
		// Can't allocate the invalid assets.
		err = suite.app.FarmKeeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.Coins{
			sdk.Coin{
				Denom:  "foo",
				Amount: sdk.NewInt(-1000),
			},
			sdk.Coin{
				Denom:  "foo",
				Amount: sdk.NewInt(-1000),
			},
		})
		suite.Error(err)
	})

	suite.Run("allocate the assets with insufficient balances", func() {
		farm, err := suite.app.FarmKeeper.NewFarm(suite.ctx)
		suite.NoError(err)
		// Can't allocate the invalid assets.
		err = suite.app.FarmKeeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, acc1, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000))))
		suite.Error(err)
	})
}

func (suite *KeeperTestSuite) TestSimpleReward() {
	suite.prepareAccounts()

	keeper := suite.app.FarmKeeper

	farm, err := keeper.NewFarm(suite.ctx)
	suite.NoError(err)

	rewards, err := keeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc1, sdk.NewInt(1))
	suite.NoError(err)
	suite.Equal(0, len(rewards))
	suite.Equal("0", suite.app.BankKeeper.GetBalance(suite.ctx, acc1, "foo").Amount.String())

	err = keeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000))))
	suite.NoError(err)

	rewards, err = keeper.WithdrawRewardsFromFarm(suite.ctx, farm.FarmId, acc1)
	suite.NoError(err)
	suite.Equal("1000foo", rewards.String())
	suite.Equal("1000", suite.app.BankKeeper.GetBalance(suite.ctx, acc1, "foo").Amount.String())

	err = keeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000))))
	suite.NoError(err)
	err = keeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000))))
	suite.NoError(err)

	rewards, err = keeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc2, sdk.NewInt(1))
	suite.NoError(err)
	suite.Equal(0, len(rewards))

	rewards, err = keeper.WithdrawRewardsFromFarm(suite.ctx, farm.FarmId, acc2)
	suite.NoError(err)
	suite.Equal(0, len(rewards))

	rewards, err = keeper.WithdrawRewardsFromFarm(suite.ctx, farm.FarmId, acc1)
	suite.NoError(err)
	suite.Equal("2000foo", rewards.String())
	suite.Equal("3000", suite.app.BankKeeper.GetBalance(suite.ctx, acc1, "foo").Amount.String())
}

func (suite *KeeperTestSuite) TestSimpleReward2() {
	suite.prepareAccounts()

	keeper := suite.app.FarmKeeper

	farm, err := keeper.NewFarm(suite.ctx)
	suite.NoError(err)

	rewards, err := keeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc1, sdk.NewInt(1))
	suite.NoError(err)
	suite.Equal(0, len(rewards))
	suite.Equal("0", suite.app.BankKeeper.GetBalance(suite.ctx, acc1, "foo").Amount.String())

	err = keeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000))))
	suite.NoError(err)

	// Until this, acc1 has the 1000foo as rewards.

	rewards, err = keeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc2, sdk.NewInt(2))
	suite.NoError(err)
	suite.Equal(0, len(rewards))
	suite.Equal("0", suite.app.BankKeeper.GetBalance(suite.ctx, acc2, "foo").Amount.String())

	err = keeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000))))
	suite.NoError(err)
	err = keeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(2000))))
	suite.NoError(err)

	// Until this, acc1 has the 2000foo as rewards. And, acc2 has the 2000foo as rewards.
	rewards, err = keeper.WithdrawRewardsFromFarm(suite.ctx, farm.FarmId, acc1)
	suite.NoError(err)
	// But has small difference...
	suite.Equal("1999foo", rewards.String())
	suite.Equal("1999", suite.app.BankKeeper.GetBalance(suite.ctx, acc1, "foo").Amount.String())
	rewards, err = keeper.WithdrawRewardsFromFarm(suite.ctx, farm.FarmId, acc2)
	suite.NoError(err)
	// But has small difference...
	suite.Equal("1999foo", rewards.String())
	suite.Equal("1999", suite.app.BankKeeper.GetBalance(suite.ctx, acc2, "foo").Amount.String())
}

type step struct {
	Allocation    sdk.Coins
	Acc1ShareDiff sdk.Int
	Acc2ShareDiff sdk.Int
	Acc3ShareDiff sdk.Int
}

func calculateSteps(steps []step) (acc1Rewards sdk.Coins, acc2Rewards sdk.Coins, acc3Rewards sdk.Coins) {
	acc1share := sdk.NewInt(0)
	acc2share := sdk.NewInt(0)
	acc3share := sdk.NewInt(0)

	acc1Rewards = sdk.Coins{}
	acc2Rewards = sdk.Coins{}
	acc3Rewards = sdk.Coins{}

	for _, step := range steps {
		if !step.Allocation.Empty() {
			totalShare := acc1share.Add(acc2share).Add(acc3share)

			if totalShare.IsPositive() {
				if !acc1share.IsZero() {
					rewards, _ := sdk.NewDecCoinsFromCoins(step.Allocation...).MulDec(acc1share.ToDec().Quo(totalShare.ToDec())).TruncateDecimal()
					acc1Rewards = acc1Rewards.Add(rewards...)
				}

				if !acc2share.IsZero() {
					rewards, _ := sdk.NewDecCoinsFromCoins(step.Allocation...).MulDec(acc2share.ToDec().Quo(totalShare.ToDec())).TruncateDecimal()
					acc2Rewards = acc2Rewards.Add(rewards...)
				}

				if !acc3share.IsZero() {
					rewards, _ := sdk.NewDecCoinsFromCoins(step.Allocation...).MulDec(acc3share.ToDec().Quo(totalShare.ToDec())).TruncateDecimal()
					acc3Rewards = acc3Rewards.Add(rewards...)
				}
			}
		}

		if !step.Acc1ShareDiff.IsNil() && !step.Acc1ShareDiff.IsZero() {
			acc1share = acc1share.Add(step.Acc1ShareDiff)
		}

		if !step.Acc2ShareDiff.IsNil() && !step.Acc2ShareDiff.IsZero() {
			acc2share = acc2share.Add(step.Acc2ShareDiff)
		}

		if !step.Acc3ShareDiff.IsNil() && !step.Acc3ShareDiff.IsZero() {
			acc3share = acc3share.Add(step.Acc3ShareDiff)
		}
	}

	return
}

func (suite *KeeperTestSuite) TestCompareDumbCalculation() {
	steps := []step{
		{
			Acc1ShareDiff: sdk.NewInt(100),
		},
		{
			Allocation: sdk.NewCoins(
				sdk.NewCoin("foo", sdk.NewInt(10000000)),
				sdk.NewCoin("bar", sdk.NewInt(10000000)),
				sdk.NewCoin("baz", sdk.NewInt(10000000)),
			),
		},
		{
			Acc2ShareDiff: sdk.NewInt(100),
		},
		{
			Allocation: sdk.NewCoins(
				sdk.NewCoin("foo", sdk.NewInt(10000000)),
				sdk.NewCoin("bar", sdk.NewInt(10000000)),
				sdk.NewCoin("baz", sdk.NewInt(10000000)),
			),
		},
		{
			Acc3ShareDiff: sdk.NewInt(100),
		},
		{
			Allocation: sdk.NewCoins(
				sdk.NewCoin("foo", sdk.NewInt(20000000)),
				sdk.NewCoin("bar", sdk.NewInt(20000000)),
				sdk.NewCoin("baz", sdk.NewInt(20000000)),
			),
		},
		{
			Acc1ShareDiff: sdk.NewInt(-10),
		},
		{
			Allocation: sdk.NewCoins(
				sdk.NewCoin("baz", sdk.NewInt(25000000)),
			),
		},
		{
			Acc1ShareDiff: sdk.NewInt(-50),
			Acc3ShareDiff: sdk.NewInt(-90),
		},
		{
			Allocation: sdk.NewCoins(
				sdk.NewCoin("foo", sdk.NewInt(25000000)),
			),
		},
		{
			Acc1ShareDiff: sdk.NewInt(-10),
		},
		{
			Allocation: sdk.NewCoins(
				sdk.NewCoin("bar", sdk.NewInt(15000000)),
			),
		},
	}

	// Use dumb calculations to find results of each step to compare with F1 distribution
	expectedAcc1Rewards, expectedAcc2Rewards, expectedAcc3Rewards := calculateSteps(steps)

	suite.prepareAccounts()

	keeper := suite.app.FarmKeeper
	farm, err := keeper.NewFarm(suite.ctx)
	suite.NoError(err)

	acc1TotalRewards := sdk.Coins{}
	acc2TotalRewards := sdk.Coins{}
	acc3TotalRewards := sdk.Coins{}

	for _, step := range steps {
		if !step.Allocation.Empty() {
			err := keeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, step.Allocation)
			suite.NoError(err)
		}

		if !step.Acc1ShareDiff.IsNil() && !step.Acc1ShareDiff.IsZero() {
			if step.Acc1ShareDiff.IsPositive() {
				rewards, err := keeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc1, step.Acc1ShareDiff)
				suite.NoError(err)
				acc1TotalRewards = acc1TotalRewards.Add(rewards...)
			} else {
				rewards, err := keeper.WithdrawShareFromFarm(suite.ctx, farm.FarmId, acc1, step.Acc1ShareDiff.Neg())
				suite.NoError(err)
				acc1TotalRewards = acc1TotalRewards.Add(rewards...)
			}
		}

		if !step.Acc2ShareDiff.IsNil() && !step.Acc2ShareDiff.IsZero() {
			if step.Acc2ShareDiff.IsPositive() {
				rewards, err := keeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc2, step.Acc2ShareDiff)
				suite.NoError(err)
				acc2TotalRewards = acc2TotalRewards.Add(rewards...)
			} else {
				rewards, err := keeper.WithdrawShareFromFarm(suite.ctx, farm.FarmId, acc2, step.Acc2ShareDiff.Neg())
				suite.NoError(err)
				acc2TotalRewards = acc2TotalRewards.Add(rewards...)
			}
		}

		if !step.Acc3ShareDiff.IsNil() && !step.Acc3ShareDiff.IsZero() {
			if step.Acc3ShareDiff.IsPositive() {
				rewards, err := keeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc3, step.Acc3ShareDiff)
				suite.NoError(err)
				acc3TotalRewards = acc3TotalRewards.Add(rewards...)
			} else {
				rewards, err := keeper.WithdrawShareFromFarm(suite.ctx, farm.FarmId, acc3, step.Acc3ShareDiff.Neg())
				suite.NoError(err)
				acc3TotalRewards = acc3TotalRewards.Add(rewards...)
			}
		}
	}

	rewards, err := keeper.WithdrawRewardsFromFarm(suite.ctx, farm.FarmId, acc1)
	suite.NoError(err)
	acc1TotalRewards = acc1TotalRewards.Add(rewards...)
	rewards, err = keeper.WithdrawRewardsFromFarm(suite.ctx, farm.FarmId, acc2)
	suite.NoError(err)
	acc2TotalRewards = acc2TotalRewards.Add(rewards...)
	rewards, err = keeper.WithdrawRewardsFromFarm(suite.ctx, farm.FarmId, acc3)
	suite.NoError(err)
	acc3TotalRewards = acc3TotalRewards.Add(rewards...)

	suite.Equal(acc1TotalRewards.String(), suite.app.BankKeeper.GetAllBalances(suite.ctx, acc1).String())
	suite.Equal(acc2TotalRewards.String(), suite.app.BankKeeper.GetAllBalances(suite.ctx, acc2).String())
	suite.Equal(acc3TotalRewards.String(), suite.app.BankKeeper.GetAllBalances(suite.ctx, acc3).String())

	// The diffence between the result of dumb calculation and farm's distribution should be very small
	deltaAcc1Rewards, _ := sdk.NewDecCoinsFromCoins(expectedAcc1Rewards...).SafeSub(sdk.NewDecCoinsFromCoins(acc1TotalRewards...))
	for _, delta := range deltaAcc1Rewards {
		suite.True(delta.Amount.Abs().LTE(sdk.OneDec()))
	}
	deltaAcc2Rewards, _ := sdk.NewDecCoinsFromCoins(expectedAcc2Rewards...).SafeSub(sdk.NewDecCoinsFromCoins(acc2TotalRewards...))
	for _, delta := range deltaAcc2Rewards {
		suite.True(delta.Amount.Abs().LTE(sdk.OneDec()))
	}
	deltaAcc3Rewards, _ := sdk.NewDecCoinsFromCoins(expectedAcc3Rewards...).SafeSub(sdk.NewDecCoinsFromCoins(acc3TotalRewards...))
	for _, delta := range deltaAcc3Rewards {
		suite.True(delta.Amount.Abs().LTE(sdk.OneDec()))
	}
}
