package usecase_test

import (
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/route"
)

// Validates that the router returns the correct routes for the given token pair.
func (s *RouterTestSuite) TestGetCandidateRoutesBFS_OSMOATOM() {
	config := defaultRouterConfig
	config.MaxPoolsPerRoute = 5
	config.MaxRoutes = 10

	router := s.setupMainnetRouter(config)

	routes, err := router.GetCandidateRoutes(UOSMO, ATOM)
	s.Require().NoError(err)

	s.Require().Equal(8, len(routes))

	// https://app.osmosis.zone/pool/1135
	s.validateExpectedPoolIDOneHopRoute(routes[0], 1135)

	// https://app.osmosis.zone/pool/1
	s.validateExpectedPoolIDOneHopRoute(routes[1], 1)
}

// Validates that the router returns the correct routes for the given token pair.
// Inverting the swap direction should return the same routes.
func (s *RouterTestSuite) TestGetCandidateRoutesBFS_OSMOstOSMO() {
	config := defaultRouterConfig
	config.MaxPoolsPerRoute = 5
	config.MaxRoutes = 10
	config.MinOSMOLiquidity = 1000

	router := s.setupMainnetRouter(config)

	routesUOSMOIn, err := router.GetCandidateRoutes(UOSMO, stOSMO)
	s.Require().NoError(err)

	// Invert
	routesstOSMOIn, err := router.GetCandidateRoutes(stOSMO, UOSMO)
	s.Require().NoError(err)

	s.Require().Equal(len(routesUOSMOIn), len(routesstOSMOIn))

	// https://info.osmosis.zone/token/stOSMO
	// Pools 833 and 1252 at the time of test creation.
	s.Require().Equal(2, len(routesstOSMOIn))
	s.validateExpectedPoolIDOneHopRoute(routesstOSMOIn[1], 833)
	s.validateExpectedPoolIDOneHopRoute(routesstOSMOIn[0], 1252)

	// https://info.osmosis.zone/token/stOSMO
	// Pools 833 and 1252 at the time of test creation.
	s.Require().Equal(2, len(routesstOSMOIn))
	s.validateExpectedPoolIDOneHopRoute(routesstOSMOIn[1], 833)
	s.validateExpectedPoolIDOneHopRoute(routesstOSMOIn[0], 1252)
}

// Validate that can find at least 1 route with no error for top 10
// pairs by volume.
func (s *RouterTestSuite) TestGetCandidateRoutesBFS_Top10VolumePairs() {
	config := defaultRouterConfig
	config.MaxPoolsPerRoute = 3
	config.MaxRoutes = 10
	router := s.setupMainnetRouter(config)

	// Manually taken from https://info.osmosis.zone/ in Nov 2023.
	top10ByVolumeDenoms := []string{
		UOSMO,
		ATOM,
		stOSMO,
		stATOM,
		USDC,
		USDCaxl,
		USDT,
		WBTC,
		ETH,
		AKT,
	}

	for i := 0; i < len(top10ByVolumeDenoms); i++ {
		for j := i + 1; j < len(top10ByVolumeDenoms); j++ {
			tokenI := top10ByVolumeDenoms[i]
			tokenJ := top10ByVolumeDenoms[j]

			routes, err := router.GetCandidateRoutes(tokenI, tokenJ)
			s.Require().NoError(err)
			s.Require().Greater(len(routes), 0, "tokenI: %s, tokenJ: %s", tokenI, tokenJ)

			routes, err = router.GetCandidateRoutes(tokenJ, tokenI)
			s.Require().NoError(err)
			s.Require().Greater(len(routes), 0, "tokenJ: %s, tokenI: %s", tokenJ, tokenI)
		}
	}
}

func (s *RouterTestSuite) validateExpectedPoolIDOneHopRoute(route route.RouteImpl, expectedPoolID uint64) {
	routePools := route.GetPools()
	s.Require().Equal(1, len(routePools))
	s.Require().Equal(expectedPoolID, routePools[0].GetId())
}
