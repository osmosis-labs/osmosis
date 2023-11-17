package route_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/pools"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/routertesting"
	concentratedmodel "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/model"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

type RouterTestSuite struct {
	routertesting.RouterTestHelper
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}

var (
	// Concentrated liquidity constants
	ETH    = routertesting.ETH
	USDC   = routertesting.USDC
	USDT   = routertesting.USDT
	Denom0 = ETH
	Denom1 = USDC

	DefaultCurrentTick = routertesting.DefaultCurrentTick

	DefaultAmt0 = routertesting.DefaultAmt0
	DefaultAmt1 = routertesting.DefaultAmt1

	DefaultCoin0 = routertesting.DefaultCoin0
	DefaultCoin1 = routertesting.DefaultCoin1

	DefaultLiquidityAmt = routertesting.DefaultLiquidityAmt

	// router specific variables
	defaultTickModel = routertesting.DefaultTickModel

	noTakerFee = routertesting.NoTakerFee

	emptyRoute = routertesting.EmptyRoute
)

var (
	DefaultTakerFee     = routertesting.DefaultTakerFee
	DefaultPoolBalances = routertesting.DefaultPoolBalances
	DefaultSpreadFactor = routertesting.DefaultSpreadFactor

	DefaultPool = routertesting.DefaultPool
	EmptyRoute  = routertesting.EmptyRoute

	// Test denoms
	DenomOne   = routertesting.DenomOne
	DenomTwo   = routertesting.DenomTwo
	DenomThree = routertesting.DenomThree
	DenomFour  = routertesting.DenomFour
	DenomFive  = routertesting.DenomFive
	DenomSix   = routertesting.DenomSix
)

// This test validates that the pools in the route are converted into a new serializable
// type for clients with the following list of fields that are returned in each pool:
// - ID
// - Type
// - Balances
// - Spread Factor
// - Token Out Denom
// - Taker Fee
func (s *RouterTestSuite) TestPrepareResultPools() {
	s.Setup()

	balancerPoolID := s.PrepareBalancerPool()

	balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, balancerPoolID)
	s.Require().NoError(err)

	testcases := map[string]struct {
		route domain.Route

		expectedPools []domain.RoutablePool
	}{
		"empty route": {
			route: emptyRoute.DeepCopy(),

			expectedPools: []domain.RoutablePool{},
		},
		"single balancer pool in route": {
			route: WithRoutePools(
				emptyRoute,
				[]domain.RoutablePool{
					mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultPool, DenomOne), balancerPool),
				},
			),

			expectedPools: []domain.RoutablePool{
				pools.NewRoutableResultPool(
					balancerPoolID,
					poolmanagertypes.Balancer,
					DefaultPoolBalances,
					DefaultSpreadFactor,
					DenomOne,
					DefaultTakerFee,
				),
			},
		},

		// TODO:
		// add tests with more pool types as well as multiple pools in the route
		// https://app.clickup.com/t/86a1cfwag
	}

	for name, tc := range testcases {
		tc := tc
		s.Run(name, func() {

			resultPools := tc.route.PrepareResultPools()

			s.ValidateRoutePools(tc.expectedPools, resultPools)
			s.ValidateRoutePools(tc.expectedPools, tc.route.GetPools())
		})
	}
}

// TestReverse validates that the routes are reversed correctly by reversing the pools
// and chaing the token out denom.
// Example in this test:
//
// Original route: DenomOne & DenomTwo (DenomOne out) -> DonemOne & DenomThree (DenomThree out) -> DenomThree & DenomFour (DenomFour out)
//
// Expected route: DenomThree & DenomFour (DenomThree out) -> DenomOne & DenomThree (DenomOne out) -> DenomOne & DenomTwo (DenomTwo out)
func (s *RouterTestSuite) TestReverse() {

	route := WithRoutePools(
		emptyRoute,
		[]domain.RoutablePool{
			// DenomOne & DenomTwo with DenomOne out
			mocks.WithChainPoolModel(mocks.WithDenoms(mocks.WithTokenOutDenom(DefaultPool, DenomOne), []string{DenomOne, DenomTwo}), &concentratedmodel.Pool{}),
			// DenomOne & DenomThree with DenomThree out
			mocks.WithChainPoolModel(mocks.WithDenoms(mocks.WithTokenOutDenom(DefaultPool, DenomThree), []string{DenomOne, DenomThree}), &concentratedmodel.Pool{}),

			// DenomThree & DenomFour with DenomFour out
			mocks.WithChainPoolModel(mocks.WithDenoms(mocks.WithTokenOutDenom(DefaultPool, DenomFour), []string{DenomThree, DenomFour}), &concentratedmodel.Pool{}),
		},
	)

	expectedRoute := WithRoutePools(
		emptyRoute,
		[]domain.RoutablePool{
			mocks.WithChainPoolModel(mocks.WithDenoms(mocks.WithTokenOutDenom(DefaultPool, DenomThree), []string{DenomThree, DenomFour}), &concentratedmodel.Pool{}),
			mocks.WithChainPoolModel(mocks.WithDenoms(mocks.WithTokenOutDenom(DefaultPool, DenomOne), []string{DenomOne, DenomThree}), &concentratedmodel.Pool{}),
			mocks.WithChainPoolModel(mocks.WithDenoms(mocks.WithTokenOutDenom(DefaultPool, DenomTwo), []string{DenomOne, DenomTwo}), &concentratedmodel.Pool{}),
		},
	)

	desiredTokenOutDenom := DenomTwo

	err := route.Reverse(desiredTokenOutDenom)
	s.Require().NoError(err)

	s.Require().Equal(expectedRoute.String(), route.String())
}

func WithRoutePools(r domain.Route, pools []domain.RoutablePool) domain.Route {
	return routertesting.WithRoutePools(r, pools)
}
