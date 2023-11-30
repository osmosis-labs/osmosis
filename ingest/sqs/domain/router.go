package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

type RoutablePool interface {
	GetId() uint64

	GetType() poolmanagertypes.PoolType

	GetPoolDenoms() []string

	GetTokenOutDenom() string

	CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (sdk.Coin, error)
	ChargeTakerFeeExactIn(tokenIn sdk.Coin) (tokenInAfterFee sdk.Coin)

	// SetTokenOutDenom sets the token out denom on the routable pool.
	SetTokenOutDenom(tokenOutDenom string)

	GetTakerFee() osmomath.Dec

	GetSpreadFactor() osmomath.Dec

	String() string
}

type RoutableResultPool interface {
	RoutablePool
	GetBalances() sdk.Coins
}

type Route interface {
	GetPools() []RoutablePool
	// AddPool adds pool to route.
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
	MaxSplitRoutes     int
	MaxSplitIterations int
	// Denominated in OSMO (not uosmo)
	MinOSMOLiquidity          int
	RouteUpdateHeightInterval int64
	RouteCacheEnabled         bool
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

var _ json.Marshaler = &TakerFeeMap{}
var _ json.Unmarshaler = &TakerFeeMap{}

// MarshalJSON implements json.Marshaler.
func (tfm TakerFeeMap) MarshalJSON() ([]byte, error) {
	serializedMap := map[string]osmomath.Dec{}
	for key, value := range tfm {
		// Convert DenomPair to a string representation
		keyString := fmt.Sprintf("%s-%s", key.Denom0, key.Denom1)
		serializedMap[keyString] = value
	}

	return json.Marshal(serializedMap)
}

// UnmarshalJSON implements json.Unmarshaler.
func (tfm TakerFeeMap) UnmarshalJSON(data []byte) error {
	var serializedMap map[string]osmomath.Dec
	if err := json.Unmarshal(data, &serializedMap); err != nil {
		return err
	}

	// Convert string keys back to DenomPair
	for keyString, value := range serializedMap {
		parts := strings.Split(keyString, "-")
		if len(parts) != 2 {
			return fmt.Errorf("invalid key format: %s", keyString)
		}
		denomPair := DenomPair{Denom0: parts[0], Denom1: parts[1]}
		(tfm)[denomPair] = value
	}

	return nil
}

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

var DefaultTakerFee = osmomath.MustNewDecFromStr("0.001000000000000000")
