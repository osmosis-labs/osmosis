package authenticator_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/v27/x/smart-account/authenticator"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/testutils"
	smartaccounttypes "github.com/osmosis-labs/osmosis/v27/x/smart-account/types"
)

type AggregatedAuthenticatorsTest struct {
	BaseAuthenticatorSuite

	AnyOfAuth        authenticator.AnyOf
	AllOfAuth        authenticator.AllOf
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
	s.AnyOfAuth = authenticator.NewAnyOf(am)
	s.AllOfAuth = authenticator.NewAllOf(am)
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
		s.OsmosisApp.GetKVStoreKey()[smartaccounttypes.StoreKey],
	)

	am.RegisterAuthenticator(s.AnyOfAuth)
	am.RegisterAuthenticator(s.AllOfAuth)
	am.RegisterAuthenticator(s.alwaysApprove)
	am.RegisterAuthenticator(s.neverApprove)
	am.RegisterAuthenticator(s.approveAndBlock)
	am.RegisterAuthenticator(s.rejectAndConfirm)
	am.RegisterAuthenticator(s.spyAuth)
}

func (s *AggregatedAuthenticatorsTest) TearDownTest() {
	os.RemoveAll(s.HomeDir)
}

func (s *AggregatedAuthenticatorsTest) TestAnyOf() {
	// Define data
	testData := []byte{}

	// Define test cases
	type testCase struct {
		name             string
		authenticators   []authenticator.Authenticator
		expectInit       bool
		expectSuccessful bool
		expectConfirm    bool
	}

	testCases := []testCase{
		{
			name:             "alwaysApprove + neverApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.neverApprove},
			expectInit:       true,
			expectSuccessful: true,
			expectConfirm:    true,
		},
		{
			name:             "neverApprove + neverApprove",
			authenticators:   []authenticator.Authenticator{s.neverApprove, s.neverApprove},
			expectInit:       true,
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "alwaysApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.alwaysApprove},
			expectInit:       true,
			expectSuccessful: true,
			expectConfirm:    true,
		},
		{
			name:             "neverApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.neverApprove, s.alwaysApprove},
			expectInit:       true,
			expectSuccessful: true,
			expectConfirm:    true,
		},
		{
			name:             "alwaysApprove + alwaysApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.alwaysApprove, s.alwaysApprove},
			expectInit:       true,
			expectSuccessful: true,
			expectConfirm:    true,
		},
		{
			name:             "alwaysApprove + alwaysApprove + neverApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.alwaysApprove, s.neverApprove},
			expectInit:       true,
			expectSuccessful: true,
			expectConfirm:    true,
		},
		{
			name:             "alwaysApprove + neverApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.neverApprove, s.alwaysApprove},
			expectInit:       true,
			expectSuccessful: true,
			expectConfirm:    true,
		},
		{
			name:             "neverApprove + neverApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.neverApprove, s.neverApprove, s.alwaysApprove},
			expectInit:       true,
			expectSuccessful: true,
			expectConfirm:    true,
		},
		{
			name:             "neverApprove + neverApprove + neverApprove",
			authenticators:   []authenticator.Authenticator{s.neverApprove, s.neverApprove, s.neverApprove},
			expectInit:       true,
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "approveAndBlock",
			authenticators:   []authenticator.Authenticator{s.approveAndBlock},
			expectInit:       false,
			expectSuccessful: true,
			expectConfirm:    false,
		},
		{
			name:             "rejectAndConfirm",
			authenticators:   []authenticator.Authenticator{s.rejectAndConfirm},
			expectInit:       false,
			expectSuccessful: false,
			expectConfirm:    true,
		},
		{
			name:             "approveAndBlock + rejectAndConfirm",
			authenticators:   []authenticator.Authenticator{s.approveAndBlock, s.rejectAndConfirm},
			expectInit:       true,
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
					Type:   auth.Type(),
					Config: testData,
				})
			}

			data, _ := json.Marshal(initData)
			initializedAuth, err := s.AnyOfAuth.Initialize(data)
			if !tc.expectInit {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				// Generate authentication request
				ak := s.OsmosisApp.AccountKeeper
				sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()
				// sample msg
				msg := &bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 1))}
				// sample tx
				tx, err := s.GenSimpleTx([]sdk.Msg{msg}, []cryptotypes.PrivKey{s.TestPrivKeys[0]})
				s.Require().NoError(err)
				request, err := authenticator.GenerateAuthenticationRequest(s.Ctx, s.OsmosisApp.AppCodec(), ak, sigModeHandler, s.TestAccAddress[0], s.TestAccAddress[0], nil, sdk.NewCoins(), msg, tx, 0, false, authenticator.SequenceMatch)
				s.Require().NoError(err)

				// Attempt to authenticate using initialized authenticator
				err = initializedAuth.Authenticate(s.Ctx, request)
				s.Require().Equal(tc.expectSuccessful, err == nil)

				err = initializedAuth.ConfirmExecution(s.Ctx, request)
				s.Require().Equal(tc.expectConfirm, err == nil)

			}
		})
	}
}

