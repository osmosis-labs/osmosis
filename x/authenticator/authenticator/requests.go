package authenticator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/iface"
)

type TrackRequest struct {
	AuthenticatorId     string         `json:"authenticator_id"`
	Account             sdk.AccAddress `json:"account"`
	Msg                 iface.LocalAny `json:"msg"`
	AuthenticatorParams []byte         `json:"authenticator_params,omitempty"`
}

type ConfirmExecutionRequest struct {
	AuthenticatorId     string         `json:"authenticator_id"`
	Account             sdk.AccAddress `json:"account"`
	Msg                 iface.LocalAny `json:"msg"`
	AuthenticatorParams []byte         `json:"authenticator_params,omitempty"`
}
