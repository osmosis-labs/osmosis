package authenticator_test

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v19/app"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/iface"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/testutils"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"strings"
	"testing"
)

type AggregatedAuthenticatorsTest struct {
	suite.Suite
	Ctx           sdk.Context
	OsmosisApp    *app.OsmosisApp
	AnyOfAuth     authenticator.AnyOfAuthenticator
	AllOfAuth     authenticator.AllOfAuthenticator
	alwaysApprove testutils.TestingAuthenticator
	neverApprove  testutils.TestingAuthenticator
}

func TestAggregatedAuthenticatorsTest(t *testing.T) {
	suite.Run(t, new(AggregatedAuthenticatorsTest))
}

func (s *AggregatedAuthenticatorsTest) SetupTest() {
	s.OsmosisApp = app.Setup(false)
	s.Ctx = s.OsmosisApp.NewContext(false, tmproto.Header{})
	s.Ctx = s.Ctx.WithGasMeter(sdk.NewGasMeter(1_000_000))
	am := authenticator.NewAuthenticatorManager()

	// Define authenticators
	s.AnyOfAuth = authenticator.NewAnyOfAuthenticator(am)
	s.AllOfAuth = authenticator.NewAllOfAuthenticator(am)
	s.alwaysApprove = testutils.TestingAuthenticator{
		Approve:        testutils.Always,
		GasConsumption: 10,
	}
	s.neverApprove = testutils.TestingAuthenticator{
		Approve:        testutils.Never,
		GasConsumption: 10,
	}

	am.RegisterAuthenticator(s.AnyOfAuth)
	am.RegisterAuthenticator(s.AllOfAuth)
	am.RegisterAuthenticator(s.alwaysApprove)
	am.RegisterAuthenticator(s.neverApprove)
}

func (s *AggregatedAuthenticatorsTest) TestAnyOfAuthenticator() {
	// Define data
	testData := []byte{}

	// Define test cases
	type testCase struct {
		name             string
		authenticators   []iface.Authenticator
		expectSuccessful bool
	}

	testCases := []testCase{
		{
			name:             "alwaysApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.neverApprove},
			expectSuccessful: true,
		},
		{
			name:             "neverApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.neverApprove},
			expectSuccessful: false,
		},
		{
			name:             "alwaysApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.alwaysApprove},
			expectSuccessful: true,
		},
		{
			name:             "neverApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.alwaysApprove},
			expectSuccessful: true,
		},
		{
			name:             "alwaysApprove + alwaysApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.alwaysApprove, s.alwaysApprove},
			expectSuccessful: true,
		},
		{
			name:             "alwaysApprove + alwaysApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.alwaysApprove, s.neverApprove},
			expectSuccessful: true,
		},
		{
			name:             "alwaysApprove + neverApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.neverApprove, s.alwaysApprove},
			expectSuccessful: true,
		},
		{
			name:             "neverApprove + neverApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.neverApprove, s.alwaysApprove},
			expectSuccessful: true,
		},
		{
			name:             "neverApprove + neverApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.neverApprove, s.neverApprove},
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

			success := initializedAuth.Authenticate(s.Ctx, nil, nil, authData)
			s.Require().Equal(tc.expectSuccessful, success.IsAuthenticated())
		})
	}
}

