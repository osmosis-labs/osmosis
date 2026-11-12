package recoverylock

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"cosmossdk.io/log"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var ac = addresscodec.NewBech32Codec("osmo")

func TestEmbeddedConfig(t *testing.T) {
	if !testBuild {
		t.Fatal("run the tests with -tags recoverylock_test")
	}
	var buf bytes.Buffer
	cfg := MustConfig(ac, nil, log.NewLogger(&buf, log.ColorOption(false)))
	for _, want := range []string{"recovery lock config", "mode=test", "activation_height=1000000000", "expiry_time=4102444800", "expiry_utc=2100-01-01T00:00:00Z",
		"recovery_address=osmo1ksdgkxhjctrgxrz6lw232f92fq7sr3hplcxk65", "fee_payers=8", "list_count=3",
		"list_sha256=27786f20c6c4ed068aaa675907be1409f1b904f99d4418d4e0680c5845e0e2f8"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("startup log lacks %q: %s", want, buf.String())
		}
	}
	if n := len(cfg.Listed()); n != 3 {
		t.Fatalf("test list has %d entries, want 3", n)
	}
	if n := len(cfg.FeePayers()); n != 8 {
		t.Fatalf("test fee payers %d, want 8", n)
	}
}

func randAddr(t *testing.T) string {
	s, err := ac.BytesToString(secp256k1.GenPrivKey().PubKey().Address())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func listOf(addrs ...string) string { return "# list\n" + strings.Join(addrs, "\n") + "\n" }

// mk builds params.json for the given list text, starting from a valid
// release configuration with fresh random addresses.
func mk(t *testing.T, list string, mut func(p *params)) []byte {
	t.Helper()
	sum := sha256.Sum256([]byte(list))
	n := 0
	for _, l := range strings.Split(list, "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
			n++
		}
	}
	// reference block 100000 heights before activation at 1 s per block
	p := params{RecoveryAddress: randAddr(t), ActivationHeight: 75_000_000, ActivationTimeEstimate: 1_800_000_000, UnbondingSeconds: 1_209_600,
		EstimateRefHeight: 74_900_000, EstimateRefTime: 1_800_000_000 - 100_000,
		ExpiryTime: 1_800_000_000 + 2*1_209_600 + ExpiryMarginSeconds, ListSHA256: hex.EncodeToString(sum[:]), ListCount: n}
	for i := 0; i < 8; i++ {
		p.FeePayers = append(p.FeePayers, randAddr(t))
	}
	if mut != nil {
		mut(&p)
	}
	b, _ := json.Marshal(p)
	return b
}

func TestConfigValidation(t *testing.T) {
	embedded := string(listedFile)
	real := listOf(randAddr(t), randAddr(t))
	modStr, _ := ac.BytesToString(authtypes.NewModuleAddress("distribution"))
	blocked := map[string]bool{modStr: false}
	long, _ := ac.BytesToString(make([]byte, 32))
	testAddr := strings.TrimSpace(strings.Split(embedded, "\n")[1])
	a1 := randAddr(t)
	cases := []struct {
		name   string
		list   string
		params []byte
		test   bool
		ok     bool
	}{
		{"embedded test values in a test build", embedded, paramsFile, true, true},
		{"embedded test values refused outside test builds", embedded, paramsFile, false, false},
		{"real values in a release build", real, mk(t, real, nil), false, true},
		{"known test address listed, flag off", listOf(testAddr), mk(t, listOf(testAddr), nil), false, false},
		{"known test address as recovery", real, mk(t, real, func(p *params) { p.RecoveryAddress = testAddr }), false, false},
		{"too few fee payers", real, mk(t, real, func(p *params) { p.FeePayers = p.FeePayers[:7] }), false, false},
		{"expiry one second short of the margin", real, mk(t, real, func(p *params) { p.ExpiryTime-- }), false, false},
		{"expiry with only one unbonding period", real, mk(t, real, func(p *params) { p.ExpiryTime = p.ActivationTimeEstimate + p.UnbondingSeconds + ExpiryMarginSeconds }), false, false},
		{"duplicate fee payer", real, mk(t, real, func(p *params) { p.FeePayers[7] = p.FeePayers[0] }), false, false},
		{"duplicate fee payer in another spelling", real, mk(t, real, func(p *params) { p.FeePayers[7] = strings.ToUpper(p.FeePayers[0]) }), false, false},
		{"duplicate fee payer in a test build", real, mk(t, real, func(p *params) { p.FeePayers[1] = p.FeePayers[0] }), true, false},
		{"missing estimate reference height", real, mk(t, real, func(p *params) { p.EstimateRefHeight = 0 }), false, false},
		{"missing estimate reference time", real, mk(t, real, func(p *params) { p.EstimateRefTime = 0 }), false, false},
		{"activation height at the reference height", real, mk(t, real, func(p *params) { p.ActivationHeight = p.EstimateRefHeight }), false, false},
		{"activation height below the reference height", real, mk(t, real, func(p *params) { p.ActivationHeight = p.EstimateRefHeight - 1 }), false, false},
		{"activation height with a digit too many (0.1 s per block)", real, mk(t, real, func(p *params) { p.ActivationHeight = p.EstimateRefHeight + 1_000_000 }), false, false},
		{"activation height with a digit too few (10 s per block)", real, mk(t, real, func(p *params) { p.ActivationHeight = p.EstimateRefHeight + 10_000 }), false, false},
		{"reference time in milliseconds", real, mk(t, real, func(p *params) { p.EstimateRefTime *= 1000 }), false, false},
		{"reference time after the activation estimate", real, mk(t, real, func(p *params) { p.EstimateRefTime = p.ActivationTimeEstimate + 1 }), false, false},
		{"0.5 s per block accepted", real, mk(t, real, func(p *params) { p.ActivationHeight = p.EstimateRefHeight + 200_000 }), false, true},
		{"just under 0.5 s per block refused", real, mk(t, real, func(p *params) { p.ActivationHeight = p.EstimateRefHeight + 200_001 }), false, false},
		{"3 s per block accepted", real, mk(t, real, func(p *params) { p.EstimateRefTime = p.ActivationTimeEstimate - 300_000 }), false, true},
		{"just over 3 s per block refused", real, mk(t, real, func(p *params) { p.EstimateRefTime = p.ActivationTimeEstimate - 300_001 }), false, false},
		{"estimate not checked in test builds", real, mk(t, real, func(p *params) { p.EstimateRefHeight = 0; p.EstimateRefTime = 0 }), true, true},
		{"missing activation time estimate", real, mk(t, real, func(p *params) { p.ActivationTimeEstimate = 0 }), false, false},
		{"missing unbonding period", real, mk(t, real, func(p *params) { p.UnbondingSeconds = 0 }), false, false},
		{"unbonding period below 14 days", real, mk(t, real, func(p *params) { p.UnbondingSeconds = MinReleaseUnbondingSeconds - 1; p.ExpiryTime = 4_000_000_000 }), false, false},
		{"later expiry is fine", real, mk(t, real, func(p *params) { p.ExpiryTime += 3600 }), false, true},
		{"duplicate line", listOf(a1, a1), mk(t, listOf(a1, a1), nil), false, false},
		{"duplicate line in another spelling", listOf(a1, strings.ToUpper(a1)), mk(t, listOf(a1, strings.ToUpper(a1)), nil), false, false},
		{"duplicate line in a test build", listOf(a1, a1), mk(t, listOf(a1, a1), nil), true, false},
		{"hash mismatch", real + "\n", mk(t, real, nil), false, false},
		{"count mismatch", real, mk(t, real, func(p *params) { p.ListCount = 5 }), false, false},
		{"module account listed", listOf(modStr), mk(t, listOf(modStr), nil), false, false},
		{"module account recovery", real, mk(t, real, func(p *params) { p.RecoveryAddress = modStr }), false, false},
		{"module account fee payer", real, mk(t, real, func(p *params) { p.FeePayers[0] = modStr }), false, false},
		{"32 byte address listed", listOf(long), mk(t, listOf(long), nil), false, false},
	}
	for _, c := range cases {
		_, _, err := parseConfig([]byte(c.list), c.params, ac, blocked, c.test)
		if (err == nil) != c.ok {
			t.Errorf("%s: ok=%v err=%v", c.name, c.ok, err)
		}
	}
}

