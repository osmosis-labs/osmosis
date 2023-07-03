package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"github.com/osmosis-labs/osmosis/v16/tests/mocks"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/stableswap"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

var (
	defaultPoolParamsStableSwap = stableswap.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}
	defaultPoolId = uint64(1)
)

// import (
// 	"math/rand"
// 	"time"

// 	"github.com/cosmos/cosmos-sdk/simapp"
// 	sdk "github.com/cosmos/cosmos-sdk/types"

// 	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
// )

// func (s *KeeperTestSuite) TestCleanupPool() {
// 	// Mint some assets to the accounts.
// 	for _, acc := range s.TestAccs {
// 		s.FundAcc(
// 			s.App.BankKeeper,
// 			s.Ctx,
// 			acc,
// 			sdk.NewCoins(
// 				sdk.NewCoin("uosmo", sdk.NewInt(1000000000)),
// 				sdk.NewCoin("foo", sdk.NewInt(1000)),
// 				sdk.NewCoin("bar", sdk.NewInt(1000)),
// 				sdk.NewCoin("baz", sdk.NewInt(1000)),
// 			),
// 		)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}

// 	poolId, err := s.App.GAMMKeeper.CreateBalancerPool(s.Ctx, acc1, defaultPoolParams, []types.PoolAsset{
// 		{
// 			Weight: sdk.NewInt(100),
// 			Token:  sdk.NewCoin("foo", sdk.NewInt(1000)),
// 		},
// 		{
// 			Weight: sdk.NewInt(100),
// 			Token:  sdk.NewCoin("bar", sdk.NewInt(1000)),
// 		},
// 		{
// 			Weight: sdk.NewInt(100),
// 			Token:  sdk.NewCoin("baz", sdk.NewInt(1000)),
// 		},
// 	}, "")
// 	s.NoError(err)

// 	for _, acc := range []sdk.AccAddress{acc2, acc3} {
// 		err = s.App.GAMMKeeper.JoinPool(s.Ctx, acc, poolId, types.OneShare.MulRaw(100), sdk.NewCoins(
// 			sdk.NewCoin("foo", sdk.NewInt(1000)),
// 			sdk.NewCoin("bar", sdk.NewInt(1000)),
// 			sdk.NewCoin("baz", sdk.NewInt(1000)),
// 		))
// 		s.NoError(err)
// 	}

// 	pool, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
// 	s.NoError(err)
// 	denom := pool.GetTotalShares().Denom
// 	totalAmount := sdk.ZeroInt()
// 	for _, acc := range s.TestAccs {
// 		coin := s.App.BankKeeper.GetBalance(s.Ctx, acc, denom)
// 		s.True(coin.Amount.Equal(types.OneShare.MulRaw(100)))
// 		totalAmount = totalAmount.Add(coin.Amount)
// 	}
// 	s.True(totalAmount.Equal(types.OneShare.MulRaw(300)))

// 	err = s.App.GAMMKeeper.CleanupBalancerPool(s.Ctx, []uint64{poolId}, []string{})
// 	s.NoError(err)
// 	for _, acc := range s.TestAccs {
// 		for _, denom := range []string{"foo", "bar", "baz"} {
// 			amt := s.App.BankKeeper.GetBalance(s.Ctx, acc, denom)
// 			s.True(amt.Amount.Equal(sdk.NewInt(1000)),
// 				"Expected equal %s: %d, %d", amt.Denom, amt.Amount.Int64(), 1000)
// 		}
// 	}
// }

// func (s *KeeperTestSuite) TestCleanupPoolRandomized() {
// 	// address => deposited coins
// 	coinOf := make(map[string]sdk.Coins)
// 	denoms := []string{"foo", "bar", "baz"}

// 	// Mint some assets to the accounts.
// 	for _, acc := range s.TestAccs {
// 		coins := make(sdk.Coins, 3)
// 		for i := range coins {
// 			amount := sdk.NewInt(rand.Int63n(1000))
// 			// give large amount of coins to the pool creator
// 			if i == 0 {
// 				amount = amount.MulRaw(10000)
// 			}
// 			coins[i] = sdk.Coin{denoms[i], amount}
// 		}
// 		coinOf[acc.String()] = coins
// 		coins = append(coins, sdk.NewCoin("uosmo", sdk.NewInt(1000000000)))

