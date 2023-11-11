package usecase

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/pools"
)

var _ domain.Route = &routeImpl{}

type routeImpl struct {
	Pools []domain.RoutablePool "json:\"pools\""
}

// PrepareResultPools implements domain.Route.
// Strips away unnecessary fields from each pool in the route,
// leaving only the data needed by client
// The following are the list of fields that are returned to the client in each pool:
// - ID
// - Type
// - Balances
// - Spread Factor
// - Token Out Denom
// - Taker Fee
// Note that it mutates the route.
// Returns the resulting pools.
func (r *routeImpl) PrepareResultPools() []domain.RoutablePool {
	for i, pool := range r.Pools {
		sqsModel := pool.GetSQSPoolModel()

		r.Pools[i] = pools.NewRoutableResultPool(
			pool.GetId(),
			pool.GetType(),
			sqsModel.Balances,
			// Note that we cannot get the SpreadFactor method on
			// the CosmWasm pool models as it does not implement it.
			// As a result, we propagate it via SQS model.
			sqsModel.SpreadFactor,
			pool.GetTokenOutDenom(),
			pool.GetTakerFee(),
		)
	}
	return r.Pools
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
	r.Pools = append(r.Pools, pools.NewRoutablePool(pool, tokenOutDenom, takerFee))
}

// CalculateTokenOutByTokenIn implements Route.
func (r *routeImpl) CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (tokenOut sdk.Coin, err error) {
	defer func() {
		// TODO: cover this by test
		if r := recover(); r != nil {
			tokenOut = sdk.Coin{}
			err = fmt.Errorf("error when calculating out by in in route: %v", r)
		}
	}()

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
