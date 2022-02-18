package keeper_test

import (
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func (suite *KeeperTestSuite) TestCleanupPool() {
	// Mint some assets to the accounts.
	for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
		err := simapp.FundAccount(
			suite.app.BankKeeper,
			suite.ctx,
			acc,
			sdk.NewCoins(
				sdk.NewCoin("uosmo", sdk.NewInt(1000000000)),
				sdk.NewCoin("foo", sdk.NewInt(1000)),
				sdk.NewCoin("bar", sdk.NewInt(1000)),
				sdk.NewCoin("baz", sdk.NewInt(1000)),
			),
		)
		if err != nil {
			panic(err)
		}
	}

	poolId, err := suite.app.GAMMKeeper.CreateBalancerPool(suite.ctx, acc1, defaultPoolParams, []types.PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("foo", sdk.NewInt(1000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("bar", sdk.NewInt(1000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("baz", sdk.NewInt(1000)),
		},
	}, "")
	suite.NoError(err)

	for _, acc := range []sdk.AccAddress{acc2, acc3} {
		err = suite.app.GAMMKeeper.JoinPool(suite.ctx, acc, poolId, types.OneShare.MulRaw(100), sdk.NewCoins(
			sdk.NewCoin("foo", sdk.NewInt(1000)),
			sdk.NewCoin("bar", sdk.NewInt(1000)),
			sdk.NewCoin("baz", sdk.NewInt(1000)),
		))
		suite.NoError(err)
	}

	pool, err := suite.app.GAMMKeeper.GetPool(suite.ctx, poolId)
	suite.NoError(err)
	denom := pool.GetTotalShares().Denom
	totalAmount := sdk.ZeroInt()
	for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
		coin := suite.app.BankKeeper.GetBalance(suite.ctx, acc, denom)
		suite.True(coin.Amount.Equal(types.OneShare.MulRaw(100)))
		totalAmount = totalAmount.Add(coin.Amount)
	}
	suite.True(totalAmount.Equal(types.OneShare.MulRaw(300)))

	err = suite.app.GAMMKeeper.CleanupBalancerPool(suite.ctx, []uint64{poolId}, []string{})
	suite.NoError(err)
	for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
		for _, denom := range []string{"foo", "bar", "baz"} {
			amt := suite.app.BankKeeper.GetBalance(suite.ctx, acc, denom)
			suite.True(amt.Amount.Equal(sdk.NewInt(1000)),
				"Expected equal %s: %d, %d", amt.Denom, amt.Amount.Int64(), 1000)
		}
	}
}

func (suite *KeeperTestSuite) TestCleanupPoolRandomized() {
	// address => deposited coins
	coinOf := make(map[string]sdk.Coins)
	denoms := []string{"foo", "bar", "baz"}

	// Mint some assets to the accounts.
	for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
		coins := make(sdk.Coins, 3)
		for i := range coins {
			amount := sdk.NewInt(rand.Int63n(1000))
			// give large amount of coins to the pool creator
			if i == 0 {
				amount = amount.MulRaw(10000)
			}
			coins[i] = sdk.Coin{denoms[i], amount}
		}
		coinOf[acc.String()] = coins
		coins = append(coins, sdk.NewCoin("uosmo", sdk.NewInt(1000000000)))

		err := simapp.FundAccount(
			suite.app.BankKeeper,
			suite.ctx,
			acc,
			coins.Sort(),
		)
		if err != nil {
			panic(err)
		}
	}

	initialAssets := []types.PoolAsset{}
	for _, coin := range coinOf[acc1.String()] {
		initialAssets = append(initialAssets, types.PoolAsset{Weight: types.OneShare.MulRaw(100), Token: coin})
	}
	poolId, err := suite.app.GAMMKeeper.CreateBalancerPool(suite.ctx, acc1, defaultPoolParams, initialAssets, "")
	suite.NoError(err)

	for _, acc := range []sdk.AccAddress{acc2, acc3} {
		err = suite.app.GAMMKeeper.JoinPool(suite.ctx, acc, poolId, types.OneShare, coinOf[acc.String()])
		suite.NoError(err)
	}

	err = suite.app.GAMMKeeper.CleanupBalancerPool(suite.ctx, []uint64{poolId}, []string{})
	suite.NoError(err)
	for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
		for _, coin := range coinOf[acc.String()] {
			amt := suite.app.BankKeeper.GetBalance(suite.ctx, acc, coin.Denom)
			// the refund could have rounding error
			suite.True(amt.Amount.Sub(coin.Amount).Abs().LTE(sdk.NewInt(2)),
				"Expected equal %s: %d, %d", amt.Denom, amt.Amount.Int64(), coin.Amount.Int64())
		}
	}
}

