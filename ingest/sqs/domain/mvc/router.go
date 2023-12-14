package mvc

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/route"
)

// RouterRepository represent the router's repository contract
type RouterRepository interface {
	GetTakerFee(ctx context.Context, denom0, denom1 string) (osmomath.Dec, error)
	GetAllTakerFees(ctx context.Context) (domain.TakerFeeMap, error)
	SetTakerFee(ctx context.Context, tx Tx, denom0, denom1 string, takerFee osmomath.Dec) error
	// SetRoutesTx sets the routes for the given denoms in the given transaction.
	// Sorts denom0 and denom1 lexicographically before setting the routes.
	// Returns error if the transaction fails.
	SetRoutesTx(ctx context.Context, tx Tx, denom0, denom1 string, routes route.CandidateRoutes) error
	// SetRoutes sets the routes for the given denoms. Creates a new transaction and executes it.
	// Sorts denom0 and denom1 lexicographically before setting the routes.
	// Returns error if the transaction fails.
	SetRoutes(ctx context.Context, denom0, denom1 string, routes route.CandidateRoutes) error
	// GetRoutes returns the routes for the given denoms.
	// Sorts denom0 and denom1 lexicographically before setting the routes.
	// Returns empty slice and no error if no routes are present.
	// Returns error if the routes are not found.
	GetRoutes(ctx context.Context, denom0, denom1 string) (route.CandidateRoutes, error)
}

// RouterUsecase represent the router's usecases
type RouterUsecase interface {
	// GetOptimalQuote returns the optimal quote for the given tokenIn and tokenOutDenom.
	GetOptimalQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (domain.Quote, error)
	// GetBestSingleRouteQuote returns the best single route quote for the given tokenIn and tokenOutDenom.
	GetBestSingleRouteQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (domain.Quote, error)
	// GetCustomQuote returns the custom quote for the given tokenIn, tokenOutDenom and poolIDs.
	// It searches for the route that contains the specified poolIDs in the given order.
	// If such route is not found it returns an error.
	GetCustomQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string, poolIDs []uint64) (domain.Quote, error)
	// GetCandidateRoutes returns the candidate routes for the given tokenIn and tokenOutDenom.
	GetCandidateRoutes(ctx context.Context, tokenInDenom, tokenOutDenom string) (route.CandidateRoutes, error)
	// GetTakerFee returns the taker fee for all token pairs in a pool.
	GetTakerFee(ctx context.Context, poolID uint64) ([]domain.TakerFeeForPair, error)
	// GetCachedCandidateRoutes returns the candidate routes for the given tokenIn and tokenOutDenom from cache.
	// It does not recompute the routes if they are not present in cache.
	// Returns error if cache is disabled.
	GetCachedCandidateRoutes(ctx context.Context, tokenInDenom, tokenOutDenom string) (route.CandidateRoutes, error)
	// StoreRoutes stores all router state in the files locally. Used for debugging.
	StoreRouterStateFiles(ctx context.Context) error
}
