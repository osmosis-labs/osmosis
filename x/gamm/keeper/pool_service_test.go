package keeper_test

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

var (
	defaultSwapFee    = sdk.MustNewDecFromStr("0.025")
	defaultExitFee    = sdk.MustNewDecFromStr("0.025")
	defaultPoolParams = balancer.PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}
	defaultFutureGovernor = ""

	// pool assets
	defaultFooAsset types.PoolAsset = types.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}
	defaultBarAsset types.PoolAsset = types.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
	}
	defaultPoolAssets []types.PoolAsset = []types.PoolAsset{defaultFooAsset, defaultBarAsset}
	defaultAcctFunds  sdk.Coins         = sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
		sdk.NewCoin("foo", sdk.NewInt(10000000)),
		sdk.NewCoin("bar", sdk.NewInt(10000000)),
		sdk.NewCoin("baz", sdk.NewInt(10000000)),
	)
)

func (suite *KeeperTestSuite) TestCreateBalancerPool() {
	params := suite.app.GAMMKeeper.GetParams(suite.ctx)

	poolCreationFeeDecCoins := sdk.DecCoins{}
	for _, coin := range params.PoolCreationFee {
		poolCreationFeeDecCoins = poolCreationFeeDecCoins.Add(sdk.NewDecCoin(coin.Denom, coin.Amount))
	}

	func() {
		keeper := suite.app.GAMMKeeper

		// Try to create pool without balances.
		_, err := keeper.CreateBalancerPool(suite.ctx, acc1, defaultPoolParams, defaultPoolAssets, defaultFutureGovernor)
		suite.Require().Error(err)
	}()

	tests := []struct {
		fn func()
	}{{
		fn: func() {
			keeper := suite.app.GAMMKeeper
			prevFeePool := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
			prevAcc1Bal := suite.app.BankKeeper.GetAllBalances(suite.ctx, acc1)
			poolId, err := keeper.CreateBalancerPool(suite.ctx, acc1, defaultPoolParams, defaultPoolAssets, defaultFutureGovernor)
			suite.Require().NoError(err)

			pool, err := keeper.GetPool(suite.ctx, poolId)
			suite.Require().NoError(err)
			suite.Require().Equal(types.InitPoolSharesSupply.String(), pool.GetTotalShares().Amount.String(),
				fmt.Sprintf("share token should be minted as %s initially", types.InitPoolSharesSupply.String()),
			)

			// check fee is correctly sent to community pool
			feePool := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
			suite.Require().Equal(feePool, prevFeePool.Add(poolCreationFeeDecCoins...))

			// check account's balance is correctly reduced
			acc1Bal := suite.app.BankKeeper.GetAllBalances(suite.ctx, acc1)
			suite.Require().Equal(acc1Bal.String(),
				prevAcc1Bal.Sub(params.PoolCreationFee).
					Sub(sdk.Coins{
						sdk.NewCoin("bar", sdk.NewInt(10000)),
						sdk.NewCoin("foo", sdk.NewInt(10000)),
					}).Add(sdk.NewCoin(types.GetPoolShareDenom(pool.GetId()), types.InitPoolSharesSupply)).String(),
			)

			liquidity := suite.app.GAMMKeeper.GetTotalLiquidity(suite.ctx)
			suite.Require().Equal("10000bar,10000foo", liquidity.String())
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			_, err := keeper.CreateBalancerPool(
				suite.ctx, acc1, balancer.PoolParams{
					SwapFee: sdk.NewDecWithPrec(-1, 2),
					ExitFee: sdk.NewDecWithPrec(1, 2),
				}, defaultPoolAssets, defaultFutureGovernor)
			suite.Require().Error(err, "can't create a pool with negative swap fee")
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			_, err := keeper.CreateBalancerPool(suite.ctx, acc1, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(-1, 2),
			}, defaultPoolAssets, defaultFutureGovernor)
			suite.Require().Error(err, "can't create a pool with negative exit fee")
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			_, err := keeper.CreateBalancerPool(suite.ctx, acc1, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []types.PoolAsset{}, defaultFutureGovernor)
			suite.Require().Error(err, "can't create the pool with empty PoolAssets")
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			_, err := keeper.CreateBalancerPool(suite.ctx, acc1, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []types.PoolAsset{{
				Weight: sdk.NewInt(0),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}}, defaultFutureGovernor)
			suite.Require().Error(err, "can't create the pool with 0 weighted PoolAsset")
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			_, err := keeper.CreateBalancerPool(suite.ctx, acc1, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []types.PoolAsset{{
				Weight: sdk.NewInt(-1),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}}, defaultFutureGovernor)
			suite.Require().Error(err, "can't create the pool with negative weighted PoolAsset")
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			_, err := keeper.CreateBalancerPool(suite.ctx, acc1, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []types.PoolAsset{{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(0)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}}, defaultFutureGovernor)
			suite.Require().Error(err, "can't create the pool with 0 balance PoolAsset")
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			_, err := keeper.CreateBalancerPool(suite.ctx, acc1, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []types.PoolAsset{{
				Weight: sdk.NewInt(100),
				Token: sdk.Coin{
					Denom:  "foo",
					Amount: sdk.NewInt(-1),
				},
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}}, defaultFutureGovernor)
			suite.Require().Error(err, "can't create the pool with negative balance PoolAsset")
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			_, err := keeper.CreateBalancerPool(suite.ctx, acc1, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []types.PoolAsset{{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}}, defaultFutureGovernor)
			suite.Require().Error(err, "can't create the pool with duplicated PoolAssets")
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			keeper.SetParams(suite.ctx, types.Params{
				PoolCreationFee: sdk.Coins{},
			})
			_, err := keeper.CreateBalancerPool(suite.ctx, acc1, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, defaultPoolAssets, defaultFutureGovernor)
			suite.Require().NoError(err)
			pools, err := keeper.GetPools(suite.ctx)
			suite.Require().Len(pools, 1)
			suite.Require().NoError(err)
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			keeper.SetParams(suite.ctx, types.Params{
				PoolCreationFee: nil,
			})
			_, err := keeper.CreateBalancerPool(suite.ctx, acc1, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, defaultPoolAssets, defaultFutureGovernor)
			suite.Require().NoError(err)
			pools, err := keeper.GetPools(suite.ctx)
			suite.Require().Len(pools, 1)
			suite.Require().NoError(err)
		},
	}}

	for _, test := range tests {
		suite.SetupTest()

		// Mint some assets to the accounts.
		for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
			err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, acc, defaultAcctFunds)
			if err != nil {
				panic(err)
			}
		}

		test.fn()
	}
}

