package usecase

import (
	"context"
	"time"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/pools"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/route"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

type poolsUseCase struct {
	contextTimeout         time.Duration
	poolsRepository        mvc.PoolsRepository
	redisRepositoryManager mvc.TxManager
}

var _ mvc.PoolsUsecase = &poolsUseCase{}

// NewPoolsUsecase will create a new pools use case object
func NewPoolsUsecase(timeout time.Duration, poolsRepository mvc.PoolsRepository, redisRepositoryManager mvc.TxManager) mvc.PoolsUsecase {
	return &poolsUseCase{
		contextTimeout:         timeout,
		poolsRepository:        poolsRepository,
		redisRepositoryManager: redisRepositoryManager,
	}
}

// GetAllPools returns all pools from the repository.
func (p *poolsUseCase) GetAllPools(ctx context.Context) ([]domain.PoolI, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	pools, err := p.poolsRepository.GetAllPools(ctx)
	if err != nil {
		return nil, err
	}

	return pools, nil
}

// GetRoutesFromCandidates implements mvc.PoolsUsecase.
func (p *poolsUseCase) GetRoutesFromCandidates(ctx context.Context, candidateRoutes route.CandidateRoutes, takerFeeMap domain.TakerFeeMap, tokenInDenom, tokenOutDenom string) ([]route.RouteImpl, error) {
	// Get all pools
	poolsData, err := p.poolsRepository.GetPools(ctx, candidateRoutes.UniquePoolIDs)
	if err != nil {
		return nil, err
	}

	// TODO: refactor get these directl from the pools repository.
	// Get conentrated pools and separately get tick model for them
	concentratedPoolIDs := make([]uint64, 0)
	for _, candidatePool := range poolsData {
		if candidatePool.GetType() == poolmanagertypes.Concentrated {
			concentratedPoolIDs = append(concentratedPoolIDs, candidatePool.GetId())
		}
	}

	// Get tick model for concentrated pools
	tickModelMap, err := p.GetTickModelMap(ctx, concentratedPoolIDs)
	if err != nil {
		return nil, err
	}

	// Convert each candidate route into the actual route with all pool data
	routes := make([]route.RouteImpl, 0, len(candidateRoutes.Routes))
	for _, candidateRoute := range candidateRoutes.Routes {
		previousTokenOutDenom := tokenInDenom
		routablePools := make([]domain.RoutablePool, 0, len(candidateRoute.Pools))
		for _, candidatePool := range candidateRoute.Pools {
			// Get the pool data for routing
			pool, ok := poolsData[candidatePool.ID]
			if !ok {
				return nil, domain.PoolNotFoundError{PoolID: candidatePool.ID}
			}

			// Get taker fee
			takerFee := takerFeeMap.GetTakerFee(previousTokenOutDenom, candidatePool.TokenOutDenom)

			if pool.GetType() == poolmanagertypes.Concentrated {
				// Get tick model for concentrated pool
				tickModel, ok := tickModelMap[pool.GetId()]
				if !ok {
					return nil, domain.ConcentratedTickModelNotSetError{
						PoolId: pool.GetId(),
					}
				}

				if err := pool.SetTickModel(&tickModel); err != nil {
					return nil, err
				}
			}

			// Create routable pool
			routablePools = append(routablePools, pools.NewRoutablePool(pool, candidatePool.TokenOutDenom, takerFee))
		}

		routes = append(routes, route.RouteImpl{
			Pools: routablePools,
		})
	}

	return routes, nil
}

// GetTickModelMap implements mvc.PoolsUsecase.
func (p *poolsUseCase) GetTickModelMap(ctx context.Context, poolIDs []uint64) (map[uint64]domain.TickModel, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	tickModelMap, err := p.poolsRepository.GetTickModelForPools(ctx, poolIDs)
	if err != nil {
		return nil, err
	}

	return tickModelMap, nil
}

// GetPool implements mvc.PoolsUsecase.
func (p *poolsUseCase) GetPool(ctx context.Context, poolID uint64) (domain.PoolI, error) {
	pools, err := p.poolsRepository.GetPools(ctx, map[uint64]struct {
	}{
		poolID: {},
	})

	if err != nil {
		return nil, err
	}

	pool, ok := pools[poolID]
	if !ok {
		return nil, domain.PoolNotFoundError{PoolID: poolID}
	}
	return pool, nil
}
