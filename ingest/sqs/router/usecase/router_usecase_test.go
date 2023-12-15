package usecase_test

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/cache"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/log"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/route"
	"github.com/osmosis-labs/osmosis/v21/x/gamm/pool-models/balancer"
)

// Tests the call to handleRoutes by mocking the router repository and pools use case
// with relevant data.
func (s *RouterTestSuite) TestHandleRoutes() {
	const (
		defaultTimeoutDuration = time.Second * 10

		tokenInDenom  = "uosmo"
		tokenOutDenom = "uion"

		minOsmoLiquidity = 10000 * usecase.OsmoPrecisionMultiplier
	)

	// Create test balancer pool

	balancerCoins := sdk.NewCoins(
		sdk.NewCoin(tokenInDenom, sdk.NewInt(1000000000000000000)),
		sdk.NewCoin(tokenOutDenom, sdk.NewInt(1000000000000000000)),
	)

	balancerPoolID := s.PrepareBalancerPoolWithCoins(balancerCoins...)
	balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, balancerPoolID)
	s.Require().NoError(err)

	defaultPool := &domain.PoolWrapper{
		ChainModel: balancerPool,
		SQSModel: domain.SQSPool{
			TotalValueLockedUSDC: osmomath.NewInt(int64(minOsmoLiquidity + 1)),
			PoolDenoms:           []string{tokenInDenom, tokenOutDenom},
			Balances:             balancerCoins,
			SpreadFactor:         DefaultSpreadFactor,
		},
	}

	var (
		defaultRoute = WithCandidateRoutePools(
			EmptyCandidateRoute,
			[]route.CandidatePool{
				{
					ID:            defaultPool.GetId(),
					TokenOutDenom: tokenOutDenom,
				},
			},
		)

		defaultSinglePools = []domain.PoolI{defaultPool}

		singleDefaultRoutes = route.CandidateRoutes{
			Routes: []route.CandidateRoute{defaultRoute},
			UniquePoolIDs: map[uint64]struct{}{
				defaultPool.GetId(): {},
			},
		}

		emptyPools = []domain.PoolI{}

		emptyRoutes = route.CandidateRoutes{}

		defaultRouterConfig = domain.RouterConfig{
			// Only these config values are relevant for this test
			// for searching for routes when none were present in cache.
			MaxPoolsPerRoute: 4,
			MaxRoutes:        4,

			// These configs are not relevant for this test.
			PreferredPoolIDs:          []uint64{},
			MaxSplitIterations:        10,
			MinOSMOLiquidity:          minOsmoLiquidity,
			RouteUpdateHeightInterval: 10,
		}
	)

	testCases := []struct {
		name string

		repositoryRoutes route.CandidateRoutes
		repositoryPools  []domain.PoolI
		takerFeeMap      domain.TakerFeeMap
		isCacheDisabled  bool

		expectedCandidateRoutes route.CandidateRoutes

		expectedError error
	}{
		{
			name: "routes in cache -> use them",

			repositoryRoutes: singleDefaultRoutes,
			repositoryPools:  emptyPools,

			expectedCandidateRoutes: singleDefaultRoutes,
		},
		{
			name: "cache is disabled in config -> recomputes routes despite having available in cache",

			repositoryRoutes: singleDefaultRoutes,
			repositoryPools:  emptyPools,
			isCacheDisabled:  true,

			expectedCandidateRoutes: emptyRoutes,
		},
		{
			name: "no routes in cache but relevant pools in store -> recomputes routes",

			repositoryRoutes: emptyRoutes,
			repositoryPools:  defaultSinglePools,

			expectedCandidateRoutes: singleDefaultRoutes,
		},
		{
			name: "no routes in cache and no relevant pools in store -> returns no routes",

			repositoryRoutes: emptyRoutes,
			repositoryPools:  emptyPools,

			expectedCandidateRoutes: emptyRoutes,
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
				Routes: map[domain.DenomPair]route.CandidateRoutes{
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

			routerUseCase := usecase.NewRouterUsecase(defaultTimeoutDuration, routerRepositoryMock, poolsUseCaseMock, domain.RouterConfig{
				RouteCacheEnabled: !tc.isCacheDisabled,
			}, &log.NoOpLogger{}, cache.New())

			routerUseCaseImpl, ok := routerUseCase.(*usecase.RouterUseCaseImpl)
			s.Require().True(ok)

			// Initialize router
			router := usecase.NewRouter(defaultRouterConfig.PreferredPoolIDs, defaultRouterConfig.MaxPoolsPerRoute, defaultRouterConfig.MaxRoutes, defaultRouterConfig.MaxSplitRoutes, defaultRouterConfig.MaxSplitIterations, defaultRouterConfig.MaxSplitIterations, &log.NoOpLogger{})
			router = usecase.WithSortedPools(router, poolsUseCaseMock.Pools)

			// System under test
			ctx := context.Background()
			actualCandidateRoutes, err := routerUseCaseImpl.HandleRoutes(ctx, router, tokenInDenom, tokenOutDenom)

			if tc.expectedError != nil {
				s.Require().EqualError(err, tc.expectedError.Error())
				s.Require().Len(actualCandidateRoutes, 0)
				return
			}

			s.Require().NoError(err)

			// Pre-set routes should be returned.

			s.Require().Equal(len(tc.expectedCandidateRoutes.Routes), len(actualCandidateRoutes.Routes))
			for i, route := range actualCandidateRoutes.Routes {
				s.Require().Equal(tc.expectedCandidateRoutes.Routes[i], route)
			}

			// For the case where the cache is disabled, the expected routes in cache
			// will be the same as the original routes in the repository.
			if tc.isCacheDisabled {
				tc.expectedCandidateRoutes = tc.repositoryRoutes
			}

			// Check that router repository was updated
			s.Require().Equal(tc.expectedCandidateRoutes, routerRepositoryMock.Routes[domain.DenomPair{Denom0: tokenOutDenom, Denom1: tokenInDenom}])
		})
	}
}

