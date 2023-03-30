package model_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
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

// TesGetSwapFee validates that swap fee is set to zero.
func (s *CosmWasmPoolSuite) TestGetSwapFee() {
	var (
		expectedSwapFee = sdk.ZeroDec()
	)

	pool := s.PrepareCosmWasmPool()

	actualSwapFee := pool.GetSwapFee(s.Ctx)

	s.Require().Equal(expectedSwapFee, actualSwapFee)
}

func (s *CosmWasmPoolSuite) TestSpotPrice() {
	var (
		expectedSpotPrice = sdk.OneDec()
	)

	pool := s.PrepareCosmWasmPool()

	actualSpotPrice, err := pool.SpotPrice(s.Ctx, denomA, denomB)
	s.Require().NoError(err)

	s.Require().Equal(expectedSpotPrice, actualSpotPrice)

	actualSpotPrice, err = pool.SpotPrice(s.Ctx, denomB, denomA)
	s.Require().NoError(err)

	s.Require().Equal(expectedSpotPrice, actualSpotPrice)
}