// 		s.FundAcc(
// 			s.App.BankKeeper,
// 			s.Ctx,
// 			acc,
// 			coins.Sort(),
// 		)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}

// 	initialAssets := []types.PoolAsset{}
// 	for _, coin := range coinOf[acc1.String()] {
// 		initialAssets = append(initialAssets, types.PoolAsset{Weight: types.OneShare.MulRaw(100), Token: coin})
// 	}
// 	poolId, err := s.App.GAMMKeeper.CreateBalancerPool(s.Ctx, acc1, defaultPoolParams, initialAssets, "")
// 	s.NoError(err)

// 	for _, acc := range []sdk.AccAddress{acc2, acc3} {
// 		err = s.App.GAMMKeeper.JoinPool(s.Ctx, acc, poolId, types.OneShare, coinOf[acc.String()])
// 		s.NoError(err)
// 	}

// 	err = s.App.GAMMKeeper.CleanupBalancerPool(s.Ctx, []uint64{poolId}, []string{})
// 	s.NoError(err)
// 	for _, acc := range s.TestAccs {
// 		for _, coin := range coinOf[acc.String()] {
// 			amt := s.App.BankKeeper.GetBalance(s.Ctx, acc, coin.Denom)
// 			// the refund could have rounding error
// 			s.True(amt.Amount.Sub(coin.Amount).Abs().LTE(sdk.NewInt(2)),
// 				"Expected equal %s: %d, %d", amt.Denom, amt.Amount.Int64(), coin.Amount.Int64())
// 		}
// 	}
// }

// func (s *KeeperTestSuite) TestCleanupPoolErrorOnSwap() {
// 	s.Ctx = s.Ctx.WithBlockTime(time.Unix(1000, 1000))
// 	s.FundAcc(
// 		s.App.BankKeeper,
// 		s.Ctx,
// 		acc1,
// 		sdk.NewCoins(
// 			sdk.NewCoin("uosmo", sdk.NewInt(1000000000)),
// 			sdk.NewCoin("foo", sdk.NewInt(1000)),
// 			sdk.NewCoin("bar", sdk.NewInt(1000)),
// 			sdk.NewCoin("baz", sdk.NewInt(1000)),
// 		),
// 	)
// 	if err != nil {
// 		panic(err)
// 	}

// 	poolId, err := s.App.GAMMKeeper.CreateBalancerPool(s.Ctx, acc1, defaultPoolParams, []types.PoolAsset{
// 		{
// 			Weight: sdk.NewInt(100),
// 			Token:  sdk.NewCoin("foo", sdk.NewInt(1000)),
// 		},
// 		{
// 			Weight: sdk.NewInt(100),
// 			Token:  sdk.NewCoin("bar", sdk.NewInt(1000)),
// 		},
// 		{
// 			Weight: sdk.NewInt(100),
// 			Token:  sdk.NewCoin("baz", sdk.NewInt(1000)),
// 		},
// 	}, "")
// 	s.NoError(err)

// 	err = s.App.GAMMKeeper.CleanupBalancerPool(s.Ctx, []uint64{poolId}, []string{})
// 	s.NoError(err)

// 	_, _, err = s.App.GAMMKeeper.SwapExactAmountIn(s.Ctx, acc1, poolId, sdk.NewCoin("foo", sdk.NewInt(1)), "bar", sdk.NewInt(1))
// 	s.Error(err)
// }

// func (s *KeeperTestSuite) TestCleanupPoolWithLockup() {
// 	s.Ctx = s.Ctx.WithBlockTime(time.Unix(1000, 1000))
// 	s.FundAcc(
// 		s.App.BankKeeper,
// 		s.Ctx,
// 		acc1,
// 		sdk.NewCoins(
// 			sdk.NewCoin("uosmo", sdk.NewInt(1000000000)),
// 			sdk.NewCoin("foo", sdk.NewInt(1000)),
// 			sdk.NewCoin("bar", sdk.NewInt(1000)),
// 			sdk.NewCoin("baz", sdk.NewInt(1000)),
// 		),
// 	)
// 	if err != nil {
// 		panic(err)
// 	}

