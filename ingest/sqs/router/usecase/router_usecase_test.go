package usecase_test

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
)

// Tests the call to handleRoutes by mocking the router repository and pools use case
// with relevant data.
func (s *RouterTestSuite) TestHandleRoutes() {
	const (
		defaultTimeoutDuration = time.Second * 10

		tokenInDenom  = "uosmo"
		tokenOutDenom = "uion"
	)

	// Create test balancer pool
	balancerPoolID := s.PrepareBalancerPoolWithCoins(
		sdk.NewCoin(tokenInDenom, sdk.NewInt(1000000000000000000)),
		sdk.NewCoin(tokenOutDenom, sdk.NewInt(1000000000000000000)),
	)
	balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, balancerPoolID)
	s.Require().NoError(err)

	var (
		defaultTakerFeeMap = domain.TakerFeeMap{
			{Denom0: tokenOutDenom, Denom1: tokenInDenom}: DefaultTakerFee,
		}

		defaultRoute = WithRoutePools(
			EmptyRoute,
			[]domain.RoutablePool{
				mocks.WithDenoms(
					mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultPool, tokenOutDenom), balancerPool),
					[]string{tokenInDenom, tokenOutDenom},
				),
			},
		)

		defaultPool = mocks.WithTokenOutDenom(
			mocks.WithTakerFee(
				mocks.WithDenoms(
					mocks.WithChainPoolModel(
						DefaultPool, balancerPool),
					[]string{tokenInDenom, tokenOutDenom}),
				DefaultTakerFee,
			),
			tokenOutDenom)

		defaultSinglePools = []domain.PoolI{defaultPool}

		singleDefaultRoutes = []domain.Route{defaultRoute}

		emptyPools = []domain.PoolI{}

		emptyRoutes = []domain.Route{}

		defaultRouterConfig = domain.RouterConfig{
			// Only these config values are relevant for this test
			// for searching for routes when none were present in cache.
			MaxPoolsPerRoute: 4,
			MaxRoutes:        4,

			// These configs are not relevant for this test.
			PreferredPoolIDs:          []uint64{},
			MaxSplitIterations:        10,
			MinOSMOLiquidity:          10000,
			RouteUpdateHeightInterval: 10,
		}
	)

	testCases := []struct {
		name string

		repositoryRoutes []domain.Route

		repositoryPools []domain.PoolI

		takerFeeMap domain.TakerFeeMap

		expectedRoutes []domain.Route

		expectedError error
	}{
		{
			name: "routes in cache -> use them",

			repositoryRoutes: singleDefaultRoutes,
			repositoryPools:  emptyPools,
			takerFeeMap:      defaultTakerFeeMap,

			expectedRoutes: singleDefaultRoutes,
		},
		{
			name: "no routes in cache but relevant pools in store -> recomputes routes",

			repositoryRoutes: emptyRoutes,
			repositoryPools:  defaultSinglePools,
			takerFeeMap:      defaultTakerFeeMap,

			expectedRoutes: singleDefaultRoutes,
		},
		{
			name: "no routes in cache and no relevant pools in store -> returns no routes",

			repositoryRoutes: emptyRoutes,
			repositoryPools:  emptyPools,
			takerFeeMap:      defaultTakerFeeMap,

			expectedRoutes: emptyRoutes,
		},
		{
			name: "errro: no taker fees set",

			repositoryRoutes: emptyRoutes,
			repositoryPools:  defaultSinglePools,
			takerFeeMap:      domain.TakerFeeMap{},

			expectedRoutes: emptyRoutes,

			expectedError: domain.TakerFeeNotFoundForDenomPairError{Denom0: tokenOutDenom, Denom1: tokenInDenom},
		},

		// TODO:
		// routes in cache but pools have more optimal -> cache is still used
		// multiple routes in cache -> use them
		// multiple rotues in pools -> use them
		// error in repository -> return error
		// error in storing routes after recomputing -> return error
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {

			routerRepositoryMock := &mocks.RedisRouterRepositoryMock{
				Routes: map[domain.DenomPair][]domain.Route{
					// These are the routes that are stored in cache and returned by the call to GetRoutes.
					{Denom0: tokenOutDenom, Denom1: tokenInDenom}: tc.repositoryRoutes,
				},

				// No need to set taker fees on the mock since they are only relevant when
				// set on the router for this test.
			}

			poolsUseCaseMock := &mocks.PoolsUsecaseMock{
				// These are the pools returned by the call to GetAllPools
				Pools: tc.repositoryPools,
			}

			routerUseCase := usecase.NewRouterUsecase(defaultTimeoutDuration, routerRepositoryMock, poolsUseCaseMock, domain.RouterConfig{}, &log.NoOpLogger{})

			routerUseCaseImpl, ok := routerUseCase.(*usecase.RouterUseCaseImpl)
			s.Require().True(ok)

			// Initialize router
			router := usecase.NewRouter(defaultRouterConfig.PreferredPoolIDs, tc.takerFeeMap, defaultRouterConfig.MaxPoolsPerRoute, defaultRouterConfig.MaxRoutes, defaultRouterConfig.MaxSplitIterations, defaultRouterConfig.MaxSplitIterations, &log.NoOpLogger{})
			router = usecase.WithSortedPools(router, poolsUseCaseMock.Pools)

			// System under test
			ctx := context.Background()
			actualRoutes, err := routerUseCaseImpl.HandleRoutes(ctx, router, tokenInDenom, tokenOutDenom)

			if tc.expectedError != nil {
				s.Require().EqualError(err, tc.expectedError.Error())
				s.Require().Len(actualRoutes, 0)
				return
			}

			s.Require().NoError(err)

			// Pre-set routes should be returned.
			s.Require().Equal(len(tc.expectedRoutes), len(actualRoutes))
			for i, route := range actualRoutes {
				s.Require().Equal(tc.expectedRoutes[i].String(), route.String())
			}

			// Check that router repository was updated
			s.Require().Equal(tc.expectedRoutes, routerRepositoryMock.Routes[domain.DenomPair{Denom0: tokenOutDenom, Denom1: tokenInDenom}])
		})
	}
}
