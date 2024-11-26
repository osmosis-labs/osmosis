package apptesting

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	poolmanager "github.com/osmosis-labs/osmosis/v27/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

func (s *KeeperTestHelper) RunBasicSwap(poolId uint64) {
	denoms, err := s.App.PoolManagerKeeper.RouteGetPoolDenoms(s.Ctx, poolId)
	s.Require().NoError(err)

	swapIn := sdk.NewCoins(sdk.NewCoin(denoms[0], osmomath.NewInt(1000)))
	s.FundAcc(s.TestAccs[0], swapIn)

	msg := poolmanagertypes.MsgSwapExactAmountIn{
		Sender:            s.TestAccs[0].String(),
		Routes:            []poolmanagertypes.SwapAmountInRoute{{PoolId: poolId, TokenOutDenom: denoms[1]}},
		TokenIn:           swapIn[0],
		TokenOutMinAmount: osmomath.ZeroInt(),
	}

	poolManagerMsgServer := poolmanager.NewMsgServerImpl(s.App.PoolManagerKeeper)
	_, err = poolManagerMsgServer.SwapExactAmountIn(s.Ctx, &msg)
	s.Require().NoError(err)
}

// CreatePoolFromType creates a basic pool of the given type for testing.
func (s *KeeperTestHelper) CreatePoolFromType(poolType poolmanagertypes.PoolType) uint64 {
	switch poolType {
	case poolmanagertypes.Balancer:
		balancerPoolID := s.PrepareBalancerPool()
		return balancerPoolID
	case poolmanagertypes.Stableswap:
		stableswapPoolID := s.PrepareBasicStableswapPool()
		return stableswapPoolID
	case poolmanagertypes.Concentrated:
		concentratedPool := s.PrepareConcentratedPool()
		return concentratedPool.GetId()
	case poolmanagertypes.CosmWasm:
		cosmwasmPool := s.PrepareCosmWasmPool()
		return cosmwasmPool.GetId()
	default:
		s.FailNow(fmt.Sprintf("unsupported pool type for this operation (%s)", poolmanagertypes.PoolType_name[int32(poolType)]))
	}

	return 0
}

// CreatePoolFromTypeWithCoins creates a pool with the given type and initialized with the given coins.
func (s *KeeperTestHelper) CreatePoolFromTypeWithCoins(poolType poolmanagertypes.PoolType, coins sdk.Coins) uint64 {
	return s.CreatePoolFromTypeWithCoinsAndSpreadFactor(poolType, coins, osmomath.ZeroDec())
}

// CreatePoolFromTypeWithCoinsAndSpreadFactor creates a pool with given type, initialized with the given coins as initial liquidity and spread factor.
func (s *KeeperTestHelper) CreatePoolFromTypeWithCoinsAndSpreadFactor(poolType poolmanagertypes.PoolType, coins sdk.Coins, spreadFactor osmomath.Dec) uint64 {
	switch poolType {
	case poolmanagertypes.Balancer:
		poolId := s.PrepareCustomBalancerPoolFromCoins(coins, balancer.PoolParams{
			SwapFee: spreadFactor,
			ExitFee: osmomath.ZeroDec(),
		})
		return poolId
	case poolmanagertypes.Concentrated:
		s.Require().Len(coins, 2)
		pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], coins[0].Denom, coins[1].Denom, DefaultTickSpacing, spreadFactor)
		s.CreateFullRangePosition(pool, coins)
		return pool.GetId()
	case poolmanagertypes.CosmWasm:
		s.Require().Len(coins, 2)
		pool := s.PrepareCustomTransmuterPool(s.TestAccs[0], []string{coins[0].Denom, coins[1].Denom})
		s.JoinTransmuterPool(s.TestAccs[0], pool.GetId(), coins)
		return pool.GetId()
	default:
		s.FailNow(fmt.Sprintf("unsupported pool type for this operation (%s)", poolmanagertypes.PoolType_name[int32(poolType)]))
	}
	return 0
}
