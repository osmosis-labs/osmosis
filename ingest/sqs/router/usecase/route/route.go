package route

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/pools"
)

var _ domain.Route = &RouteImpl{}

type RouteImpl struct {
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
func (r *RouteImpl) PrepareResultPools() []domain.RoutablePool {
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
func (r *RouteImpl) GetPools() []domain.RoutablePool {
	return r.Pools
}

// DeepCopy implements Route.
func (r RouteImpl) DeepCopy() domain.Route {
	poolsCopy := make([]domain.RoutablePool, len(r.Pools))

	for i, pool := range r.Pools {
		tokenOutDenom := pool.GetTokenOutDenom()
		takerFee := pool.GetTakerFee()
		// TODO: pool is actually not deep copied but this should be fine temporarily
		poolsCopy[i] = pools.NewRoutablePool(pool, tokenOutDenom, takerFee)
	}

	return &RouteImpl{
		Pools: poolsCopy,
	}
}

func (r *RouteImpl) AddPool(pool domain.PoolI, tokenOutDenom string, takerFee osmomath.Dec) {
	r.Pools = append(r.Pools, pools.NewRoutablePool(pool, tokenOutDenom, takerFee))
}

// CalculateTokenOutByTokenIn implements Route.
func (r *RouteImpl) CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (tokenOut sdk.Coin, err error) {
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
func (r *RouteImpl) String() string {
	var strBuilder strings.Builder
	for _, pool := range r.Pools {
		_, err := strBuilder.WriteString(fmt.Sprintf("{{%s %s}}", pool.String(), pool.GetTokenOutDenom()))
		if err != nil {
			panic(err)
		}
	}

	return strBuilder.String()
}

// GetTokenOutDenom implements domain.Route.
// Returns token out denom of the last pool in the route.
// If route is empty, returns empty string.
func (r *RouteImpl) GetTokenOutDenom() string {
	if len(r.Pools) == 0 {
		return ""
	}

	return r.Pools[len(r.Pools)-1].GetTokenOutDenom()
}

// Reverse reverses the route, making it become as if the current tokenOutDenom is tokenInDenom.
// Relies on providing the desiredTokenInDenom which must be the tokenOutDenom of the first pool in the route.
// Errors if the token out denom of the previous pool is equal to the token out denom of the current pool.
// Errors if the token out denom of the previous pool is not contained in the pool denoms of the current pool.
// Returns nil on success.
func (r *RouteImpl) Reverse(desiredTokenInDenom string) error {
	previousTokenOut := desiredTokenInDenom

	for i, currentPool := range r.Pools {

		nextTokenOutDenom := currentPool.GetTokenOutDenom()

		// Validate
		if previousTokenOut == nextTokenOutDenom {
			return fmt.Errorf("previous token out denom equals next token out denom (%s), %s, pool index (%d)", previousTokenOut, r.Pools, i)
		}

		currentPoolDenoms := currentPool.GetPoolDenoms()
		if !osmoutils.Contains(currentPoolDenoms, previousTokenOut) {
			return fmt.Errorf("previous token out denom %s not in pool denoms %v, route (%s)", nextTokenOutDenom, currentPoolDenoms, r.Pools)
		}

		currentPool.SetTokenOutDenom(previousTokenOut)

		previousTokenOut = nextTokenOutDenom
	}

	r.Pools = osmoutils.ReverseSlice(r.Pools)

	return nil
}
