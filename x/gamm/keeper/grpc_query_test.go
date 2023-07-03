package keeper_test

import (
	gocontext "context"
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/stableswap"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/v2types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

func (s *KeeperTestSuite) TestCalcExitPoolCoinsFromShares() {
	queryClient := s.queryClient
	ctx := s.Ctx
	poolId := s.PrepareBalancerPool()
	exitFee := sdk.ZeroDec()

	testCases := []struct {
		name          string
		poolId        uint64
		shareInAmount sdk.Int
		expectedErr   error
	}{
		{
			"valid test case",
			poolId,
			sdk.NewInt(1000000000000000000),
			nil,
		},
		{
			"pool id does not exist",
			poolId + 1,
			sdk.NewInt(1000000000000000000),
			types.ErrPoolNotFound,
		},
		{
			"zero share in amount",
			poolId,
			sdk.ZeroInt(),
			errorsmod.Wrapf(types.ErrInvalidMathApprox, "share ratio is zero or negative"),
		},
		{
			"negative share in amount",
			poolId,
			sdk.NewInt(-10000),
			errorsmod.Wrapf(types.ErrInvalidMathApprox, "share ratio is zero or negative"),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := queryClient.CalcExitPoolCoinsFromShares(gocontext.Background(), &types.QueryCalcExitPoolCoinsFromSharesRequest{
				PoolId:        tc.poolId,
				ShareInAmount: tc.shareInAmount,
			})
			if tc.expectedErr == nil {
				poolRes, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
					PoolId: tc.poolId,
				})
				s.Require().NoError(err)

				var pool types.CFMMPoolI
				err = s.App.InterfaceRegistry().UnpackAny(poolRes.Pool, &pool)
				s.Require().NoError(err)

				exitCoins, err := pool.CalcExitPoolCoinsFromShares(ctx, tc.shareInAmount, exitFee)
				s.Require().NoError(err)

				// For each coin in exitCoins we are looking for a match in our response
				// We need to find exactly len(out) such matches
				coins_checked := 0
				for _, coin := range exitCoins {
					for _, actual_coin := range out.TokensOut {
						if coin.Denom == actual_coin.Denom {
							s.Require().Equal(coin.Amount, actual_coin.Amount)
							coins_checked++
						}
					}
				}
				s.Require().Equal(out.TokensOut, exitCoins)
			} else {
				s.Require().ErrorIs(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCalcJoinPoolNoSwapShares() {
	queryClient := s.queryClient
	ctx := s.Ctx
	poolId := s.PrepareBalancerPool()
	spreadFactor := sdk.ZeroDec()

	testCases := []struct {
		name        string
		poolId      uint64
		tokensIn    sdk.Coins
		expectedErr error
	}{
		{
			"valid uneven multi asset join test case",
			poolId,
			sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(5000000)), sdk.NewCoin("bar", sdk.NewInt(5000000)), sdk.NewCoin("baz", sdk.NewInt(5000000)), sdk.NewCoin("uosmo", sdk.NewInt(5000000))),
			nil,
		},
		{
			"valid even multi asset join test case",
			poolId,
			sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(500000)), sdk.NewCoin("bar", sdk.NewInt(1000000)), sdk.NewCoin("baz", sdk.NewInt(1500000)), sdk.NewCoin("uosmo", sdk.NewInt(2000000))),
			nil,
		},
		{
			"invalid single asset join test case",
			poolId,
			sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000000))),
			errors.New("no-swap joins require LP'ing with all assets in pool"),
		},
		{
			"pool id does not exist",
			poolId + 1,
			sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000000))),
			types.PoolDoesNotExistError{PoolId: poolId + 1},
		},
		{
			"token in denom does not exist",
			poolId,
			sdk.NewCoins(sdk.NewCoin("random", sdk.NewInt(10000))),
			errorsmod.Wrapf(types.ErrDenomNotFoundInPool, "input denoms must already exist in the pool (%s)", "random"),
		},
		{
			"join pool with incorrect amount of assets",
			poolId,
			sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000)), sdk.NewCoin("bar", sdk.NewInt(10000))),
			errors.New("no-swap joins require LP'ing with all assets in pool"),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := queryClient.CalcJoinPoolNoSwapShares(gocontext.Background(), &types.QueryCalcJoinPoolNoSwapSharesRequest{
				PoolId:   tc.poolId,
				TokensIn: tc.tokensIn,
			})
			if tc.expectedErr == nil {
				poolRes, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
					PoolId: tc.poolId,
				})
				s.Require().NoError(err)

				var pool types.CFMMPoolI
				err = s.App.InterfaceRegistry().UnpackAny(poolRes.Pool, &pool)
				s.Require().NoError(err)

				numShares, numLiquidity, err := pool.CalcJoinPoolNoSwapShares(ctx, tc.tokensIn, spreadFactor)
				s.Require().NoError(err)
				s.Require().Equal(numShares, out.SharesOut)
				s.Require().Equal(numLiquidity, out.TokensOut)
			} else {
				s.Require().EqualError(err, tc.expectedErr.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestPoolsWithFilter() {
	var (
		defaultAcctFunds sdk.Coins = sdk.NewCoins(
			sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
			sdk.NewCoin("foo", sdk.NewInt(10000000)),
			sdk.NewCoin("bar", sdk.NewInt(10000000)),
			sdk.NewCoin("baz", sdk.NewInt(10000000)),
		)
		defaultPoolParams = balancer.PoolParams{
			SwapFee: sdk.ZeroDec(),
			ExitFee: sdk.ZeroDec(),
		}
	)

	testCases := []struct {
		name                        string
		num_pools                   int
		expected_num_pools_response int
		min_liquidity               string
		pool_type                   string
		poolAssets                  []balancer.PoolAsset
		expectedErr                 bool
	}{
		{
			name:                        "valid tc with both filters for min liquidity and pool type",
			num_pools:                   1,
			expected_num_pools_response: 1,
			min_liquidity:               "50000foo, 50000bar",
			pool_type:                   "Balancer",
			poolAssets: []balancer.PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
				},
				{
					Weight: sdk.NewInt(200),
					Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
				},
			},
			expectedErr: false,
		},
		{
			name:                        "only min liquidity specified (too high for pools - return 0 pools)",
			num_pools:                   1,
			expected_num_pools_response: 0,
			min_liquidity:               "500000000foo, 500000000bar",
			poolAssets: []balancer.PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
				},
				{
					Weight: sdk.NewInt(200),
					Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
				},
			},
			expectedErr: false,
		},
		{
			name:                        "wrong pool type specified",
			num_pools:                   1,
			expected_num_pools_response: 0,
			min_liquidity:               "",
			pool_type:                   "balaswap",
			poolAssets: []balancer.PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
				},
				{
					Weight: sdk.NewInt(200),
					Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
				},
			},
			expectedErr: false,
		},
		{
			name:                        "2 parameters specified + single-token min liquidity",
			num_pools:                   1,
			expected_num_pools_response: 4,
			min_liquidity:               "500foo",
			pool_type:                   "Balancer",
			poolAssets: []balancer.PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
				},
				{
					Weight: sdk.NewInt(200),
					Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
				},
			},
			expectedErr: false,
		},
		{
			name:                        "min_liquidity denom not present in pool",
			num_pools:                   1,
			expected_num_pools_response: 0,
			min_liquidity:               "500whoami",
			poolAssets: []balancer.PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
				},
				{
					Weight: sdk.NewInt(200),
					Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
				},
			},
			expectedErr: false,
		},
		{
			name:                        "only min liquidity specified - valid",
			num_pools:                   1,
			expected_num_pools_response: 6,
			min_liquidity:               "0foo,0bar",
			poolAssets: []balancer.PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
				},
				{
					Weight: sdk.NewInt(200),
					Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
				},
			},
			expectedErr: false,
		},
		{
			name:                        "only valid pool type specified",
			num_pools:                   1,
			expected_num_pools_response: 7,
			pool_type:                   "Balancer",
			poolAssets: []balancer.PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
				},
				{
					Weight: sdk.NewInt(200),
					Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
				},
			},
			expectedErr: false,
		},
		{
			name:                        "invalid min liquidity specified",
			num_pools:                   1,
			expected_num_pools_response: 1,
			min_liquidity:               "wrong300foo",
			pool_type:                   "Balancer",
			poolAssets: []balancer.PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
				},
				{
					Weight: sdk.NewInt(200),
					Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
				},
			},
			expectedErr: true,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			for i := 0; i < tc.num_pools; i++ {
				s.prepareCustomBalancerPool(
					defaultAcctFunds,
					tc.poolAssets,
					defaultPoolParams,
				)
			}
			res, err := s.queryClient.PoolsWithFilter(s.Ctx.Context(), &types.QueryPoolsWithFilterRequest{
				MinLiquidity: tc.min_liquidity,
				PoolType:     tc.pool_type,
			})
			if tc.expectedErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expected_num_pools_response, len(res.Pools))
			}
		})
	}
}