func (s *AggregatedAuthenticatorsTest) TestAllOf() {
	// Define data
	testData := []byte{}

	// Define test cases
	type testCase struct {
		name             string
		authenticators   []authenticator.Authenticator
		expectInit       bool
		expectSuccessful bool
		expectConfirm    bool
	}

	testCases := []testCase{
		{
			name:             "alwaysApprove + neverApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.neverApprove},
			expectInit:       true,
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "neverApprove + neverApprove",
			authenticators:   []authenticator.Authenticator{s.neverApprove, s.neverApprove},
			expectInit:       true,
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "alwaysApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.alwaysApprove},
			expectInit:       true,
			expectSuccessful: true,
			expectConfirm:    true,
		},
		{
			name:             "neverApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.neverApprove, s.alwaysApprove},
			expectInit:       true,
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "alwaysApprove + alwaysApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.alwaysApprove, s.alwaysApprove},
			expectInit:       true,
			expectSuccessful: true,
			expectConfirm:    true,
		},
		{
			name:             "alwaysApprove + alwaysApprove + neverApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.alwaysApprove, s.neverApprove},
			expectInit:       true,
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "alwaysApprove + neverApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.alwaysApprove, s.neverApprove, s.alwaysApprove},
			expectInit:       true,
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "neverApprove + neverApprove + alwaysApprove",
			authenticators:   []authenticator.Authenticator{s.neverApprove, s.neverApprove, s.alwaysApprove},
			expectInit:       true,
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "neverApprove + neverApprove + neverApprove",
			authenticators:   []authenticator.Authenticator{s.neverApprove, s.neverApprove, s.neverApprove},
			expectInit:       true,
			expectSuccessful: false,
			expectConfirm:    false,
		},
		{
			name:             "approveAndBlock",
			authenticators:   []authenticator.Authenticator{s.approveAndBlock},
			expectInit:       false,
			expectSuccessful: true,
			expectConfirm:    false,
		},
		{
			name:             "rejectAndConfirm",
			authenticators:   []authenticator.Authenticator{s.rejectAndConfirm},
			expectInit:       false,
			expectSuccessful: false,
			expectConfirm:    true,
		},
		{
			name:             "approveAndBlock + rejectAndConfirm",
			authenticators:   []authenticator.Authenticator{s.approveAndBlock, s.rejectAndConfirm},
			expectInit:       true,
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
					Type:   auth.Type(),
					Config: testData,
				})
			}

			data, _ := json.Marshal(initData)
			initializedAuth, err := s.AllOfAuth.Initialize(data)
			if !tc.expectInit {
				s.Require().Error(err)

			} else {
				s.Require().NoError(err)

				// Generate authentication request
				ak := s.OsmosisApp.AccountKeeper
				sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()

				// sample msg
				msg := &bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 1))}
				// sample tx
				tx, err := s.GenSimpleTx([]sdk.Msg{msg}, []cryptotypes.PrivKey{s.TestPrivKeys[0]})
				s.Require().NoError(err)
				cdc := s.OsmosisApp.AppCodec()
				request, err := authenticator.GenerateAuthenticationRequest(s.Ctx, cdc, ak, sigModeHandler, s.TestAccAddress[0], s.TestAccAddress[0], nil, sdk.NewCoins(), msg, tx, 0, false, authenticator.SequenceMatch)
				s.Require().NoError(err)

				// Attempt to authenticate using initialized authenticator
				err = initializedAuth.Authenticate(s.Ctx, request)
				s.Require().Equal(tc.expectSuccessful, err == nil)

				err = initializedAuth.ConfirmExecution(s.Ctx, request)
				s.Require().Equal(tc.expectConfirm, err == nil)

			}
		})
	}
}