// 	poolId, err := s.App.GAMMKeeper.CreateBalancerPool(s.Ctx, acc1, defaultPoolParams, []types.PoolAsset{
// 		{
// 			Weight: sdk.NewInt(100),
// 			Token:  sdk.NewCoin("foo", sdk.NewInt(1000)),
// 		},
// 		{
// 			Weight: sdk.NewInt(100),
// 			Token:  sdk.NewCoin("bar", sdk.NewInt(1000)),
// 		},
// 		{
// 			Weight: sdk.NewInt(100),
// 			Token:  sdk.NewCoin("baz", sdk.NewInt(1000)),
// 		},
// 	}, "")
// 	s.NoError(err)

// 	_, err = s.App.LockupKeeper.LockTokens(s.Ctx, acc1, sdk.Coins{sdk.NewCoin(types.GetPoolShareDenom(poolId), types.InitPoolSharesSupply)}, time.Hour)
// 	s.NoError(err)

// 	for _, lock := range s.App.LockupKeeper.GetLocksDenom(s.Ctx, types.GetPoolShareDenom(poolId)) {
// 		err = s.App.LockupKeeper.ForceUnlock(s.Ctx, lock)
// 		s.NoError(err)
// 	}

// 	err = s.App.GAMMKeeper.CleanupBalancerPool(s.Ctx, []uint64{poolId}, []string{})
// 	s.NoError(err)
// 	for _, coin := range []string{"foo", "bar", "baz"} {
// 		amt := s.App.BankKeeper.GetBalance(s.Ctx, acc1, coin)
// 		// the refund could have rounding error
// 		s.True(amt.Amount.Equal(sdk.NewInt(1000)) || amt.Amount.Equal(sdk.NewInt(1000).SubRaw(1)),
// 			"Expected equal %s: %d, %d", amt.Denom, amt.Amount.Int64(), sdk.NewInt(1000).Int64())
// 	}
// }

// TestGetPoolAndPoke tests that the right pools is returned from GetPoolAndPoke.
// For the pools implementing the weighted extension, asserts that PokePool is called.
func (s *KeeperTestSuite) TestGetPoolAndPoke() {
	const (
		startTime = 1000
		blockTime = startTime + 100
	)

	// N.B.: We make a copy because SmoothWeightChangeParams get mutated.
	// We would like to avoid mutating global pool assets that are used in other tests.
	defaultPoolAssetsCopy := make([]balancer.PoolAsset, 2)
	copy(defaultPoolAssetsCopy, defaultPoolAssets)

	startPoolWeightAssets := []balancer.PoolAsset{
		{
			Weight: defaultPoolAssets[0].Weight.Quo(sdk.NewInt(2)),
			Token:  defaultPoolAssets[0].Token,
		},
		{
			Weight: defaultPoolAssets[1].Weight.Mul(sdk.NewInt(3)),
			Token:  defaultPoolAssets[1].Token,
		},
	}

	tests := map[string]struct {
		isPokePool bool
		poolId     uint64
	}{
		"weighted pool - change weights": {
			isPokePool: true,
			poolId: s.prepareCustomBalancerPool(defaultAcctFunds, startPoolWeightAssets, balancer.PoolParams{
				SwapFee: defaultSpreadFactor,
				ExitFee: defaultZeroExitFee,
				SmoothWeightChangeParams: &balancer.SmoothWeightChangeParams{
					StartTime:          time.Unix(startTime, 0), // start time is before block time so the weights should change
					Duration:           time.Hour,
					InitialPoolWeights: startPoolWeightAssets,
					TargetPoolWeights:  defaultPoolAssetsCopy,
				},
			}),
		},
		"non weighted pool": {
			poolId: s.prepareCustomStableswapPool(
				defaultAcctFunds,
				stableswap.PoolParams{
					SwapFee: defaultSpreadFactor,
					ExitFee: defaultZeroExitFee,
				},
				sdk.NewCoins(sdk.NewCoin(defaultAcctFunds[0].Denom, defaultAcctFunds[0].Amount.QuoRaw(2)), sdk.NewCoin(defaultAcctFunds[1].Denom, defaultAcctFunds[1].Amount.QuoRaw(2))),
				[]uint64{1, 1},
			),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			k := s.App.GAMMKeeper
			ctx := s.Ctx.WithBlockTime(time.Unix(blockTime, 0))

			pool, err := k.GetPoolAndPoke(ctx, tc.poolId)

			s.Require().NoError(err)
			s.Require().Equal(tc.poolId, pool.GetId())

			if tc.isPokePool {
				pokePool, ok := pool.(types.WeightedPoolExtension)
				s.Require().True(ok)

				poolAssetWeight0, err := pokePool.GetTokenWeight(startPoolWeightAssets[0].Token.Denom)
				s.Require().NoError(err)

				poolAssetWeight1, err := pokePool.GetTokenWeight(startPoolWeightAssets[1].Token.Denom)
				s.Require().NoError(err)

				s.Require().NotEqual(startPoolWeightAssets[0].Weight, poolAssetWeight0)
				s.Require().NotEqual(startPoolWeightAssets[1].Weight, poolAssetWeight1)
				return
			}

			_, ok := pool.(types.WeightedPoolExtension)
			s.Require().False(ok)
		})
	}
}

