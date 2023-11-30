package usecase_test

import (
	"context"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/coinutil"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/log"
	poolsusecase "github.com/osmosis-labs/osmosis/v21/ingest/sqs/pools/usecase"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase"
	routerusecase "github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/route"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/routertesting"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

const (
	defaultPoolID = uint64(1)
	UOSMO         = "uosmo"
	ATOM          = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	stOSMO        = "ibc/D176154B0C63D1F9C6DCFB4F70349EBF2E2B5A87A05902F57A6AE92B863E9AEC"
	stATOM        = "ibc/C140AFD542AE77BD7DCC83F13FDD8C5E5BB8C4929785E6EC2F4C636F98F17901"
	USDC          = "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"
	USDCaxl       = "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858"
	USDT          = "ibc/4ABBEF4C8926DDDB320AE5188CFD63267ABBCEFC0583E4AE05D6E5AA2401DDAB"
	WBTC          = "ibc/D1542AA8762DB13087D8364F3EA6509FD6F009A34F00426AF9E4F9FA85CBBF1F"
	ETH           = "ibc/EA1D43981D5C9A1C4AAEA9C23BB1D4FA126BA9BC7020A25E0AE4AA841EA25DC5"
	AKT           = "ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4"
	UMEE          = "ibc/67795E528DF67C5606FC20F824EA39A6EF55BA133F4DC79C90A8C47A0901E17C"
	UION          = "uion"
)

// TODO: copy exists in candidate_routes_test.go - share & reuse
var (
	DefaultTakerFee     = osmomath.MustNewDecFromStr("0.002")
	DefaultPoolBalances = sdk.NewCoins(
		sdk.NewCoin(DenomOne, DefaultAmt0),
		sdk.NewCoin(DenomTwo, DefaultAmt1),
	)
	DefaultSpreadFactor = osmomath.MustNewDecFromStr("0.005")

	DefaultMockPool = &mocks.MockRoutablePool{
		ID:                   defaultPoolID,
		Denoms:               []string{DenomOne, DenomTwo},
		TotalValueLockedUSDC: osmomath.NewInt(10),
		PoolType:             poolmanagertypes.Balancer,
		Balances:             DefaultPoolBalances,
		TakerFee:             DefaultTakerFee,
		SpreadFactor:         DefaultSpreadFactor,
	}
	EmptyRoute          = route.RouteImpl{}
	EmptyCandidateRoute = route.CandidateRoute{}

	// Test denoms
	DenomOne   = routertesting.DenomOne
	DenomTwo   = routertesting.DenomTwo
	DenomThree = routertesting.DenomThree
	DenomFour  = routertesting.DenomFour
	DenomFive  = routertesting.DenomFive
	DenomSix   = routertesting.DenomSix
)

// This test validates that we are able to split over multiple routes.
// We know that higher liquidity pools should be more optimal than the lower liquidity pools
// all else equal.
// Given that, we initialize several pools with different amounts of liquidity.
// We define an expected order of the amountsIn and Out for the split routes based on liquidity.
// Lastly, we assert that the actual order of the split routes is the same as the expected order.
func (s *RouterTestSuite) TestGetBestSplitRoutesQuote() {
	type routeWithOrder struct {
		route domain.SplitRoute
		order int
	}

	s.Setup()

	xLiquidity := sdk.NewCoins(
		sdk.NewCoin(DenomOne, sdk.NewInt(1_000_000_000_000)),
		sdk.NewCoin(DenomTwo, sdk.NewInt(2_000_000_000_000)),
	)

	// X Liquidity
	defaultBalancerPoolID := s.PrepareBalancerPoolWithCoins(xLiquidity...)

	// 2X liquidity
	// Note that the second pool has more liquidity than the first so it should be preferred
	secondBalancerPoolIDSameDenoms := s.PrepareBalancerPoolWithCoins(coinutil.MulRaw(xLiquidity, 2)...)

	// 4X liquidity
	// Note that the third pool has more liquidity than first and second so it should be preferred
	thirdBalancerPoolIDSameDenoms := s.PrepareBalancerPoolWithCoins(coinutil.MulRaw(xLiquidity, 4)...)

	// Get the defaultBalancerPool from the store
	defaultBalancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, defaultBalancerPoolID)
	s.Require().NoError(err)

	// Get the secondBalancerPool from the store
	secondBalancerPoolSameDenoms, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, secondBalancerPoolIDSameDenoms)
	s.Require().NoError(err)

	// // Get the thirdBalancerPool from the store
	thirdBalancerPoolSameDenoms, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, thirdBalancerPoolIDSameDenoms)
	s.Require().NoError(err)

	tests := map[string]struct {
		maxSplitIterations int

		routes      []route.RouteImpl
		tokenIn     sdk.Coin
		expectError error

		expectedTokenOutDenom string

		// Ascending order in terms of which route is preferred
		// and uses the largest amount of the token in
		expectedProportionInOrder []int
	}{
		"valid single route": {
			routes: []route.RouteImpl{
				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultMockPool, DenomOne), defaultBalancerPool),
				})},
			tokenIn: sdk.NewCoin(DenomTwo, sdk.NewInt(100)),

			expectedTokenOutDenom: DenomOne,

			expectedProportionInOrder: []int{0},
		},
		"valid two route single hop": {
			routes: []route.RouteImpl{
				// Route 1
				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultMockPool, DenomOne), defaultBalancerPool),
				}),

				// Route 2
				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithPoolID(mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultMockPool, DenomOne), secondBalancerPoolSameDenoms), 2),
				}),
			},

			maxSplitIterations: 10,

			tokenIn: sdk.NewCoin(DenomTwo, sdk.NewInt(5_000_000)),

			expectedTokenOutDenom: DenomOne,

			// Route 2 is preferred because it has 2x the liquidity of Route 1
			expectedProportionInOrder: []int{0, 1},
		},
		"valid three route single hop": {
			routes: []route.RouteImpl{
				// Route 1
				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultMockPool, DenomOne), defaultBalancerPool),
				}),

				// Route 2
				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithPoolID(mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultMockPool, DenomOne), thirdBalancerPoolSameDenoms), 3),
				}),

				// Route 3
				WithRoutePools(route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithPoolID(mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultMockPool, DenomOne), secondBalancerPoolSameDenoms), 2),
				}),
			},

			maxSplitIterations: 17,

			tokenIn: sdk.NewCoin(DenomTwo, sdk.NewInt(56_789_321)),

			expectedTokenOutDenom: DenomOne,

			// Route 2 is preferred because it has 4x the liquidity of Route 1
			// and 2X the liquidity of Route 3
			expectedProportionInOrder: []int{2, 0, 1},
		},

		// TODO: cover error cases
		// TODO: multi route multi hop
		// TODO: assert that split ratios are correct
	}

	for name, tc := range tests {
		s.Run(name, func() {

			logger, err := log.NewLogger(false, "", "")
			s.Require().NoError(err)

			r := routerusecase.NewRouter([]uint64{}, 0, 0, 0, tc.maxSplitIterations, 0, logger)

			quote, err := r.GetSplitQuote(tc.routes, tc.tokenIn)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(tc.expectError, err)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.tokenIn, quote.GetAmountIn())

			quoteCoinOut := quote.GetAmountOut()
			// We only validate that some amount is returned. The correctness of the amount is to be calculated at a different level
			// of abstraction.
			s.Require().NotNil(quoteCoinOut)

			// Validate that amounts in in the quote split routes add up to the original amount in
			routes := quote.GetRoute()
			actualTotalFromSplits := sdk.ZeroInt()
			for _, splitRoute := range routes {
				actualTotalFromSplits = actualTotalFromSplits.Add(splitRoute.GetAmountIn())
			}

			// Error tolerance of 1 to account for the rounding differences
			errTolerance := osmomath.ErrTolerance{
				AdditiveTolerance: osmomath.OneDec(),
			}
			osmoassert.Equal(s.T(), errTolerance, tc.tokenIn.Amount, actualTotalFromSplits)

			// Route must not be nil
			actualRoutes := quote.GetRoute()
			s.Require().NotNil(actualRoutes)

			s.Require().Equal(len(tc.expectedProportionInOrder), len(actualRoutes))

			routesWithOrder := make([]routeWithOrder, len(actualRoutes))
			for i, route := range actualRoutes {
				routesWithOrder[i] = routeWithOrder{
					route: route,
					order: tc.expectedProportionInOrder[i],
				}
			}

			// Sort actual routes in the expected order
			// and assert that the token in is strictly decreasing
			sort.Slice(routesWithOrder, func(i, j int) bool {
				return routesWithOrder[i].order < routesWithOrder[j].order
			})
			s.Require().NotEmpty(routesWithOrder)

			// Iterate over sorted routes with order and validate that the
			// amounts in and out are strictly decreasing as per the expected order.
			previousRouteAmountIn := routesWithOrder[0].route.GetAmountIn()
			previousRouteAmountOut := routesWithOrder[0].route.GetAmountOut()
			for i := 1; i < len(routesWithOrder)-1; i++ {
				currentRouteAmountIn := routesWithOrder[i].route.GetAmountIn()
				currentRouteAmountOut := routesWithOrder[i].route.GetAmountOut()

				// Both in and out amounts must be strictly decreasing
				s.Require().True(previousRouteAmountIn.GT(currentRouteAmountIn))
				s.Require().True(previousRouteAmountOut.GT(currentRouteAmountOut))
			}
		})
	}
}