func (suite *KeeperTestSuite) TestJoinPool() {
	tests := []struct {
		fn func(poolId uint64)
	}{
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper
				balancesBefore := suite.app.BankKeeper.GetAllBalances(suite.ctx, acc2)
				err := keeper.JoinPool(suite.ctx, acc2, poolId, types.OneShare.MulRaw(50), sdk.Coins{})
				suite.Require().NoError(err)
				suite.Require().Equal(types.OneShare.MulRaw(50).String(), suite.app.BankKeeper.GetBalance(suite.ctx, acc2, "gamm/pool/1").Amount.String())
				balancesAfter := suite.app.BankKeeper.GetAllBalances(suite.ctx, acc2)

				deltaBalances, _ := balancesBefore.SafeSub(balancesAfter)
				// The pool was created with the 10000foo, 10000bar, and the pool share was minted as 100000000gamm/pool/1.
				// Thus, to get the 50*OneShare gamm/pool/1, (10000foo, 10000bar) * (1 / 2) balances should be provided.
				suite.Require().Equal("5000", deltaBalances.AmountOf("foo").String())
				suite.Require().Equal("5000", deltaBalances.AmountOf("bar").String())

				liquidity := suite.app.GAMMKeeper.GetTotalLiquidity(suite.ctx)
				suite.Require().Equal("15000bar,15000foo", liquidity.String())
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper
				err := keeper.JoinPool(suite.ctx, acc2, poolId, sdk.NewInt(0), sdk.Coins{})
				suite.Require().Error(err, "can't join the pool with requesting 0 share amount")
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper
				err := keeper.JoinPool(suite.ctx, acc2, poolId, sdk.NewInt(-1), sdk.Coins{})
				suite.Require().Error(err, "can't join the pool with requesting negative share amount")
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper
				// Test the "tokenInMaxs"
				// In this case, to get the 50 * OneShare amount of share token, the foo, bar token are expected to be provided as 5000 amounts.
				err := keeper.JoinPool(suite.ctx, acc2, poolId, types.OneShare.MulRaw(50), sdk.Coins{
					sdk.NewCoin("foo", sdk.NewInt(4999)),
				})
				suite.Require().Error(err)
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper
				// Test the "tokenInMaxs"
				// In this case, to get the 50 * OneShare amount of share token, the foo, bar token are expected to be provided as 5000 amounts.
				err := keeper.JoinPool(suite.ctx, acc2, poolId, types.OneShare.MulRaw(50), sdk.Coins{
					sdk.NewCoin("foo", sdk.NewInt(5000)),
				})
				suite.Require().NoError(err)

				liquidity := suite.app.GAMMKeeper.GetTotalLiquidity(suite.ctx)
				suite.Require().Equal("15000bar,15000foo", liquidity.String())
			},
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		// Mint some assets to the accounts.
		for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
			err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, acc, defaultAcctFunds)
			if err != nil {
				panic(err)
			}
		}

		// Create the pool at first
		poolId, err := suite.app.GAMMKeeper.CreateBalancerPool(suite.ctx, acc1, balancer.PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}, defaultPoolAssets, defaultFutureGovernor)
		suite.Require().NoError(err)

		test.fn(poolId)
	}
}

