package domain_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v28/app/apptesting"
	"github.com/osmosis-labs/osmosis/v28/ingest/indexer/domain"
)

type PairTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

func TestPairTestSuite(t *testing.T) {
	suite.Run(t, new(PairTestSuite))
}

// TestShouldFilterDenom tests the ShouldFilterDenom function.
func (suite *PairTestSuite) TestShouldFilterDenom() {

	tests := []struct {
		name                      string
		poolDenom                 string
		expectedShouldFilterDenom bool
	}{
		{
			name:                      "CL pool 1066",
			poolDenom:                 "cl/pool/1066",
			expectedShouldFilterDenom: true,
		},
		{
			name:                      "Gamm pool 5",
			poolDenom:                 "gamm/pool/5",
			expectedShouldFilterDenom: true,
		},
		{
			name:                      "Transmuter pool",
			poolDenom:                 "factory/osmo10c8y69yylnlwrhu32ralf08ekladhfknfqrjsy9yqc9ml8mlxpqq2sttzk/transmuter/poolshare",
			expectedShouldFilterDenom: false,
		},
		{
			name:                      "Empty string",
			poolDenom:                 "",
			expectedShouldFilterDenom: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.Require().Equal(tt.expectedShouldFilterDenom, domain.ShouldFilterDenom(tt.poolDenom))
		})
	}
}

// TestIsMultiDenom tests the IsMultiDenom function.
func (suite *PairTestSuite) TestIsMultiDenom() {
	tests := []struct {
		name                 string
		denoms               []string
		expectedIsMultiDenom bool
	}{
		{
			name:                 "Single denom",
			denoms:               []string{"ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7"},
			expectedIsMultiDenom: false,
		},
		{
			name:                 "Double denoms",
			denoms:               []string{"uosmo", "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7"},
			expectedIsMultiDenom: false,
		},
		{
			name:                 "Multi denoms",
			denoms:               []string{"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", "ibc/E6931F78057F7CC5DA0FD6CEF82FF39373A6E0452BF1FD76910B93292CF356C1", "uosmo"},
			expectedIsMultiDenom: true,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.Require().Equal(tt.expectedIsMultiDenom, domain.IsMultiDenom(tt.denoms))
		})
	}
}