func (suite *KeeperTestSuite) TestCleanupPoolErrorOnSwap() {
	suite.ctx = suite.ctx.WithBlockTime(time.Unix(1000, 1000))
	err := simapp.FundAccount(
		suite.app.BankKeeper,
		suite.ctx,
		acc1,
		sdk.NewCoins(
			sdk.NewCoin("uosmo", sdk.NewInt(1000000000)),
			sdk.NewCoin("foo", sdk.NewInt(1000)),
			sdk.NewCoin("bar", sdk.NewInt(1000)),
			sdk.NewCoin("baz", sdk.NewInt(1000)),
		),
	)
	if err != nil {
		panic(err)
	}

	poolId, err := suite.app.GAMMKeeper.CreateBalancerPool(suite.ctx, acc1, defaultPoolParams, []types.PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("foo", sdk.NewInt(1000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("bar", sdk.NewInt(1000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("baz", sdk.NewInt(1000)),
		},
	}, "")
	suite.NoError(err)

	err = suite.app.GAMMKeeper.CleanupBalancerPool(suite.ctx, []uint64{poolId}, []string{})
	suite.NoError(err)

	_, _, err = suite.app.GAMMKeeper.SwapExactAmountIn(suite.ctx, acc1, poolId, sdk.NewCoin("foo", sdk.NewInt(1)), "bar", sdk.NewInt(1))
	suite.Error(err)
}

func (suite *KeeperTestSuite) TestCleanupPoolWithLockup() {
	suite.ctx = suite.ctx.WithBlockTime(time.Unix(1000, 1000))
	err := simapp.FundAccount(
		suite.app.BankKeeper,
		suite.ctx,
		acc1,
		sdk.NewCoins(
			sdk.NewCoin("uosmo", sdk.NewInt(1000000000)),
			sdk.NewCoin("foo", sdk.NewInt(1000)),
			sdk.NewCoin("bar", sdk.NewInt(1000)),
			sdk.NewCoin("baz", sdk.NewInt(1000)),
		),
	)
	if err != nil {
		panic(err)
	}

	poolId, err := suite.app.GAMMKeeper.CreateBalancerPool(suite.ctx, acc1, defaultPoolParams, []types.PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("foo", sdk.NewInt(1000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("bar", sdk.NewInt(1000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("baz", sdk.NewInt(1000)),
		},
	}, "")
	suite.NoError(err)

	_, err = suite.app.LockupKeeper.LockTokens(suite.ctx, acc1, sdk.Coins{sdk.NewCoin(types.GetPoolShareDenom(poolId), types.InitPoolSharesSupply)}, time.Hour)
	suite.NoError(err)

	for _, lock := range suite.app.LockupKeeper.GetLocksDenom(suite.ctx, types.GetPoolShareDenom(poolId)) {
		err = suite.app.LockupKeeper.ForceUnlock(suite.ctx, lock)
		suite.NoError(err)
	}

	err = suite.app.GAMMKeeper.CleanupBalancerPool(suite.ctx, []uint64{poolId}, []string{})
	suite.NoError(err)
	for _, coin := range []string{"foo", "bar", "baz"} {
		amt := suite.app.BankKeeper.GetBalance(suite.ctx, acc1, coin)
		// the refund could have rounding error
		suite.True(amt.Amount.Equal(sdk.NewInt(1000)) || amt.Amount.Equal(sdk.NewInt(1000).SubRaw(1)),
			"Expected equal %s: %d, %d", amt.Denom, amt.Amount.Int64(), sdk.NewInt(1000).Int64())
	}
}