func (s *AggregatedAuthenticatorsTest) TestAllOfAuthenticator() {
	// Define data
	testData := []byte{}

	// Define test cases
	type testCase struct {
		name             string
		authenticators   []iface.Authenticator
		expectSuccessful bool
	}

	testCases := []testCase{
		{
			name:             "alwaysApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.neverApprove},
			expectSuccessful: false,
		},
		{
			name:             "neverApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.neverApprove},
			expectSuccessful: false,
		},
		{
			name:             "alwaysApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.alwaysApprove},
			expectSuccessful: true,
		},
		{
			name:             "neverApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.alwaysApprove},
			expectSuccessful: false,
		},
		{
			name:             "alwaysApprove + alwaysApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.alwaysApprove, s.alwaysApprove},
			expectSuccessful: true,
		},
		{
			name:             "alwaysApprove + alwaysApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.alwaysApprove, s.neverApprove},
			expectSuccessful: false,
		},
		{
			name:             "alwaysApprove + neverApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.neverApprove, s.alwaysApprove},
			expectSuccessful: false,
		},
		{
			name:             "neverApprove + neverApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.neverApprove, s.alwaysApprove},
			expectSuccessful: false,
		},
		{
			name:             "neverApprove + neverApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.neverApprove, s.neverApprove},
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
			initializedAuth, err := s.AllOfAuth.Initialize(data)
			s.Require().NoError(err)

			// Attempt to authenticate using initialized authenticator
			authData, err := initializedAuth.GetAuthenticationData(s.Ctx, tx, -1, false)
			s.Require().NoError(err)

			success := initializedAuth.Authenticate(s.Ctx, nil, nil, authData)
			s.Require().Equal(tc.expectSuccessful, success.IsAuthenticated())
		})
	}
}

type testAuth struct {
	name          string
	authenticator iface.Authenticator
	subAuths      []testAuth
}

func (s *AggregatedAuthenticatorsTest) TestComposedAuthenticator() {
	testData := []byte{}

	// Helper function to create a name and a list of testAuths.
	createAuth := func(prefix string, sub ...testAuth) testAuth {
		var names []string
		for _, s := range sub {
			names = append(names, s.name)
		}
		name := prefix + "(" + strings.Join(names, ", ") + ")"
		var auth iface.Authenticator
		if prefix == "AnyOf" {
			auth = s.AnyOfAuth
		} else {
			auth = s.AllOfAuth
		}
		return testAuth{name: name, authenticator: auth, subAuths: sub}
	}

	// Shorthand functions using the helper.
	AnyOf := func(sub ...testAuth) testAuth {
		return createAuth("AnyOf", sub...)
	}

	AllOf := func(sub ...testAuth) testAuth {
		return createAuth("AllOf", sub...)
	}

	always := testAuth{name: "always", authenticator: s.alwaysApprove}
	never := testAuth{name: "never", authenticator: s.neverApprove}

	type testCase struct {
		auth    testAuth
		success bool
	}

	testCases := []testCase{
		//{
		//	auth:    AnyOf(AllOf(always, always), AnyOf(always, never)),
		//	success: true,
		//},
		//{
		//	auth:    AllOf(AnyOf(always, never), AnyOf(never, always)),
		//	success: true,
		//},
		//{
		//	auth:    AllOf(AnyOf(never, never, never), AnyOf(never, always, never)),
		//	success: false,
		//},
		//{
		//	auth:    AnyOf(AnyOf(never, never, never), AnyOf(never, never, never)),
		//	success: false,
		//},
		{
			auth:    AnyOf(AnyOf(never, never, never), AnyOf(never, never, never), AllOf(always)),
			success: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.auth.name, func(t *testing.T) {
			data, err := marshalAuth(tc.auth, testData)
			s.Require().NoError(err)

			initializedTop, err := tc.auth.authenticator.Initialize(data)
			s.Require().NoError(err)

			var tx sdk.Tx
			authData, err := initializedTop.GetAuthenticationData(s.Ctx, tx, -1, false)
			s.Require().NoError(err)

			success := initializedTop.Authenticate(s.Ctx, nil, nil, authData)
			s.Require().Equal(tc.success, success.IsAuthenticated())
		})
	}
}

func marshalAuth(ta testAuth, testData []byte) ([]byte, error) {
	initData := []authenticator.InitializationData{}

	for _, sub := range ta.subAuths {
		subData, err := marshalAuth(sub, testData)
		if err != nil {
			return nil, err
		}
		initData = append(initData, authenticator.InitializationData{
			AuthenticatorType: sub.authenticator.Type(),
			Data:              subData,
		})
	}

	return json.Marshal(initData)
}
