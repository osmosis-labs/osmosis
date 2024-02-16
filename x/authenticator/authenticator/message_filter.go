package authenticator

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v21/x/authenticator/iface"
)

var _ iface.Authenticator = &MessageFilterAuthenticator{}

type MessageFilterAuthenticator struct {
	MsgTypes string
}

func NewMessageFilterAuthenticator() MessageFilterAuthenticator {
	return MessageFilterAuthenticator{}
}

func (m MessageFilterAuthenticator) Type() string {
	return "MessageFilterAuthenticator"
}

func (m MessageFilterAuthenticator) StaticGas() uint64 {
	return 0
}

func (m MessageFilterAuthenticator) Initialize(data []byte) (iface.Authenticator, error) {
	m.MsgTypes = string(data)
	return m, nil
}

func (m MessageFilterAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg) error {
	return nil
}

func (m MessageFilterAuthenticator) Authenticate(ctx sdk.Context, request iface.AuthenticationRequest) iface.AuthenticationResult {
	// Check that the string is a substring of the MsgTypes
	isAllowed := strings.Contains(m.MsgTypes, request.Msg.TypeURL)
	if !isAllowed {
		return iface.NotAuthenticated()
	}
	return iface.Authenticated()
}

func (m MessageFilterAuthenticator) ConfirmExecution(ctx sdk.Context, request iface.AuthenticationRequest) iface.ConfirmationResult {
	return iface.Confirm()
}

func (m MessageFilterAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("data is empty")
	}

	return nil
}

func (m MessageFilterAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
}
