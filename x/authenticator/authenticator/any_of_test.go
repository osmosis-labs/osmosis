package authenticator_test

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v19/app"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/testutils"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

type AnyOfAuthenticationSuite struct {
	suite.Suite
	Ctx        sdk.Context
	OsmosisApp *app.OsmosisApp
	AnyOfAuth  authenticator.AnyOfAuthenticator

	alwaysApprove testutils.TestingAuthenticator
	neverApprove  testutils.TestingAuthenticator
}

func TestAnyOfAuthenticationSuite(t *testing.T) {
	suite.Run(t, new(AnyOfAuthenticationSuite))
}

func (s *AnyOfAuthenticationSuite) SetupTest() {
	s.OsmosisApp = app.Setup(false)
	s.Ctx = s.OsmosisApp.NewContext(false, tmproto.Header{})
	s.Ctx = s.Ctx.WithGasMeter(sdk.NewGasMeter(1_000_000))
	am := authenticator.NewAuthenticatorManager()
	// Define authenticators
	s.AnyOfAuth = authenticator.NewAnyOfAuthenticator(am)
	s.alwaysApprove = testutils.TestingAuthenticator{
		Approve:        testutils.Always,
		GasConsumption: 10,
	}
	s.neverApprove = testutils.TestingAuthenticator{
		Approve:        testutils.Never,
		GasConsumption: 10,
	}
	am.RegisterAuthenticator(s.AnyOfAuth)
	am.RegisterAuthenticator(s.alwaysApprove)
	am.RegisterAuthenticator(s.neverApprove)
}

func (s *AnyOfAuthenticationSuite) TestAnyOfAuthenticator() {
	// Define data
	testData := []byte{}

	// Define test cases
	type testCase struct {
		name             string
		authenticators   []authenticator.Authenticator
		expectSuccessful bool
	}

	testCases := []testCase{
		{
			name:             "alwaysApprove + neverApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.neverApprove},
			expectSuccessful: true,
		},
		{
			name:             "neverApprove + neverApprove",
			authenticators:   []authenticator.Authenticator{s.neverApprove, s.neverApprove},
			expectSuccessful: false,
		},
		{
			name:             "alwaysApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.alwaysApprove},
			expectSuccessful: true,
		},
		{
			name:             "neverApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.neverApprove, s.alwaysApprove},
			expectSuccessful: true,
		},
		{
			name:             "alwaysApprove + alwaysApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.alwaysApprove, s.alwaysApprove},
			expectSuccessful: true,
		},
		{
			name:             "alwaysApprove + alwaysApprove + neverApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.alwaysApprove, s.neverApprove},
			expectSuccessful: true,
		},
		{
			name:             "alwaysApprove + neverApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.neverApprove, s.alwaysApprove},
			expectSuccessful: true,
		},
		{
			name:             "neverApprove + neverApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.neverApprove, s.neverApprove, s.alwaysApprove},
			expectSuccessful: true,
		},
		{
			name:             "neverApprove + neverApprove + neverApprove",
			authenticators:   []authenticator.Authenticator{s.neverApprove, s.neverApprove, s.neverApprove},
			expectSuccessful: false,
		},
	}

	// Simulating a transaction
	var tx sdk.Tx
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			// Convert the authenticators to InitializationData
			initData := []authenticator.InitializationData{}
			for _, auth := range tc.authenticators {
				initData = append(initData, authenticator.InitializationData{
					AuthenticatorType: auth.Type(),
					Data:              testData,
				})
			}

			data, _ := json.Marshal(initData)
			initializedAuth, err := s.AnyOfAuth.Initialize(data)
			s.Require().NoError(err)

			// Attempt to authenticate using initialized authenticator
			authData, err := initializedAuth.GetAuthenticationData(s.Ctx, tx, -1, false)
			s.Require().NoError(err)

			success, err := initializedAuth.Authenticate(s.Ctx, nil, authData)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectSuccessful, success)
		})
	}
}