func TestKnownTestAddressesCoverEmbedded(t *testing.T) {
	known := knownTestAddresses()
	var p params
	if err := json.Unmarshal(paramsFile, &p); err != nil {
		t.Fatal(err)
	}
	all := append([]string{p.RecoveryAddress}, p.FeePayers...)
	for _, l := range strings.Split(string(listedFile), "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
			all = append(all, l)
		}
	}
	for _, s := range all {
		b, err := ac.StringToBytes(s)
		if err != nil || !known[string(b)] {
			t.Errorf("%s is not in the known test address set", s)
		}
	}
}

// TestKnownTestAddressesCoverLocalNetwork pins the attacker and funder test
// addresses used on a local test network.
func TestKnownTestAddressesCoverLocalNetwork(t *testing.T) {
	known := knownTestAddresses()
	for _, s := range []string{"osmo1jx2vma8xjyzr2ugrp66z7h3l78zhq4m08egdx5", "osmo1rhu75nrrtswygzgxlcmtp2kny6maa9q058py2d"} {
		b, err := ac.StringToBytes(s)
		if err != nil || !known[string(b)] {
			t.Errorf("%s is not in the known test address set", s)
		}
	}
}

// TestParamsStrict: params.json is parsed strictly. Each case starts from a
// valid release configuration and must be refused.
func TestParamsStrict(t *testing.T) {
	real := listOf(randAddr(t), randAddr(t))
	good := string(mk(t, real, nil))
	if _, _, err := parseConfig([]byte(real), []byte(good), ac, nil, false); err != nil {
		t.Fatalf("base case must pass: %v", err)
	}
	body := strings.TrimSuffix(good, "}")
	cases := map[string]string{
		// encoding/json would take the last value and accept these
		"duplicate key (later value wins in encoding/json)": body + `,"test_values":false}`,
		"duplicate key with an unsafe later value":          body + `,"test_values":true}`,
		"unknown key":    body + `,"expiry":1}`,
		"wrong-case key": strings.Replace(good, `"expiry_time"`, `"Expiry_Time"`, 1),
		"wrong-case key in addition to the right one": body + `,"EXPIRY_TIME":4000000000}`,
		"trailing data":    good + ` {}`,
		"trailing garbage": good + `x`,
		"wrong type":       strings.Replace(good, `"list_count":2`, `"list_count":"2"`, 1),
		"not an object":    `[` + good + `]`,
		"truncated":        body,
	}
	for name, js := range cases {
		if js == good {
			t.Fatalf("%s: case did not change the input", name)
		}
		if _, _, err := parseConfig([]byte(real), []byte(js), ac, nil, false); err == nil {
			t.Errorf("%s: must be refused", name)
		}
		if _, _, err := parseConfig([]byte(real), []byte(js), ac, nil, true); err == nil {
			t.Errorf("%s: must be refused in test builds too", name)
		}
	}
}
