package apptesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/stableswap"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

const (
	BAR   = "bar"
	BAZ   = "baz"
	FOO   = "foo"
	UOSMO = "uosmo"
	STAKE = "stake"
)

var DefaultAcctFunds sdk.Coins = sdk.NewCoins(
	sdk.NewCoin(UOSMO, sdk.NewInt(10000000000)),
	sdk.NewCoin(FOO, sdk.NewInt(10000000000)),
	sdk.NewCoin(BAR, sdk.NewInt(10000000000)),
	sdk.NewCoin(BAZ, sdk.NewInt(10000000000)),
)

var DefaultPoolAssets = []balancer.PoolAsset{
	{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin(FOO, sdk.NewInt(5000000)),
	},
	{
		Weight: sdk.NewInt(200),
		Token:  sdk.NewCoin(BAR, sdk.NewInt(5000000)),
	},
	{
		Weight: sdk.NewInt(300),
		Token:  sdk.NewCoin(BAZ, sdk.NewInt(5000000)),
	},
	{
		Weight: sdk.NewInt(400),
		Token:  sdk.NewCoin(UOSMO, sdk.NewInt(5000000)),
	},
}

var DefaultStableswapLiquidity = sdk.NewCoins(
	sdk.NewCoin(FOO, sdk.NewInt(10000000)),
	sdk.NewCoin(BAR, sdk.NewInt(10000000)),
	sdk.NewCoin(BAZ, sdk.NewInt(10000000)),
)

var ImbalancedStableswapLiquidity = sdk.NewCoins(
	sdk.NewCoin(FOO, sdk.NewInt(10_000_000_000)),
	sdk.NewCoin(BAR, sdk.NewInt(20_000_000_000)),
	sdk.NewCoin(BAZ, sdk.NewInt(30_000_000_000)),
)

// PrepareBalancerPoolWithCoins returns a balancer pool
// consisted of given coins with equal weight.
func (s *KeeperTestHelper) PrepareBalancerPoolWithCoins(coins ...sdk.Coin) uint64 {
	weights := make([]int64, len(coins))
	for i := 0; i < len(coins); i++ {
		weights[i] = 1
	}
	return s.PrepareBalancerPoolWithCoinsAndWeights(coins, weights)
}

// PrepareBalancerPoolWithCoins returns a balancer pool
// PrepareBalancerPoolWithCoinsAndWeights returns a balancer pool
// consisted of given coins with the specified weights.
func (s *KeeperTestHelper) PrepareBalancerPoolWithCoinsAndWeights(coins sdk.Coins, weights []int64) uint64 {
	var poolAssets []balancer.PoolAsset
	for i, coin := range coins {
		poolAsset := balancer.PoolAsset{
			Weight: sdk.NewInt(weights[i]),
			Token:  coin,
		}
		poolAssets = append(poolAssets, poolAsset)
	}

	return s.PrepareCustomBalancerPool(poolAssets, balancer.PoolParams{
		SwapFee: sdk.ZeroDec(),
		ExitFee: sdk.ZeroDec(),
	})
}

// PrepareBalancerPool returns a Balancer pool's pool-ID with pool params set in PrepareBalancerPoolWithPoolParams.
func (s *KeeperTestHelper) PrepareBalancerPool() uint64 {
	poolId := s.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
		SwapFee: sdk.NewDec(0),
		ExitFee: sdk.NewDec(0),
	})

	spotPrice, err := s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, FOO, BAR)
	s.NoError(err)
	s.Equal(sdk.NewDec(2).String(), spotPrice.String())
	spotPrice, err = s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, BAR, BAZ)
	s.NoError(err)
	s.Equal(sdk.NewDecWithPrec(15, 1).String(), spotPrice.String())
	spotPrice, err = s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, BAZ, FOO)
	s.NoError(err)
	oneThird := sdk.NewDec(1).Quo(sdk.NewDec(3))
	sp := oneThird.MulInt(gammtypes.SpotPriceSigFigs).RoundInt().ToDec().QuoInt(gammtypes.SpotPriceSigFigs)
	s.Equal(sp.String(), spotPrice.String())

	return poolId
}

// PrepareMultipleBalancerPools returns X Balancer pool's with X being provided by the user.
func (s *KeeperTestHelper) PrepareMultipleBalancerPools(poolsToCreate uint16) []uint64 {
	var poolIds []uint64
	for i := uint16(0); i < poolsToCreate; i++ {
		poolId := s.PrepareBalancerPool()
		poolIds = append(poolIds, poolId)
	}

	return poolIds
}

