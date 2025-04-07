package commondomain_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	commondomain "github.com/osmosis-labs/osmosis/v27/ingest/common/domain"
)

type CommonDomainTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

func TestCommonDomainTestSuite(t *testing.T) {
	suite.Run(t, new(CommonDomainTestSuite))
}

// Validates the invariant that the block process strategy manager.
// When initialized, should push all data.
// If the block process strategy manager has observed an error, it should push all data.
// If the block process strategy manager has not observed an error after pushing all data, it should push only changed data.
func (s *CommonDomainTestSuite) TestBlockProcessStrategyManager() {

	blockStrategyManager := commondomain.NewBlockProcessStrategyManager()

	// ShouldPushAllData should return true when initialized
	s.Require().True(blockStrategyManager.ShouldPushAllData())

	blockStrategyManager.MarkInitialDataIngested()

	// ShouldPushAllData should return false after MarkInitialDataIngested
	s.Require().False(blockStrategyManager.ShouldPushAllData())

	blockStrategyManager.MarkErrorObserved()

	// ShouldPushAllData should return true after MarkErrorObserved
	s.Require().True(blockStrategyManager.ShouldPushAllData())

	blockStrategyManager.MarkInitialDataIngested()

	// ShouldPushAllData should return false after MarkInitialDataIngested again
	s.Require().False(blockStrategyManager.ShouldPushAllData())

	blockStrategyManager.MarkInitialDataIngested()

	// Unchanged after MarkInitialDataIngested twice
	s.Require().False(blockStrategyManager.ShouldPushAllData())
}
