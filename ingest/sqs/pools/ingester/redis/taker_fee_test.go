package redis_test

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/pools/ingester/redis"
)

var (
	// The taker fee that is set for specific pairs
	defaultCustomTakerFee = osmomath.NewDecWithPrec(3, 2)

	// Another custom taker fee set for specific pairs
	otherCustomTakerFee = osmomath.NewDecWithPrec(4, 2)

	// The taker fee taken from the pool manager params
	defaultPoolManagerTakerFee = osmomath.NewDecWithPrec(5, 2)
)

// Tests that the taker fee is correctly retrieved for the given denoms
// and the map is correctly mutated.
func (s *IngesterTestSuite) TestRetrieveTakerFeeToMapIfNotExists() {
	type denomPairTakerFee struct {
		denomPair domain.DenomPair
		takerFee  osmomath.Dec
	}

	tests := map[string]struct {
		preSetTakerFeePairs []denomPairTakerFee

		denoms                         []string
		denomPairToTakerFeeMap         domain.TakerFeeMap
		expectError                    error
		expectedDenomPairToTakerFeeMap domain.TakerFeeMap
	}{
		"one denom pair, taker fee is not in the map, pre-set taker fee": {
			preSetTakerFeePairs: []denomPairTakerFee{
				{
					denomPair: domain.DenomPair{
						Denom0: USDC,
						Denom1: USDT,
					},
					takerFee: defaultCustomTakerFee,
				},
			},

			denoms: []string{USDC, USDT},

			denomPairToTakerFeeMap: domain.TakerFeeMap{},

			expectedDenomPairToTakerFeeMap: domain.TakerFeeMap{
				{
					Denom0: USDC,
					Denom1: USDT,
				}: defaultCustomTakerFee,
			},
		},
		"one denom pair, taker fee is in the map, pre-set taker fee": {
			preSetTakerFeePairs: []denomPairTakerFee{
				{
					denomPair: domain.DenomPair{
						Denom0: USDC,
						Denom1: USDT,
					},
					// Note that this is value A
					takerFee: defaultCustomTakerFee,
				},
			},

			denoms: []string{USDC, USDT},

			denomPairToTakerFeeMap: domain.TakerFeeMap{
				{
					Denom0: USDC,
					Denom1: USDT,
					// Value B is already in the map
				}: otherCustomTakerFee,
			},

			expectedDenomPairToTakerFeeMap: domain.TakerFeeMap{
				{
					Denom0: USDC,
					Denom1: USDT,
					// As a result, value A from state is ignored.
				}: otherCustomTakerFee,
			},
		},
		"one denom pair, taker fee is not in the map, do not pre-set taker fee (take from params)": {
			// No pre-set
			preSetTakerFeePairs: []denomPairTakerFee{},

			denoms: []string{USDC, USDT},

			denomPairToTakerFeeMap: domain.TakerFeeMap{},

			expectedDenomPairToTakerFeeMap: domain.TakerFeeMap{
				{
					Denom0: USDC,
					Denom1: USDT,
				}: defaultPoolManagerTakerFee,
			},
		},
		"three denom pairs, one taker fee is from pre-set, one from params and one is already in the map": {
			preSetTakerFeePairs: []denomPairTakerFee{
				{
					denomPair: domain.DenomPair{
						Denom0: USDC,
						Denom1: USDT,
					},
					takerFee: defaultCustomTakerFee,
				},
			},

			denoms: []string{USDC, USDT, USDW},

			denomPairToTakerFeeMap: domain.TakerFeeMap{
				{
					Denom0: USDT,
					Denom1: USDW,
				}: otherCustomTakerFee,
			},

			expectedDenomPairToTakerFeeMap: domain.TakerFeeMap{
				{
					Denom0: USDC,
					Denom1: USDT,
				}: defaultCustomTakerFee,
				{
					Denom0: USDC,
					Denom1: USDW,
				}: defaultPoolManagerTakerFee,
				{
					Denom0: USDT,
					Denom1: USDW,
				}: otherCustomTakerFee,
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.Setup()

			// Set default poolmanager taker fee that is different from the default taker fee.
			s.setDefaultPoolManagerTakerFee()

			// Pre-set taker fees for testing.
			for _, takerFeePair := range tc.preSetTakerFeePairs {
				s.App.PoolManagerKeeper.SetDenomPairTakerFee(s.Ctx, takerFeePair.denomPair.Denom0, takerFeePair.denomPair.Denom1, takerFeePair.takerFee)
			}

			err := redis.RetrieveTakerFeeToMapIfNotExists(s.Ctx, tc.denoms, tc.denomPairToTakerFeeMap, s.App.PoolManagerKeeper)

			if tc.expectError != nil {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			s.Require().Equal(tc.expectedDenomPairToTakerFeeMap, tc.denomPairToTakerFeeMap)
		})
	}
}

// Sets default poolmanager taker fee
func (s *IngesterTestSuite) setDefaultPoolManagerTakerFee() {
	poolmanagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolmanagerParams.TakerFeeParams.DefaultTakerFee = defaultPoolManagerTakerFee
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolmanagerParams)
}