func (s *KeeperTestHelper) PrepareBasicStableswapPool() uint64 {
	// Mint some assets to the account.
	s.FundAcc(s.TestAccs[0], DefaultAcctFunds)

	params := stableswap.PoolParams{
		SwapFee: sdk.NewDec(0),
		ExitFee: sdk.NewDec(0),
	}

	msg := stableswap.NewMsgCreateStableswapPool(s.TestAccs[0], params, DefaultStableswapLiquidity, []uint64{}, "")
	poolId, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, msg)
	s.NoError(err)
	return poolId
}

func (s *KeeperTestHelper) PrepareImbalancedStableswapPool() uint64 {
	// Mint some assets to the account.
	s.FundAcc(s.TestAccs[0], ImbalancedStableswapLiquidity)

	params := stableswap.PoolParams{
		SwapFee: sdk.NewDec(0),
		ExitFee: sdk.NewDec(0),
	}

	msg := stableswap.NewMsgCreateStableswapPool(s.TestAccs[0], params, ImbalancedStableswapLiquidity, []uint64{1, 1, 1}, "")
	poolId, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, msg)
	s.NoError(err)
	return poolId
}

// PrepareBalancerPoolWithPoolParams sets up a Balancer pool with poolParams.
// Uses default pool assets.
func (s *KeeperTestHelper) PrepareBalancerPoolWithPoolParams(poolParams balancer.PoolParams) uint64 {
	s.FundAcc(s.TestAccs[0], DefaultAcctFunds)
	return s.PrepareCustomBalancerPool(DefaultPoolAssets, poolParams)
}

// PrepareCustomBalancerPool sets up a Balancer pool with an array of assets and given parameters
func (s *KeeperTestHelper) PrepareCustomBalancerPool(assets []balancer.PoolAsset, params balancer.PoolParams) uint64 {
	// Add coins for pool creation fee + coins needed to mint balances
	fundCoins := sdk.NewCoins(sdk.NewCoin(UOSMO, sdk.NewInt(10000000000)))
	for _, a := range assets {
		fundCoins = fundCoins.Add(a.Token)
	}
	s.FundAcc(s.TestAccs[0], fundCoins)

	msg := balancer.NewMsgCreateBalancerPool(s.TestAccs[0], params, assets, "")
	poolId, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, msg)
	s.NoError(err)
	return poolId
}

// PrepareCustomBalancerPoolFromCoins sets up a Balancer pool with an array of coins and given parameters
// The coins are converted to pool assets where each asset has a weight of 1.
func (s *KeeperTestHelper) PrepareCustomBalancerPoolFromCoins(coins sdk.Coins, params balancer.PoolParams) uint64 {
	var poolAssets []balancer.PoolAsset
	for _, coin := range coins {
		poolAsset := balancer.PoolAsset{
			Weight: sdk.NewInt(1),
			Token:  coin,
		}
		poolAssets = append(poolAssets, poolAsset)
	}

	return s.PrepareCustomBalancerPool(poolAssets, params)
}

// Modify spotprice of a pool to target spotprice
func (s *KeeperTestHelper) ModifySpotPrice(poolID uint64, targetSpotPrice sdk.Dec, baseDenom string) {
	var quoteDenom string
	int64Max := int64(^uint64(0) >> 1)

	s.Require().Positive(targetSpotPrice)
	s.Require().Greater(gammtypes.MaxSpotPrice, targetSpotPrice)
	pool, _ := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolID)
	denoms, err := s.App.GAMMKeeper.GetPoolDenoms(s.Ctx, poolID)
	s.Require().NoError(err)
	if denoms[0] == baseDenom {
		quoteDenom = denoms[1]
	} else {
		quoteDenom = denoms[0]
	}

	amountTrade := s.CalcAmoutOfTokenToGetTargetPrice(s.Ctx, pool, targetSpotPrice, baseDenom, quoteDenom)
	if amountTrade.IsPositive() {
		swapIn := sdk.NewCoins(sdk.NewCoin(quoteDenom, sdk.NewInt(amountTrade.RoundInt64())))
		s.FundAcc(s.TestAccs[0], swapIn)
		msg := gammtypes.MsgSwapExactAmountIn{
			Sender:            s.TestAccs[0].String(),
			Routes:            []poolmanagertypes.SwapAmountInRoute{{PoolId: poolID, TokenOutDenom: baseDenom}},
			TokenIn:           swapIn[0],
			TokenOutMinAmount: sdk.ZeroInt(),
		}

		gammMsgServer := gammkeeper.NewMsgServerImpl(s.App.GAMMKeeper)
		_, err = gammMsgServer.SwapExactAmountIn(sdk.WrapSDKContext(s.Ctx), &msg)
		s.Require().NoError(err)
	} else {
		swapOut := sdk.NewCoins(sdk.NewCoin(quoteDenom, sdk.NewInt(amountTrade.RoundInt64()).Abs()))
		swapFee := pool.GetSwapFee(s.Ctx)
		tokenIn, err := pool.CalcInAmtGivenOut(s.Ctx, swapOut, baseDenom, swapFee)
		s.Require().NoError(err)
		s.FundAcc(s.TestAccs[0], sdk.NewCoins(tokenIn))
		msg := gammtypes.MsgSwapExactAmountOut{
			Sender:           s.TestAccs[0].String(),
			Routes:           []poolmanagertypes.SwapAmountOutRoute{{PoolId: poolID, TokenInDenom: baseDenom}},
			TokenInMaxAmount: sdk.NewInt(int64Max),
			TokenOut:         swapOut[0],
		}

		gammMsgServer := gammkeeper.NewMsgServerImpl(s.App.GAMMKeeper)
		_, err = gammMsgServer.SwapExactAmountOut(sdk.WrapSDKContext(s.Ctx), &msg)
		s.Require().NoError(err)
	}
}

