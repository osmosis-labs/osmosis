package domain_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/pools"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/route"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/routertesting"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
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

	balancerPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoins(
		sdk.NewCoin(DenomOne, sdk.NewInt(1_000_000_000)),
		sdk.NewCoin(DenomTwo, sdk.NewInt(1_000_000_000)),
	)...)

	balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, balancerPoolID)
	s.Require().NoError(err)

	testcases := map[string]struct {
		route route.RouteImpl

		expectedPools []domain.RoutablePool
	}{
		"empty route": {
			route: emptyRoute,

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

			// Note: token in is chosen arbitrarily since it is irrelevant for this test
			_, _, err := tc.route.PrepareResultPools(sdk.NewCoin(DenomTwo, DefaultAmt0))
			s.Require().NoError(err)

			s.ValidateRoutePools(tc.expectedPools, tc.route.GetPools())
		})
	}
}

func WithRoutePools(r route.RouteImpl, pools []domain.RoutablePool) route.RouteImpl {
	return routertesting.WithRoutePools(r, pools)
}
