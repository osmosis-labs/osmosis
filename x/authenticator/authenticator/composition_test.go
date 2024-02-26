package authenticator_test

import (
	"encoding/json"
	"strings"
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/iface"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/testutils"
)

type AggregatedAuthenticatorsTest struct {
	BaseAuthenticatorSuite

	AnyOfAuth        authenticator.AnyOfAuthenticator
	AllOfAuth        authenticator.AllOfAuthenticator
	alwaysApprove    testutils.TestingAuthenticator
	neverApprove     testutils.TestingAuthenticator
	approveAndBlock  testutils.TestingAuthenticator
	rejectAndConfirm testutils.TestingAuthenticator
}

func TestAggregatedAuthenticatorsTest(t *testing.T) {
	suite.Run(t, new(AggregatedAuthenticatorsTest))
}

func (s *AggregatedAuthenticatorsTest) SetupTest() {
	s.SetupKeys()
	am := authenticator.NewAuthenticatorManager()

	// Define authenticators
	s.AnyOfAuth = authenticator.NewAnyOfAuthenticator(am)
	s.AllOfAuth = authenticator.NewAllOfAuthenticator(am)
	s.alwaysApprove = testutils.TestingAuthenticator{
		Approve:        testutils.Always,
		GasConsumption: 10,
		Confirm:        testutils.Always,
	}
	s.neverApprove = testutils.TestingAuthenticator{
		Approve:        testutils.Never,
		GasConsumption: 10,
		Confirm:        testutils.Never,
	}
	s.approveAndBlock = testutils.TestingAuthenticator{
		Approve:        testutils.Always,
		GasConsumption: 10,
		Confirm:        testutils.Never,
	}
	s.rejectAndConfirm = testutils.TestingAuthenticator{
		Approve:        testutils.Never,
		GasConsumption: 10,
		Confirm:        testutils.Always,
	}

	am.RegisterAuthenticator(s.AnyOfAuth)
	am.RegisterAuthenticator(s.AllOfAuth)
	am.RegisterAuthenticator(s.alwaysApprove)
	am.RegisterAuthenticator(s.neverApprove)
	am.RegisterAuthenticator(s.approveAndBlock)
	am.RegisterAuthenticator(s.rejectAndConfirm)
}