func (s *KeeperTestSuite) TestCalcJoinPoolShares() {
	queryClient := s.queryClient
	ctx := s.Ctx
	poolId := s.PrepareBalancerPool()
	spreadFactor := sdk.ZeroDec()

	testCases := []struct {
		name        string
		poolId      uint64
		tokensIn    sdk.Coins
		expectedErr error
	}{
		{
			"valid uneven multi asset join test case",
			poolId,
			sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(5000000)), sdk.NewCoin("bar", sdk.NewInt(5000000)), sdk.NewCoin("baz", sdk.NewInt(5000000)), sdk.NewCoin("uosmo", sdk.NewInt(5000000))),
			nil,
		},
		{
			"valid even multi asset join test case",
			poolId,
			sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(500000)), sdk.NewCoin("bar", sdk.NewInt(1000000)), sdk.NewCoin("baz", sdk.NewInt(1500000)), sdk.NewCoin("uosmo", sdk.NewInt(2000000))),
			nil,
		},
		{
			"valid single asset join test case",
			poolId,
			sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000000))),
			nil,
		},
		{
			"pool id does not exist",
			poolId + 1,
			sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000000))),
			types.PoolDoesNotExistError{PoolId: poolId + 1},
		},
		{
			"token in denom does not exist",
			poolId,
			sdk.NewCoins(sdk.NewCoin("random", sdk.NewInt(10000))),
			errorsmod.Wrapf(types.ErrDenomNotFoundInPool, "input denoms must already exist in the pool (%s)", "random"),
		},
		{
			"join pool with incorrect amount of assets",
			poolId,
			sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000)), sdk.NewCoin("bar", sdk.NewInt(10000))),
			errors.New("balancer pool only supports LP'ing with one asset or all assets in pool"),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := queryClient.CalcJoinPoolShares(gocontext.Background(), &types.QueryCalcJoinPoolSharesRequest{
				PoolId:   tc.poolId,
				TokensIn: tc.tokensIn,
			})
			if tc.expectedErr == nil {
				poolRes, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
					PoolId: tc.poolId,
				})
				s.Require().NoError(err)

				var pool types.CFMMPoolI
				err = s.App.InterfaceRegistry().UnpackAny(poolRes.Pool, &pool)
				s.Require().NoError(err)

				numShares, numLiquidity, err := pool.CalcJoinPoolShares(ctx, tc.tokensIn, spreadFactor)
				s.Require().NoError(err)
				s.Require().Equal(numShares, out.ShareOutAmount)
				s.Require().Equal(numLiquidity, out.TokensOut)
			} else {
				s.Require().EqualError(err, tc.expectedErr.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryPool() {
	queryClient := s.queryClient

	// Invalid param
	_, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{})
	s.Require().Error(err)

	// Pool not exist
	_, err = queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
		PoolId: 1,
	})
	s.Require().Error(err)

	for i := 0; i < 10; i++ {
		poolId := s.PrepareBalancerPool()
		poolRes, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
			PoolId: poolId,
		})
		s.Require().NoError(err)
		var pool types.CFMMPoolI
		err = s.App.InterfaceRegistry().UnpackAny(poolRes.Pool, &pool)
		s.Require().NoError(err)
		s.Require().Equal(poolId, pool.GetId())
		s.Require().Equal(poolmanagertypes.NewPoolAddress(poolId).String(), pool.GetAddress().String())
	}
}