// Tests that routes that overlap in pools IDs get filtered out.
// Tests that the order of the routes is in decreasing priority.
// That is, if routes A and B overlap where A comes before B, then B is filtered out.
// Additionally, tests that overlapping within the same route has no effect on filtering.
// Lastly, validates that if a route overlaps with subsequent routes in the list but gets filtered out,
// then subesequent routes are not affected by filtering.
func (s *RouterTestSuite) TestFilterDuplicatePoolIDRoutes() {
	var (
		deafaultPool = &mocks.MockRoutablePool{ID: defaultPoolID}

		otherPool = &mocks.MockRoutablePool{ID: defaultPoolID + 1}

		defaultSingleRoute = WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
			deafaultPool,
		})
	)

	tests := map[string]struct {
		routes []route.RouteImpl

		expectedRoutes []route.RouteImpl
	}{
		"empty routes": {
			routes:         []route.RouteImpl{},
			expectedRoutes: []route.RouteImpl{},
		},

		"single route single pool": {
			routes: []route.RouteImpl{
				defaultSingleRoute,
			},

			expectedRoutes: []route.RouteImpl{
				defaultSingleRoute,
			},
		},

		"single route two different pools": {
			routes: []route.RouteImpl{
				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					deafaultPool,
					otherPool,
				}),
			},

			expectedRoutes: []route.RouteImpl{
				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					deafaultPool,
					otherPool,
				}),
			},
		},

		// Note that filtering only happens if pool ID duplicated across different routes.
		// Duplicate pool IDs within the same route are filtered out at a different step
		// in the router logic.
		"single route two same pools (have no effect on filtering)": {
			routes: []route.RouteImpl{
				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					deafaultPool,
					deafaultPool,
				}),
			},

			expectedRoutes: []route.RouteImpl{
				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					deafaultPool,
					deafaultPool,
				}),
			},
		},

		"two single hop routes and no duplicates": {
			routes: []route.RouteImpl{
				defaultSingleRoute,

				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					otherPool,
				}),
			},

			expectedRoutes: []route.RouteImpl{
				defaultSingleRoute,

				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					otherPool,
				}),
			},
		},

		"two single hop routes with duplicates (second filtered)": {
			routes: []route.RouteImpl{
				defaultSingleRoute,

				defaultSingleRoute,
			},

			expectedRoutes: []route.RouteImpl{
				defaultSingleRoute,
			},
		},

		"three route. first and second overlap. second and third overlap. second is filtered out but not third": {
			routes: []route.RouteImpl{
				defaultSingleRoute,

				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					deafaultPool, // first and second overlap
					otherPool,    // second and third overlap
				}),

				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					otherPool,
				}),
			},

			expectedRoutes: []route.RouteImpl{
				defaultSingleRoute,

				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					otherPool,
				}),
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {

			actualRoutes := usecase.FilterDuplicatePoolIDRoutes(tc.routes)

			s.Require().Equal(len(tc.expectedRoutes), len(actualRoutes))
		})
	}
}

