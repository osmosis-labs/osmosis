package domain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

type RoutablePool interface {
	PoolI
	GetTokenOutDenom() string
	CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (sdk.Coin, error)

	String() string
}

type Route interface {
	GetPools() []RoutablePool
	DeepCopy() Route
	AddPool(pool PoolI, tokenOut string)
	// CalculateTokenOutByTokenIn calculates the token out amount given the token in amount.
	// Returns error if the calculation fails.
	CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (sdk.Coin, error)

	GetTokenOutDenom() string

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
}

type RouterConfig struct {
	PreferredPoolIDs   []uint64
	MaxPoolsPerRoute   int
	MaxRoutes          int
	MaxSplitIterations int
	// Denominated in OSMO (not uosmo)
	MinOSMOLiquidity int
}
