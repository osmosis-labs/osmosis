package apptesting

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanager "github.com/osmosis-labs/osmosis/v15/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

func (s *KeeperTestHelper) RunBasicSwap(poolId uint64) {
	denoms, err := s.App.PoolManagerKeeper.RouteGetPoolDenoms(s.Ctx, poolId)
	s.Require().NoError(err)

	swapIn := sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(1000)))
	s.FundAcc(s.TestAccs[0], swapIn)

	msg := poolmanagertypes.MsgSwapExactAmountIn{
		Sender:            s.TestAccs[0].String(),
		Routes:            []poolmanagertypes.SwapAmountInRoute{{PoolId: poolId, TokenOutDenom: denoms[1]}},
		TokenIn:           swapIn[0],
		TokenOutMinAmount: sdk.ZeroInt(),
	}

	poolManagerMsgServer := poolmanager.NewMsgServerImpl(s.App.PoolManagerKeeper)
	_, err = poolManagerMsgServer.SwapExactAmountIn(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err)
}

// CreatePoolFromType creates a basic pool of the given type for testing.
func (s *KeeperTestHelper) CreatePoolFromType(poolType poolmanagertypes.PoolType) {
	switch poolType {
	case poolmanagertypes.Balancer:
		s.PrepareBalancerPool()
		return
	case poolmanagertypes.Stableswap:
		s.PrepareBasicStableswapPool()
		return
	case poolmanagertypes.Concentrated:
		s.PrepareConcentratedPool()
		return
	}
}

// CreatePoolFromTypeWithCoins creates a pool with the given type and initialized with the given coins.
func (s *KeeperTestHelper) CreatePoolFromTypeWithCoins(poolType poolmanagertypes.PoolType, coins sdk.Coins) uint64 {
	var poolId uint64
	if poolType == poolmanagertypes.Balancer {
		poolId = s.PrepareBalancerPoolWithCoins(coins...)
	} else if poolType == poolmanagertypes.Concentrated {
		s.Require().Len(coins, 2)
		clPool := s.PrepareConcentratedPoolWithCoins(coins[0].Denom, coins[1].Denom)
		s.CreateFullRangePosition(clPool, coins)
		poolId = clPool.GetId()
	} else {
		s.FailNow(fmt.Sprintf("unsupported pool type for this operation (%s)", poolmanagertypes.PoolType_name[int32(poolType)]))
	}
	return poolId
}