func (suite *KeeperTestSuite) TestExitPool() {
	tests := []struct {
		fn func(poolId uint64)
	}{
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper
				// Acc2 has no share token.
				_, err := keeper.ExitPool(suite.ctx, acc2, poolId, types.OneShare.MulRaw(50), sdk.Coins{})
				suite.Require().Error(err)
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				balancesBefore := suite.app.BankKeeper.GetAllBalances(suite.ctx, acc1)
				_, err := keeper.ExitPool(suite.ctx, acc1, poolId, types.InitPoolSharesSupply.QuoRaw(2), sdk.Coins{})
				suite.Require().NoError(err)
				// (100 - 50) * OneShare should remain.
				suite.Require().Equal(types.InitPoolSharesSupply.QuoRaw(2).String(), suite.app.BankKeeper.GetBalance(suite.ctx, acc1, "gamm/pool/1").Amount.String())
				balancesAfter := suite.app.BankKeeper.GetAllBalances(suite.ctx, acc1)

				deltaBalances, _ := balancesBefore.SafeSub(balancesAfter)
				// The pool was created with the 10000foo, 10000bar, and the pool share was minted as 100*OneShare gamm/pool/1.
				// Thus, to refund the 50*OneShare gamm/pool/1, (10000foo, 10000bar) * (1 / 2) balances should be refunded.
				suite.Require().Equal("-5000", deltaBalances.AmountOf("foo").String())
				suite.Require().Equal("-5000", deltaBalances.AmountOf("bar").String())
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				_, err := keeper.ExitPool(suite.ctx, acc1, poolId, sdk.NewInt(0), sdk.Coins{})
				suite.Require().Error(err, "can't join the pool with requesting 0 share amount")
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				_, err := keeper.ExitPool(suite.ctx, acc1, poolId, sdk.NewInt(-1), sdk.Coins{})
				suite.Require().Error(err, "can't join the pool with requesting negative share amount")
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				// Test the "tokenOutMins"
				// In this case, to refund the 50000000 amount of share token, the foo, bar token are expected to be refunded as 5000 amounts.
				_, err := keeper.ExitPool(suite.ctx, acc1, poolId, types.OneShare.MulRaw(50), sdk.Coins{
					sdk.NewCoin("foo", sdk.NewInt(5001)),
				})
				suite.Require().Error(err)
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				// Test the "tokenOutMins"
				// In this case, to refund the 50000000 amount of share token, the foo, bar token are expected to be refunded as 5000 amounts.
				_, err := keeper.ExitPool(suite.ctx, acc1, poolId, types.OneShare.MulRaw(50), sdk.Coins{
					sdk.NewCoin("foo", sdk.NewInt(5000)),
				})
				suite.Require().NoError(err)
			},
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		// Mint some assets to the accounts.
		for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
			err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, acc, defaultAcctFunds)
			if err != nil {
				panic(err)
			}

			// Create the pool at first
			poolId, err := suite.app.GAMMKeeper.CreateBalancerPool(suite.ctx, acc1, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDec(0),
			}, defaultPoolAssets, defaultFutureGovernor)
			suite.Require().NoError(err)

			test.fn(poolId)
		}
	}
}

