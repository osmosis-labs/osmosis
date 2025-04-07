package authenticator

import (
	"encoding/json"
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
)

var _ Authenticator = &MessageFilter{}

// MessageFilter filters incoming messages based on a predefined JSON pattern.
// It allows for complex pattern matching to support advanced authentication flows.
type MessageFilter struct {
	encCfg  appparams.EncodingConfig
	pattern []byte
}

// NewMessageFilter creates a new MessageFilter with the provided EncodingConfig.
func NewMessageFilter(encCfg appparams.EncodingConfig) MessageFilter {
	return MessageFilter{
		encCfg: encCfg,
	}
}

// Type returns the type of the authenticator.
func (m MessageFilter) Type() string {
	return "MessageFilter"
}

// StaticGas returns the static gas amount for the authenticator. Currently, it's set to zero.
func (m MessageFilter) StaticGas() uint64 {
	return 0
}

// Initialize sets up the authenticator with the given data, which should be a valid JSON pattern for message filtering.
func (m MessageFilter) Initialize(config []byte) (Authenticator, error) {
	var jsonData json.RawMessage
	err := json.Unmarshal(config, &jsonData)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid json representation of message")
	}
	m.pattern = config
	return m, nil
}

// Track is a no-op in this implementation but can be used to track message handling.
func (m MessageFilter) Track(ctx sdk.Context, request AuthenticationRequest) error {
	return nil
}

// Authenticate checks if the provided message conforms to the set JSON pattern. It returns an AuthenticationResult based on the evaluation.
func (m MessageFilter) Authenticate(ctx sdk.Context, request AuthenticationRequest) error {
	// Get the concrete message from the interface registry
	protoResponseType, err := m.encCfg.InterfaceRegistry.Resolve(request.Msg.TypeURL)
	if err != nil {
		return errorsmod.Wrap(err, "failed to resolve message type")
	}

	// Unmarshal to bytes to the concrete proto message
	err = m.encCfg.Marshaler.Unmarshal(request.Msg.Value, protoResponseType)
	if err != nil {
		return errorsmod.Wrap(err, "failed to unmarshal message")
	}

	// Convert the proto message to JSON bytes for comparison to the Initialized data from the store
	jsonBz, err := m.encCfg.Marshaler.MarshalInterfaceJSON(protoResponseType)
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshal message to JSON")
	}

	// Check that the encoding is a superset of the pattern
	err = IsJsonSuperset(m.pattern, jsonBz)
	if err != nil {
		return errorsmod.Wrap(err, "message does not match pattern")
	}
	return nil
}

// ConfirmExecution confirms the execution of a message. Currently, it always confirms.
func (m MessageFilter) ConfirmExecution(ctx sdk.Context, request AuthenticationRequest) error {
	return nil
}

// OnAuthenticatorAdded performs additional checks when an authenticator is added. Specifically, it ensures numbers in JSON are encoded as strings.
func (m MessageFilter) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	var jsonData json.RawMessage
	err := json.Unmarshal(config, &jsonData)
	if err != nil {
		return errorsmod.Wrap(err, "invalid json representation of message")
	}
	hasFloats, err := containsFloats(config)
	if err != nil {
		return errorsmod.Wrap(err, "invalid json representation of message") // This should never happen
	}
	if hasFloats {
		return fmt.Errorf("invalid json representation of message. Numbers should be encoded as strings")
	}
	return nil
}

// OnAuthenticatorRemoved is a no-op in this implementation but can be used when an authenticator is removed.
func (m MessageFilter) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	return nil
}

// containsFloats checks if the given JSON data contains any floating point numbers.
func containsFloats(data []byte) (bool, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return false, err
	}

	return checkForFloats(v), nil
}

// checkForFloats recursively checks if any value in the given data is a floating point number.
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

// isSuperset checks if the first JSON structure is a superset of the second JSON structure.
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
			if decA, err := osmomath.NewDecFromStr(av); err == nil {
				if decB, err := osmomath.NewDecFromStr(bv); err == nil {
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

// IsJsonSuperset checks if the first JSON byte array is a superset of the second JSON byte array.
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
