package authenticator

import (
	"encoding/json"
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v21/x/authenticator/iface"
)

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
		return nil, errorsmod.Wrap(err, "invalid json representation of message")
	}
	m.pattern = data
	return m, nil
}

func (m MessageFilterAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticatorId uint64) error {
	return nil
}

type EncodedMsg struct {
	MsgType string          `json:"type"`
	Value   json.RawMessage `json:"value"`
}

func containsFloats(data []byte) (bool, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return false, err
	}

	return checkForFloats(v), nil
}

func checkForFloats(v interface{}) bool {
	switch vv := v.(type) {
	case map[string]interface{}:
		for _, val := range vv {
			if checkForFloats(val) {
				return true
			}
		}
	case []interface{}:
		for _, val := range vv {
			if checkForFloats(val) {
				return true
			}
		}
	case float64:
		return true
	}
	return false
}

func isSuperset(a, b interface{}) error {
	switch av := a.(type) {
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected map, got %T", b)
		}

		// Sort the keys of 'a' to avoid non-determinism
		var sortedKeysA []string
		for key := range av {
			sortedKeysA = append(sortedKeysA, key)
		}
		sort.Strings(sortedKeysA)

		for _, key := range sortedKeysA {
			valA := av[key]
			valB, exists := bv[key]
			if !exists {
				return fmt.Errorf("key %s missing from second map", key)
			}
			if err := isSuperset(valA, valB); err != nil {
				return err
			}
		}

	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok {
			return fmt.Errorf("expected slice, got %T", b)
		}

		// TODO: Do we want to allow subset/superset checks here or require both slices to be the same?
		// If we want to allow this, maybe we want to either:
		//      1. Treat them like a set
		// 	    2. Ensure all elements of A are in B and in the same order, but B can have more elements
		if len(av) > len(bv) {
			return fmt.Errorf("first slice has more elements than second slice")
		}

		for i, valA := range av {
			if err := isSuperset(valA, bv[i]); err != nil {
				return err
			}
		}

	case string:
		if bv, ok := b.(string); ok {
			// Attempt to treat strings as numbers if they look like numbers
			if decA, err := sdk.NewDecFromStr(av); err == nil {
				if decB, err := sdk.NewDecFromStr(bv); err == nil {
					if !decA.Equal(decB) {
						return fmt.Errorf("numbers do not match: %s != %s", decA, decB)
					}
				}
			}
			if av != bv {
				return fmt.Errorf("strings do not match: %s != %s", av, bv)
			}
		} else {
			return fmt.Errorf("expected string, got %T", b)
		}

	case bool:
		if bv, ok := b.(bool); !ok || av != bv {
			return fmt.Errorf("booleans do not match or wrong type: %v != %v", av, b)
		}

	case nil:
		if b != nil {
			return fmt.Errorf("expected null, got %T", b)
		}

	case float64:
		return fmt.Errorf("numbers encoded as floats are not allowed. They should be encoded as strings")

	default:
		return fmt.Errorf("unexpected type %T", a)
	}

	return nil
}

func IsJsonSuperset(a, b []byte) error {
	var av, bv interface{}

	errA := json.Unmarshal(a, &av)
	if errA != nil {
		return fmt.Errorf("error unmarshaling first JSON: %v", errA)
	}
	errB := json.Unmarshal(b, &bv)
	if errB != nil {
		return fmt.Errorf("error unmarshaling second JSON: %v", errB)
	}

	return isSuperset(av, bv)
}

type SimplifiedMsg struct {
	MsgType string          `json:"type"`
	Value   json.RawMessage `json:"value"`
}

func (m MessageFilterAuthenticator) Authenticate(ctx sdk.Context, request iface.AuthenticationRequest) iface.AuthenticationResult {
	encodedMsg, err := json.Marshal(SimplifiedMsg{
		MsgType: request.Msg.TypeURL,
		Value:   request.Msg.Value,
	})

	if err != nil {
		return iface.NotAuthenticated()
	}
	// Check that the encoding is a superset of the pattern
	err = IsJsonSuperset(m.pattern, encodedMsg)
	if err != nil {
		// emit an event with the error
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"AuthenticatorError",
				sdk.NewAttribute("error", err.Error()),
			),
		)

		return iface.NotAuthenticated()
	}
	return iface.Authenticated()
}

func (m MessageFilterAuthenticator) ConfirmExecution(ctx sdk.Context, request iface.AuthenticationRequest) iface.ConfirmationResult {
	return iface.Confirm()
}

func (m MessageFilterAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	var jsonData json.RawMessage
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		return errorsmod.Wrap(err, "invalid json representation of message")
	}
	hasFloats, err := containsFloats(data)
	if err != nil {
		return errorsmod.Wrap(err, "invalid json representation of message") // This should never happen
	}
	if hasFloats {
		return fmt.Errorf("invalid json representation of message. Numbers should be encoded as strings")
	}
	return nil
}

func (m MessageFilterAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
}
