package authenticator

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/nsf/jsondiff"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/iface"
)

// Compile time type assertion for the SignatureData using the
// PassKeyAuthenticator struct
var _ iface.Authenticator = &MessageFilterAuthenticator{}

type MessageFilterAuthenticator struct {
	cdc codec.Codec

	pattern []byte
}

func NewMessageFilterAuthenticator(cdc codec.Codec) MessageFilterAuthenticator {
	return MessageFilterAuthenticator{cdc: cdc}
}

func (m MessageFilterAuthenticator) Type() string {
	return "MessageFilterAuthenticator"
}

func (m MessageFilterAuthenticator) StaticGas() uint64 {
	return 0
}

func (m MessageFilterAuthenticator) Initialize(data []byte) (iface.Authenticator, error) {
	var jsonData json.RawMessage
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid json representation of message")
	}
	m.pattern = data
	return m, nil
}

func (m MessageFilterAuthenticator) GetAuthenticationData(ctx sdk.Context, tx sdk.Tx, messageIndex int, simulate bool) (iface.AuthenticatorData, error) {
	return iface.EmptyAuthenticationData{}, nil
}

func (m MessageFilterAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg) error {
	return nil
}

type EncodedMsg struct {
	MsgType string          `json:"type"`
	Msg     json.RawMessage `json:"value"`
}

func (m MessageFilterAuthenticator) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.AuthenticationResult {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return iface.NotAuthenticated()
	}

	msgAsAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return iface.NotAuthenticated()
	}

	encodedMsg, err := json.Marshal(EncodedMsg{
		MsgType: msgAsAny.TypeUrl,
		Msg:     jsonMsg,
	})
	if err != nil {
		return iface.NotAuthenticated()
	}
	fmt.Println(string(encodedMsg))
	fmt.Println(string(m.pattern))

	opts := jsondiff.DefaultJSONOptions()
	// Is encodedMsg a superset of m.pattern?
	diff, _ := jsondiff.Compare(encodedMsg, m.pattern, &opts)
	if diff == jsondiff.FullMatch || diff == jsondiff.SupersetMatch {
		return iface.Authenticated()
	}
	return iface.NotAuthenticated()
}

func (m MessageFilterAuthenticator) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.ConfirmationResult {
	return iface.Confirm()
}

func (m MessageFilterAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	err := json.Unmarshal(data, &m.pattern)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid json representation of message")
	}
	return nil
}

func (m MessageFilterAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
}
