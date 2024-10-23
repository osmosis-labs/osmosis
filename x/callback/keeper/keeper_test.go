package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	e2eTesting "github.com/osmosis-labs/osmosis/v26/tests/e2e/testing"
)

type KeeperTestSuite struct {
	suite.Suite

	chain *e2eTesting.TestChain
}

func (s *KeeperTestSuite) SetupTest() {
	s.chain = e2eTesting.NewTestChain(s.T(), 1)
}

func TestCallbackKeeper(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
