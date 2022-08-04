package ibc_rate_limit_test

import (
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/stretchr/testify/suite"
	ibc_rate_limit "github.com/osmosis-labs/osmosis/v10/x/ibc-rate-limit"

	"testing"
)

type MiddlewareTestSuite struct {
	suite.Suite

	// Uncommenting this line (and the import) makes everything fail
	coordinator *ibctesting.Coordinator
	RateLimitMiddlware ibc_rate_limit.RateLimitMiddleware
}

func (suite *MiddlewareTestSuite) SetupTest() {
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

// Uncommenting this line (and the import) makes everything fail
func NewTransferPath(chainA, chainB *ibctesting.TestChain) {}

func (suite *MiddlewareTestSuite) TestSendPacket() {
	suite.T().Log("Say bye")
}
