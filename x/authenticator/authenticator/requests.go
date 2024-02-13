package authenticator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v21/x/authenticator/iface"
)

type TrackRequest struct {
	Account             sdk.AccAddress `json:"account"`
	Msg                 iface.LocalAny `json:"msg"`
	AuthenticatorParams []byte         `json:"authenticator_params,omitempty"`
}

type ConfirmExecutionRequest struct {
	Account             sdk.AccAddress `json:"account"`
	Msg                 iface.LocalAny `json:"msg"`
	AuthenticatorParams []byte         `json:"authenticator_params,omitempty"`
}