func (s *KeeperTestSuite) TestQueryPools() {
	queryClient := s.queryClient

	for i := 0; i < 10; i++ {
		poolId := s.PrepareBalancerPool()
		poolRes, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
			PoolId: poolId,
		})
		s.Require().NoError(err)
		var pool types.CFMMPoolI
		err = s.App.InterfaceRegistry().UnpackAny(poolRes.Pool, &pool)
		s.Require().NoError(err)
		s.Require().Equal(poolId, pool.GetId())
		s.Require().Equal(poolmanagertypes.NewPoolAddress(poolId).String(), pool.GetAddress().String())
	}

	res, err := queryClient.Pools(gocontext.Background(), &types.QueryPoolsRequest{
		Pagination: &query.PageRequest{
			Key:        nil,
			Limit:      1,
			CountTotal: false,
		},
	})
	s.Require().NoError(err)
	s.Require().Equal(1, len(res.Pools))
	for _, r := range res.Pools {
		var pool types.CFMMPoolI
		err = s.App.InterfaceRegistry().UnpackAny(r, &pool)
		s.Require().NoError(err)
		s.Require().Equal(poolmanagertypes.NewPoolAddress(uint64(1)).String(), pool.GetAddress().String())
		s.Require().Equal(uint64(1), pool.GetId())
	}

	res, err = queryClient.Pools(gocontext.Background(), &types.QueryPoolsRequest{
		Pagination: &query.PageRequest{
			Key:        nil,
			Limit:      5,
			CountTotal: false,
		},
	})
	s.Require().NoError(err)
	s.Require().Equal(5, len(res.Pools))
	for i, r := range res.Pools {
		var pool types.CFMMPoolI
		err = s.App.InterfaceRegistry().UnpackAny(r, &pool)
		s.Require().NoError(err)
		s.Require().Equal(poolmanagertypes.NewPoolAddress(uint64(i+1)).String(), pool.GetAddress().String())
		s.Require().Equal(uint64(i+1), pool.GetId())
	}
}