func (s *RouterTestSuite) TestConvertRankedToCandidateRoutes() {

	tests := map[string]struct {
		rankedRoutes []route.RouteImpl

		expectedCandidateRoutes route.CandidateRoutes
	}{
		"empty ranked routes": {
			rankedRoutes: []route.RouteImpl{},

			expectedCandidateRoutes: route.CandidateRoutes{
				Routes:        []route.CandidateRoute{},
				UniquePoolIDs: map[uint64]struct{}{},
			},
		},
		"single route": {
			rankedRoutes: []route.RouteImpl{
				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithPoolID(mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultMockPool, DenomOne), &balancer.Pool{}), defaultPoolID),
				}),
			},

			expectedCandidateRoutes: route.CandidateRoutes{
				Routes: []route.CandidateRoute{
					WithCandidateRoutePools(route.CandidateRoute{}, []route.CandidatePool{
						{
							ID:            defaultPoolID,
							TokenOutDenom: DenomOne,
						},
					}),
				},
				UniquePoolIDs: map[uint64]struct{}{
					defaultPoolID: {},
				},
			},
		},
		"two routes": {
			rankedRoutes: []route.RouteImpl{
				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithPoolID(mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultMockPool, DenomOne), &balancer.Pool{}), defaultPoolID),
				}),
				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithPoolID(mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultMockPool, DenomOne), &balancer.Pool{}), defaultPoolID+1),
				}),
			},

			expectedCandidateRoutes: route.CandidateRoutes{
				Routes: []route.CandidateRoute{
					WithCandidateRoutePools(route.CandidateRoute{}, []route.CandidatePool{
						{
							ID:            defaultPoolID,
							TokenOutDenom: DenomOne,
						},
					}),
					WithCandidateRoutePools(route.CandidateRoute{}, []route.CandidatePool{
						{
							ID:            defaultPoolID + 1,
							TokenOutDenom: DenomOne,
						},
					}),
				},
				UniquePoolIDs: map[uint64]struct{}{
					defaultPoolID:     {},
					defaultPoolID + 1: {},
				},
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {

			actualCandidateRoutes := usecase.ConvertRankedToCandidateRoutes(tc.rankedRoutes)

			s.Require().Equal(tc.expectedCandidateRoutes, actualCandidateRoutes)
		})
	}
}

