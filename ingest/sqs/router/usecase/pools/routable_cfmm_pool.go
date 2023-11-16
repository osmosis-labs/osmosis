package pools

import (
	"encoding/json"
	"fmt"
	"reflect"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager"

	gammtypes "github.com/osmosis-labs/osmosis/v20/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

var _ domain.RoutablePool = &routableCFMMPoolImpl{}

type routableCFMMPoolImpl struct {
	domain.PoolI
	TokenOutDenom string       "json:\"token_out_denom\""
	TakerFee      osmomath.Dec "json:\"taker_fee\""
}

// MarshalJSON implements domain.RoutablePool.
func (r *routableCFMMPoolImpl) MarshalJSON() ([]byte, error) {
	var (
		routablePoolSerialized routableSerializedPool
		err                    error
	)
	bz, err := json.Marshal(r.PoolI)
	if err != nil {
		return nil, err
	}

	routablePoolSerialized.PoolData = bz
	routablePoolSerialized.TokenOutDenom = r.TokenOutDenom
	routablePoolSerialized.TakerFee = r.TakerFee

	return json.Marshal(routablePoolSerialized)
}

// UnmarshalJSON implements domain.RoutablePool.
func (r *routableCFMMPoolImpl) UnmarshalJSON(data []byte) error {
	var routablePoolSerialized routableSerializedPool
	err := json.Unmarshal(data, &routablePoolSerialized)
	if err != nil {
		return err
	}

	r.PoolI = &domain.PoolWrapper{}

	err = json.Unmarshal(routablePoolSerialized.PoolData, r.PoolI)
	if err != nil {
		return err
	}

	r.TokenOutDenom = routablePoolSerialized.TokenOutDenom
	r.TakerFee = routablePoolSerialized.TakerFee

	return nil
}

// NewRoutablePool creates a new RoutablePool.
func NewRoutablePool(pool domain.PoolI, tokenOutDenom string, takerFee osmomath.Dec) domain.RoutablePool {
	if pool.GetType() == poolmanagertypes.Concentrated {
		return &routableConcentratedPoolImpl{
			PoolI:         pool,
			TokenOutDenom: tokenOutDenom,
			TakerFee:      takerFee,
		}
	}

	if pool.GetType() == poolmanagertypes.CosmWasm {
		return &routableTransmuterPoolImpl{
			PoolI:         pool,
			TokenOutDenom: tokenOutDenom,
			TakerFee:      takerFee,
		}
	}

	return &routableCFMMPoolImpl{
		PoolI:         pool,
		TokenOutDenom: tokenOutDenom,
		TakerFee:      takerFee,
	}
}

// CalculateTokenOutByTokenIn implements RoutablePool.
func (r *routableCFMMPoolImpl) CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (sdk.Coin, error) {
	poolType := r.GetType()

	if poolType != poolmanagertypes.Balancer && poolType != poolmanagertypes.Stableswap {
		return sdk.Coin{}, domain.InvalidPoolTypeError{PoolType: int32(poolType)}
	}

	osmosisPool := r.PoolI.GetUnderlyingPool()

	// Cast to CFMM extension
	cfmmPool, ok := osmosisPool.(gammtypes.CFMMPoolI)
	if !ok {
		return sdk.Coin{}, domain.FailedToCastPoolModelError{
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
func (r *routableCFMMPoolImpl) GetTokenOutDenom() string {
	return r.TokenOutDenom
}

// String implements domain.RoutablePool.
func (r *routableCFMMPoolImpl) String() string {
	return fmt.Sprintf("pool (%d), pool type (%d), pool denoms (%v)", r.PoolI.GetId(), r.PoolI.GetType(), r.PoolI.GetPoolDenoms())
}

// ChargeTakerFee implements domain.RoutablePool.
// Charges the taker fee for the given token in and returns the token in after the fee has been charged.
func (r *routableCFMMPoolImpl) ChargeTakerFeeExactIn(tokenIn sdk.Coin) (tokenInAfterFee sdk.Coin) {
	tokenInAfterTakerFee, _ := poolmanager.CalcTakerFeeExactIn(tokenIn, r.TakerFee)
	return tokenInAfterTakerFee
}

// GetTakerFee implements domain.RoutablePool.
func (r *routableCFMMPoolImpl) GetTakerFee() math.LegacyDec {
	return r.TakerFee
}

// UnsmarshalJSON implements domain.RoutablePool.
func (r *routableCFMMPoolImpl) UnsmarshalJSON(b []byte) error {
	return r.UnsmarshalJSON(b)
}