func (s *KeeperTestSuite) TestPoolType() {
	poolIdBalancer := s.PrepareBalancerPool()
	poolIdStableswap := s.PrepareBasicStableswapPool()

	// error when querying invalid pool ID
	_, err := s.queryClient.PoolType(gocontext.Background(), &types.QueryPoolTypeRequest{PoolId: poolIdStableswap + 1})
	s.Require().Error(err)

	res, err := s.queryClient.PoolType(gocontext.Background(), &types.QueryPoolTypeRequest{PoolId: poolIdBalancer})
	s.Require().NoError(err)
	s.Require().Equal(balancer.PoolTypeName, res.PoolType)

	res, err = s.queryClient.PoolType(gocontext.Background(),
		&types.QueryPoolTypeRequest{PoolId: poolIdStableswap})
	s.Require().NoError(err)
	s.Require().Equal(stableswap.PoolTypeName, res.PoolType)
}

func (s *KeeperTestSuite) TestQueryNumPools1() {
	res, err := s.queryClient.NumPools(gocontext.Background(), &types.QueryNumPoolsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(uint64(0), res.NumPools)
}

func (s *KeeperTestSuite) TestQueryNumPools2() {
	for i := 0; i < 10; i++ {
		s.PrepareBalancerPool()
	}

	res, err := s.queryClient.NumPools(gocontext.Background(), &types.QueryNumPoolsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(uint64(10), res.NumPools)
}

func (s *KeeperTestSuite) TestQueryTotalPoolLiquidity() {
	queryClient := s.queryClient

	// Pool not exist
	_, err := queryClient.TotalPoolLiquidity(gocontext.Background(), &types.QueryTotalPoolLiquidityRequest{PoolId: 1})
	s.Require().Error(err)

	poolId := s.PrepareBalancerPool()

	res, err := queryClient.TotalPoolLiquidity(gocontext.Background(), &types.QueryTotalPoolLiquidityRequest{PoolId: poolId})
	s.Require().NoError(err)
	expectedCoins := sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(5000000)), sdk.NewCoin("bar", sdk.NewInt(5000000)), sdk.NewCoin("baz", sdk.NewInt(5000000)), sdk.NewCoin("uosmo", sdk.NewInt(5000000)))
	s.Require().Equal(res.Liquidity, expectedCoins)
}

