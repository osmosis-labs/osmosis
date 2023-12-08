package route

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/pools"
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
		r.Pools[i] = pools.NewRoutableResultPool(
			pool.GetId(),
			pool.GetType(),
			pool.GetSpreadFactor(),
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
		// Charge taker fee
		tokenIn = pool.ChargeTakerFeeExactIn(tokenIn)

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
