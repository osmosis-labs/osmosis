package keeper_test

import (
	"github.com/c-osmosis/osmosis/x/gamm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestCreatePool() {
	func() {
		keeper := suite.app.GAMMKeeper

		// Try to create pool without balances.
		_, err := keeper.CreatePool(suite.ctx, acc1, types.PoolParamsWithoutLock{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}, []types.PoolAsset{{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
		}, {
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
		}})
		suite.Require().Error(err)
	}()

	tests := []struct {
		fn func()
	}{{
		fn: func() {
			keeper := suite.app.GAMMKeeper
			poolId, err := keeper.CreatePool(suite.ctx, acc1, types.PoolParamsWithoutLock{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []types.PoolAsset{{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}})
			suite.Require().NoError(err)

			pool, err := keeper.GetPool(suite.ctx, poolId)
			suite.Require().NoError(err)
			suite.Require().Equal("100000000", pool.GetTotalShare().Amount.String(), "share token should be minted as 100*10^6 initially")
		},
	}, {
		fn: func() {
			suite.Require().Panics(func() {
				keeper := suite.app.GAMMKeeper
				keeper.CreatePool(suite.ctx, acc1, types.PoolParamsWithoutLock{
					SwapFee: sdk.NewDecWithPrec(-1, 2),
					ExitFee: sdk.NewDecWithPrec(1, 2),
				}, []types.PoolAsset{{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
				}, {
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
				}})
			}, "can't create the pool with negative swap fee")
		},
	}, {
		fn: func() {
			suite.Require().Panics(func() {
				keeper := suite.app.GAMMKeeper
				keeper.CreatePool(suite.ctx, acc1, types.PoolParamsWithoutLock{
					SwapFee: sdk.NewDecWithPrec(1, 2),
					ExitFee: sdk.NewDecWithPrec(-1, 2),
				}, []types.PoolAsset{{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
				}, {
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
				}})
			}, "can't create the pool with negative exit fee")
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			_, err := keeper.CreatePool(suite.ctx, acc1, types.PoolParamsWithoutLock{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []types.PoolAsset{})
			suite.Require().Error(err, "can't create the pool with empty PoolAssets")
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			_, err := keeper.CreatePool(suite.ctx, acc1, types.PoolParamsWithoutLock{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []types.PoolAsset{{
				Weight: sdk.NewInt(0),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}})
			suite.Require().Error(err, "can't create the pool with 0 weighted PoolAsset")
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			_, err := keeper.CreatePool(suite.ctx, acc1, types.PoolParamsWithoutLock{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []types.PoolAsset{{
				Weight: sdk.NewInt(-1),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}})
			suite.Require().Error(err, "can't create the pool with negative weighted PoolAsset")
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			_, err := keeper.CreatePool(suite.ctx, acc1, types.PoolParamsWithoutLock{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []types.PoolAsset{{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(0)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}})
			suite.Require().Error(err, "can't create the pool with 0 balance PoolAsset")
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			_, err := keeper.CreatePool(suite.ctx, acc1, types.PoolParamsWithoutLock{
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
			}})
			suite.Require().Error(err, "can't create the pool with negative balance PoolAsset")
		},
	}, {
		fn: func() {
			keeper := suite.app.GAMMKeeper
			_, err := keeper.CreatePool(suite.ctx, acc1, types.PoolParamsWithoutLock{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []types.PoolAsset{{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}})
			suite.Require().Error(err, "can't create the pool with duplicated PoolAssets")
		},
	}}

	for _, test := range tests {
		suite.SetupTest()

		// Mint some assets to the accounts.
		for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
			err := suite.app.BankKeeper.AddCoins(
				suite.ctx,
				acc,
				sdk.NewCoins(
					sdk.NewCoin("foo", sdk.NewInt(10000000)),
					sdk.NewCoin("bar", sdk.NewInt(10000000)),
					sdk.NewCoin("baz", sdk.NewInt(10000000)),
				),
			)
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
				err := keeper.JoinPool(suite.ctx, acc2, poolId, sdk.NewInt(50000000), sdk.Coins{})
				suite.Require().NoError(err)
				suite.Require().Equal("50000000", suite.app.BankKeeper.GetBalance(suite.ctx, acc2, "gamm/pool/1").Amount.String())
				balancesAfter := suite.app.BankKeeper.GetAllBalances(suite.ctx, acc2)

				deltaBalances, _ := balancesBefore.SafeSub(balancesAfter)
				// The pool was created with the 10000foo, 10000bar, and the pool share was minted as 100000000gamm/pool/1.
				// Thus, to get the 50000000gamm/pool/1, (10000foo, 10000bar) * (1 / 2) balances should be provided.
				suite.Require().Equal("5000", deltaBalances.AmountOf("foo").String())
				suite.Require().Equal("5000", deltaBalances.AmountOf("bar").String())
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
				// In this case, to get the 50000000 amount of share token, the foo, bar token are expected to be provided as 5000 amounts.
				err := keeper.JoinPool(suite.ctx, acc2, poolId, sdk.NewInt(50000000), sdk.Coins{
					sdk.NewCoin("foo", sdk.NewInt(4999)),
				})
				suite.Require().Error(err)
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper
				// Test the "tokenInMaxs"
				// In this case, to get the 50000000 amount of share token, the foo, bar token are expected to be provided as 5000 amounts.
				err := keeper.JoinPool(suite.ctx, acc2, poolId, sdk.NewInt(50000000), sdk.Coins{
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
			err := suite.app.BankKeeper.AddCoins(
				suite.ctx,
				acc,
				sdk.NewCoins(
					sdk.NewCoin("foo", sdk.NewInt(10000000)),
					sdk.NewCoin("bar", sdk.NewInt(10000000)),
					sdk.NewCoin("baz", sdk.NewInt(10000000)),
				),
			)
			if err != nil {
				panic(err)
			}
		}

		// Create the pool at first
		poolId, err := suite.app.GAMMKeeper.CreatePool(suite.ctx, acc1, types.PoolParamsWithoutLock{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}, []types.PoolAsset{{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
		}, {
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
		}})
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
				err := keeper.ExitPool(suite.ctx, acc2, poolId, sdk.NewInt(50000000), sdk.Coins{})
				suite.Require().Error(err)
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				balancesBefore := suite.app.BankKeeper.GetAllBalances(suite.ctx, acc1)
				err := keeper.ExitPool(suite.ctx, acc1, poolId, sdk.NewInt(50000000), sdk.Coins{})
				suite.Require().NoError(err)
				// (100 - 50) * 10^6 should be remain.
				suite.Require().Equal("50000000", suite.app.BankKeeper.GetBalance(suite.ctx, acc1, "gamm/pool/1").Amount.String())
				balancesAfter := suite.app.BankKeeper.GetAllBalances(suite.ctx, acc1)

				deltaBalances, _ := balancesBefore.SafeSub(balancesAfter)
				// The pool was created with the 10000foo, 10000bar, and the pool share was minted as 100000000gamm/pool/1.
				// Thus, to refund the 50000000gamm/pool/1, (10000foo, 10000bar) * (1 / 2) balances should be refunded.
				suite.Require().Equal("-5000", deltaBalances.AmountOf("foo").String())
				suite.Require().Equal("-5000", deltaBalances.AmountOf("bar").String())
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				err := keeper.ExitPool(suite.ctx, acc1, poolId, sdk.NewInt(0), sdk.Coins{})
				suite.Require().Error(err, "can't join the pool with requesting 0 share amount")
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				err := keeper.ExitPool(suite.ctx, acc1, poolId, sdk.NewInt(-1), sdk.Coins{})
				suite.Require().Error(err, "can't join the pool with requesting negative share amount")
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				// Test the "tokenOutMins"
				// In this case, to refund the 50000000 amount of share token, the foo, bar token are expected to be refunded as 5000 amounts.
				err := keeper.ExitPool(suite.ctx, acc1, poolId, sdk.NewInt(50000000), sdk.Coins{
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
				err := keeper.ExitPool(suite.ctx, acc1, poolId, sdk.NewInt(50000000), sdk.Coins{
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
			err := suite.app.BankKeeper.AddCoins(
				suite.ctx,
				acc,
				sdk.NewCoins(
					sdk.NewCoin("foo", sdk.NewInt(10000000)),
					sdk.NewCoin("bar", sdk.NewInt(10000000)),
					sdk.NewCoin("baz", sdk.NewInt(10000000)),
				),
			)
			if err != nil {
				panic(err)
			}

			// Create the pool at first
			poolId, err := suite.app.GAMMKeeper.CreatePool(suite.ctx, acc1, types.PoolParamsWithoutLock{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDec(0),
			}, []types.PoolAsset{{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}})
			suite.Require().NoError(err)

			test.fn(poolId)
		}
	}
}