func (s *KeeperTestSuite) TestQueryTotalShares() {
	queryClient := s.queryClient

	// Pool not exist
	_, err := queryClient.TotalShares(gocontext.Background(), &types.QueryTotalSharesRequest{PoolId: 1})
	s.Require().Error(err)

	poolId := s.PrepareBalancerPool()

	// Share Token would be minted as 100.000000000000000000 share token initially.
	res, err := queryClient.TotalShares(gocontext.Background(), &types.QueryTotalSharesRequest{PoolId: poolId})
	s.Require().NoError(err)
	s.Require().Equal(types.InitPoolSharesSupply.String(), res.TotalShares.Amount.String())

	// Mint more share token.
	// TODO: Change this test structure. perhaps JoinPoolExactShareAmountOut can be used once written
	// pool, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
	// s.Require().NoError(err)
	// err = s.App.GAMMKeeper.MintPoolShareToAccount(s.Ctx, pool, acc1, types.OneShare.MulRaw(10))
	// s.Require().NoError(err)
	// s.Require().NoError(s.App.GAMMKeeper.SetPool(s.Ctx, pool))

	// res, err = queryClient.TotalShares(gocontext.Background(), &types.QueryTotalSharesRequest{PoolId: poolId})
	// s.Require().NoError(err)
	// s.Require().Equal(types.InitPoolSharesSupply.Add(types.OneShare.MulRaw(10)).String(), res.TotalShares.Amount.String())
}

func (s *KeeperTestSuite) TestQueryBalancerPoolTotalLiquidity() {
	queryClient := s.queryClient

	// Pool not exist
	res, err := queryClient.TotalLiquidity(gocontext.Background(), &types.QueryTotalLiquidityRequest{})
	s.Require().NoError(err)
	s.Require().Equal("", sdk.Coins(res.Liquidity).String())

	_ = s.PrepareBalancerPool()

	// create pool
	res, err = queryClient.TotalLiquidity(gocontext.Background(), &types.QueryTotalLiquidityRequest{})
	s.Require().NoError(err)
	s.Require().Equal("5000000bar,5000000baz,5000000foo,5000000uosmo", sdk.Coins(res.Liquidity).String())
}

// TODO: Come fix
// func (s *KeeperTestSuite) TestQueryBalancerPoolPoolAssets() {
// 	queryClient := s.queryClient

// 	// Pool not exist
// 	_, err := queryClient.PoolAssets(gocontext.Background(), &types.QueryPoolAssetsRequest{PoolId: 1})
// 	s.Require().Error(err)

// 	poolId := s.PrepareBalancerPool()

// 	res, err := queryClient.PoolAssets(gocontext.Background(), &types.QueryPoolAssetsRequest{PoolId: poolId})
// 	s.Require().NoError(err)

