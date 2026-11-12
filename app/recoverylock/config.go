// Package recoverylock wires the recovery lock into Osmosis: configuration and
// Osmosis exit rules on top of the chain independent core (./core) and the
// cosmos-sdk v0.50 adapter (./sdk50).
//
// The configuration is compiled in from listed.txt (one bech32 address per
// line) and params.json. The committed values are TEST VALUES for a local
// chain (test_values: true). They are accepted only by builds with
// -tags recoverylock_test; every other build (including the release) refuses
// them, refuses the known test addresses and requires the expiry margin. The
// real list, recovery address, fee payers, activation height, activation time
// estimate, estimate reference block, unbonding period, expiry time, list hash
// and list count are supplied at release time.
package recoverylock

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	"github.com/osmosis-labs/osmosis/v32/app/recoverylock/core"
)

//go:embed listed.txt
var listedFile []byte

//go:embed params.json
var paramsFile []byte

// MinReleaseFeePayers is the minimum fee payer set size in release builds.
const MinReleaseFeePayers = 8

// ExpiryMarginSeconds is the margin required in release builds on top of two
// unbonding periods after the estimated activation time:
// expiry_time >= activation_time_estimate + 2 x unbonding_seconds + 7 days.
// Two periods because the attacker can fill all 7 unbonding entries of a
// delegation before the halt; an undelegation by the recovery operator can
// then only start when the first of those entries completes, up to one
// unbonding period after activation, and completes one period later.
const ExpiryMarginSeconds = 7 * 24 * 3600

// MinMillisPerBlock and MaxMillisPerBlock bound the block time implied by
// the activation estimate in release builds:
// (activation_time_estimate - estimate_ref_time) / (activation_height -
// estimate_ref_height) must lie in [0.5 s, 3 s]. Osmosis mainnet block header
// times give 0.83 s per block over heights 73836404 to 74836404 and 1.27 s per
// block over 71836404 to 73836404. The band accepts that drift with room on
// both sides and refuses the mistakes that matter: a height or time with a
// digit too many or too few (a factor of 10), a time in milliseconds, or a
// height of another chain that is far from the Osmosis height.
const (
	MinMillisPerBlock = 500
	MaxMillisPerBlock = 3000
)

// MinReleaseUnbondingSeconds is the Osmosis mainnet staking unbonding_time
// (1209600s, 14 days).
// unbonding_seconds in params.json may not be lower, so a typo cannot shrink
// the expiry margin.
const MinReleaseUnbondingSeconds = 14 * 24 * 3600

// testSeeds are the seeds of all test keys used for test values (private
// key = sha256 of the seed). Their addresses are refused outside test builds.
var testSeeds = func() []string {
	// recovery and attacker keys of the app tests, attacker and funder keys of
	// the local network tests
	s := []string{"recoverylock-test-recovery", "attacker-test-key", "attacker-e2e", "funder-e2e"}
	for i := 0; i < 3; i++ {
		s = append(s, fmt.Sprintf("recoverylock-test-listed-%d", i))
	}
	for i := 0; i < 16; i++ {
		s = append(s, fmt.Sprintf("recoverylock-test-fee-payer-%d", i))
	}
	return s
}()

// testHashAddrs are further test address seeds that are not key based
// (address = first 20 bytes of sha256 of the seed). Their addresses are
// refused outside test builds too.
var testHashAddrs = []string{"test-recovery", "test-listed-0", "test-listed-1", "test-listed-2", "test-fee-payer", "test-fee-payer-0",
	"test-fee-payer-1", "test-fee-payer-2", "test-fee-payer-3", "test-fee-payer-4", "test-fee-payer-5", "test-fee-payer-6", "test-fee-payer-7"}

// knownTestAddresses returns the raw bytes of every known test address.
func knownTestAddresses() map[string]bool {
	out := map[string]bool{}
	for _, seed := range testSeeds {
		h := sha256.Sum256([]byte(seed))
		out[string((&secp256k1.PrivKey{Key: h[:]}).PubKey().Address())] = true
	}
	for _, seed := range testHashAddrs {
		h := sha256.Sum256([]byte(seed))
		out[string(h[:20])] = true
	}
	return out
}

