package testutils

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v19/x/authenticator/iface"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ iface.Authenticator     = &TestingAuthenticator{}
	_ iface.AuthenticatorData = &TestingAuthenticatorData{}
)

type ApproveOn int

const (
	Always ApproveOn = iota
	Never
)

type (
	TestingAuthenticatorData struct{}
	TestingAuthenticator     struct {
		Approve        ApproveOn
		GasConsumption int
		BlockAddition  bool
		BlockRemoval   bool
	}
)

func (t TestingAuthenticator) Type() string {
	var when string
	if t.Approve == Always {
		when = "Always"
	} else {
		when = "Never"
	}
	return "TestingAuthenticator" + when + fmt.Sprintf("GasConsumption%d", t.GasConsumption) + fmt.Sprintf("BlockAddition%t", t.BlockAddition) + fmt.Sprintf("BlockRemoval%t", t.BlockRemoval)
}

func (t TestingAuthenticator) StaticGas() uint64 {
	return uint64(t.GasConsumption)
}

func (t TestingAuthenticator) Initialize(data []byte) (iface.Authenticator, error) {
	return t, nil
}

func (t TestingAuthenticator) GetAuthenticationData(ctx sdk.Context, tx sdk.Tx, messageIndex int, simulate bool) (iface.AuthenticatorData, error) {
	return TestingAuthenticatorData{}, nil
}

func (t TestingAuthenticator) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.AuthenticationResult {
	if t.Approve == Always {
		return iface.Authenticated()
	} else {
		return iface.NotAuthenticated()
	}
}

func (t TestingAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg) error {
	return nil
}

func (t TestingAuthenticator) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.ConfirmationResult {
	return iface.Confirm()
}

func (t TestingAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	if t.BlockAddition {
		return fmt.Errorf("authenticator could not be added")
	}
	return nil
}

func (t TestingAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	if t.BlockRemoval {
		return fmt.Errorf("authenticator could not be removed")
	}
	return nil
}
