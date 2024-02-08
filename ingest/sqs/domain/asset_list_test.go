package domain_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v23/app/apptesting"
	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/domain"
)

type AssetListTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

func TestAssetListTestSuite(t *testing.T) {
	suite.Run(t, new(AssetListTestSuite))
}

func (s *AssetListTestSuite) TestParseExponents() {
	s.T().Skip("skip the test that does network call and is used for debugging")

	const (
		assetListFileURL = "https://raw.githubusercontent.com/osmosis-labs/assetlists/main/osmosis-1/osmosis-1.assetlist.json"

		defaultCosmosExponent = 6
	)
	tokensMap, err := domain.GetTokensFromChainRegistry(assetListFileURL)
	s.Require().NoError(err)
	s.Require().NotEmpty(tokensMap)

	// ATOM is present
	atomMainnetDenom := "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	atomToken, ok := tokensMap[atomMainnetDenom]
	s.Require().True(ok)
	s.Require().Equal(defaultCosmosExponent, atomToken.Precision)
	s.Require().Equal(atomMainnetDenom, atomToken.ChainDenom)

	// ION is present
	ionMainnetDenom := "uion"
	ionToken, ok := tokensMap[ionMainnetDenom]
	s.Require().True(ok)
	s.Require().Equal(defaultCosmosExponent, ionToken.Precision)
	s.Require().Equal(ionMainnetDenom, ionToken.ChainDenom)

	// IBCX is presnet
	ibcxMainnetDenom := "factory/osmo14klwqgkmackvx2tqa0trtg69dmy0nrg4ntq4gjgw2za4734r5seqjqm4gm/uibcx"
	ibcxToken, ok := tokensMap[ibcxMainnetDenom]
	s.Require().True(ok)
	s.Require().Equal(defaultCosmosExponent, ibcxToken.Precision)
	s.Require().Equal(ibcxMainnetDenom, ibcxToken.ChainDenom)
}