func (s *KeeperTestSuite) TestAsCFMMPool() {
	ctrl := gomock.NewController(s.T())

	tests := map[string]struct {
		pool        poolmanagertypes.PoolI
		expectError bool
	}{
		"cfmm pool": {
			pool: mocks.NewMockCFMMPoolI(ctrl),
		},
		"non cfmm pool": {
			pool:        mocks.NewMockConcentratedPoolExtension(ctrl),
			expectError: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			pool, err := keeper.AsCFMMPool(tc.pool)

			if tc.expectError {
				s.Require().Error(err)
				s.Require().Nil(pool)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(pool)
			s.Require().Equal(tc.pool, pool)
		})
	}
}

// TestMarshalUnmarshalPool tests that by changing the interfaces
// that we marshal to and unmarshal from store, we do not
// change the underlying bytes. This shows that migrations are
// not necessary.
func (s *KeeperTestSuite) TestMarshalUnmarshalPool() {
	s.SetupTest()
	k := s.App.GAMMKeeper

	balancerPoolId := s.PrepareBalancerPool()
	balancerPool, err := k.GetPoolAndPoke(s.Ctx, balancerPoolId)
	s.Require().NoError(err)

	stableswapPoolId := s.PrepareBasicStableswapPool()
	stableswapPool, err := k.GetPoolAndPoke(s.Ctx, stableswapPoolId)
	s.Require().NoError(err)

	tests := []struct {
		name string
		pool types.CFMMPoolI
	}{
		{
			name: "balancer",
			pool: balancerPool,
		},
		{
			name: "stableswap",
			pool: stableswapPool,
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			var poolI poolmanagertypes.PoolI = tc.pool
			var cfmmPoolI types.CFMMPoolI = tc.pool

			// Marshal poolI as PoolI
			bzPoolI, err := k.MarshalPool(poolI)
			s.Require().NoError(err)

			// Marshal cfmmPoolI as PoolI
			bzCfmmPoolI, err := k.MarshalPool(cfmmPoolI)
			s.Require().NoError(err)

			s.Require().Equal(bzPoolI, bzCfmmPoolI)

			// Unmarshal bzPoolI as CFMMPoolI
			unmarshalBzPoolIAsCfmmPoolI, err := k.UnmarshalPool(bzPoolI)
			s.Require().NoError(err)

			// Unmarshal bzPoolI as PoolI
			unmarshalBzPoolIAsPoolI, err := k.UnmarshalPoolLegacy(bzPoolI)
			s.Require().NoError(err)

			s.Require().Equal(unmarshalBzPoolIAsCfmmPoolI, unmarshalBzPoolIAsPoolI)

			// Unmarshal bzCfmmPoolI as CFMMPoolI
			unmarshalBzCfmmPoolIAsCfmmPoolI, err := k.UnmarshalPool(bzCfmmPoolI)
			s.Require().NoError(err)

			// Unmarshal bzCfmmPoolI as PoolI
			unmarshalBzCfmmPoolIAsPoolI, err := k.UnmarshalPoolLegacy(bzCfmmPoolI)
			s.Require().NoError(err)

			// bzCfmmPoolI as CFMMPoolI equals bzCfmmPoolI as PoolI
			s.Require().Equal(unmarshalBzCfmmPoolIAsCfmmPoolI, unmarshalBzCfmmPoolIAsPoolI)

			// All unmarshalled combinations are equal.
			s.Require().Equal(unmarshalBzPoolIAsCfmmPoolI, unmarshalBzCfmmPoolIAsCfmmPoolI)
		})
	}
}

