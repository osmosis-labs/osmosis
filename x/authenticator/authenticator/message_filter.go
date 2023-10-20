package authenticator

import (
	"encoding/json"
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
	pattern []byte
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
	Value   json.RawMessage `json:"value"`
}

// I don't think we actually need this since numbers get encoded as strings, but adding it to avoid non-determinism if they were to exist
func compareNumbersAsDecs(a, b json.Number) bool {
	decA, errA := sdk.NewDecFromStr(string(a))
	decB, errB := sdk.NewDecFromStr(string(b))

	// If any of the numbers couldn't be parsed, consider them non-matching
	if errA != nil || errB != nil {
		return false
	}

	return decA.Equal(decB)
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
		Value:   jsonMsg,
	})
	if err != nil {
		return iface.NotAuthenticated()
	}

	opts := jsondiff.DefaultJSONOptions()
	opts.CompareNumbers = compareNumbersAsDecs

	// Check that the encoding is a superset of the pattern
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
