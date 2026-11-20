// Package core holds the chain independent part of the recovery lock:
// configuration, the active window, the listed set, the allowed message
// registry and the strict check of contract exit messages. It has no
// dependency on any Cosmos SDK version.
package core

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
)

// AddrLen is the only accepted address length (secp256k1 and Ethereum style
// account addresses). Contract addresses are 32 bytes and fail this check.
// Module account addresses are also 20 bytes; the chain configuration must
// refuse them separately. None of them may be listed.
const AddrLen = 20

// MaxFeePayers caps the fee payer set.
const MaxFeePayers = 16

// Config is compiled into the binary of each chain.
type Config struct {
	listed   map[string]struct{}
	sorted   [][]byte
	recovery []byte
	payers   map[string]struct{}
	start    int64
	expiry   int64
}

// NewConfig validates and builds a configuration.
//   - listed: account address bytes of the compromised wallets
//   - recovery: address bytes of the only allowed destination
//   - feePayers: address bytes of the accounts allowed to pay fees for
//     transactions that involve a listed address (1 to MaxFeePayers)
//   - activationHeight: first block height at which the lock applies
//   - expiryUnix: block time (unix seconds) from which the lock stops applying
func NewConfig(listed [][]byte, recovery []byte, feePayers [][]byte, activationHeight, expiryUnix int64) (*Config, error) {
	if len(listed) == 0 {
		return nil, errors.New("recovery lock: empty list")
	}
	if len(feePayers) == 0 || len(feePayers) > MaxFeePayers {
		return nil, fmt.Errorf("recovery lock: need 1 to %d fee payers, got %d", MaxFeePayers, len(feePayers))
	}
	if activationHeight <= 0 || expiryUnix <= 0 {
		return nil, fmt.Errorf("recovery lock: invalid activation height %d or expiry time %d", activationHeight, expiryUnix)
	}
	if err := checkLen("recovery address", recovery); err != nil {
		return nil, err
	}
	c := &Config{
		listed:   make(map[string]struct{}, len(listed)),
		recovery: clone(recovery),
		payers:   make(map[string]struct{}, len(feePayers)),
		start:    activationHeight,
		expiry:   expiryUnix,
	}
	for i, a := range listed {
		if err := checkLen(fmt.Sprintf("listed address %d", i), a); err != nil {
			return nil, err
		}
		if _, dup := c.listed[string(a)]; dup {
			continue
		}
		c.listed[string(a)] = struct{}{}
		c.sorted = append(c.sorted, clone(a))
	}
	for i, p := range feePayers {
		if err := checkLen(fmt.Sprintf("fee payer %d", i), p); err != nil {
			return nil, err
		}
		if c.IsListed(p) || bytes.Equal(p, recovery) {
			return nil, errors.New("recovery lock: a fee payer must not be listed or be the recovery address")
		}
		c.payers[string(p)] = struct{}{}
	}
	if c.IsListed(c.recovery) {
		return nil, errors.New("recovery lock: the recovery address must not be listed")
	}
	sort.Slice(c.sorted, func(i, j int) bool { return bytes.Compare(c.sorted[i], c.sorted[j]) < 0 })
	return c, nil
}

func checkLen(what string, b []byte) error {
	if len(b) != AddrLen {
		return fmt.Errorf("recovery lock: %s has %d bytes, want %d", what, len(b), AddrLen)
	}
	return nil
}

func clone(b []byte) []byte { return append([]byte(nil), b...) }

// Active reports whether the lock applies to a block at this height and block
// time. Callers must check this first and do nothing else when it is false
// (no state reads, no gas, no events).
func (c *Config) Active(height, unixTime int64) bool {
	return height >= c.start && unixTime < c.expiry
}

// HeightActive reports whether the height is at or after activation, ignoring
// expiry. Used only where the block time is unknown (CheckTx right after a
// restart), which makes the mempool stricter, never consensus.
func (c *Config) HeightActive(height int64) bool { return height >= c.start }

// IsListed reports whether addr is a compromised wallet.
func (c *Config) IsListed(addr []byte) bool {
	_, ok := c.listed[string(addr)]
	return ok
}

// IsRecovery reports whether addr is the recovery address.
func (c *Config) IsRecovery(addr []byte) bool { return bytes.Equal(addr, c.recovery) }

// IsFeePayer reports whether addr is one of the configured fee payers.
func (c *Config) IsFeePayer(addr []byte) bool {
	_, ok := c.payers[string(addr)]
	return ok
}

// Recovery returns a copy of the recovery address bytes.
func (c *Config) Recovery() []byte { return clone(c.recovery) }

// FeePayers returns the fee payers in ascending byte order.
func (c *Config) FeePayers() [][]byte {
	out := make([][]byte, 0, len(c.payers))
	for p := range c.payers {
		out = append(out, []byte(p))
	}
	sort.Slice(out, func(i, j int) bool { return bytes.Compare(out[i], out[j]) < 0 })
	return out
}

// Listed returns the listed addresses in ascending byte order.
func (c *Config) Listed() [][]byte {
	out := make([][]byte, len(c.sorted))
	for i, a := range c.sorted {
		out[i] = clone(a)
	}
	return out
}

// Window returns the activation height and the expiry time (unix seconds).
func (c *Config) Window() (int64, int64) { return c.start, c.expiry }