func (s *KeeperTestSuite) TestSetStableSwapScalingFactors() {
	controllerAddr := s.TestAccs[0]
	failAddr := s.TestAccs[1]

	testcases := []struct {
		name             string
		poolId           uint64
		scalingFactors   []uint64
		sender           sdk.AccAddress
		expError         error
		isStableSwapPool bool
	}{
		{
			name:             "Error: Pool does not exist",
			poolId:           2,
			scalingFactors:   []uint64{1, 1},
			sender:           controllerAddr,
			expError:         types.PoolDoesNotExistError{PoolId: defaultPoolId + 1},
			isStableSwapPool: false,
		},
		{
			name:             "Error: Pool id is not of type stableswap pool",
			poolId:           1,
			scalingFactors:   []uint64{1, 1},
			sender:           controllerAddr,
			expError:         fmt.Errorf("pool id 1 is not of type stableswap pool"),
			isStableSwapPool: false,
		},
		{
			name:             "Error: Can not set scaling factors",
			poolId:           1,
			scalingFactors:   []uint64{1, 1},
			sender:           failAddr,
			expError:         types.ErrNotScalingFactorGovernor,
			isStableSwapPool: true,
		},
		{
			name:             "Valid case",
			poolId:           1,
			scalingFactors:   []uint64{1, 1},
			sender:           controllerAddr,
			isStableSwapPool: true,
		},
	}
	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()
			if tc.isStableSwapPool == true {
				poolId := s.prepareCustomStableswapPool(
					defaultAcctFunds,
					stableswap.PoolParams{
						SwapFee: defaultSpreadFactor,
						ExitFee: defaultZeroExitFee,
					},
					sdk.NewCoins(sdk.NewCoin(defaultAcctFunds[0].Denom, defaultAcctFunds[0].Amount.QuoRaw(2)), sdk.NewCoin(defaultAcctFunds[1].Denom, defaultAcctFunds[1].Amount.QuoRaw(2))),
					tc.scalingFactors,
				)
				pool, _ := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolId)
				stableswapPool, _ := pool.(*stableswap.Pool)
				stableswapPool.ScalingFactorController = controllerAddr.String()
				err := s.App.GAMMKeeper.SetPool(s.Ctx, stableswapPool)
				s.Require().NoError(err)
			} else {
				s.prepareCustomBalancerPool(
					defaultAcctFunds,
					defaultPoolAssets,
					defaultPoolParams)
			}
			err := s.App.GAMMKeeper.SetStableSwapScalingFactors(s.Ctx, tc.poolId, tc.scalingFactors, tc.sender.String())
			if tc.expError != nil {
				s.Require().Error(err)
				s.Require().EqualError(err, tc.expError.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetMaximalNoSwapLPAmount() {
	tests := map[string]struct {
		poolId              uint64
		shareOutAmount      sdk.Int
		expectedLpLiquidity sdk.Coins
		err                 error
	}{
		"Balancer pool: Share ratio is zero": {
			poolId: suite.prepareCustomBalancerPool(defaultAcctFunds, defaultPoolAssets, balancer.PoolParams{
				SwapFee: defaultSpreadFactor,
				ExitFee: defaultZeroExitFee,
			}),
			shareOutAmount: sdk.ZeroInt(),
			err:            types.ErrInvalidMathApprox,
		},

		"Balancer pool: Share ratio is negative": {
			poolId: suite.prepareCustomBalancerPool(defaultAcctFunds, defaultPoolAssets, balancer.PoolParams{
				SwapFee: defaultSpreadFactor,
				ExitFee: defaultZeroExitFee,
			}),
			shareOutAmount: sdk.NewInt(-1),
			err:            types.ErrInvalidMathApprox,
		},

		"Balancer pool: Pass": {
			poolId: suite.prepareCustomBalancerPool(defaultAcctFunds, defaultPoolAssets, balancer.PoolParams{
				SwapFee: defaultSpreadFactor,
				ExitFee: defaultZeroExitFee,
			}),

			// totalShare:   100_000_000_000_000_000_000
			// shareOutAmount: 8_000_000_000_000_000_000
			// shareRatio = shareOutAmount/totalShare = 0.08
			// Amount of tokens in poolAssets:
			// 		- defaultPoolAssets[1].Token.Amount: 10000
			//  	- defaultPoolAssets[0].Token.Amount: 10000
			shareOutAmount: sdk.NewInt(8_000_000_000_000_000_000),
			expectedLpLiquidity: sdk.Coins{
				sdk.NewInt64Coin(defaultPoolAssets[1].Token.Denom, 800),
				sdk.NewInt64Coin(defaultPoolAssets[0].Token.Denom, 800),
			},
		},

		"Balancer pool: Pass with ceiling result": {
			poolId: suite.prepareCustomBalancerPool(defaultAcctFunds, defaultPoolAssets, balancer.PoolParams{
				SwapFee: defaultSpreadFactor,
				ExitFee: defaultZeroExitFee,
			}),

			// totalShare:   100_000_000_000_000_000_000
			// shareOutAmount: 8_888_000_000_000_000_000
			// shareRatio = shareOutAmount/totalShare = 0.08888
			// Amount of tokens in poolAssets:
			// 		- defaultPoolAssets[1].Token.Amount: 10000
			//  	- defaultPoolAssets[0].Token.Amount: 10000
			shareOutAmount: sdk.NewInt(8_888_000_000_000_000_000),
			expectedLpLiquidity: sdk.Coins{
				sdk.NewInt64Coin(defaultPoolAssets[1].Token.Denom, 889),
				sdk.NewInt64Coin(defaultPoolAssets[0].Token.Denom, 889),
			},
		},

		"Stableswap pool: Share ratio is zero with even two-asset": {
			poolId: suite.prepareCustomStableswapPool(
				defaultAcctFunds,
				stableswap.PoolParams{
					SwapFee: defaultSpreadFactor,
					ExitFee: defaultZeroExitFee,
				},
				sdk.NewCoins(sdk.NewCoin(defaultAcctFunds[1].Denom, defaultAcctFunds[1].Amount), sdk.NewCoin(defaultAcctFunds[2].Denom, defaultAcctFunds[2].Amount)),
				[]uint64{1, 1},
			),
			shareOutAmount: sdk.ZeroInt(),
			err:            types.ErrInvalidMathApprox,
		},

		"Stableswap pool: Share ratio is zero with uneven two-asset": {
			poolId: suite.prepareCustomStableswapPool(
				defaultAcctFunds,
				stableswap.PoolParams{
					SwapFee: defaultSpreadFactor,
					ExitFee: defaultZeroExitFee,
				},
				sdk.NewCoins(sdk.NewCoin(defaultAcctFunds[1].Denom, defaultAcctFunds[1].Amount.QuoRaw(2)), sdk.NewCoin(defaultAcctFunds[2].Denom, defaultAcctFunds[2].Amount)),
				[]uint64{1, 1},
			),
			shareOutAmount: sdk.ZeroInt(),
			err:            types.ErrInvalidMathApprox,
		},

		"Stableswap pool: Share ratio is negative with even two-asset": {
			poolId: suite.prepareCustomStableswapPool(
				defaultAcctFunds,
				stableswap.PoolParams{
					SwapFee: defaultSpreadFactor,
					ExitFee: defaultZeroExitFee,
				},
				sdk.NewCoins(sdk.NewCoin(defaultAcctFunds[1].Denom, defaultAcctFunds[1].Amount), sdk.NewCoin(defaultAcctFunds[2].Denom, defaultAcctFunds[2].Amount)),
				[]uint64{1, 1},
			),
			shareOutAmount: sdk.NewInt(-1),
			err:            types.ErrInvalidMathApprox,
		},

		"Stableswap pool: Share ratio is negative with uneven two-asset": {
			poolId: suite.prepareCustomStableswapPool(
				defaultAcctFunds,
				stableswap.PoolParams{
					SwapFee: defaultSpreadFactor,
					ExitFee: defaultZeroExitFee,
				},
				sdk.NewCoins(sdk.NewCoin(defaultAcctFunds[1].Denom, defaultAcctFunds[1].Amount.QuoRaw(2)), sdk.NewCoin(defaultAcctFunds[2].Denom, defaultAcctFunds[2].Amount)),
				[]uint64{1, 1},
			),
			shareOutAmount: sdk.NewInt(-1),
			err:            types.ErrInvalidMathApprox,
		},

		"Stableswap pool: Pass with even two-asset": {
			poolId: suite.prepareCustomStableswapPool(
				defaultAcctFunds,
				stableswap.PoolParams{
					SwapFee: defaultSpreadFactor,
					ExitFee: defaultZeroExitFee,
				},
				sdk.NewCoins(sdk.NewCoin(defaultAcctFunds[1].Denom, defaultAcctFunds[1].Amount), sdk.NewCoin(defaultAcctFunds[2].Denom, defaultAcctFunds[2].Amount)),
				[]uint64{1, 1},
			),
			shareOutAmount: sdk.NewInt(8_000_000_000_000_000_000),
			expectedLpLiquidity: sdk.Coins{
				sdk.NewInt64Coin(defaultAcctFunds[1].Denom, 800000),
				sdk.NewInt64Coin(defaultAcctFunds[2].Denom, 800000),
			},
		},

		"Stableswap pool: Pass with even two-asset, ceiling result": {
			poolId: suite.prepareCustomStableswapPool(
				defaultAcctFunds,
				stableswap.PoolParams{
					SwapFee: defaultSpreadFactor,
					ExitFee: defaultZeroExitFee,
				},
				sdk.NewCoins(sdk.NewCoin(defaultAcctFunds[1].Denom, defaultAcctFunds[1].Amount), sdk.NewCoin(defaultAcctFunds[2].Denom, defaultAcctFunds[2].Amount)),
				[]uint64{1, 1},
			),
			// totalShare:   100_000_000_000_000_000_000
			// shareOutAmount: 8_888_888_000_000_000_000
			// shareRatio = shareOutAmount/totalShare = 0.08888888
			// Amount of tokens in liquidity:
			// 		- defaultAcctFunds[1].Amount: 10000000
			//  	- defaultAcctFunds[2].Amount: 10000000
			shareOutAmount: sdk.NewInt(8_888_888_000_000_000_000),
			expectedLpLiquidity: sdk.Coins{
				sdk.NewInt64Coin(defaultAcctFunds[1].Denom, 888889),
				sdk.NewInt64Coin(defaultAcctFunds[2].Denom, 888889),
			},
		},

		"Stableswap pool: Pass with uneven two-asset": {
			poolId: suite.prepareCustomStableswapPool(
				defaultAcctFunds,
				stableswap.PoolParams{
					SwapFee: defaultSpreadFactor,
					ExitFee: defaultZeroExitFee,
				},
				sdk.NewCoins(sdk.NewCoin(defaultAcctFunds[1].Denom, defaultAcctFunds[1].Amount.QuoRaw(2)), sdk.NewCoin(defaultAcctFunds[2].Denom, defaultAcctFunds[2].Amount)),
				[]uint64{1, 1},
			),
			shareOutAmount: sdk.NewInt(8_000_000_000_000_000_000),
			expectedLpLiquidity: sdk.Coins{
				sdk.NewInt64Coin(defaultAcctFunds[1].Denom, 400000),
				sdk.NewInt64Coin(defaultAcctFunds[2].Denom, 800000),
			},
		},

		"Stableswap pool: Pass with uneven two-asset, ceiling result": {
			poolId: suite.prepareCustomStableswapPool(
				defaultAcctFunds,
				stableswap.PoolParams{
					SwapFee: defaultSpreadFactor,
					ExitFee: defaultZeroExitFee,
				},
				sdk.NewCoins(sdk.NewCoin(defaultAcctFunds[1].Denom, defaultAcctFunds[1].Amount.QuoRaw(2)), sdk.NewCoin(defaultAcctFunds[2].Denom, defaultAcctFunds[2].Amount)),
				[]uint64{1, 1},
			),
			// totalShare:   100_000_000_000_000_000_000
			// shareOutAmount: 8_888_888_000_000_000_000
			// shareRatio = shareOutAmount/totalShare = 0.08888888
			// Amount of tokens in liquidity:
			// 		- defaultAcctFunds[1].Amount: 10000000
			//  	- defaultAcctFunds[2].Amount: 5000000
			shareOutAmount: sdk.NewInt(8_888_888_000_000_000_000),
			expectedLpLiquidity: sdk.Coins{
				sdk.NewInt64Coin(defaultAcctFunds[1].Denom, 444445),
				sdk.NewInt64Coin(defaultAcctFunds[2].Denom, 888889),
			},
		},
	}
	for name, tc := range tests {
		suite.Run(name, func() {
			k := suite.App.GAMMKeeper

			pool, err := k.GetPoolAndPoke(suite.Ctx, tc.poolId)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.poolId, pool.GetId())

			neededLpLiquidity, err := keeper.GetMaximalNoSwapLPAmount(suite.Ctx, pool, tc.shareOutAmount)
			if tc.err != nil {
				suite.Require().Error(err)
				msgError := fmt.Sprintf("Too few shares out wanted. (debug: getMaximalNoSwapLPAmount share ratio is zero or negative): %s", tc.err)
				suite.Require().EqualError(err, msgError)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(neededLpLiquidity, tc.expectedLpLiquidity)
			}

		})
	}
}

func (suite *KeeperTestSuite) TestGetTotalPoolShares() {
	tests := map[string]struct {
		sharesJoined   sdk.Int
		poolNotCreated bool

		expectedError error
	}{
		"happy path: default balancer pool": {
			sharesJoined: sdk.ZeroInt(),
		},
		"Multiple LPs with shares exist": {
			sharesJoined: types.OneShare,
		},
		"error: pool does not exist": {
			sharesJoined:   sdk.ZeroInt(),
			poolNotCreated: true,
			expectedError:  types.PoolDoesNotExistError{PoolId: uint64(0)},
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			suite.SetupTest()
			gammKeeper := suite.App.GAMMKeeper
			testAccount := suite.TestAccs[0]

			// --- Setup ---

			// Mint some assets to the accounts.
			balancerPoolId := uint64(0)
			if !tc.poolNotCreated {
				balancerPoolId = suite.PrepareBalancerPool()
			}

			sharesJoined := sdk.ZeroInt()
			if !tc.sharesJoined.Equal(sdk.ZeroInt()) {
				suite.FundAcc(testAccount, defaultAcctFunds)
				_, sharesActualJoined, err := gammKeeper.JoinPoolNoSwap(suite.Ctx, testAccount, balancerPoolId, tc.sharesJoined, sdk.Coins{})
				suite.Require().NoError(err)
				sharesJoined = sharesActualJoined
			}

			// --- System under test ---

			totalShares, err := gammKeeper.GetTotalPoolShares(suite.Ctx, balancerPoolId)

			// --- Assertions ---

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}

			suite.Require().NoError(err)
			suite.Require().Equal(types.InitPoolSharesSupply.Add(sharesJoined), totalShares)
		})
	}
}