// This test ensures strict route validation.
// See individual test cases for details.
func (s *RouterTestSuite) TestValidateAndFilterRoutes() {

	defaultDenomOneTwoOutTwoPool := usecase.CandidatePoolWrapper{
		CandidatePool: route.CandidatePool{
			ID:            defaultPoolID,
			TokenOutDenom: DenomTwo,
		},
		PoolDenoms: []string{DenomOne, DenomTwo},
	}

	tests := map[string]struct {
		routes                    [][]usecase.CandidatePoolWrapper
		tokenInDenom              string
		expectError               error
		expectFiltered            bool
		expectFilteredRouteLength int
	}{
		"valid single route single hop": {
			routes: [][]usecase.CandidatePoolWrapper{
				{
					defaultDenomOneTwoOutTwoPool,
				},
			},

			tokenInDenom: DenomOne,
		},
		"valid single route multi-hop": {
			routes: [][]usecase.CandidatePoolWrapper{
				{
					defaultDenomOneTwoOutTwoPool,
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID + 1,
							TokenOutDenom: DenomThree,
						},
						PoolDenoms: []string{DenomTwo, DenomThree},
					},
				},
			},

			tokenInDenom: DenomOne,
		},
		"valid multi route": {
			routes: [][]usecase.CandidatePoolWrapper{
				{
					defaultDenomOneTwoOutTwoPool,
				},
				{
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID + 1,
							TokenOutDenom: DenomThree,
						},
						PoolDenoms: []string{DenomOne, DenomThree},
					},
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID + 2,
							TokenOutDenom: DenomTwo,
						},
						PoolDenoms: []string{DenomTwo, DenomThree},
					},
				},
			},

			tokenInDenom: DenomOne,
		},

		// errors

		"error: no pools in route": {
			routes: [][]usecase.CandidatePoolWrapper{
				{},
			},

			tokenInDenom: DenomTwo,

			expectError: usecase.NoPoolsInRouteError{RouteIndex: 0},
		},
		"error: token out mismatch between multiple routes": {
			routes: [][]usecase.CandidatePoolWrapper{
				{
					defaultDenomOneTwoOutTwoPool,
				},
				{
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID + 1,
							TokenOutDenom: DenomThree,
						},
						PoolDenoms: []string{DenomTwo, DenomThree},
					},
				},
			},

			tokenInDenom: DenomTwo,

			expectError: usecase.TokenOutMismatchBetweenRoutesError{TokenOutDenomRouteA: DenomTwo, TokenOutDenomRouteB: DenomThree},
		},
		"error: token in matches token out": {
			routes: [][]usecase.CandidatePoolWrapper{
				{
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID + 1,
							TokenOutDenom: DenomOne,
						},
						PoolDenoms: []string{DenomOne, DenomTwo},
					},
				},
			},

			tokenInDenom: DenomOne,

			expectError: usecase.TokenOutDenomMatchesTokenInDenomError{Denom: DenomOne},
		},
		"error: token in does not match pool denoms": {
			routes: [][]usecase.CandidatePoolWrapper{
				{
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID,
							TokenOutDenom: DenomOne,
						},
						PoolDenoms: []string{DenomOne, DenomTwo},
					},
				},
			},
			tokenInDenom: DenomThree,

			expectError: usecase.PreviousTokenOutDenomNotInPoolError{RouteIndex: 0, PoolId: DefaultMockPool.GetId(), PreviousTokenOutDenom: DenomThree},
		},
		"error: token out does not match pool denoms": {
			routes: [][]usecase.CandidatePoolWrapper{
				{
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID,
							TokenOutDenom: DenomThree,
						},
						PoolDenoms: []string{DenomOne, DenomTwo},
					},
				},
			},
			tokenInDenom: DenomOne,

			expectError: usecase.CurrentTokenOutDenomNotInPoolError{RouteIndex: 0, PoolId: DefaultMockPool.GetId(), CurrentTokenOutDenom: DenomThree},
		},

		// Routes filtered
		"filtered: token in is in the route": {
			routes: [][]usecase.CandidatePoolWrapper{
				{
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID,
							TokenOutDenom: DenomTwo,
						},
						PoolDenoms: []string{DenomOne, DenomTwo},
					},
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID + 1,
							TokenOutDenom: DenomTwo,
						},
						PoolDenoms: []string{DenomTwo, DenomFour},
					},
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID + 2,
							TokenOutDenom: DenomFour,
						},
						PoolDenoms: []string{DenomTwo, DenomFour},
					},
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID + 3,
							TokenOutDenom: DenomThree,
						},
						PoolDenoms: []string{DenomFour, DenomOne},
					},
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID + 4,
							TokenOutDenom: DenomThree,
						},
						PoolDenoms: []string{DenomOne, DenomThree},
					},
				},
			},
			tokenInDenom: DenomOne,

			expectFiltered: true,
		},
		"filtered: token out is in the route": {
			routes: [][]usecase.CandidatePoolWrapper{
				{
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID,
							TokenOutDenom: DenomTwo,
						},
						PoolDenoms: []string{DenomOne, DenomTwo},
					},
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID + 1,
							TokenOutDenom: DenomTwo,
						},
						PoolDenoms: []string{DenomTwo, DenomFour},
					},
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID + 2,
							TokenOutDenom: DenomTwo,
						},
						PoolDenoms: []string{DenomTwo, DenomFour},
					},
				},
			},
			tokenInDenom: DenomOne,

			expectFiltered: true,
		},
		"filtered: same pool id within only route": {
			routes: [][]usecase.CandidatePoolWrapper{
				{
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID,
							TokenOutDenom: DenomTwo,
						},
						PoolDenoms: []string{DenomOne, DenomTwo},
					},
					{
						CandidatePool: route.CandidatePool{
							ID:            defaultPoolID,
							TokenOutDenom: DenomFour,
						},
						PoolDenoms: []string{DenomTwo, DenomFour},
					},
				},
			},

			tokenInDenom: DenomOne,

			expectFiltered: true,
		},
		"not filtered: same pool id between routes": {
			routes: [][]usecase.CandidatePoolWrapper{
				{
					defaultDenomOneTwoOutTwoPool,
				},
				{
					defaultDenomOneTwoOutTwoPool,
				},
			},
			tokenInDenom: DenomOne,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {

			router := routerusecase.NewRouter([]uint64{}, 0, 0, 0, 0, 0, &log.NoOpLogger{})

			filteredCandidateRoutes, err := router.ValidateAndFilterRoutes(tc.routes, tc.tokenInDenom)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)

			if tc.expectFiltered {
				s.Require().NotEqual(len(tc.routes), len(filteredCandidateRoutes.Routes))
				s.Require().Len(filteredCandidateRoutes.Routes, tc.expectFilteredRouteLength)
				return
			}

			s.Require().Equal(len(tc.routes), len(filteredCandidateRoutes.Routes))
		})
	}
}

