package pools

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	cwpoolmodel "github.com/osmosis-labs/osmosis/v21/x/cosmwasmpool/model"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

var _ domain.RoutablePool = &routableTransmuterPoolImpl{}

type routableTransmuterPoolImpl struct {
	ChainPool     *cwpoolmodel.CosmWasmPool "json:\"pool\""
	Balances      sdk.Coins                 "json:\"balances\""
	TokenOutDenom string                    "json:\"token_out_denom\""
	TakerFee      osmomath.Dec              "json:\"taker_fee\""
	SpreadFactor  osmomath.Dec              "json:\"spread_factor\""
}

// GetId implements domain.RoutablePool.
func (r *routableTransmuterPoolImpl) GetId() uint64 {
	return r.ChainPool.PoolId
}

// GetPoolDenoms implements domain.RoutablePool.
func (r *routableTransmuterPoolImpl) GetPoolDenoms() []string {
	return osmoutils.CoinsDenoms(r.Balances)
}

// GetType implements domain.RoutablePool.
func (*routableTransmuterPoolImpl) GetType() poolmanagertypes.PoolType {
	return poolmanagertypes.CosmWasm
}

// GetSpreadFactor implements domain.RoutablePool.
func (r *routableTransmuterPoolImpl) GetSpreadFactor() math.LegacyDec {
	return r.SpreadFactor
}

// CalculateTokenOutByTokenIn implements domain.RoutablePool.
// It calculates the amount of token out given the amount of token in for a transmuter pool.
// Transmuter pool allows no slippage swaps. It just returns the same amount of token out as token in
// Returns error if:
// - the underlying chain pool set on the routable pool is not of transmuter type
// - the token in amount is greater than the balance of the token in
// - the token in amount is greater than the balance of the token out
func (r *routableTransmuterPoolImpl) CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (sdk.Coin, error) {
	poolType := r.GetType()

	// Esnure that the pool is concentrated
	if poolType != poolmanagertypes.CosmWasm {
		return sdk.Coin{}, domain.InvalidPoolTypeError{PoolType: int32(poolType)}
	}

	balances := r.Balances

	// Validate token in balance
	if err := validateBalance(tokenIn.Amount, balances, tokenIn.Denom); err != nil {
		return sdk.Coin{}, err
	}

	// Validate token out balance
	if err := validateBalance(tokenIn.Amount, balances, r.TokenOutDenom); err != nil {
		return sdk.Coin{}, err
	}

	// No slippage swaps - just return the same amount of token out as token in
	// as long as there is enough liquidity in the pool.
	return sdk.NewCoin(r.TokenOutDenom, tokenIn.Amount), nil
}

// GetTokenOutDenom implements RoutablePool.
func (r *routableTransmuterPoolImpl) GetTokenOutDenom() string {
	return r.TokenOutDenom
}

// String implements domain.RoutablePool.
func (r *routableTransmuterPoolImpl) String() string {
	return fmt.Sprintf("pool (%d), pool type (%d), pool denoms (%v), token out (%s)", r.ChainPool.PoolId, poolmanagertypes.CosmWasm, r.GetPoolDenoms(), r.TokenOutDenom)
}

// ChargeTakerFeeExactIn implements domain.RoutablePool.
// Returns tokenInAmount and does not charge any fee for transmuter pools.
func (r *routableTransmuterPoolImpl) ChargeTakerFeeExactIn(tokenIn sdk.Coin) (inAmountAfterFee sdk.Coin) {
	return tokenIn
}

// validateBalance validates that the balance of the denom to validate is greater than the token in amount.
// Returns nil on success, error otherwise.
func validateBalance(tokenInAmount osmomath.Int, balances sdk.Coins, denomToValidate string) error {
	balanceToValidate := balances.AmountOf(denomToValidate)
	if tokenInAmount.GT(balanceToValidate) {
		return domain.TransmuterInsufficientBalanceError{
			Denom:         denomToValidate,
			BalanceAmount: balanceToValidate.String(),
			Amount:        tokenInAmount.String(),
		}
	}

	return nil
}

// GetTakerFee implements domain.RoutablePool.
func (r *routableTransmuterPoolImpl) GetTakerFee() math.LegacyDec {
	return r.TakerFee
}

// SetTokenOutDenom implements domain.RoutablePool.
func (r *routableTransmuterPoolImpl) SetTokenOutDenom(tokenOutDenom string) {
	r.TokenOutDenom = tokenOutDenom
}
