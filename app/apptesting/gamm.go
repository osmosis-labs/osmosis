package apptesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v11/x/gamm/types"
)

var DefaultAcctFunds sdk.Coins = sdk.NewCoins(
	sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
	sdk.NewCoin("foo", sdk.NewInt(10000000)),
	sdk.NewCoin("bar", sdk.NewInt(10000000)),
	sdk.NewCoin("baz", sdk.NewInt(10000000)),
)

// PrepareBalancerPoolWithCoins returns a balancer pool
// consisted of given coins with equal weight.
func (s *KeeperTestHelper) PrepareBalancerPoolWithCoins(coins ...sdk.Coin) uint64 {
	var poolAssets []balancer.PoolAsset
	for _, coin := range coins {
		poolAsset := balancer.PoolAsset{
			Weight: sdk.NewInt(1),
			Token:  coin,
		}
		poolAssets = append(poolAssets, poolAsset)
	}

	return s.PrepareBalancerPoolWithPoolAsset(poolAssets)
}

// PrepareBalancerPool returns a Balancer pool's pool-ID with pool params set in PrepareBalancerPoolWithPoolParams.
func (s *KeeperTestHelper) PrepareBalancerPool() uint64 {
	poolId := s.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
		SwapFee: sdk.NewDec(0),
		ExitFee: sdk.NewDec(0),
	})

	spotPrice, err := s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, "foo", "bar")
	s.NoError(err)
	s.Equal(sdk.NewDec(2).String(), spotPrice.String())
	spotPrice, err = s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, "bar", "baz")
	s.NoError(err)
	s.Equal(sdk.NewDecWithPrec(15, 1).String(), spotPrice.String())
	spotPrice, err = s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, "baz", "foo")
	s.NoError(err)
	oneThird := sdk.NewDec(1).Quo(sdk.NewDec(3))
	sp := oneThird.MulInt(gammtypes.SigFigs).RoundInt().ToDec().QuoInt(gammtypes.SigFigs)
	s.Equal(sp.String(), spotPrice.String())

	return poolId
}

// PrepareBalancerPoolWithPoolParams sets up a Balancer pool with poolParams.
func (s *KeeperTestHelper) PrepareBalancerPoolWithPoolParams(poolParams balancer.PoolParams) uint64 {
	// Mint some assets to the account.
	s.FundAcc(s.TestAccs[0], DefaultAcctFunds)

	poolAssets := []balancer.PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
		},
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
		},
		{
			Weight: sdk.NewInt(300),
			Token:  sdk.NewCoin("baz", sdk.NewInt(5000000)),
		},
	}
	msg := balancer.NewMsgCreateBalancerPool(s.TestAccs[0], poolParams, poolAssets, "")
	poolId, err := s.App.GAMMKeeper.CreatePool(s.Ctx, msg)
	s.NoError(err)
	return poolId
}

// PrepareBalancerPoolWithPoolAsset sets up a Balancer pool with an array of assets.
func (s *KeeperTestHelper) PrepareBalancerPoolWithPoolAsset(assets []balancer.PoolAsset) uint64 {
	// Add coins for pool creation fee + coins needed to mint balances
	fundCoins := sdk.Coins{sdk.NewCoin("uosmo", sdk.NewInt(10000000000))}
	for _, a := range assets {
		fundCoins = fundCoins.Add(a.Token)
	}
	s.FundAcc(s.TestAccs[0], fundCoins)

	msg := balancer.NewMsgCreateBalancerPool(s.TestAccs[0], balancer.PoolParams{
		SwapFee: sdk.ZeroDec(),
		ExitFee: sdk.ZeroDec(),
	}, assets, "")
	poolId, err := s.App.GAMMKeeper.CreatePool(s.Ctx, msg)
	s.NoError(err)
	return poolId
}

func (s *KeeperTestHelper) RunBasicSwap(poolId uint64) {
	denoms, err := s.App.GAMMKeeper.GetPoolDenoms(s.Ctx, poolId)
	s.Require().NoError(err)

	swapIn := sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(1000)))
	s.FundAcc(s.TestAccs[0], swapIn)

	msg := gammtypes.MsgSwapExactAmountIn{
		Sender:            string(s.TestAccs[0]),
		Routes:            []gammtypes.SwapAmountInRoute{{PoolId: poolId, TokenOutDenom: denoms[1]}},
		TokenIn:           swapIn[0],
		TokenOutMinAmount: sdk.ZeroInt(),
	}
	// TODO: switch to message
	_, err = s.App.GAMMKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], poolId, msg.TokenIn, denoms[1], msg.TokenOutMinAmount)
	s.Require().NoError(err)
}

func (s *KeeperTestHelper) RunBasicJoinPool(poolId uint64) {
	denoms, err := s.App.GAMMKeeper.GetPoolDenoms(s.Ctx, poolId)
	s.Require().NoError(err)

	for _, denom := range denoms {
		s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(10000000))))
	}

	pool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolId)
	s.Require().NoError(err)
	totalPoolShare := pool.GetTotalShares()
	totalPoolShare.Quo(sdk.NewInt(100000))

	tokenIn, _, err := s.App.GAMMKeeper.JoinPoolNoSwap(s.Ctx, s.TestAccs[0], poolId, totalPoolShare.Quo(sdk.NewInt(100000)), sdk.Coins{})
	s.Require().NoError(err)
	s.FundAcc(s.TestAccs[0], tokenIn)
}
