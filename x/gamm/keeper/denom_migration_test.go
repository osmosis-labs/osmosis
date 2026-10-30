package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/x/gamm/pool-models/balancer"
)

// TestReplaceBalancerPoolDenomUnsafe checks that the denom is renamed in the pool assets and in a
// smooth weight change, that every list stays sorted, and that invalid replacements change nothing.
func (s *KeeperTestSuite) TestReplaceBalancerPoolDenomUnsafe() {
	assets := func(weightA, weightB int64) []balancer.PoolAsset {
		return []balancer.PoolAsset{
			{Token: sdk.NewInt64Coin("aaa", 1_000_000), Weight: osmomath.NewInt(weightA)},
			{Token: sdk.NewInt64Coin("mmm", 2_000_000), Weight: osmomath.NewInt(weightB)},
		}
	}
	funds := sdk.NewCoins(sdk.NewInt64Coin("aaa", 10_000_000), sdk.NewInt64Coin("mmm", 10_000_000), sdk.NewInt64Coin("uosmo", 10_000_000_000))
	poolId := s.prepareCustomBalancerPool(funds, assets(1, 3), balancer.PoolParams{
		SwapFee: defaultSpreadFactor,
		ExitFee: defaultZeroExitFee,
		SmoothWeightChangeParams: &balancer.SmoothWeightChangeParams{
			StartTime:          s.Ctx.BlockTime().Add(time.Hour),
			Duration:           time.Hour,
			InitialPoolWeights: assets(1, 3),
			TargetPoolWeights:  assets(3, 1),
		},
	})
	before, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolId)
	s.Require().NoError(err)

	for name, args := range map[string][2]string{
		"same denom":        {"aaa", "aaa"},
		"empty new denom":   {"aaa", ""},
		"old denom missing": {"zzz", "nnn"},
		"new denom present": {"aaa", "mmm"},
	} {
		s.Require().Error(s.App.GAMMKeeper.ReplaceBalancerPoolDenomUnsafe(s.Ctx, poolId, args[0], args[1]), name)
	}
	unchanged, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolId)
	s.Require().NoError(err)
	s.Require().Equal(before, unchanged)

	// "zzz" sorts after "mmm", so the renamed asset moves to the end of every list.
	s.Require().NoError(s.App.GAMMKeeper.ReplaceBalancerPoolDenomUnsafe(s.Ctx, poolId, "aaa", "zzz"))
	pool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolId)
	s.Require().NoError(err)
	balancerPool := pool.(*balancer.Pool)
	change := balancerPool.PoolParams.SmoothWeightChangeParams
	for _, list := range [][]balancer.PoolAsset{balancerPool.PoolAssets, change.InitialPoolWeights, change.TargetPoolWeights} {
		s.Require().Equal([]string{"mmm", "zzz"}, []string{list[0].Token.Denom, list[1].Token.Denom})
	}
	s.Require().Equal(osmomath.NewInt(1_000_000), balancerPool.PoolAssets[1].Token.Amount)
	s.Require().Equal(before.(*balancer.Pool).PoolAssets[0].Weight, balancerPool.PoolAssets[1].Weight)
	s.Require().Equal(before.GetTotalShares(), pool.GetTotalShares())
}