// Validates that the ranked route cache functions as expected for optimal quotes.
// This test is set up by focusing on ATOM / OSMO mainnet state pool.
// We restrict the number of routes via config.
//
// As of today there are 3 major ATOM / OSMO pools:
// Pool ID 1: https://app.osmosis.zone/pool/1 (balancer) 0.2% spread factor and 20M of liquidity to date
// Pool ID 1135: https://app.osmosis.zone/pool/1135 (concentrated) 0.2% spread factor and 14M of liquidity to date
// Pool ID 1265: https://app.osmosis.zone/pool/1265 (concentrated) 0.05% spread factor and 224K of liquidity to date
//
// Based on this state, the small amounts of token in should go through pool 1265
// Medium amounts of token in should go through pool 1135
// and large amounts of token in should go through pool 1.
//
// For the purposes of testing cache, we focus on a small amount of token in (1_000_000 uosmo), expecting pool 1265 to be returned.
// We will, however, tweak the cache by test case to force other pools to be returned and ensure that the cache is used.
func (s *RouterTestSuite) TestGetOptimalQuote_Cache() {
	const (
		defaultTokenInDenom  = UOSMO
		defaultTokenOutDenom = ATOM

		// See test description above for details about
		// the pools.
		poolIDOneBalancer      = uint64(1)
		poolID1135Concentrated = uint64(1135)
		poolID1265Concentrated = uint64(1265)
	)

	var (
		defaultAmountIn = osmomath.NewInt(1_000_000)
	)

	tests := map[string]struct {
		preCachedRoutes              route.CandidateRoutes
		cacheOrderOfMagnitudeTokenIn int

		cacheExpiryDuration time.Duration

		amountIn osmomath.Int

		expectedRoutePoolID uint64
	}{
		"cache is not set, computes routes": {
			amountIn: defaultAmountIn,

			// For the default amount in, we expect pool 1265 to be returned.
			// See test description above for details.
			expectedRoutePoolID: poolID1265Concentrated,
		},
		"cache is set to balancer - overwrites computed": {
			amountIn: defaultAmountIn,

			preCachedRoutes: route.CandidateRoutes{
				Routes: []route.CandidateRoute{
					{
						Pools: []route.CandidatePool{
							{
								ID:            poolIDOneBalancer,
								TokenOutDenom: ATOM,
							},
						},
					},
				},
				UniquePoolIDs: map[uint64]struct{}{
					poolIDOneBalancer: {},
				},
			},

			cacheOrderOfMagnitudeTokenIn: osmomath.OrderOfMagnitude(defaultAmountIn.ToLegacyDec()),

			cacheExpiryDuration: time.Hour,

			// We expect balancer because it is cached.
			expectedRoutePoolID: poolIDOneBalancer,
		},
		"cache is set to balancer but for a different order of magnitude - computes new routes": {
			amountIn: defaultAmountIn,

			preCachedRoutes: route.CandidateRoutes{
				Routes: []route.CandidateRoute{
					{
						Pools: []route.CandidatePool{
							{
								ID:            poolIDOneBalancer,
								TokenOutDenom: ATOM,
							},
						},
					},
				},
				UniquePoolIDs: map[uint64]struct{}{
					poolIDOneBalancer: {},
				},
			},

			// Note that we multiply the order of magnitude by 10 so cache is not applied for this amount in.
			cacheOrderOfMagnitudeTokenIn: osmomath.OrderOfMagnitude(defaultAmountIn.ToLegacyDec().MulInt64(10)),

			cacheExpiryDuration: time.Hour,

			// We expect pool 1265 because the cache is not applied.
			expectedRoutePoolID: poolID1265Concentrated,
		},
		"cache is expired - overwrites computed": {
			amountIn: defaultAmountIn,

			preCachedRoutes: route.CandidateRoutes{
				Routes: []route.CandidateRoute{
					{
						Pools: []route.CandidatePool{
							{
								ID:            poolIDOneBalancer,
								TokenOutDenom: ATOM,
							},
						},
					},
				},
				UniquePoolIDs: map[uint64]struct{}{
					poolIDOneBalancer: {},
				},
			},

			cacheOrderOfMagnitudeTokenIn: osmomath.OrderOfMagnitude(defaultAmountIn.ToLegacyDec()),

			// Note: we rely on the fact that the it takes more than 1 nanosecond from the test set up to
			// test execution.
			cacheExpiryDuration: time.Nanosecond,

			// We expect pool 1265 because the cache with balancer pool expires.
			expectedRoutePoolID: poolID1265Concentrated,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			// Setup router config
			config := defaultRouterConfig
			// Note that we set one max route for ease of testing caching specifically.
			config.MaxRoutes = 1

			// Setup mainnet router
			router, tickMap, takerFeeMap := s.setupMainnetRouter(config)

			rankedRouteCache := cache.New()

			if len(tc.preCachedRoutes.Routes) > 0 {
				rankedRouteCache.Set(usecase.FormatRankedRouteCacheKey(defaultTokenInDenom, defaultTokenOutDenom, tc.cacheOrderOfMagnitudeTokenIn), tc.preCachedRoutes, tc.cacheExpiryDuration)
			}

			// Mock router use case.
			routerUsecase, _ := s.setupRouterAndPoolsUsecase(router, defaultTokenInDenom, defaultTokenOutDenom, tickMap, takerFeeMap, rankedRouteCache)

			// System under test
			quote, err := routerUsecase.GetOptimalQuote(context.Background(), sdk.NewCoin(defaultTokenInDenom, tc.amountIn), defaultTokenOutDenom)

			// We only validate that error does not occur without actually validating the quote.
			s.Require().NoError(err)

			// By construction, this test always expects 1 route
			quoteRoutes := quote.GetRoute()
			s.Require().Len(quoteRoutes, 1)

			// By construction, this test always expects 1 pool
			routePools := quoteRoutes[0].GetPools()
			s.Require().Len(routePools, 1)

			// Validate that the pool ID is the expected one
			s.Require().Equal(tc.expectedRoutePoolID, routePools[0].GetId())

			// Validate that the quote is not nil
			s.Require().NotNil(quote.GetAmountOut())
		})
	}
}
