package authenticator_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/iface"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/testutils"
	authenticatortypes "github.com/osmosis-labs/osmosis/v23/x/authenticator/types"
)

type AggregatedAuthenticatorsTest struct {
	BaseAuthenticatorSuite

	AnyOfAuth        authenticator.AnyOfAuthenticator
	AllOfAuth        authenticator.AllOfAuthenticator
	alwaysApprove    testutils.TestingAuthenticator
	neverApprove     testutils.TestingAuthenticator
	approveAndBlock  testutils.TestingAuthenticator
	rejectAndConfirm testutils.TestingAuthenticator
	spyAuth          testutils.SpyAuthenticator
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
	s.spyAuth = testutils.NewSpyAuthenticator(
		s.OsmosisApp.GetKVStoreKey()[authenticatortypes.AuthenticatorStoreKey],
	)

	am.RegisterAuthenticator(s.AnyOfAuth)
	am.RegisterAuthenticator(s.AllOfAuth)
	am.RegisterAuthenticator(s.alwaysApprove)
	am.RegisterAuthenticator(s.neverApprove)
	am.RegisterAuthenticator(s.approveAndBlock)
	am.RegisterAuthenticator(s.rejectAndConfirm)
	am.RegisterAuthenticator(s.spyAuth)
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
			expectConfirm:    true,
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
			expectConfirm:    true,
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
			expectConfirm:    true,
		},
		{
			name:             "alwaysApprove + neverApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.alwaysApprove, s.neverApprove, s.alwaysApprove},
			expectSuccessful: true,
			expectConfirm:    true,
		},
		{
			name:             "neverApprove + neverApprove + alwaysApprove",
			authenticators:   []iface.Authenticator{s.neverApprove, s.neverApprove, s.alwaysApprove},
			expectSuccessful: true,
			expectConfirm:    true,
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
			expectConfirm:    true,
		},
	}

	// Simulating a transaction
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			// Convert the authenticators to InitializationData
			initData := []authenticator.SubAuthenticatorInitData{}
			for _, auth := range tc.authenticators {
				initData = append(initData, authenticator.SubAuthenticatorInitData{
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
			request, err := authenticator.GenerateAuthenticationData(s.Ctx, ak, sigModeHandler, s.TestAccAddress[0], s.TestAccAddress[0], msg, tx, 0, false, authenticator.SequenceMatch)
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
			initData := []authenticator.SubAuthenticatorInitData{}
			for _, auth := range tc.authenticators {
				initData = append(initData, authenticator.SubAuthenticatorInitData{
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
			request, err := authenticator.GenerateAuthenticationData(s.Ctx, ak, sigModeHandler, s.TestAccAddress[0], s.TestAccAddress[0], msg, tx, 0, false, authenticator.SequenceMatch)
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
			request, err := authenticator.GenerateAuthenticationData(s.Ctx, ak, sigModeHandler, s.TestAccAddress[0], s.TestAccAddress[0], msg, tx, 0, false, authenticator.SequenceMatch)
			s.Require().NoError(err)

			err = initializedTop.Authenticate(s.Ctx, request)
			s.Require().Equal(tc.success, err == nil)
		})
	}
}

type CompositeSpyAuth struct {
	AnyOf []*CompositeSpyAuth
	AllOf []*CompositeSpyAuth
	Name  string
}

func allOf(sub ...*CompositeSpyAuth) *CompositeSpyAuth {
	return &CompositeSpyAuth{AllOf: sub}
}

func anyOf(sub ...*CompositeSpyAuth) *CompositeSpyAuth {
	return &CompositeSpyAuth{AnyOf: sub}
}

func root(name string) *CompositeSpyAuth {
	return &CompositeSpyAuth{Name: name}
}

func (csa *CompositeSpyAuth) Type() string {
	am := authenticator.NewAuthenticatorManager()

	if len(csa.Name) > 0 {
		return testutils.SpyAuthenticator{}.Type()
	} else if len(csa.AnyOf) > 0 {
		return authenticator.NewAnyOfAuthenticator(am).Type()
	} else if len(csa.AllOf) > 0 {
		return authenticator.NewAllOfAuthenticator(am).Type()
	}

	panic("unreachable")
}

func (csa *CompositeSpyAuth) buildInitData() ([]byte, error) {

	// root
	if len(csa.Name) > 0 {
		spyData := testutils.SpyAuthenticatorData{
			Name: csa.Name,
		}
		return json.Marshal(spyData)
	} else if len(csa.AnyOf) > 0 {
		var initData []authenticator.SubAuthenticatorInitData
		for _, subAuth := range csa.AnyOf {
			data, err := subAuth.buildInitData()
			if err != nil {
				return nil, err
			}

			initData = append(initData, authenticator.SubAuthenticatorInitData{
				AuthenticatorType: subAuth.Type(),
				Data:              data,
			})
		}

		return json.Marshal(initData)
	} else if len(csa.AllOf) > 0 {
		var initData []authenticator.SubAuthenticatorInitData
		for _, subAuth := range csa.AllOf {
			data, err := subAuth.buildInitData()
			if err != nil {
				return nil, err
			}

			initData = append(initData, authenticator.SubAuthenticatorInitData{
				AuthenticatorType: subAuth.Type(),
				Data:              data,
			})
		}

		return json.Marshal(initData)

	}

	return nil, fmt.Errorf("unreachable")
}

func (s *AggregatedAuthenticatorsTest) TestNestedAuthenticatorCalls() {
	// Define test cases
	type testCase struct {
		name          string
		compositeAuth CompositeSpyAuth
		names         []string
		id            string
		expectedIds   []string
	}

	testCases := []testCase{
		{
			name:          "AllOf(a)",
			compositeAuth: *allOf(root("a")),
			id:            "1",
			names:         []string{"a"},
			expectedIds:   []string{"1.0"},
		},
		{
			name:          "AllOf(a, b)",
			compositeAuth: *allOf(root("a"), root("b")),
			id:            "2",
			names:         []string{"a", "b"},
			expectedIds:   []string{"2.0", "2.1"},
		},
		{
			name:          "AllOf(AnyOf(a, b), c)",
			compositeAuth: *allOf(anyOf(root("a"), root("b")), root("c")),
			id:            "3",
			names:         []string{"a", "c"}, // b is not called because anyOf is short-circuited
			expectedIds:   []string{"3.0.0", "3.1"},
		},
		{
			name:          "AnyOf(AllOf(a, b), c)",
			compositeAuth: *anyOf(allOf(root("a"), root("b")), root("c")),
			id:            "4",
			names:         []string{"a", "b"}, // c is not called because allOf is short-circuited
			expectedIds:   []string{"4.0.0", "4.0.1"},
		},
		{
			name:          "AnyOf(c, AllOf(a, b))",
			compositeAuth: *anyOf(root("c"), allOf(root("a"), root("b"))),
			id:            "5",
			names:         []string{"c"}, // a and b are not called because allOf is short-circuited
			expectedIds:   []string{"5.0"},
		},
		{
			name:          "AnyOf(AllOf(AnyOf(a, b), c), AnyOf(d, e))",
			compositeAuth: *anyOf(allOf(anyOf(root("a"), root("b")), root("c")), anyOf(root("d"), root("e"))),
			id:            "6",
			names:         []string{"a"}, // b,c,d and e are not called because allOf is short-circuited
			expectedIds:   []string{"6.0.0.0"},
		},
	}

	for _, tc := range testCases {
		s.Ctx = s.Ctx.WithGasMeter(sdk.NewGasMeter(2_000_000))
		data, err := tc.compositeAuth.buildInitData()
		s.Require().NoError(err)

		var auth iface.Authenticator
		if len(tc.compositeAuth.AllOf) > 0 {
			auth, err = s.AllOfAuth.Initialize(data)
			s.Require().NoError(err)
		} else if len(tc.compositeAuth.AnyOf) > 0 {
			auth, err = s.AnyOfAuth.Initialize(data)
			s.Require().NoError(err)
		} else {
			panic("invalid compositeAuth")
		}

		// reset all spy authenticators that the test is checking
		for _, name := range tc.names {
			spy := testutils.SpyAuthenticator{KvStoreKey: s.spyAuth.KvStoreKey, Name: name}
			spy.ResetLatestCalls(s.Ctx)
		}

		msg := &bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 1))}

		encodedMsg, err := codectypes.NewAnyWithValue(msg)
		s.Require().NoError(err, "Should encode Any value successfully")

		// mock the authentication request
		authReq := iface.AuthenticationRequest{
			AuthenticatorId:     tc.id,
			Account:             s.TestAccAddress[0],
			FeePayer:            s.TestAccAddress[0],
			Msg:                 iface.LocalAny{TypeURL: encodedMsg.TypeUrl, Value: encodedMsg.Value},
			MsgIndex:            0,
			Signature:           []byte{1, 1, 1, 1, 1},
			SignModeTxData:      iface.SignModeData{Direct: []byte{1, 1, 1, 1, 1}},
			SignatureData:       iface.SimplifiedSignatureData{Signers: []sdk.AccAddress{s.TestAccAddress[0]}, Signatures: [][]byte{{1, 1, 1, 1, 1}}},
			Simulate:            false,
			AuthenticatorParams: []byte{1, 1, 1, 1, 1},
		}

		// make calls
		auth.OnAuthenticatorAdded(s.Ctx, authReq.Account, data, authReq.AuthenticatorId)
		auth.Authenticate(s.Ctx, authReq)
		auth.Track(s.Ctx, authReq.Account, authReq.FeePayer, msg, authReq.MsgIndex, tc.id)
		auth.ConfirmExecution(s.Ctx, authReq)
		auth.OnAuthenticatorRemoved(s.Ctx, authReq.Account, data, authReq.AuthenticatorId)

		// Check that the spy authenticator was called with the expected data
		for i, name := range tc.names {
			expectedAuthReq := authReq
			expectedAuthReq.AuthenticatorId = tc.expectedIds[i]

			spy := testutils.SpyAuthenticator{KvStoreKey: s.spyAuth.KvStoreKey, Name: name}
			latestCalls := spy.GetLatestCalls(s.Ctx)

			spyData, err := json.Marshal(testutils.SpyAuthenticatorData{Name: name})
			s.Require().NoError(err, "Should marshal spy data successfully")

			s.Require().Equal(
				testutils.SpyAddRequest{
					Account:         expectedAuthReq.Account,
					Data:            spyData,
					AuthenticatorId: expectedAuthReq.AuthenticatorId,
				},
				latestCalls.OnAuthenticatorAdded,
			)
			s.Require().Equal(expectedAuthReq, latestCalls.Authenticate)
			s.Require().Equal(
				testutils.SpyTrackRequest{
					AuthenticatorId: expectedAuthReq.AuthenticatorId,
					Account:         expectedAuthReq.Account,
					Msg:             expectedAuthReq.Msg,
					MsgIndex:        expectedAuthReq.MsgIndex,
				},
				latestCalls.Track,
			)
			s.Require().Equal(expectedAuthReq, latestCalls.ConfirmExecution)
			s.Require().Equal(
				testutils.SpyRemoveRequest{
					Account:         expectedAuthReq.Account,
					Data:            spyData,
					AuthenticatorId: expectedAuthReq.AuthenticatorId,
				},
				latestCalls.OnAuthenticatorRemoved,
			)
		}
	}

}

func marshalAuth(ta testAuth, testData []byte) ([]byte, error) {
	initData := []authenticator.SubAuthenticatorInitData{}

	for _, sub := range ta.subAuths {
		subData, err := marshalAuth(sub, testData)
		if err != nil {
			return nil, err
		}
		initData = append(initData, authenticator.SubAuthenticatorInitData{
			AuthenticatorType: sub.authenticator.Type(),
			Data:              subData,
		})
	}

	return json.Marshal(initData)
}
