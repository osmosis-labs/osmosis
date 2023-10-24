package usecase

import (
	"strings"

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
func (r *routeImpl) CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (tokenOut sdk.Coin, err error) {
	for _, pool := range r.pools {
		tokenOut, err = pool.CalculateTokenOutByTokenIn(tokenIn)
		if err != nil {
			return sdk.Coin{}, err
		}

		tokenIn = tokenOut
	}

	return tokenOut, nil
}

// String implements domain.Route.
func (r *routeImpl) String() string {
	var strBuilder strings.Builder
	for _, pool := range r.pools {
		_, err := strBuilder.WriteString(pool.String())
		if err != nil {
			panic(err)
		}
	}

	return strBuilder.String()
}

// GetTokenOutDenom implements domain.Route.
// Returns token out denom of the last pool in the route.
// If route is empty, returns empty string.
func (r *routeImpl) GetTokenOutDenom() string {
	if len(r.pools) == 0 {
		return ""
	}

	return r.pools[len(r.pools)-1].GetTokenOutDenom()
}