type params struct {
	RecoveryAddress  string   `json:"recovery_address"`
	FeePayers        []string `json:"fee_payers"`
	ActivationHeight int64    `json:"activation_height"`
	ExpiryTime       int64    `json:"expiry_time"`
	// ActivationTimeEstimate is the expected block time (unix seconds) of the
	// activation height; UnbondingSeconds is the chain's live staking
	// unbonding_time. Release builds require
	// expiry_time >= activation_time_estimate + 2 x unbonding_seconds + 7 days
	// (see ExpiryMarginSeconds).
	//
	// EstimateRefHeight and EstimateRefTime are a real mainnet block height
	// and its block time (unix seconds) taken when the config is built.
	// Release builds require activation_height > estimate_ref_height and the
	// implied block time (activation_time_estimate - estimate_ref_time) /
	// (activation_height - estimate_ref_height) within [MinMillisPerBlock,
	// MaxMillisPerBlock], which cross-checks the activation height against
	// the time estimate.
	//
	// Estimate method (at release time): take the latest block (height h0,
	// time t0; these are estimate_ref_height and estimate_ref_time) and the
	// block 10000 heights earlier (time t1) from a node; average block time
	// b = (t0 - t1) / 10000; estimate = t0 + (H - h0) * b rounded up to the
	// hour, where H is the activation height. Use the later of this and the
	// planned halt time. The 7 day margin absorbs block time drift and a slow
	// restart at the halt.
	ActivationTimeEstimate int64  `json:"activation_time_estimate"`
	UnbondingSeconds       int64  `json:"unbonding_seconds"`
	EstimateRefHeight      int64  `json:"estimate_ref_height"`
	EstimateRefTime        int64  `json:"estimate_ref_time"`
	ListSHA256             string `json:"list_sha256"`
	ListCount              int    `json:"list_count"`
	TestValues             bool   `json:"test_values"`
}

// paramKeys are the exact JSON keys of params (case sensitive).
var paramKeys = func() map[string]bool {
	out := map[string]bool{}
	t := reflect.TypeOf(params{})
	for i := 0; i < t.NumField(); i++ {
		out[strings.Split(t.Field(i).Tag.Get("json"), ",")[0]] = true
	}
	return out
}()

// parseParams decodes params.json strictly: duplicate keys at any depth,
// unknown keys, keys in another case than the exact field names, values of
// the wrong type and trailing data are refused (encoding/json alone takes the
// last of duplicate keys, ignores unknown keys and matches keys case
// insensitively). A null value leaves the field at its zero value, which
// fails every later check that needs a value; for test_values it means false.
func parseParams(b []byte) (params, error) {
	var p params
	v, err := core.StrictJSON(b)
	if err != nil {
		return p, err
	}
	top, ok := v.(map[string]any)
	if !ok {
		return p, fmt.Errorf("params must be a JSON object")
	}
	for k := range top {
		if !paramKeys[k] {
			return p, fmt.Errorf("unknown key %q", k)
		}
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return p, err
	}
	return p, nil
}

// MustConfig parses and validates the embedded configuration and panics on
// any problem, so a misconfigured binary never starts. blocked holds the
// app's blocked addresses: the module accounts keyed by bech32, and
// sanctioned addresses keyed by 0x hex, which this bech32 lookup does not
// match. No module account may be listed, be the recovery address or be a
// fee payer.
//
// It logs the values validators should check against the published release
// (activation height, expiry time, recovery address, fee payer count,
// list_count, list_sha256 and test or release mode).
func MustConfig(ac address.Codec, blocked map[string]bool, logger log.Logger) *core.Config {
	cfg, p, err := parseConfig(listedFile, paramsFile, ac, blocked, testBuild)
	if err != nil {
		panic("recovery lock config: " + err.Error())
	}
	mode := "release"
	if testBuild {
		mode = "test"
	}
	if logger != nil {
		start, expiry := cfg.Window()
		logger.Info("recovery lock config",
			"mode", mode,
			"activation_height", start,
			"expiry_time", expiry,
			"expiry_utc", time.Unix(expiry, 0).UTC().Format(time.RFC3339),
			"recovery_address", p.RecoveryAddress,
			"fee_payers", len(cfg.FeePayers()),
			"list_count", p.ListCount,
			"list_sha256", strings.ToLower(p.ListSHA256),
		)
	}
	return cfg
}

// checkEstimate cross-checks the activation height against the activation
// time estimate (release builds).
func checkEstimate(p params) error {
	if p.EstimateRefHeight <= 0 || p.EstimateRefTime <= 0 {
		return fmt.Errorf("estimate_ref_height and estimate_ref_time are required")
	}
	if p.ActivationHeight <= p.EstimateRefHeight {
		return fmt.Errorf("activation_height %d must be above estimate_ref_height %d", p.ActivationHeight, p.EstimateRefHeight)
	}
	dh := p.ActivationHeight - p.EstimateRefHeight
	dt := p.ActivationTimeEstimate - p.EstimateRefTime
	// bounds keep the products below 2^63
	const limit = int64(1) << 40
	if dt <= 0 || dt >= limit || dh >= limit {
		return fmt.Errorf("activation_time_estimate %d is not plausible after estimate_ref_time %d", p.ActivationTimeEstimate, p.EstimateRefTime)
	}
	if dt*1000 < MinMillisPerBlock*dh || dt*1000 > MaxMillisPerBlock*dh {
		return fmt.Errorf("implied block time %d ms ((activation_time_estimate - estimate_ref_time) / (activation_height - estimate_ref_height)) is outside %d to %d ms",
			dt*1000/dh, MinMillisPerBlock, MaxMillisPerBlock)
	}
	return nil
}

