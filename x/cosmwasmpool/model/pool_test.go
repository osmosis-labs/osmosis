package model_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v22/app/apptesting"
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

	s.Ctx = s.Ctx.WithGasMeter(sdk.NewGasMeter(1000000))

	actualSpotPrice, err := pool.SpotPrice(s.Ctx, denomA, denomB)
	s.Require().NoError(err)

	// Validate that the gas was charged on the input context
	endGas := s.Ctx.GasMeter().GasConsumed()
	s.Require().NotZero(endGas)

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
