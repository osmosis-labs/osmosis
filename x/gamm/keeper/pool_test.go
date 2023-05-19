package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"github.com/osmosis-labs/osmosis/v15/tests/mocks"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/stableswap"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

var (
	defaultPoolParamsStableSwap = stableswap.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}
	defaultPoolId = uint64(1)
)

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
				SwapFee: defaultSwapFee,
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
					SwapFee: defaultSwapFee,
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

func (s *KeeperTestSuite) TestConvertToCFMMPool() {
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

			pool, err := keeper.ConvertToCFMMPool(tc.pool)

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
						SwapFee: defaultSwapFee,
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