type testAuth struct {
	name          string
	authenticator authenticator.Authenticator
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
		var auth authenticator.Authenticator
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
		{
			auth:    AnyOf(AllOf(always, always), AnyOf(always, never)),
			success: true,
		},
		{
			auth:    AllOf(AnyOf(always, never), AnyOf(never, always)),
			success: true,
		},
		{
			auth:    AllOf(AnyOf(never, never, never), AnyOf(never, always, never)),
			success: false,
		},
		{
			auth:    AnyOf(AnyOf(never, never, never), AnyOf(never, never, never)),
			success: false,
		},
		{
			auth:    AnyOf(AnyOf(never, never, never), AnyOf(never, never, never), AllOf(always, always)),
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
			request, err := authenticator.GenerateAuthenticationRequest(s.Ctx, s.OsmosisApp.AppCodec(), ak, sigModeHandler, s.TestAccAddress[0], s.TestAccAddress[0], nil, sdk.NewCoins(), msg, tx, 0, false, authenticator.SequenceMatch)
			s.Require().NoError(err)

			err = initializedTop.Authenticate(s.Ctx, request)
			s.Require().Equal(tc.success, err == nil)
		})
	}
}

type CompositeSpyAuth struct {
	anyOf []*CompositeSpyAuth
	allOf []*CompositeSpyAuth

	name        string
	failureFlag testutils.FailureFlag
}

func allOf(sub ...*CompositeSpyAuth) *CompositeSpyAuth {
	return &CompositeSpyAuth{allOf: sub}
}

func anyOf(sub ...*CompositeSpyAuth) *CompositeSpyAuth {
	return &CompositeSpyAuth{anyOf: sub}
}

func spy(name string) *CompositeSpyAuth {
	return &CompositeSpyAuth{name: name, failureFlag: 0}
}

func spyWithFailure(name string, failureFlag testutils.FailureFlag) *CompositeSpyAuth {
	return &CompositeSpyAuth{name: name, failureFlag: failureFlag}
}

func (csa *CompositeSpyAuth) Type() string {
	am := authenticator.NewAuthenticatorManager()

	if len(csa.name) > 0 {
		return testutils.SpyAuthenticator{}.Type()
	} else if len(csa.anyOf) > 0 {
		return authenticator.NewAnyOf(am).Type()
	} else if len(csa.allOf) > 0 {
		return authenticator.NewAllOf(am).Type()
	}

	panic("unreachable")
}

func (csa *CompositeSpyAuth) isSpy() bool {
	return len(csa.name) > 0
}

func (csa *CompositeSpyAuth) isAnyOf() bool {
	return len(csa.anyOf) > 0
}

func (csa *CompositeSpyAuth) isAllOf() bool {
	return len(csa.allOf) > 0
}

