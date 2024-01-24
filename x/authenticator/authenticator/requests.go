package authenticator

import sdk "github.com/cosmos/cosmos-sdk/types"

type SignModeData struct {
	Direct  []byte `json:"sign_mode_direct"`
	Textual string `json:"sign_mode_textual"`
}

type LocalAny struct {
	TypeURL string `json:"type_url"`
	Value   []byte `json:"value"`
}

type ExplicitTxData struct {
	ChainID       string     `json:"chain_id"`
	AccountNumber uint64     `json:"account_number"`
	Sequence      uint64     `json:"sequence"`
	TimeoutHeight uint64     `json:"timeout_height"`
	Msgs          []LocalAny `json:"msgs"`
	Memo          string     `json:"memo"`
}

type SimplifiedSignatureData struct {
	Signers    []sdk.AccAddress `json:"signers"`
	Signatures [][]byte         `json:"signatures"`
}

type AuthenticationRequest struct {
	Account             sdk.AccAddress          `json:"account"`
	Msg                 LocalAny                `json:"msg"`
	Signature           []byte                  `json:"signature"` // Only allowing messages with a single signer
	SignModeTxData      SignModeData            `json:"sign_mode_tx_data"`
	TxData              ExplicitTxData          `json:"tx_data"`
	SignatureData       SimplifiedSignatureData `json:"signature_data"`
	Simulate            bool                    `json:"simulate"`
	AuthenticatorParams []byte                  `json:"authenticator_params,omitempty"`
}

type TrackRequest struct {
	Account sdk.AccAddress `json:"account"`
	Msg     LocalAny       `json:"msg"`
}

type ConfirmExecutionRequest struct {
	Account sdk.AccAddress `json:"account"`
	Msg     LocalAny       `json:"msg"`
}
