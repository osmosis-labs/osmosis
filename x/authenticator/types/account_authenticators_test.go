package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/osmosis-labs/osmosis/v19/x/authenticator/types"
)

// Mock Authenticator for testing purposes
type MockAuthenticator struct {
	authType string
}

func (m MockAuthenticator) GetAuthenticationData(tx sdk.Tx, messageIndex uint8, simulate bool) (types.AuthenticatorData, error) {
	return "mock", nil
}

func (m MockAuthenticator) Authenticate(ctx sdk.Context, msg sdk.Msg, authenticationData types.AuthenticatorData) (bool, error) {
	return true, nil
}

func (m MockAuthenticator) ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticated bool, authenticationData types.AuthenticatorData) bool {
	return true
}

func (m MockAuthenticator) Type() string {
	return m.authType
}

var _ types.Authenticator = MockAuthenticator{}

func TestInitializeAuthenticators(t *testing.T) {
	types.ResetAuthenticators() // Reset the global state
	auth1 := MockAuthenticator{"type1"}
	auth2 := MockAuthenticator{"type2"}

	types.InitializeAuthenticators([]types.Authenticator{auth1, auth2})

	authenticators := types.GetRegisteredAuthenticators()
	require.Equal(t, 2, len(authenticators))
	require.Contains(t, authenticators, auth1)
	require.Contains(t, authenticators, auth2)
}

func TestRegisterAuthenticator(t *testing.T) {
	types.ResetAuthenticators() // Reset the global state
	auth3 := MockAuthenticator{"type3"}
	types.RegisterAuthenticator(auth3)
	require.True(t, types.IsAuthenticatorTypeRegistered("type3"))
}

func TestUnregisterAuthenticator(t *testing.T) {
	types.ResetAuthenticators() // Reset the global state
	auth2 := MockAuthenticator{"type2"}
	types.RegisterAuthenticator(auth2) // Register first to ensure it's there
	require.True(t, types.IsAuthenticatorTypeRegistered("type2"))
	types.UnregisterAuthenticator(auth2)
	require.False(t, types.IsAuthenticatorTypeRegistered("type2"))
}

func TestGetRegisteredAuthenticators(t *testing.T) {
	types.ResetAuthenticators() // Reset the global state
	expectedAuthTypes := []string{"type1", "type3"}
	unexpectedAuthTypes := []string{"type2"}

	authenticators := types.GetRegisteredAuthenticators()

	for _, auth := range authenticators {
		authType := auth.Type()
		require.Contains(t, expectedAuthTypes, authType)
		require.NotContains(t, unexpectedAuthTypes, authType)
	}
}

func TestAsAuthenticator(t *testing.T) {
	types.ResetAuthenticators() // Reset the global state

	// Register mock authenticator
	auth1 := MockAuthenticator{"type1"}
	types.RegisterAuthenticator(auth1)

	// Check if a registered authenticator type is recognized
	accountAuth := types.AccountAuthenticator{Type: "type1"}
	require.NotNil(t, accountAuth.AsAuthenticator(), "Expected a valid Authenticator for 'type1'")
	require.Equal(t, "type1", accountAuth.AsAuthenticator().Type())

	// Check for an unregistered authenticator type
	accountAuth = types.AccountAuthenticator{Type: "typeX"}
	require.Nil(t, accountAuth.AsAuthenticator(), "Didn't expect a valid Authenticator for 'typeX'")
}

// Second mock that always fails authentication
type MockAuthenticatorFail struct {
	authType string
}

func (m MockAuthenticatorFail) GetAuthenticationData(tx sdk.Tx, messageIndex uint8, simulate bool) (types.AuthenticatorData, error) {
	return "mock-fail", nil
}

func (m MockAuthenticatorFail) Authenticate(ctx sdk.Context, msg sdk.Msg, authenticationData types.AuthenticatorData) (bool, error) {
	return false, nil
}

func (m MockAuthenticatorFail) ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticated bool, authenticationData types.AuthenticatorData) bool {
	return true
}

func (m MockAuthenticatorFail) Type() string {
	return m.authType
}

// Ensure our mocks implement the Authenticator interface
var _ types.Authenticator = MockAuthenticator{}
var _ types.Authenticator = MockAuthenticatorFail{}

// Tests for the mocks behavior
func TestMockAuthenticators(t *testing.T) {
	types.ResetAuthenticators() // Reset the global state
	// Create instances of our mocks
	mockPass := MockAuthenticator{"type-pass"}
	mockFail := MockAuthenticatorFail{"type-fail"}

	// You may need to mock sdk.Tx, sdk.Msg, and sdk.Context based on their implementations
	var mockTx sdk.Tx
	var mockMsg sdk.Msg
	var mockCtx sdk.Context

	// Testing mockPass
	dataPass, _ := mockPass.GetAuthenticationData(mockTx, 0, false)
	isAuthenticatedPass, _ := mockPass.Authenticate(mockCtx, mockMsg, dataPass)
	require.True(t, isAuthenticatedPass)

	// Testing mockFail
	dataFail, _ := mockFail.GetAuthenticationData(mockTx, 0, false)
	isAuthenticatedFail, _ := mockFail.Authenticate(mockCtx, mockMsg, dataFail)
	require.False(t, isAuthenticatedFail)
}
