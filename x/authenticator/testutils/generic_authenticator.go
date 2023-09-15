package testutils

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
)

var _ authenticator.Authenticator = &TestingAuthenticator{}
var _ authenticator.AuthenticatorData = &TestingAuthenticatorData{}

type ApproveOn int

const (
	Always ApproveOn = iota
	Never
)

type TestingAuthenticatorData struct{}
type TestingAuthenticator struct {
	Approve        ApproveOn
	GasConsumption int
}

func (t TestingAuthenticator) Type() string {
	var when string
	if t.Approve == Always {
		when = "Always"
	} else {
		when = "Never"
	}
	return "TestingAuthenticator" + when + fmt.Sprintf("GasConsumption%d", t.GasConsumption)
}

func (t TestingAuthenticator) StaticGas() uint64 {
	return uint64(t.GasConsumption)
}

func (t TestingAuthenticator) Initialize(data []byte) (authenticator.Authenticator, error) {
	return t, nil
}

func (t TestingAuthenticator) GetAuthenticationData(ctx sdk.Context, tx sdk.Tx, messageIndex int8, simulate bool) (authenticator.AuthenticatorData, error) {
	return TestingAuthenticatorData{}, nil
}

func (t TestingAuthenticator) Authenticate(ctx sdk.Context, msg sdk.Msg, authenticationData authenticator.AuthenticatorData) (bool, error) {
	if t.Approve == Always {
		return true, nil
	} else {
		return false, nil
	}
}

func (t TestingAuthenticator) AuthenticationFailed(ctx sdk.Context, authenticatorData authenticator.AuthenticatorData, msg sdk.Msg) {
}

func (t TestingAuthenticator) ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticated bool, authenticationData authenticator.AuthenticatorData) bool {
	return true
}
