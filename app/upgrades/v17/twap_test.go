package v17_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v17/app/apptesting"
	v17 "github.com/osmosis-labs/osmosis/v17/app/upgrades/v17"
)

type TwapUpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *TwapUpgradeTestSuite) SetupTest() {
	suite.Setup()
}

func TestTwapUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(TwapUpgradeTestSuite))
}

func (s *TwapUpgradeTestSuite) TestFlipTwapSpotPriceRecords() {
	tests := map[string]struct {
		poolIds []uint64
	}{
		"success": {
			poolIds: []uint64{1},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()

			s.PrepareMultipleConcentratedPools(5)

			// perform bunch of swaps
			// let x amount of time run by
			// check twap data
			s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 24))
			err := v17.FlipTwapSpotPriceRecords(s.Ctx, tc.poolIds, &s.App.AppKeepers)
			s.Require().NoError(err)
		})
	}
}
