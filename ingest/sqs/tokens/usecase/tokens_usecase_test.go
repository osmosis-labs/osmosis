package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v20/app/apptesting"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/tokens/usecase"
)

type TokensUseCaseTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

func TestTokensUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(TokensUseCaseTestSuite))
}

func (s *TokensUseCaseTestSuite) TestParseExponents() {
	s.T().Skip("skip the test that does network call and is used for debugging")

	const assetListFileURL = "https://raw.githubusercontent.com/osmosis-labs/assetlists/main/osmosis-1/osmosis-1.assetlist.json"

	tokensMap, err := usecase.GetTokensFromChainRegistry(assetListFileURL)
	s.Require().NoError(err)
	s.Require().NotEmpty(tokensMap)
}
