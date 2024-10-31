package model_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
)

type CosmWasmPoolSuite struct {
	apptesting.KeeperTestHelper
}

const (
	denomA = "axlusdc"
	denomB = "gravusdc"
)

func TestPoolModuleSuite(t *testing.T) {
	suite.Run(t, new(CosmWasmPoolSuite))
}

func (s *CosmWasmPoolSuite) SetupTest() {
	s.Setup()
}

// TestGetSpreadFactor validates that spread factor is set to zero.
func (s *CosmWasmPoolSuite) TestGetSpreadFactor() {
	var (
		expectedSwapFee = osmomath.ZeroDec()
	)

	pool := s.PrepareCosmWasmPool()

	actualSwapFee := pool.GetSpreadFactor(s.Ctx)

	s.Require().Equal(expectedSwapFee, actualSwapFee)
}

// TestSpotPrice validates that spot price is returned as one.
func (s *CosmWasmPoolSuite) TestSpotPrice() {
	var expectedSpotPrice = osmomath.OneBigDec()

	pool := s.PrepareCosmWasmPool()

	s.Ctx = s.Ctx.WithGasMeter(storetypes.NewGasMeter(100000000))

	const (
		// Charge gas before the system under test method and make sure it is not dropped
		gasChargeBefore = 1000000

		// Charge gas after the system under test method and make sure it is not dropped
		gasChargeAfter = 5555555
	)

	s.Ctx.GasMeter().ConsumeGas(gasChargeBefore, "gas charge before")

	actualSpotPrice, err := pool.SpotPrice(s.Ctx, denomA, denomB)
	s.Require().NoError(err)

	s.Ctx.GasMeter().ConsumeGas(gasChargeAfter, "gas charge after")

	// Validate that the gas was charged on the input context
	gasConsumed := s.Ctx.GasMeter().GasConsumed()
	s.Require().NotZero(gasConsumed)

	// Make sure that gas charge before and after is not dropped
	gasConsumed = gasConsumed - gasChargeBefore - gasChargeAfter
	s.Require().Greater(gasConsumed, uint64(0))

	s.Require().Equal(expectedSpotPrice, actualSpotPrice)

	actualSpotPrice, err = pool.SpotPrice(s.Ctx, denomB, denomA)
	s.Require().NoError(err)

	s.Require().Equal(expectedSpotPrice, actualSpotPrice)
}

// TestGetPoolDenoms validates that pool denoms are returned correctly.
func (s *CosmWasmPoolSuite) TestGetPoolDenoms() {
	cwPool := s.PrepareCosmWasmPool()
	poolDenoms := cwPool.GetPoolDenoms(s.Ctx)
	s.Require().Equal([]string{apptesting.DefaultTransmuterDenomA, apptesting.DefaultTransmuterDenomB}, poolDenoms)
}
