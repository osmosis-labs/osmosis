package pools

import (
	"encoding/json"
	"errors"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

var (
	_ domain.RoutablePool       = &routableResultPoolImpl{}
	_ domain.RoutableResultPool = &routableResultPoolImpl{}
)

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

type auxRoutableResultPoolImpl struct {
	ID            uint64                    "json:\"id\""
	Type          poolmanagertypes.PoolType "json:\"type\""
	Balances      sdk.Coins                 "json:\"balances\""
	SpreadFactor  osmomath.Dec              "json:\"spread_factor\""
	TokenOutDenom string                    "json:\"token_out_denom\""
	TakerFee      osmomath.Dec              "json:\"taker_fee\""
}

// MarshalJSON implements domain.RoutablePool.
func (r *routableResultPoolImpl) MarshalJSON() ([]byte, error) {
	aux := auxRoutableResultPoolImpl{}
	aux.ID = r.ID
	aux.Type = r.Type
	aux.Balances = r.Balances
	aux.SpreadFactor = r.SpreadFactor
	aux.TokenOutDenom = r.TokenOutDenom
	aux.TakerFee = r.TakerFee

	return json.Marshal(aux)
}

// UnmarshalJSON implements domain.RoutablePool.
func (r *routableResultPoolImpl) UnmarshalJSON([]byte) error {
	aux := auxRoutableResultPoolImpl{}
	err := json.Unmarshal([]byte{}, &aux)
	if err != nil {
		return err
	}

	r.ID = aux.ID
	r.Type = aux.Type
	r.Balances = aux.Balances
	r.SpreadFactor = aux.SpreadFactor
	r.TokenOutDenom = aux.TokenOutDenom
	r.TakerFee = aux.TakerFee

	return nil
}

// NewRoutableResultPool returns the new routable result pool with the given parameters.
func NewRoutableResultPool(ID uint64, poolType poolmanagertypes.PoolType, balances sdk.Coins, spreadFactor osmomath.Dec, tokenOutDenom string, takerFee osmomath.Dec) domain.RoutablePool {
	return &routableResultPoolImpl{
		ID:            ID,
		Type:          poolType,
		Balances:      balances,
		SpreadFactor:  spreadFactor,
		TokenOutDenom: tokenOutDenom,
		TakerFee:      takerFee,
	}
}

// GetId implements domain.RoutablePool.
func (r *routableResultPoolImpl) GetId() uint64 {
	return r.ID
}

// GetPoolDenoms implements domain.RoutablePool.
func (r *routableResultPoolImpl) GetPoolDenoms() []string {
	denoms := make([]string, len(r.Balances))
	for i, balance := range r.Balances {
		denoms[i] = balance.Denom
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
func (r *routableResultPoolImpl) GetTotalValueLockedUOSMO() math.Int {
	return osmomath.Int{}
}

// GetType implements domain.RoutablePool.
func (r *routableResultPoolImpl) GetType() poolmanagertypes.PoolType {
	return r.Type
}

// GetUnderlyingPool implements domain.RoutablePool.
func (r *routableResultPoolImpl) GetUnderlyingPool() poolmanagertypes.PoolI {
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
func (r *routableResultPoolImpl) GetTokenOutDenom() string {
	return r.TokenOutDenom
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

// GetBalances implements domain.RoutableResultPool.
func (r *routableResultPoolImpl) GetBalances() sdk.Coins {
	return r.Balances
}