// 	/*
// 		{
// 			Weight: sdk.NewInt(200 * GuaranteedWeightPrecision),
// 			Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
// 		},
// 		{
// 			Weight: sdk.NewInt(300 * GuaranteedWeightPrecision),
// 			Token:  sdk.NewCoin("baz", sdk.NewInt(5000000)),
// 		},
// 		{
// 			Weight: sdk.NewInt(100 * GuaranteedWeightPrecision),
// 			Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
// 		},
// 	*/
// 	PoolAssets := res.PoolAssets
// 	s.Require().Equal(3, len(PoolAssets))

// 	s.Require().Equal(sdk.NewInt(200*types.GuaranteedWeightPrecision), PoolAssets[0].Weight)
// 	s.Require().Equal(sdk.NewInt(300*types.GuaranteedWeightPrecision), PoolAssets[1].Weight)
// 	s.Require().Equal(sdk.NewInt(100*types.GuaranteedWeightPrecision), PoolAssets[2].Weight)

// 	s.Require().Equal("5000000bar", PoolAssets[0].Token.String())
// 	s.Require().Equal("5000000baz", PoolAssets[1].Token.String())
// 	s.Require().Equal("5000000foo", PoolAssets[2].Token.String())
// }

func (s *KeeperTestSuite) TestQueryBalancerPoolSpotPrice() {
	queryClient := s.queryClient
	poolID := s.PrepareBalancerPool()

	testCases := []struct {
		name      string
		req       *types.QuerySpotPriceRequest
		expectErr bool
		result    string
	}{
		{
			name: "non-existent pool",
			req: &types.QuerySpotPriceRequest{
				PoolId:          0,
				BaseAssetDenom:  "foo",
				QuoteAssetDenom: "bar",
			},
			expectErr: true,
		},
		{
			name: "missing asset denoms",
			req: &types.QuerySpotPriceRequest{
				PoolId: poolID,
			},
			expectErr: true,
		},
		{
			name: "missing pool ID and quote denom",
			req: &types.QuerySpotPriceRequest{
				BaseAssetDenom: "foo",
			},
			expectErr: true,
		},
		{
			name: "missing pool ID and base denom",
			req: &types.QuerySpotPriceRequest{
				QuoteAssetDenom: "bar",
			},
			expectErr: true,
		},
		{
			name: "valid request for foo/bar",
			req: &types.QuerySpotPriceRequest{
				PoolId:          poolID,
				BaseAssetDenom:  "foo",
				QuoteAssetDenom: "bar",
			},
			result: sdk.NewDec(2).String(),
		},
		{
			name: "valid request for bar/baz",
			req: &types.QuerySpotPriceRequest{
				PoolId:          poolID,
				BaseAssetDenom:  "bar",
				QuoteAssetDenom: "baz",
			},
			result: sdk.NewDecWithPrec(15, 1).String(),
		},
		{
			name: "valid request for baz/foo",
			req: &types.QuerySpotPriceRequest{
				PoolId:          poolID,
				BaseAssetDenom:  "baz",
				QuoteAssetDenom: "foo",
			},
			result: sdk.MustNewDecFromStr("0.333333330000000000").String(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			result, err := queryClient.SpotPrice(gocontext.Background(), tc.req)
			if tc.expectErr {
				s.Require().Error(err, "expected error")
			} else {
				s.Require().NoError(err, "unexpected error")
				s.Require().Equal(tc.result, result.SpotPrice)
			}
		})
	}
}