func (s *KeeperTestHelper) RunBasicExit(poolId uint64) {
	shareInAmount := sdk.NewInt(100)
	tokenOutMins := sdk.NewCoins()

	msg := gammtypes.MsgExitPool{
		Sender:        s.TestAccs[0].String(),
		PoolId:        poolId,
		ShareInAmount: shareInAmount,
		TokenOutMins:  tokenOutMins,
	}

	gammMsgServer := gammkeeper.NewMsgServerImpl(s.App.GAMMKeeper)
	_, err := gammMsgServer.ExitPool(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err)
}

func (s *KeeperTestHelper) RunBasicJoin(poolId uint64) {
	pool, _ := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolId)
	denoms, err := s.App.GAMMKeeper.GetPoolDenoms(s.Ctx, poolId)
	s.Require().NoError(err)

	tokenIn := sdk.NewCoins()
	for _, denom := range denoms {
		tokenIn = tokenIn.Add(sdk.NewCoin(denom, sdk.NewInt(10000000)))
	}

	s.FundAcc(s.TestAccs[0], sdk.NewCoins(tokenIn...))

	totalPoolShare := pool.GetTotalShares()
	msg := gammtypes.MsgJoinPool{
		Sender:         s.TestAccs[0].String(),
		PoolId:         poolId,
		ShareOutAmount: totalPoolShare.Quo(sdk.NewInt(100000)),
		TokenInMaxs:    tokenIn,
	}

	gammMsgServer := gammkeeper.NewMsgServerImpl(s.App.GAMMKeeper)
	_, err = gammMsgServer.JoinPool(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err)
}

func (s *KeeperTestHelper) CalcAmoutOfTokenToGetTargetPrice(ctx sdk.Context, pool gammtypes.CFMMPoolI, targetSpotPrice sdk.Dec, baseDenom, quoteDenom string) (amountTrade sdk.Dec) {
	blPool, ok := pool.(*balancer.Pool)
	s.Require().True(ok)
	quoteAsset, _ := blPool.GetPoolAsset(quoteDenom)
	baseAsset, err := blPool.GetPoolAsset(baseDenom)
	s.Require().NoError(err)

	s.Require().NotEqual(baseAsset.Weight, sdk.ZeroInt())
	s.Require().NotEqual(quoteAsset.Weight, sdk.ZeroInt())

	spotPriceNow, err := blPool.SpotPrice(ctx, baseDenom, quoteDenom)
	s.Require().NoError(err)

	// Amount of quote token need to trade to get target spot price
	// AmoutQuoteTokenNeedToTrade = AmoutQuoTokenNow * ((targetSpotPrice/spotPriceNow)^((weight_base/(weight_base + weight_quote))) -1 )

	ratioPrice := targetSpotPrice.Quo(spotPriceNow)
	ratioWeight := (baseAsset.Weight.ToDec()).Quo(baseAsset.Weight.ToDec().Add(quoteAsset.Weight.ToDec()))

	amountTrade = quoteAsset.Token.Amount.ToDec().Mul(osmomath.Pow(ratioPrice, ratioWeight).Sub(sdk.OneDec()))

	return amountTrade
}