func parseConfig(listed, paramsJSON []byte, ac address.Codec, blocked map[string]bool, test bool) (*core.Config, params, error) {
	p, err := parseParams(paramsJSON)
	if err != nil {
		return nil, p, fmt.Errorf("params: %w", err)
	}
	cfg, err := build(listed, p, ac, blocked, test)
	return cfg, p, err
}

func build(listed []byte, p params, ac address.Codec, blocked map[string]bool, test bool) (*core.Config, error) {
	if !test {
		if p.TestValues {
			return nil, fmt.Errorf("test values outside a recoverylock_test build")
		}
		if len(p.FeePayers) < MinReleaseFeePayers {
			return nil, fmt.Errorf("release builds need at least %d fee payers, got %d", MinReleaseFeePayers, len(p.FeePayers))
		}
		if p.ActivationTimeEstimate <= 0 {
			return nil, fmt.Errorf("activation_time_estimate is required")
		}
		if p.UnbondingSeconds < MinReleaseUnbondingSeconds {
			return nil, fmt.Errorf("unbonding_seconds %d is below the mainnet unbonding time %d", p.UnbondingSeconds, MinReleaseUnbondingSeconds)
		}
		if p.UnbondingSeconds > 365*24*3600 {
			return nil, fmt.Errorf("unbonding_seconds %d is not plausible", p.UnbondingSeconds)
		}
		if err := checkEstimate(p); err != nil {
			return nil, err
		}
		if min := p.ActivationTimeEstimate + 2*p.UnbondingSeconds + ExpiryMarginSeconds; p.ExpiryTime < min {
			return nil, fmt.Errorf("expiry_time %d is earlier than activation estimate + 2 x unbonding + 7 days (%d)", p.ExpiryTime, min)
		}
	}
	known := knownTestAddresses()
	sum := sha256.Sum256(listed)
	if got := hex.EncodeToString(sum[:]); !strings.EqualFold(got, p.ListSHA256) {
		return nil, fmt.Errorf("listed.txt sha256 %s does not match params %s", got, p.ListSHA256)
	}
	decode := func(what, s string) ([]byte, error) {
		b, err := ac.StringToBytes(s)
		if err != nil {
			return nil, fmt.Errorf("%s %q: %w", what, s, err)
		}
		if len(b) != core.AddrLen {
			return nil, fmt.Errorf("%s %q has %d bytes, want %d", what, s, len(b), core.AddrLen)
		}
		if !test && known[string(b)] {
			return nil, fmt.Errorf("%s %q is a known test address", what, s)
		}
		canon, err := ac.BytesToString(b)
		if err != nil {
			return nil, err
		}
		if _, isBlocked := blocked[canon]; isBlocked {
			return nil, fmt.Errorf("%s %q is a module account or blocked address", what, s)
		}
		return b, nil
	}
	var list [][]byte
	seen := map[string]int{}
	for i, line := range strings.Split(string(listed), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		b, err := decode(fmt.Sprintf("list line %d", i+1), line)
		if err != nil {
			return nil, err
		}
		// Duplicates are compared as address bytes, so a second spelling of
		// the same address (for example upper case bech32) is caught too.
		if prev, dup := seen[string(b)]; dup {
			return nil, fmt.Errorf("listed.txt line %d duplicates line %d", i+1, prev)
		}
		seen[string(b)] = i + 1
		list = append(list, b)
	}
	if len(list) != p.ListCount {
		return nil, fmt.Errorf("listed.txt has %d addresses, params say %d", len(list), p.ListCount)
	}
	rec, err := decode("recovery address", p.RecoveryAddress)
	if err != nil {
		return nil, err
	}
	var payers [][]byte
	seenPayer := map[string]int{}
	for i, s := range p.FeePayers {
		b, err := decode(fmt.Sprintf("fee payer %d", i), s)
		if err != nil {
			return nil, err
		}
		// compared as address bytes, like listed lines, so the release
		// minimum counts distinct fee payers
		if prev, dup := seenPayer[string(b)]; dup {
			return nil, fmt.Errorf("fee payer %d duplicates fee payer %d", i, prev)
		}
		seenPayer[string(b)] = i
		payers = append(payers, b)
	}
	return core.NewConfig(list, rec, payers, p.ActivationHeight, p.ExpiryTime)
}