func (csa *CompositeSpyAuth) buildInitData() ([]byte, error) {

	// root
	if len(csa.name) > 0 {
		spyData := testutils.SpyAuthenticatorData{
			Name:    csa.name,
			Failure: csa.failureFlag,
		}
		return json.Marshal(spyData)
	} else if len(csa.anyOf) > 0 {
		var initData []authenticator.SubAuthenticatorInitData
		for _, subAuth := range csa.anyOf {
			data, err := subAuth.buildInitData()
			if err != nil {
				return nil, err
			}

			initData = append(initData, authenticator.SubAuthenticatorInitData{
				Type:   subAuth.Type(),
				Config: data,
			})
		}

		return json.Marshal(initData)
	} else if len(csa.allOf) > 0 {
		var initData []authenticator.SubAuthenticatorInitData
		for _, subAuth := range csa.allOf {
			data, err := subAuth.buildInitData()
			if err != nil {
				return nil, err
			}

			initData = append(initData, authenticator.SubAuthenticatorInitData{
				Type:   subAuth.Type(),
				Config: data,
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
			name:          "AllOf(a, b)",
			compositeAuth: *allOf(spy("a"), spy("b")),
			id:            "2",
			names:         []string{"a", "b"},
			expectedIds:   []string{"2.0", "2.1"},
		},
		{
			name:          "AllOf(AnyOf(a, b), c)",
			compositeAuth: *allOf(anyOf(spy("a"), spy("b")), spy("c")),
			id:            "3",
			names:         []string{"a", "c"}, // b is not called because anyOf is short-circuited
			expectedIds:   []string{"3.0.0", "3.1"},
		},
		{
			name:          "AnyOf(AllOf(a, b), c)",
			compositeAuth: *anyOf(allOf(spy("a"), spy("b")), spy("c")),
			id:            "4",
			names:         []string{"a", "b"}, // c is not called because allOf is short-circuited
			expectedIds:   []string{"4.0.0", "4.0.1"},
		},
		{
			name:          "AnyOf(c, AllOf(a, b))",
			compositeAuth: *anyOf(spy("c"), allOf(spy("a"), spy("b"))),
			id:            "5",
			names:         []string{"c"}, // a and b are not called because allOf is short-circuited
			expectedIds:   []string{"5.0"},
		},
		{
			name:          "AnyOf(AllOf(AnyOf(a, b), c), AnyOf(d, e))",
			compositeAuth: *anyOf(allOf(anyOf(spy("a"), spy("b")), spy("c")), anyOf(spy("d"), spy("e"))),
			id:            "6",
			names:         []string{"a"}, // b,c,d and e are not called because allOf is short-circuited
			expectedIds:   []string{"6.0.0.0"},
		},
	}

	for _, tc := range testCases {
		originalCtx := s.Ctx
		s.Ctx, _ = s.Ctx.WithGasMeter(storetypes.NewGasMeter(2_000_000)).CacheContext()
		data, err := tc.compositeAuth.buildInitData()
		s.Require().NoError(err)

		var auth authenticator.Authenticator
		if tc.compositeAuth.isAllOf() {
			auth, err = s.AllOfAuth.Initialize(data)
			s.Require().NoError(err)
		} else if tc.compositeAuth.isAnyOf() {
			auth, err = s.AnyOfAuth.Initialize(data)
			s.Require().NoError(err)
		} else {
			panic("top lv must be allOf or anyOf")
		}

		msg := &bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 1))}

		encodedMsg, err := codectypes.NewAnyWithValue(msg)
		s.Require().NoError(err, "Should encode Any value successfully")

		// mock the authentication request
		authReq := authenticator.AuthenticationRequest{
			AuthenticatorId:     tc.id,
			Account:             s.TestAccAddress[0],
			FeePayer:            s.TestAccAddress[0],
			FeeGranter:          nil,
			Fee:                 sdk.NewCoins(),
			Msg:                 authenticator.LocalAny{TypeURL: encodedMsg.TypeUrl, Value: encodedMsg.Value},
			MsgIndex:            0,
			Signature:           []byte{1, 1, 1, 1, 1},
			SignModeTxData:      authenticator.SignModeData{Direct: []byte{1, 1, 1, 1, 1}},
			SignatureData:       authenticator.SimplifiedSignatureData{Signers: []sdk.AccAddress{s.TestAccAddress[0]}, Signatures: [][]byte{{1, 1, 1, 1, 1}}},
			Simulate:            false,
			AuthenticatorParams: []byte{1, 1, 1, 1, 1},
		}

		// make calls
		auth.OnAuthenticatorAdded(s.Ctx, authReq.Account, data, authReq.AuthenticatorId)
		auth.Authenticate(s.Ctx, authReq)
		auth.Track(s.Ctx, authReq)
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

		s.Ctx = originalCtx
	}

}

