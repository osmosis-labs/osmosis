package usecase

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

var _ domain.Route = &routeImpl{}

type routeImpl struct {
	pools []domain.RoutablePool
}

// GetPools implements Route.
func (r *routeImpl) GetPools() []domain.RoutablePool {
	return r.pools
}

func (r routeImpl) DeepCopy() domain.Route {
	poolsCopy := make([]domain.RoutablePool, len(r.pools))
	copy(poolsCopy, r.pools)
	return &routeImpl{
		pools: poolsCopy,
	}
}

func (r *routeImpl) AddPool(pool domain.PoolI, tokenOutDenom string) {
	routablePool := &routablePoolImpl{
		PoolI:         pool,
		tokenOutDenom: tokenOutDenom,
	}
	r.pools = append(r.pools, routablePool)
}

// CalculateTokenOutByTokenIn implements Route.
func (r *routeImpl) CalculateTokenOutByTokenIn(tokenIn sdk.Coin, tokenOutDenom string) (tokenOut sdk.Coin, err error) {
	for _, pool := range r.pools {
		tokenOut, err = pool.CalculateTokenOutByTokenIn(tokenIn)
		if err != nil {
			return sdk.Coin{}, err
		}

		tokenIn = tokenOut
	}

	return tokenOut, nil
}