func (suite *KeeperTestSuite) TestActiveBalancerPool() {
	type testCase struct {
		blockTime  time.Time
		expectPass bool
	}

	testCases := []testCase{
		{time.Unix(1000, 0), true},
		{time.Unix(2000, 0), true},
	}

	for _, tc := range testCases {
		suite.SetupTest()

		// Mint some assets to the accounts.
		for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
			err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, acc, defaultAcctFunds)
			if err != nil {
				panic(err)
			}

			// Create the pool at first
			poolId := suite.prepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: sdk.NewDec(0),
				ExitFee: sdk.NewDec(0),
			})
			suite.ctx = suite.ctx.WithBlockTime(tc.blockTime)

			// uneffected by start time
			err = suite.app.GAMMKeeper.JoinPool(suite.ctx, acc1, poolId, types.OneShare.MulRaw(50), sdk.Coins{})
			suite.Require().NoError(err)
			_, err = suite.app.GAMMKeeper.ExitPool(suite.ctx, acc1, poolId, types.InitPoolSharesSupply.QuoRaw(2), sdk.Coins{})
			suite.Require().NoError(err)

			foocoin := sdk.NewCoin("foo", sdk.NewInt(10))

			if tc.expectPass {
				_, err = suite.app.GAMMKeeper.JoinSwapExternAmountIn(suite.ctx, acc1, poolId, foocoin, sdk.ZeroInt())
				suite.Require().NoError(err)
				_, err = suite.app.GAMMKeeper.JoinSwapShareAmountOut(suite.ctx, acc1, poolId, "foo", types.OneShare.MulRaw(10), sdk.NewInt(1000000000000000000))
				suite.Require().NoError(err)
				_, err = suite.app.GAMMKeeper.ExitSwapShareAmountIn(suite.ctx, acc1, poolId, "foo", types.OneShare.MulRaw(10), sdk.ZeroInt())
				suite.Require().NoError(err)
				_, err = suite.app.GAMMKeeper.ExitSwapExternAmountOut(suite.ctx, acc1, poolId, foocoin, sdk.NewInt(1000000000000000000))
				suite.Require().NoError(err)
			} else {
				_, err = suite.app.GAMMKeeper.JoinSwapExternAmountIn(suite.ctx, acc1, poolId, foocoin, sdk.ZeroInt())
				suite.Require().Error(err)
				_, err = suite.app.GAMMKeeper.JoinSwapShareAmountOut(suite.ctx, acc1, poolId, "foo", types.OneShare.MulRaw(10), sdk.NewInt(1000000000000000000))
				suite.Require().Error(err)
				_, err = suite.app.GAMMKeeper.ExitSwapShareAmountIn(suite.ctx, acc1, poolId, "foo", types.OneShare.MulRaw(10), sdk.ZeroInt())
				suite.Require().Error(err)
				_, err = suite.app.GAMMKeeper.ExitSwapExternAmountOut(suite.ctx, acc1, poolId, foocoin, sdk.NewInt(1000000000000000000))
				suite.Require().Error(err)
			}
		}
	}
}