// any_of can have failed sub-authenticators's confirm_execution but not failing the whole transaction
// that means that the failed sub-authenticators could write to the store if not handled properly
// which we don't want to happen since it breaks the semantics of not committing failed tx state
func (s *AggregatedAuthenticatorsTest) TestAnyOfNotWritingFailedSubAuthState() {
	// Define test cases
	type testCase struct {
		name            string
		compositeAuth   CompositeSpyAuth
		names           []string
		isStateReverted []bool
	}

	testCases := []testCase{
		{
			name:            "AnyOf(fail, pass)",
			compositeAuth:   *anyOf(spyWithFailure("fail", testutils.CONFIRM_EXECUTION_FAIL), spy("pass")),
			names:           []string{"fail", "pass"},
			isStateReverted: []bool{true, false},
		},
		{
			name: "AnyOf(fail_1, fail_2, pass)",
			compositeAuth: *anyOf(
				spyWithFailure("fail_1", testutils.CONFIRM_EXECUTION_FAIL),
				spyWithFailure("fail_2", testutils.CONFIRM_EXECUTION_FAIL),
				spy("pass"),
			),
			names:           []string{"fail_1", "fail_2", "pass"},
			isStateReverted: []bool{true, true, false},
		},
		{
			name: "AnyOf(fail, AllOf(fail_2, pass_1), pass_2)",
			compositeAuth: *anyOf(
				spyWithFailure("fail_1", testutils.CONFIRM_EXECUTION_FAIL),
				allOf(
					spyWithFailure("fail_2", testutils.CONFIRM_EXECUTION_FAIL),
					spy("pass_1"),
				),
				spy("pass_2"),
			),
			names: []string{"fail_1", "fail_2", "pass_1", "pass_2"},
			// pass_1 reverted since it's inside all of with failed auth
			isStateReverted: []bool{true, true, true, false},
		},
		{
			name: "AllOf(pass_1, AnyOf(fail_1, pass_2)",
			compositeAuth: *allOf(
				spy("pass_1"),
				anyOf(
					spyWithFailure("fail_1", testutils.CONFIRM_EXECUTION_FAIL),
					spy("pass_2"),
				),
			),
			names:           []string{"pass_1", "fail_1", "pass_2"},
			isStateReverted: []bool{false, true, false},
		},
		{
			name: "AnyOf(AnyOf(AllOf(pass_1, fail_1), pass_2), pass_3)",
			compositeAuth: *anyOf(
				anyOf(
					allOf(
						spy("pass_1"),
						spyWithFailure("fail_1", testutils.CONFIRM_EXECUTION_FAIL),
					),
					spy("pass_2"),
				),
				spy("pass_3"),
			),
			// pass_3 is short circuited, so ignored here
			names:           []string{"pass_1", "fail_1", "pass_2"},
			isStateReverted: []bool{true, true, false},
		},
	}

	for _, tc := range testCases {
		originalCtx := s.Ctx
		s.Ctx, _ = s.Ctx.WithGasMeter(storetypes.NewGasMeter(2_000_000)).CacheContext()
		data, err := tc.compositeAuth.buildInitData()
		s.Require().NoError(err)

		var auth authenticator.Authenticator
		if tc.compositeAuth.isAllOf() {
			auth, err = s.AllOfAuth.Initialize(data)
			s.Require().NoError(err)
		} else if tc.compositeAuth.isAnyOf() {
			auth, err = s.AnyOfAuth.Initialize(data)
			s.Require().NoError(err)
		} else {
			panic("top lv must be  anyOf")
		}

		msg := &bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 1))}

		encodedMsg, err := codectypes.NewAnyWithValue(msg)
		s.Require().NoError(err, "Should encode Any value successfully")

		// mock the authentication request
		authReq := authenticator.AuthenticationRequest{
			AuthenticatorId:     "1",
			Account:             s.TestAccAddress[0],
			FeePayer:            s.TestAccAddress[0],
			Msg:                 authenticator.LocalAny{TypeURL: encodedMsg.TypeUrl, Value: encodedMsg.Value},
			MsgIndex:            0,
			Signature:           []byte{1, 1, 1, 1, 1},
			SignModeTxData:      authenticator.SignModeData{Direct: []byte{1, 1, 1, 1, 1}},
			SignatureData:       authenticator.SimplifiedSignatureData{Signers: []sdk.AccAddress{s.TestAccAddress[0]}, Signatures: [][]byte{{1, 1, 1, 1, 1}}},
			Simulate:            false,
			AuthenticatorParams: []byte{1, 1, 1, 1, 1},
		}

		// make calls
		auth.ConfirmExecution(s.Ctx, authReq)

		for i, name := range tc.names {
			spy := testutils.SpyAuthenticator{KvStoreKey: s.spyAuth.KvStoreKey, Name: name}
			latestCalls := spy.GetLatestCalls(s.Ctx)
			latestConfirmExecuion := latestCalls.ConfirmExecution

			// NOTE: This assertion relying on the fact that latestCalls are stored in kv store
			// so if state is reverted, latest call will be empty.
			// This is not ideal since it could be interpreted as not being called at all
			// but it's good enough for now.
			//
			// Maybe using in mem naked Map to track calls instead and having kvstore
			// for actual state tracking would be better
			if tc.isStateReverted[i] {
				s.Require().Empty(latestConfirmExecuion)
			} else {
				s.Require().NotEmpty(latestConfirmExecuion)
			}

		}

		s.Ctx = originalCtx
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
			Type:   sub.authenticator.Type(),
			Config: subData,
		})
	}

	return json.Marshal(initData)
}