// Validates that USDT for UMEE quote does not cause an error.
// This pair originally caused an error due to the lack of filtering that was
// added later.
func (s *RouterTestSuite) TestGetBestSplitRoutesQuote_Mainnet_USDTUMEE() {
	config := defaultRouterConfig
	config.MaxPoolsPerRoute = 5
	config.MaxRoutes = 10

	var (
		amountIn = osmomath.NewInt(1000_000_000)
	)

	router, tickMap, takerFeeMap := s.setupMainnetRouter(config)

	routes := s.constructRoutesFromMainnetPools(router, USDT, UMEE, tickMap, takerFeeMap)

	quote, err := router.GetOptimalQuote(sdk.NewCoin(USDT, amountIn), routes)

	// We only validate that error does not occur without actually validating the quote.
	s.Require().NoError(err)

	// Expecting 2 routes based on the mainnet state
	// 1: 1205
	// 2: 1110 -> 1077
	quoteRoutes := quote.GetRoute()
	s.Require().Len(quoteRoutes, 3)
}

func (s *RouterTestSuite) TestGetBestSplitRoutesQuote_Mainnet_UOSMOUION() {
	config := defaultRouterConfig
	config.MaxPoolsPerRoute = 5
	config.MaxRoutes = 10

	var (
		amountIn = osmomath.NewInt(5000000)
	)

	router, tickMap, takerFeeMap := s.setupMainnetRouter(config)

	routes := s.constructRoutesFromMainnetPools(router, UOSMO, UION, tickMap, takerFeeMap)

	quote, err := router.GetOptimalQuote(sdk.NewCoin(UOSMO, amountIn), routes)

	// We only validate that error does not occur without actually validating the quote.
	s.Require().NoError(err)

	// Expecting 1 route based on mainnet state
	s.Require().Len(quote.GetRoute(), 1)

	// Should be pool 2 based on state at the time of writing.
	// If new state is added, review the quote against mainnet.
	s.Require().Len(quote.GetRoute()[0].GetPools(), 1)

	// Validate that the pool is pool 2
	s.Require().Equal(uint64(2), quote.GetRoute()[0].GetPools()[0].GetId())
}

