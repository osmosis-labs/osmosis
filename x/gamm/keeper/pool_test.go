package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"github.com/osmosis-labs/osmosis/v15/tests/mocks"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

var (
	defaultPoolAssetsStableSwap = sdk.Coins{
		sdk.NewCoin("atom", sdk.NewInt(100)),
		sdk.NewCoin("osmo", sdk.NewInt(100)),
	}
	defaultPoolId                        = uint64(1)
	defaultAcctFundsStableSwap sdk.Coins = sdk.NewCoins(
		sdk.NewCoin("udym", sdk.NewInt(10000000000)),
		sdk.NewCoin("atom", sdk.NewInt(100)),
		sdk.NewCoin("osmo", sdk.NewInt(100)),
	)
)

// TestGetPoolAndPoke tests that the right pools is returned from GetPoolAndPoke.
// For the pools implementing the weighted extension, asserts that PokePool is called.
func (suite *KeeperTestSuite) TestGetPoolAndPoke() {
	const (
		startTime = 1000
		blockTime = startTime + 100
	)

	// N.B.: We make a copy because SmoothWeightChangeParams get mutated.
	// We would like to avoid mutating global pool assets that are used in other tests.
	defaultPoolAssetsCopy := make([]balancertypes.PoolAsset, 2)
	copy(defaultPoolAssetsCopy, defaultPoolAssets)

	startPoolWeightAssets := []balancertypes.PoolAsset{
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
			poolId: suite.prepareCustomBalancerPool(defaultAcctFunds, startPoolWeightAssets, balancer.PoolParams{
				SwapFee: defaultSwapFee,
				ExitFee: defaultExitFee,
				SmoothWeightChangeParams: &balancer.SmoothWeightChangeParams{
					StartTime:          time.Unix(startTime, 0), // start time is before block time so the weights should change
					Duration:           time.Hour,
					InitialPoolWeights: startPoolWeightAssets,
					TargetPoolWeights:  defaultPoolAssetsCopy,
				},
			}),
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			k := suite.App.GAMMKeeper
			ctx := suite.Ctx.WithBlockTime(time.Unix(blockTime, 0))

			pool, err := k.GetPoolAndPoke(ctx, tc.poolId)

			suite.Require().NoError(err)
			suite.Require().Equal(tc.poolId, pool.GetId())

			if tc.isPokePool {
				pokePool, ok := pool.(types.WeightedPoolExtension)
				suite.Require().True(ok)

				poolAssetWeight0, err := pokePool.GetTokenWeight(startPoolWeightAssets[0].Token.Denom)
				suite.Require().NoError(err)

				poolAssetWeight1, err := pokePool.GetTokenWeight(startPoolWeightAssets[1].Token.Denom)
				suite.Require().NoError(err)

				suite.Require().NotEqual(startPoolWeightAssets[0].Weight, poolAssetWeight0)
				suite.Require().NotEqual(startPoolWeightAssets[1].Weight, poolAssetWeight1)
				return
			}

			_, ok := pool.(types.WeightedPoolExtension)
			suite.Require().False(ok)
		})
	}
}

func (suite *KeeperTestSuite) TestConvertToCFMMPool() {
	ctrl := gomock.NewController(suite.T())

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
		suite.Run(name, func() {
			suite.SetupTest()

			pool, err := keeper.ConvertToCFMMPool(tc.pool)

			if tc.expectError {
				suite.Require().Error(err)
				suite.Require().Nil(pool)
				return
			}

			suite.Require().NoError(err)
			suite.Require().NotNil(pool)
			suite.Require().Equal(tc.pool, pool)
		})
	}
}

// TestMarshalUnmarshalPool tests that by changing the interfaces
// that we marshal to and unmarshal from store, we do not
// change the underlying bytes. This shows that migrations are
// not necessary.
func (suite *KeeperTestSuite) TestMarshalUnmarshalPool() {

	suite.SetupTest()
	k := suite.App.GAMMKeeper

	balancerPoolId := suite.PrepareBalancerPool()
	balancerPool, err := k.GetPoolAndPoke(suite.Ctx, balancerPoolId)
	suite.Require().NoError(err)
	suite.Require().NoError(err)

	tests := []struct {
		name string
		pool types.CFMMPoolI
	}{
		{
			name: "balancer",
			pool: balancerPool,
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			var poolI poolmanagertypes.PoolI = tc.pool
			var cfmmPoolI types.CFMMPoolI = tc.pool

			// Marshal poolI as PoolI
			bzPoolI, err := k.MarshalPool(poolI)
			suite.Require().NoError(err)

			// Marshal cfmmPoolI as PoolI
			bzCfmmPoolI, err := k.MarshalPool(cfmmPoolI)
			suite.Require().NoError(err)

			suite.Require().Equal(bzPoolI, bzCfmmPoolI)

			// Unmarshal bzPoolI as CFMMPoolI
			unmarshalBzPoolIAsCfmmPoolI, err := k.UnmarshalPool(bzPoolI)
			suite.Require().NoError(err)

			// Unmarshal bzPoolI as PoolI
			unmarshalBzPoolIAsPoolI, err := k.UnmarshalPoolLegacy(bzPoolI)
			suite.Require().NoError(err)

			suite.Require().Equal(unmarshalBzPoolIAsCfmmPoolI, unmarshalBzPoolIAsPoolI)

			// Unmarshal bzCfmmPoolI as CFMMPoolI
			unmarshalBzCfmmPoolIAsCfmmPoolI, err := k.UnmarshalPool(bzCfmmPoolI)
			suite.Require().NoError(err)

			// Unmarshal bzCfmmPoolI as PoolI
			unmarshalBzCfmmPoolIAsPoolI, err := k.UnmarshalPoolLegacy(bzCfmmPoolI)
			suite.Require().NoError(err)

			// bzCfmmPoolI as CFMMPoolI equals bzCfmmPoolI as PoolI
			suite.Require().Equal(unmarshalBzCfmmPoolIAsCfmmPoolI, unmarshalBzCfmmPoolIAsPoolI)

			// All unmarshalled combinations are equal.
			suite.Require().Equal(unmarshalBzPoolIAsCfmmPoolI, unmarshalBzCfmmPoolIAsCfmmPoolI)
		})
	}
}
