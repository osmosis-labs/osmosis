package usecase

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"

	gammtypes "github.com/osmosis-labs/osmosis/v20/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

var _ domain.RoutablePool = &routablePoolImpl{}

type routablePoolImpl struct {
	domain.PoolI
	TokenOutDenom string "json:\"token_out_denom\""
}

// NewRoutablePool creates a new RoutablePool.
func NewRoutablePool(pool domain.PoolI, tokenOutDenom string) domain.RoutablePool {
	return &routablePoolImpl{
		PoolI:         pool,
		TokenOutDenom: tokenOutDenom,
	}
}

// CalculateTokenOutByTokenIn implements RoutablePool.
func (r *routablePoolImpl) CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (sdk.Coin, error) {
	poolType := r.GetType()

	if poolType != poolmanagertypes.Balancer {
		return sdk.Coin{}, OnlyBalancerPoolsSupportedError{ActualType: int32(poolType)}
	}

	osmosisPool := r.PoolI.GetUnderlyingPool()

	// Cast to CFMM extension
	cfmmPool, ok := osmosisPool.(gammtypes.CFMMPoolI)
	if !ok {
		return sdk.Coin{}, FailedToCastPoolModelError{
			ExpectedModel: reflect.TypeOf((gammtypes.CFMMPoolI)(nil)).Elem().Name(),
			ActualModel:   reflect.TypeOf(r).Elem().Name(),
		}
	}

	// TODO: remove context from interface as it is unusded
	tokenOut, err := cfmmPool.CalcOutAmtGivenIn(sdk.Context{}, sdk.NewCoins(tokenIn), r.TokenOutDenom, cfmmPool.GetSpreadFactor(sdk.Context{}))
	if err != nil {
		return sdk.Coin{}, err
	}

	return tokenOut, nil
}

// GetTokenOutDenom implements RoutablePool.
func (rp *routablePoolImpl) GetTokenOutDenom() string {
	return rp.TokenOutDenom
}

// String implements domain.RoutablePool.
func (r *routablePoolImpl) String() string {
	return fmt.Sprintf("pool (%d), pool type (%d), pool denoms (%v)", r.PoolI.GetId(), r.PoolI.GetType(), r.PoolI.GetPoolDenoms())
}
