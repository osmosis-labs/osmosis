package domain

import (
	"context"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

type RoutablePool interface {
	json.Marshaler
	json.Unmarshaler

	PoolI
	GetTokenOutDenom() string
	CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (sdk.Coin, error)
	ChargeTakerFeeExactIn(tokenIn sdk.Coin) (tokenInAfterFee sdk.Coin)

	GetTakerFee() osmomath.Dec

	String() string
}

type RoutableResultPool interface {
	RoutablePool
	GetBalances() sdk.Coins
}

type Route interface {
	json.Unmarshaler
	GetPools() []RoutablePool
	DeepCopy() Route
	AddPool(pool PoolI, tokenOut string, takerFee osmomath.Dec)
	// CalculateTokenOutByTokenIn calculates the token out amount given the token in amount.
	// Returns error if the calculation fails.
	CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (sdk.Coin, error)

	GetTokenOutDenom() string

	// PrepareResultPools strips away unnecessary fields
	// from each pool in the route,
	// leaving only the data needed by client
	// Note that it mutates the route.
	// Returns the resulting pools.
	PrepareResultPools() []RoutablePool

	String() string
}

// RouterRepository represent the router's repository contract
type RouterRepository interface {
	GetTakerFee(ctx context.Context, denom0, denom1 string) (osmomath.Dec, error)
	GetAllTakerFees(ctx context.Context) (TakerFeeMap, error)
	SetTakerFee(ctx context.Context, tx Tx, denom0, denom1 string, takerFee osmomath.Dec) error
	// SetRoutesTx sets the routes for the given denoms in the given transaction.
	// Sorts denom0 and denom1 lexicographically before setting the routes.
	// Returns error if the transaction fails.
	SetRoutesTx(ctx context.Context, tx Tx, denom0, denom1 string, routes []Route) error
	// SetRoutes sets the routes for the given denoms. Creates a new transaction and executes it.
	// Sorts denom0 and denom1 lexicographically before setting the routes.
	// Returns error if the transaction fails.
	SetRoutes(ctx context.Context, denom0, denom1 string, routes []Route) error
	// GetRoutes returns the routes for the given denoms.
	// Sorts denom0 and denom1 lexicographically before setting the routes.
	// Returns empty slice and no error if no routes are present.
	// Returns error if the routes are not found.
	GetRoutes(ctx context.Context, denom0, denom1 string) ([]Route, error)
}

// RouterUsecase represent the router's usecases
type RouterUsecase interface {
	// GetOptimalQuote returns the optimal quote for the given tokenIn and tokenOutDenom.
	GetOptimalQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (Quote, error)
	// GetBestSingleRouteQuote returns the best single route quote for the given tokenIn and tokenOutDenom.
	GetBestSingleRouteQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (Quote, error)
	// GetCandidateRoutes returns the candidate routes for the given tokenIn and tokenOutDenom.
	GetCandidateRoutes(ctx context.Context, tokenInDenom, tokenOutDenom string) ([]Route, error)
}

type SplitRoute interface {
	Route
	GetAmountIn() osmomath.Int
	GetAmountOut() osmomath.Int
}

type Quote interface {
	GetAmountIn() sdk.Coin
	GetAmountOut() osmomath.Int
	GetRoute() []SplitRoute
	GetEffectiveSpreadFactor() osmomath.Dec

	// PrepareResult mutates the quote to prepare
	// it with the data formatted for output to the client.
	PrepareResult() ([]SplitRoute, osmomath.Dec)

	String() string
}

type RouterConfig struct {
	PreferredPoolIDs   []uint64
	MaxPoolsPerRoute   int
	MaxRoutes          int
	MaxSplitIterations int
	// Denominated in OSMO (not uosmo)
	MinOSMOLiquidity          int
	RouteUpdateHeightInterval int64
}

// DenomPair encapsulates a pair of denoms.
// The order of the denoms ius that Denom0 precedes
// Denom1 lexicographically.
type DenomPair struct {
	Denom0 string
	Denom1 string
}

// TakerFeeMap is a map of DenomPair to taker fee.
// It sorts the denoms lexicographically before looking up the taker fee.
type TakerFeeMap map[DenomPair]osmomath.Dec

// Has returns true if the taker fee for the given denoms is found.
// It sorts the denoms lexicographically before looking up the taker fee.
func (tfm TakerFeeMap) Has(denom0, denom1 string) bool {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	_, found := tfm[DenomPair{Denom0: denom0, Denom1: denom1}]
	return found
}

// GetTakerFee returns the taker fee for the given denoms.
// It sorts the denoms lexicographically before looking up the taker fee.
// Returns error if the taker fee is not found.
func (tfm TakerFeeMap) GetTakerFee(denom0, denom1 string) (osmomath.Dec, error) {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	takerFee, found := tfm[DenomPair{Denom0: denom0, Denom1: denom1}]

	if !found {
		return osmomath.Dec{}, TakerFeeNotFoundForDenomPairError{
			Denom0: denom0,
			Denom1: denom1,
		}
	}

	return takerFee, nil
}

// SetTakerFee sets the taker fee for the given denoms.
// It sorts the denoms lexicographically before setting the taker fee.
func (tfm TakerFeeMap) SetTakerFee(denom0, denom1 string, takerFee osmomath.Dec) {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	tfm[DenomPair{Denom0: denom0, Denom1: denom1}] = takerFee
}
