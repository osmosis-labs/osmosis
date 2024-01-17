package pools

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"

	"github.com/osmosis-labs/osmosis/v21/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"

	"github.com/osmosis-labs/osmosis/v21/x/gamm/pool-models/balancer"
)

var _ domain.RoutablePool = &routableBalancerPoolImpl{}

type routableBalancerPoolImpl struct {
	ChainPool     *balancer.Pool "json:\"pool\""
	TokenOutDenom string         "json:\"token_out_denom\""
	TakerFee      osmomath.Dec   "json:\"taker_fee\""
}

// CalculateTokenOutByTokenIn implements RoutablePool.
func (r *routableBalancerPoolImpl) CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (sdk.Coin, error) {
	tokenOut, err := r.ChainPool.CalcOutAmtGivenIn(sdk.Context{}, sdk.NewCoins(tokenIn), r.TokenOutDenom, r.GetSpreadFactor())
	if err != nil {
		return sdk.Coin{}, err
	}

	return tokenOut, nil
}

// GetTokenOutDenom implements RoutablePool.
func (r *routableBalancerPoolImpl) GetTokenOutDenom() string {
	return r.TokenOutDenom
}

// String implements domain.RoutablePool.
func (r *routableBalancerPoolImpl) String() string {
	return fmt.Sprintf("pool (%d), pool type (%d), pool denoms (%v), token out (%s)", r.ChainPool.Id, poolmanagertypes.Balancer, r.ChainPool.GetPoolDenoms(sdk.Context{}), r.TokenOutDenom)
}

// ChargeTakerFee implements domain.RoutablePool.
// Charges the taker fee for the given token in and returns the token in after the fee has been charged.
func (r *routableBalancerPoolImpl) ChargeTakerFeeExactIn(tokenIn sdk.Coin) (tokenInAfterFee sdk.Coin) {
	tokenInAfterTakerFee, _ := poolmanager.CalcTakerFeeExactIn(tokenIn, r.TakerFee)
	return tokenInAfterTakerFee
}

// GetTakerFee implements domain.RoutablePool.
func (r *routableBalancerPoolImpl) GetTakerFee() math.LegacyDec {
	return r.TakerFee
}

// SetTokenOutDenom implements domain.RoutablePool.
func (r *routableBalancerPoolImpl) SetTokenOutDenom(tokenOutDenom string) {
	r.TokenOutDenom = tokenOutDenom
}

// GetSpreadFactor implements domain.RoutablePool.
func (r *routableBalancerPoolImpl) GetSpreadFactor() math.LegacyDec {
	return r.ChainPool.GetSpreadFactor(sdk.Context{})
}

// GetId implements domain.RoutablePool.
func (r *routableBalancerPoolImpl) GetId() uint64 {
	return r.ChainPool.Id
}

// GetPoolDenoms implements domain.RoutablePool.
func (r *routableBalancerPoolImpl) GetPoolDenoms() []string {
	return r.ChainPool.GetPoolDenoms(sdk.Context{})
}

// GetType implements domain.RoutablePool.
func (*routableBalancerPoolImpl) GetType() poolmanagertypes.PoolType {
	return poolmanagertypes.Balancer
}
