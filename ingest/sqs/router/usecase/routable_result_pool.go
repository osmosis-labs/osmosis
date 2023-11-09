package usecase

import (
	"errors"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

var _ domain.RoutablePool = &routableCFMMPoolImpl{}

// routableResultPoolImpl is a generalized implementation that is returned to the client
// side in quotes. It contains all the relevant pool data needed for Osmosis frontend
type routableResultPoolImpl struct {
	ID            uint64                    "json:\"id\""
	Type          poolmanagertypes.PoolType "json:\"type\""
	Balances      sdk.Coins                 "json:\"balances\""
	SpreadFactor  osmomath.Dec              "json:\"spread_factor\""
	TokenOutDenom string                    "json:\"token_out_denom\""
	TakerFee      osmomath.Dec              "json:\"taker_fee\""
}

// GetId implements domain.RoutablePool.
func (r *routableResultPoolImpl) GetId() uint64 {
	return r.ID
}

// GetPoolDenoms implements domain.RoutablePool.
func (r *routableResultPoolImpl) GetPoolDenoms() []string {
	denoms := make([]string, len(r.Balances))
	for _, balance := range r.Balances {
		denoms = append(denoms, balance.Denom)
	}

	return denoms
}

// GetSQSPoolModel implements domain.RoutablePool.
func (r *routableResultPoolImpl) GetSQSPoolModel() domain.SQSPool {
	return domain.SQSPool{
		Balances:     r.Balances,
		PoolDenoms:   r.GetPoolDenoms(),
		SpreadFactor: r.SpreadFactor,
	}
}

// GetTickModel implements domain.RoutablePool.
func (r *routableResultPoolImpl) GetTickModel() (*domain.TickModel, error) {
	return nil, errors.New("not implemented")
}

// GetTotalValueLockedUOSMO implements domain.RoutablePool.
func (*routableResultPoolImpl) GetTotalValueLockedUOSMO() math.Int {
	return osmomath.Int{}
}

// GetType implements domain.RoutablePool.
func (r *routableResultPoolImpl) GetType() poolmanagertypes.PoolType {
	return r.Type
}

// GetUnderlyingPool implements domain.RoutablePool.
func (*routableResultPoolImpl) GetUnderlyingPool() poolmanagertypes.PoolI {
	return nil
}

// Validate implements domain.RoutablePool.
func (*routableResultPoolImpl) Validate(minUOSMOTVL math.Int) error {
	return nil
}

// CalculateTokenOutByTokenIn implements RoutablePool.
func (r *routableResultPoolImpl) CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (sdk.Coin, error) {
	return sdk.Coin{}, errors.New("not implemented")
}

// GetTokenOutDenom implements RoutablePool.
func (rp *routableResultPoolImpl) GetTokenOutDenom() string {
	return rp.TokenOutDenom
}

// String implements domain.RoutablePool.
func (r *routableResultPoolImpl) String() string {
	return fmt.Sprintf("pool (%d), pool type (%d), pool denoms (%v)", r.GetId(), r.GetType(), r.GetPoolDenoms())
}

// ChargeTakerFee implements domain.RoutablePool.
// Charges the taker fee for the given token in and returns the token in after the fee has been charged.
func (r *routableResultPoolImpl) ChargeTakerFeeExactIn(tokenIn sdk.Coin) (tokenInAfterFee sdk.Coin) {
	tokenInAfterTakerFee, _ := poolmanager.CalcTakerFeeExactIn(tokenIn, r.TakerFee)
	return tokenInAfterTakerFee
}

// GetTakerFee implements domain.RoutablePool.
func (r *routableResultPoolImpl) GetTakerFee() math.LegacyDec {
	return r.TakerFee
}
