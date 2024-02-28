package authenticator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TrackRequest struct {
	AuthenticatorId     string         `json:"authenticator_id"`
	Account             sdk.AccAddress `json:"account"`
	Msg                 LocalAny       `json:"msg"`
	MsgIndex            uint64         `json:"msg_index"`
	AuthenticatorParams []byte         `json:"authenticator_params,omitempty"`
}

type ConfirmExecutionRequest struct {
	AuthenticatorId     string         `json:"authenticator_id"`
	Account             sdk.AccAddress `json:"account"`
	Msg                 LocalAny       `json:"msg"`
	MsgIndex            uint64         `json:"msg_index"`
	AuthenticatorParams []byte         `json:"authenticator_params,omitempty"`
}

type AuthenticationRequest struct {
	AuthenticatorId string         `json:"authenticator_id"`
	Account         sdk.AccAddress `json:"account"`
	Msg             LocalAny       `json:"msg"`

	// Since array size is int, and size depends on the system architecture,
	// we use uint64 to cover all available architectures.
	// It is unsigned, so at this point, it can't be negative.
	MsgIndex uint64 `json:"msg_index"`

	// Only allowing messages with a single signer
	Signature           []byte                  `json:"signature"`
	SignModeTxData      SignModeData            `json:"sign_mode_tx_data"`
	TxData              ExplicitTxData          `json:"tx_data"`
	SignatureData       SimplifiedSignatureData `json:"signature_data"`
	Simulate            bool                    `json:"simulate"`
	AuthenticatorParams []byte                  `json:"authenticator_params,omitempty"`
}
