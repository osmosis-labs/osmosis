package types_test

import (
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/iface"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/types"
)

// Mock Authenticator for testing purposes
type MockAuthenticator struct {
	authType string
}

func (m MockAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
}

func (m MockAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg) error {
	return nil
}

func (m MockAuthenticator) Initialize(data []byte) (iface.Authenticator, error) {
	return m, nil
}

func (m MockAuthenticator) StaticGas() uint64 {
	return 1000
}

func (m MockAuthenticator) GetAuthenticationData(ctx sdk.Context, tx sdk.Tx, messageIndex int8, simulate bool) (iface.AuthenticatorData, error) {
	return "mock", nil
}

func (m MockAuthenticator) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.AuthenticationResult {
	return iface.Authenticated()
}

func (m MockAuthenticator) AuthenticationFailed(ctx sdk.Context, authenticatorData iface.AuthenticatorData, msg sdk.Msg) {
}

func (m MockAuthenticator) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.ConfirmationResult {
	return iface.Confirm()
}

func (m MockAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
}

func (m MockAuthenticator) Type() string {
	return m.authType
}

var _ iface.Authenticator = MockAuthenticator{}

func TestInitializeAuthenticators(t *testing.T) {
	am := authenticator.NewAuthenticatorManager()
	auth1 := MockAuthenticator{"type1"}
	auth2 := MockAuthenticator{"type2"}

	am.InitializeAuthenticators([]iface.Authenticator{auth1, auth2})

	authenticators := am.GetRegisteredAuthenticators()
	require.Equal(t, 2, len(authenticators))
	require.Contains(t, authenticators, auth1)
	require.Contains(t, authenticators, auth2)
}

func TestRegisterAuthenticator(t *testing.T) {
	am := authenticator.NewAuthenticatorManager()
	auth3 := MockAuthenticator{"type3"}
	am.RegisterAuthenticator(auth3)
	require.True(t, am.IsAuthenticatorTypeRegistered("type3"))
}

func TestUnregisterAuthenticator(t *testing.T) {
	am := authenticator.NewAuthenticatorManager()
	auth2 := MockAuthenticator{"type2"}
	am.RegisterAuthenticator(auth2) // Register first to ensure it's there
	require.True(t, am.IsAuthenticatorTypeRegistered("type2"))
	am.UnregisterAuthenticator(auth2)
	require.False(t, am.IsAuthenticatorTypeRegistered("type2"))
}

func TestGetRegisteredAuthenticators(t *testing.T) {
	am := authenticator.NewAuthenticatorManager()
	expectedAuthTypes := []string{"type1", "type3"}
	unexpectedAuthTypes := []string{"type2"}

	authenticators := am.GetRegisteredAuthenticators()

	for _, auth := range authenticators {
		authType := auth.Type()
		require.Contains(t, expectedAuthTypes, authType)
		require.NotContains(t, unexpectedAuthTypes, authType)
	}
}

func TestAsAuthenticator(t *testing.T) {
	am := authenticator.NewAuthenticatorManager()

	// Register mock authenticator
	auth1 := MockAuthenticator{"type1"}
	am.RegisterAuthenticator(auth1)

	// Check if a registered authenticator type is recognized
	accountAuth := types.AccountAuthenticator{Type: "type1"}
	require.NotNil(t, accountAuth.AsAuthenticator(am), "Expected a valid Authenticator for 'type1'")
	require.Equal(t, "type1", accountAuth.AsAuthenticator(am).Type())

	// Check for an unregistered authenticator type
	accountAuth = types.AccountAuthenticator{Type: "typeX"}
	require.Nil(t, accountAuth.AsAuthenticator(am), "Didn't expect a valid Authenticator for 'typeX'")
}

// Second mock that always fails authentication
type MockAuthenticatorFail struct {
	authType string
}

func (m MockAuthenticatorFail) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
}

func (m MockAuthenticatorFail) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
}

func (m MockAuthenticatorFail) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg) error {
	return nil
}

func (m MockAuthenticatorFail) Initialize(data []byte) (iface.Authenticator, error) {
	return m, nil
}

func (m MockAuthenticatorFail) StaticGas() uint64 {
	return 1000
}

func (m MockAuthenticatorFail) GetAuthenticationData(ctx sdk.Context, tx sdk.Tx, messageIndex int8, simulate bool) (iface.AuthenticatorData, error) {
	return "mock-fail", nil
}

func (m MockAuthenticatorFail) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.AuthenticationResult {
	return iface.NotAuthenticated()
}

func (m MockAuthenticatorFail) AuthenticationFailed(ctx sdk.Context, authenticatorData iface.AuthenticatorData, msg sdk.Msg) {
}

func (m MockAuthenticatorFail) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.ConfirmationResult {
	return iface.Confirm()
}

func (m MockAuthenticatorFail) Type() string {
	return m.authType
}

// Ensure our mocks implement the Authenticator interface
var _ iface.Authenticator = MockAuthenticator{}
var _ iface.Authenticator = MockAuthenticatorFail{}

// Tests for the mocks behavior
func TestMockAuthenticators(t *testing.T) {
	// Create instances of our mocks
	mockPass := MockAuthenticator{"type-pass"}
	mockFail := MockAuthenticatorFail{"type-fail"}

	// You may need to mock sdk.Tx, sdk.Msg, and sdk.Context based on their implementations
	var mockTx sdk.Tx
	var mockMsg sdk.Msg
	var mockCtx sdk.Context

	// Testing mockPass
	dataPass, _ := mockPass.GetAuthenticationData(mockCtx, mockTx, 0, false)
	isAuthenticatedPass := mockPass.Authenticate(mockCtx, nil, mockMsg, dataPass)
	require.True(t, isAuthenticatedPass.IsAuthenticated())

	// Testing mockFail
	dataFail, _ := mockFail.GetAuthenticationData(mockCtx, mockTx, 0, false)
	isAuthenticatedFail := mockFail.Authenticate(mockCtx, nil, mockMsg, dataFail)
	require.False(t, isAuthenticatedFail.IsAuthenticated())
}