func (s *KeeperTestSuite) TestV2QueryBalancerPoolSpotPrice() {
	v2queryClient := v2types.NewQueryClient(s.QueryHelper)
	coins := sdk.NewCoins(
		sdk.NewInt64Coin("tokenA", 1000),
		sdk.NewInt64Coin("tokenB", 2000),
		sdk.NewInt64Coin("tokenC", 3000),
		sdk.NewInt64Coin("tokenD", 4000),
		sdk.NewInt64Coin("tokenE", 4000), // 4000 intentional
	)
	poolID := s.PrepareBalancerPoolWithCoins(coins...)

	testCases := []struct {
		name      string
		req       *v2types.QuerySpotPriceRequest
		expectErr bool
		result    string
	}{
		{
			name: "non-existent pool",
			req: &v2types.QuerySpotPriceRequest{
				PoolId:          0,
				BaseAssetDenom:  "tokenA",
				QuoteAssetDenom: "tokenB",
			},
			expectErr: true,
		},
		{
			name: "missing asset denoms",
			req: &v2types.QuerySpotPriceRequest{
				PoolId: poolID,
			},
			expectErr: true,
		},
		{
			name: "missing pool ID and quote denom",
			req: &v2types.QuerySpotPriceRequest{
				BaseAssetDenom: "tokenA",
			},
			expectErr: true,
		},
		{
			name: "missing pool ID and base denom",
			req: &v2types.QuerySpotPriceRequest{
				QuoteAssetDenom: "tokenB",
			},
			expectErr: true,
		},
		{
			name: "tokenA in terms of tokenB",
			req: &v2types.QuerySpotPriceRequest{
				PoolId:          poolID,
				BaseAssetDenom:  "tokenA",
				QuoteAssetDenom: "tokenB",
			},
			result: sdk.NewDec(2).String(),
		},
		{
			name: "tokenB in terms of tokenA",
			req: &v2types.QuerySpotPriceRequest{
				PoolId:          poolID,
				BaseAssetDenom:  "tokenB",
				QuoteAssetDenom: "tokenA",
			},
			result: sdk.NewDecWithPrec(5, 1).String(),
		},
		{
			name: "tokenC in terms of tokenD (rounded decimal of 4/3)",
			req: &v2types.QuerySpotPriceRequest{
				PoolId:          poolID,
				BaseAssetDenom:  "tokenC",
				QuoteAssetDenom: "tokenD",
			},
			result: sdk.MustNewDecFromStr("1.333333330000000000").String(),
		},
		{
			name: "tokenD in terms of tokenE (1)",
			req: &v2types.QuerySpotPriceRequest{
				PoolId:          poolID,
				BaseAssetDenom:  "tokenD",
				QuoteAssetDenom: "tokenE",
			},
			result: sdk.OneDec().String(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			result, err := v2queryClient.SpotPrice(gocontext.Background(), tc.req)
			if tc.expectErr {
				s.Require().Error(err, "expected error")
			} else {
				s.Require().NoError(err, "unexpected error")
				s.Require().Equal(tc.result, result.SpotPrice)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryStableswapPoolSpotPrice() {
	queryClient := s.queryClient
	poolIDEven := s.PrepareBasicStableswapPool()
	poolIDUneven := s.PrepareImbalancedStableswapPool()

	testCases := []struct {
		name      string
		req       *types.QuerySpotPriceRequest
		expectErr bool
		result    string
	}{
		{
			name: "non-existent pool",
			req: &types.QuerySpotPriceRequest{
				PoolId:          0,
				BaseAssetDenom:  "foo",
				QuoteAssetDenom: "bar",
			},
			expectErr: true,
		},
		{
			name: "missing asset denoms",
			req: &types.QuerySpotPriceRequest{
				PoolId: poolIDEven,
			},
			expectErr: true,
		},
		{
			name: "missing pool ID and quote denom",
			req: &types.QuerySpotPriceRequest{
				BaseAssetDenom: "foo",
			},
			expectErr: true,
		},
		{
			name: "missing pool ID and base denom",
			req: &types.QuerySpotPriceRequest{
				QuoteAssetDenom: "bar",
			},
			expectErr: true,
		},
		{
			name: "valid request for foo/bar in even pool",
			req: &types.QuerySpotPriceRequest{
				PoolId:          poolIDEven,
				BaseAssetDenom:  "bar",
				QuoteAssetDenom: "foo",
			},
			result: "1.000000000000000000",
		},
		{
			name: "foo in terms of bar for in a 1:2:3, foo bar baz pool",
			req: &types.QuerySpotPriceRequest{
				PoolId:          poolIDUneven,
				BaseAssetDenom:  "bar",
				QuoteAssetDenom: "foo",
			},
			result: "1.454545450000000000",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			result, err := queryClient.SpotPrice(gocontext.Background(), tc.req)
			if tc.expectErr {
				s.Require().Error(err, "expected error")
			} else {
				s.Require().NoError(err, "unexpected error")
				// We allow for a small geometric error due to our spot price being an approximation
				expectedSpotPrice := sdk.MustNewDecFromStr(tc.result)
				actualSpotPrice := sdk.MustNewDecFromStr(result.SpotPrice)
				diff := (expectedSpotPrice.Sub(actualSpotPrice)).Abs()
				errTerm := diff.Quo(sdk.MinDec(expectedSpotPrice, actualSpotPrice))

				s.Require().True(errTerm.LT(sdk.NewDecWithPrec(1, 3)), "Expected: %d, Actual: %d", expectedSpotPrice, actualSpotPrice)
			}
		})
	}
}

func (s *KeeperTestSuite) TestV2QueryStableswapPoolSpotPrice() {
	v2queryClient := v2types.NewQueryClient(s.QueryHelper)
	poolIDEven := s.PrepareBasicStableswapPool()
	poolIDUneven := s.PrepareImbalancedStableswapPool()

	testCases := []struct {
		name      string
		req       *v2types.QuerySpotPriceRequest
		expectErr bool
		result    string
	}{
		{
			name: "non-existent pool",
			req: &v2types.QuerySpotPriceRequest{
				PoolId:          0,
				BaseAssetDenom:  "foo",
				QuoteAssetDenom: "bar",
			},
			expectErr: true,
		},
		{
			name: "missing asset denoms",
			req: &v2types.QuerySpotPriceRequest{
				PoolId: poolIDEven,
			},
			expectErr: true,
		},
		{
			name: "missing pool ID and quote denom",
			req: &v2types.QuerySpotPriceRequest{
				BaseAssetDenom: "foo",
			},
			expectErr: true,
		},
		{
			name: "missing pool ID and base denom",
			req: &v2types.QuerySpotPriceRequest{
				QuoteAssetDenom: "bar",
			},
			expectErr: true,
		},
		{
			name: "foo in terms of bar in even pool",
			req: &v2types.QuerySpotPriceRequest{
				PoolId:          poolIDEven,
				BaseAssetDenom:  "bar",
				QuoteAssetDenom: "foo",
			},
			result: "1.000000000000000000",
		},
		{
			name: "foo in terms of bar in uneven pool",
			req: &v2types.QuerySpotPriceRequest{
				PoolId:          poolIDUneven,
				BaseAssetDenom:  "foo",
				QuoteAssetDenom: "bar",
			},
			result: "1.454545450000000000",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			result, err := v2queryClient.SpotPrice(gocontext.Background(), tc.req)
			if tc.expectErr {
				s.Require().Error(err, "expected error")
			} else {
				s.Require().NoError(err, "unexpected error")

				// We allow for a small geometric error due to our spot price being an approximation
				expectedSpotPrice := sdk.MustNewDecFromStr(tc.result)
				actualSpotPrice := sdk.MustNewDecFromStr(result.SpotPrice)
				diff := (expectedSpotPrice.Sub(actualSpotPrice)).Abs()
				errTerm := diff.Quo(sdk.MinDec(expectedSpotPrice, actualSpotPrice))

				s.Require().True(errTerm.LT(sdk.NewDecWithPrec(1, 3)), "Expected: %d, Actual: %d", expectedSpotPrice, actualSpotPrice)
			}
		})
	}
}