func (s *RouterTestSuite) TestGetBestSplitRoutesQuote_Mainnet_USDTATOM() {
	config := defaultRouterConfig
	config.MaxPoolsPerRoute = 4
	config.MaxRoutes = 5
	config.MaxSplitRoutes = 3
	config.MaxSplitIterations = 10

	var (
		amountIn = osmomath.NewInt(1000000)
	)

	router, tickMap, takerFeeMap := s.setupMainnetRouter(config)

	routes := s.constructRoutesFromMainnetPools(router, USDT, ATOM, tickMap, takerFeeMap)

	quote, err := router.GetOptimalQuote(sdk.NewCoin(USDT, amountIn), routes)

	// We only validate that error does not occur without actually validating the quote.
	s.Require().NoError(err)

	s.Require().NotNil(quote)

	s.Require().Greater(len(quote.GetRoute()), 0)
}

func (s *RouterTestSuite) TestGetBestSplitRoutesQuote_Mainnet_AKTUMEE() {
	config := defaultRouterConfig
	config.MaxPoolsPerRoute = 4
	config.MaxRoutes = 10
	config.MaxSplitRoutes = 3

	var (
		amountIn = osmomath.NewInt(100_000_000)
	)

	router, tickMap, takerFeeMap := s.setupMainnetRouter(config)

	routes := s.constructRoutesFromMainnetPools(router, AKT, UMEE, tickMap, takerFeeMap)

	quote, err := router.GetOptimalQuote(sdk.NewCoin(AKT, amountIn), routes)

	// We only validate that error does not occur without actually validating the quote.
	s.Require().NoError(err)

	// Expecting 2 routes based on mainnet state
	s.Require().Len(quote.GetRoute(), 2)

	// TODO: need to compare this quote against mainnet state

	route := quote.GetRoute()[0]
	// Expecting 2 pools in the route
	s.Require().Len(route.GetPools(), 2)

	// Validate that the pool is pool 1093
	s.Require().Equal(uint64(1093), route.GetPools()[0].GetId())

	// Validate that the pool is pool 1110
	s.Require().Equal(uint64(1110), route.GetPools()[1].GetId())
}

