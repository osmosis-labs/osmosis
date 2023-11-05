package usecase

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

var _ domain.Route = &routeImpl{}

type routeImpl struct {
	Pools []domain.RoutablePool
}

// GetPools implements Route.
func (r *routeImpl) GetPools() []domain.RoutablePool {
	return r.Pools
}

func (r routeImpl) DeepCopy() domain.Route {
	poolsCopy := make([]domain.RoutablePool, len(r.Pools))
	copy(poolsCopy, r.Pools)
	return &routeImpl{
		Pools: poolsCopy,
	}
}

func (r *routeImpl) AddPool(pool domain.PoolI, tokenOutDenom string, takerFee osmomath.Dec) {
	routablePool := NewRoutablePool(pool, tokenOutDenom, takerFee)
	r.Pools = append(r.Pools, routablePool)
}

// CalculateTokenOutByTokenIn implements Route.
func (r *routeImpl) CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (tokenOut sdk.Coin, err error) {
	for _, pool := range r.Pools {
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
	for _, pool := range r.Pools {
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
	if len(r.Pools) == 0 {
		return ""
	}

	return r.Pools[len(r.Pools)-1].GetTokenOutDenom()
}
