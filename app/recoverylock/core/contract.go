package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

// Contract exits: a listed wallet may execute a contract only with one of the
// message shapes below. In cw20 and cw721 contracts they move a balance or a
// token to the recovery address or remove rights the wallet granted earlier.
// The contract itself is not checked: any contract that accepts one of these
// shapes runs with the wallet as sender. The message must be exactly one of
//
//	{"transfer":{"recipient":<recovery>,"amount":"<uint>"}}             (cw20)
//	{"transfer_nft":{"recipient":<recovery>,"token_id":"<string>"}}     (cw721)
//	{"decrease_allowance":{"spender":<address>,"amount":"<uint>"}}      (cw20)
//	{"revoke":{"spender":<address>,"token_id":"<string>"}}              (cw721)
//	{"revoke_all":{"operator":<address>}}                               (cw721)
//
// with no funds attached. Field names follow cw20-base 1.0.1 and cw721-base
// 0.16.0 (the contracts used in the app tests). decrease_allowance must not
// carry the optional expires field (setting an expiry is not a removal of
// rights). <address> is any address the chain's address codec accepts; the
// recipient of a transfer must equal the canonical recovery address byte for
// byte. The JSON is parsed strictly: duplicate keys at any level, unknown or
// missing fields, other value types and trailing data are refused. Funds are
// checked by the chain adapter.

var uintRe = regexp.MustCompile(`^[0-9]{1,39}$`)

type fieldKind int

const (
	fieldRecovery fieldKind = iota // must equal the recovery address
	fieldAddr                      // any valid address
	fieldUint                      // decimal integer string
	fieldString                    // non-empty string
)

// contractExits lists the allowed execute messages and their exact fields.
var contractExits = map[string]map[string]fieldKind{
	"transfer":           {"recipient": fieldRecovery, "amount": fieldUint},
	"transfer_nft":       {"recipient": fieldRecovery, "token_id": fieldString},
	"decrease_allowance": {"spender": fieldAddr, "amount": fieldUint},
	"revoke":             {"spender": fieldAddr, "token_id": fieldString},
	"revoke_all":         {"operator": fieldAddr},
}

// CheckContractExit checks a contract execute message against the allowed
// shapes. recovery is the canonical bech32 recovery address; validAddr
// reports whether a string is a valid account address for the chain.
func CheckContractExit(msg []byte, recovery string, validAddr func(string) error) error {
	v, err := StrictJSON(msg)
	if err != nil {
		return fmt.Errorf("contract message: %w", err)
	}
	top, ok := v.(map[string]any)
	if !ok || len(top) != 1 {
		return errors.New("contract message must be an object with exactly one key")
	}
	for k, x := range top {
		fields, ok := contractExits[k]
		if !ok {
			return fmt.Errorf("contract message %q is not allowed (only transfer and transfer_nft to the recovery address, decrease_allowance, revoke, revoke_all)", k)
		}
		body, ok := x.(map[string]any)
		if !ok {
			return fmt.Errorf("contract message %q must hold an object", k)
		}
		if len(body) != len(fields) {
			return fmt.Errorf("contract message %q must have exactly the fields %s", k, fieldNames(fields))
		}
		for name, kind := range fields {
			raw, present := body[name]
			if !present {
				return fmt.Errorf("contract message %q must have exactly the fields %s", k, fieldNames(fields))
			}
			val, ok := raw.(string)
			if !ok || val == "" {
				return fmt.Errorf("contract message %q field %s must be a non-empty string", k, name)
			}
			switch kind {
			case fieldRecovery:
				if val != recovery {
					return fmt.Errorf("contract message %q field %s must be the recovery address", k, name)
				}
			case fieldAddr:
				if validAddr == nil {
					return errors.New("no address check configured")
				}
				if err := validAddr(val); err != nil {
					return fmt.Errorf("contract message %q field %s: %w", k, name, err)
				}
			case fieldUint:
				if !uintRe.MatchString(val) {
					return fmt.Errorf("contract message %q field %s must be a decimal integer string", k, name)
				}
			}
		}
	}
	return nil
}

func fieldNames(f map[string]fieldKind) string {
	names := make([]string, 0, len(f))
	for n := range f {
		names = append(names, n)
	}
	sort.Strings(names)
	return strings.Join(names, " and ")
}

// StrictJSON decodes one JSON value and refuses duplicate object keys at any
// depth and any data after the value. Numbers are kept as json.Number.
func StrictJSON(b []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	v, err := strictValue(dec, 0)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, errors.New("trailing data after the JSON value")
	}
	return v, nil
}

const maxJSONDepth = 32

func strictValue(dec *json.Decoder, depth int) (any, error) {
	if depth > maxJSONDepth {
		return nil, errors.New("JSON nested too deeply")
	}
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch d := t.(type) {
	case json.Delim:
		switch d {
		case '{':
			m := map[string]any{}
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return nil, err
				}
				k, ok := kt.(string)
				if !ok {
					return nil, errors.New("object key is not a string")
				}
				if _, dup := m[k]; dup {
					return nil, fmt.Errorf("duplicate JSON key %q", k)
				}
				x, err := strictValue(dec, depth+1)
				if err != nil {
					return nil, err
				}
				m[k] = x
			}
			if _, err := dec.Token(); err != nil { // '}'
				return nil, err
			}
			return m, nil
		case '[':
			var a []any
			for dec.More() {
				x, err := strictValue(dec, depth+1)
				if err != nil {
					return nil, err
				}
				a = append(a, x)
			}
			if _, err := dec.Token(); err != nil { // ']'
				return nil, err
			}
			return a, nil
		}
		return nil, fmt.Errorf("unexpected %v", d)
	default:
		return t, nil
	}
}