// Generates routes from mainnet state by:
// - instrumenting pool repository mock with pools and ticks
// - setting this mock on the pools use case
// - setting the pool use case on the router (called during GetCandidateRoutes() method)
// - converting candidate routes to routes with all the necessary data.
// COTRACT: router is initialized with setupMainnetRouter(...) or setupDefaultMainnetRouter(...)
func (s *RouterTestSuite) constructRoutesFromMainnetPools(router *routerusecase.Router, tokenInDenom, tokenOutDenom string, tickMap map[uint64]domain.TickModel, takerFeeMap domain.TakerFeeMap) []route.RouteImpl {
	// Setup router repository mock
	routerRepositoryMock := mocks.RedisRouterRepositoryMock{}
	routerusecase.WithRouterRepository(router, &routerRepositoryMock)

	// Setup pools usecase mock.
	poolsRepositoryMock := mocks.RedisPoolsRepositoryMock{
		Pools:     router.GetSortedPools(),
		TickModel: tickMap,
	}
	poolsUsecase := poolsusecase.NewPoolsUsecase(time.Hour, &poolsRepositoryMock, nil)
	routerusecase.WithPoolsUsecase(router, poolsUsecase)

	candidateRoutes, err := router.GetCandidateRoutes(tokenInDenom, tokenOutDenom)
	s.Require().NoError(err)

	routes, err := poolsUsecase.GetRoutesFromCandidates(context.Background(), candidateRoutes, takerFeeMap, tokenInDenom, tokenOutDenom)
	s.Require().NoError(err)

	return routes
}