func (s *AggregatedAuthenticatorsTest) TestAnyOfAuthenticator() {
	// Define data
	testData := []byte{}

	// Define test cases
	type testCase struct {
		name             string
		authenticators   []iface.Authenticator
		expectSuccessful bool
		expectConfirm    bool
	}

	testCases := []testCase{
		{
			name:             "alwaysApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.neverApprove},
			expectSuccessful: true,
			expectConfirm:    false,
		},
		{
			name:             "neverApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.neverApprove},
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "alwaysApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.alwaysApprove},
			expectSuccessful: true,
			expectConfirm:    true,
		},
		{
			name:             "neverApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.alwaysApprove},
			expectSuccessful: true,
			expectConfirm:    false,
		},
		{
			name:             "alwaysApprove + alwaysApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.alwaysApprove, s.alwaysApprove},
			expectSuccessful: true,
			expectConfirm:    true,
		},
		{
			name:             "alwaysApprove + alwaysApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.alwaysApprove, s.neverApprove},
			expectSuccessful: true,
			expectConfirm:    false,
		},
		{
			name:             "alwaysApprove + neverApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.neverApprove, s.alwaysApprove},
			expectSuccessful: true,
			expectConfirm:    false,
		},
		{
			name:             "neverApprove + neverApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.neverApprove, s.alwaysApprove},
			expectSuccessful: true,
			expectConfirm:    false,
		},
		{
			name:             "neverApprove + neverApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.neverApprove, s.neverApprove},
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "approveAndBlock",
			authenticators:   []iface.Authenticator{s.approveAndBlock},
			expectSuccessful: true,
			expectConfirm:    false,
		},
		{
			name:             "rejectAndConfirm",
			authenticators:   []iface.Authenticator{s.rejectAndConfirm},
			expectSuccessful: false,
			expectConfirm:    true,
		},
		{
			name:             "approveAndBlock + rejectAndConfirm",
			authenticators:   []iface.Authenticator{s.approveAndBlock, s.rejectAndConfirm},
			expectSuccessful: true,
			expectConfirm:    false,
		},
	}

	// Simulating a transaction
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

			// Generate authentication request
			ak := s.OsmosisApp.AccountKeeper
			sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()
			// sample msg
			msg := &bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 1))}
			// sample tx
			tx, err := s.GenSimpleTx([]sdk.Msg{msg}, []cryptotypes.PrivKey{s.TestPrivKeys[0]})
			s.Require().NoError(err)
			request, err := authenticator.GenerateAuthenticationData(s.Ctx, ak, sigModeHandler, s.TestAccAddress[0], msg, tx, 0, false, authenticator.SequenceMatch)
			s.Require().NoError(err)

			// Attempt to authenticate using initialized authenticator
			err = initializedAuth.Authenticate(s.Ctx, request)
			s.Require().Equal(tc.expectSuccessful, err == nil)

			err = initializedAuth.ConfirmExecution(s.Ctx, request)
			s.Require().Equal(tc.expectConfirm, err == nil)
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
		expectConfirm    bool
	}

	testCases := []testCase{
		{
			name:             "alwaysApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.neverApprove},
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "neverApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.neverApprove},
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "alwaysApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.alwaysApprove},
			expectSuccessful: true,
			expectConfirm:    true,
		},
		{
			name:             "neverApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.alwaysApprove},
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "alwaysApprove + alwaysApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.alwaysApprove, s.alwaysApprove},
			expectSuccessful: true,
			expectConfirm:    true,
		},
		{
			name:             "alwaysApprove + alwaysApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.alwaysApprove, s.neverApprove},
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "alwaysApprove + neverApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.neverApprove, s.alwaysApprove},
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "neverApprove + neverApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.neverApprove, s.alwaysApprove},
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "neverApprove + neverApprove + neverApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.neverApprove, s.neverApprove},
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "approveAndBlock",
			authenticators:   []iface.Authenticator{s.approveAndBlock},
			expectSuccessful: true,
			expectConfirm:    false,
		},
		{
			name:             "rejectAndConfirm",
			authenticators:   []iface.Authenticator{s.rejectAndConfirm},
			expectSuccessful: false,
			expectConfirm:    true,
		},
		{
			name:             "approveAndBlock + rejectAndConfirm",
			authenticators:   []iface.Authenticator{s.approveAndBlock, s.rejectAndConfirm},
			expectSuccessful: false,
			expectConfirm:    false,
		},
	}

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

			// Generate authentication request
			ak := s.OsmosisApp.AccountKeeper
			sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()

			// sample msg
			msg := &bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 1))}
			// sample tx
			tx, err := s.GenSimpleTx([]sdk.Msg{msg}, []cryptotypes.PrivKey{s.TestPrivKeys[0]})
			s.Require().NoError(err)
			request, err := authenticator.GenerateAuthenticationData(s.Ctx, ak, sigModeHandler, s.TestAccAddress[0], msg, tx, 0, false, authenticator.SequenceMatch)
			s.Require().NoError(err)

			// Attempt to authenticate using initialized authenticator
			err = initializedAuth.Authenticate(s.Ctx, request)
			s.Require().Equal(tc.expectSuccessful, err == nil)

			err = initializedAuth.ConfirmExecution(s.Ctx, request)
			s.Require().Equal(tc.expectConfirm, err == nil)
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

			// Generate authentication request
			ak := s.OsmosisApp.AccountKeeper
			sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()
			// sample msg
			msg := &bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 1))}
			// sample tx
			tx, err := s.GenSimpleTx([]sdk.Msg{msg}, []cryptotypes.PrivKey{s.TestPrivKeys[0]})
			s.Require().NoError(err)
			request, err := authenticator.GenerateAuthenticationData(s.Ctx, ak, sigModeHandler, s.TestAccAddress[0], msg, tx, 0, false, authenticator.SequenceMatch)
			s.Require().NoError(err)

			err = initializedTop.Authenticate(s.Ctx, request)
			s.Require().Equal(tc.success, err == nil)
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
