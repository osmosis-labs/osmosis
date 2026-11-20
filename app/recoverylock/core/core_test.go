package core

import (
	"bytes"
	"errors"
	"testing"
)

func a(b byte) []byte { return bytes.Repeat([]byte{b}, AddrLen) }

func cfg(t *testing.T) *Config {
	t.Helper()
	c, err := NewConfig([][]byte{a(3), a(1), a(2), a(1)}, a(9), [][]byte{a(8), a(7)}, 100, 5000)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestValidation(t *testing.T) {
	bad := []struct {
		name               string
		list               [][]byte
		rec                []byte
		payers             [][]byte
		height, expiryTime int64
	}{
		{"empty list", nil, a(9), [][]byte{a(8)}, 1, 2},
		{"listed recovery", [][]byte{a(1)}, a(1), [][]byte{a(8)}, 1, 2},
		{"listed payer", [][]byte{a(1)}, a(9), [][]byte{a(1)}, 1, 2},
		{"payer is recovery", [][]byte{a(1)}, a(9), [][]byte{a(9)}, 1, 2},
		{"no payers", [][]byte{a(1)}, a(9), nil, 1, 2},
		{"too many payers", [][]byte{a(1)}, a(9), [][]byte{a(10), a(11), a(12), a(13), a(14), a(15), a(16), a(17), a(18), a(19), a(20), a(21), a(22), a(23), a(24), a(25), a(26)}, 1, 2},
		{"short listed", [][]byte{{1}}, a(9), [][]byte{a(8)}, 1, 2},
		{"long listed (module or contract)", [][]byte{bytes.Repeat([]byte{1}, 32)}, a(9), [][]byte{a(8)}, 1, 2},
		{"short recovery", [][]byte{a(1)}, []byte{9}, [][]byte{a(8)}, 1, 2},
		{"short payer", [][]byte{a(1)}, a(9), [][]byte{{8}}, 1, 2},
		{"zero height", [][]byte{a(1)}, a(9), [][]byte{a(8)}, 0, 2},
		{"zero expiry", [][]byte{a(1)}, a(9), [][]byte{a(8)}, 1, 0},
	}
	for _, c := range bad {
		if _, err := NewConfig(c.list, c.rec, c.payers, c.height, c.expiryTime); err == nil {
			t.Errorf("%s: accepted", c.name)
		}
	}
}

func TestWindowAndListing(t *testing.T) {
	c := cfg(t)
	cases := []struct {
		h, ts int64
		want  bool
	}{{99, 10, false}, {100, 10, true}, {100000, 4999, true}, {100000, 5000, false}, {100, 6000, false}}
	for _, x := range cases {
		if c.Active(x.h, x.ts) != x.want {
			t.Errorf("Active(%d,%d) != %v", x.h, x.ts, x.want)
		}
	}
	if !c.HeightActive(100) || c.HeightActive(99) {
		t.Fatal("HeightActive")
	}
	if !c.IsListed(a(2)) || c.IsListed(a(4)) {
		t.Fatal("listing")
	}
	if !c.IsFeePayer(a(7)) || !c.IsFeePayer(a(8)) || c.IsFeePayer(a(9)) {
		t.Fatal("fee payers")
	}
	l := c.Listed()
	if len(l) != 3 || l[0][0] != 1 || l[1][0] != 2 || l[2][0] != 3 {
		t.Fatalf("sorted unique list wrong: %v", l)
	}
	l[0][0] = 77
	if !c.IsListed(a(1)) {
		t.Fatal("Listed must return copies")
	}
	if p := c.FeePayers(); len(p) != 2 || p[0][0] != 7 {
		t.Fatalf("fee payers order %v", p)
	}
}

func TestRegistry(t *testing.T) {
	c := cfg(t)
	r := NewRegistry()
	r.Register("/ok", func(*Config, any, any, []byte) error { return nil })
	r.Register("/no", func(*Config, any, any, []byte) error { return errors.New("no") })
	if err := r.Check(c, nil, "/ok", nil, a(1)); err != nil {
		t.Fatal(err)
	}
	if err := r.Check(c, nil, "/no", nil, a(1)); err == nil {
		t.Fatal("rule rejection ignored")
	}
	if err := r.Check(c, nil, "/unknown", nil, a(1)); err == nil {
		t.Fatal("unregistered type allowed")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("duplicate registration must panic")
		}
	}()
	r.Register("/ok", nil)
}

func TestContractExit(t *testing.T) {
	const rec = "osmo1ksdgkxhjctrgxrz6lw232f92fq7sr3hplcxk65"
	const other = "osmo1jx2vma8xjyzr2ugrp66z7h3l78zhq4m08egdx5"
	valid := func(a string) error {
		if a == rec || a == other {
			return nil
		}
		return errors.New("invalid address")
	}
	cases := []struct {
		msg string
		ok  bool
	}{
		{`{"transfer":{"recipient":"` + rec + `","amount":"1000"}}`, true},
		{`{"transfer_nft":{"recipient":"` + rec + `","token_id":"7"}}`, true},
		{` {"transfer_nft" : {"token_id":"a b","recipient":"` + rec + `"}} `, true},
		{`{"transfer":{"recipient":"osmo1jx2vma8xjyzr2ugrp66z7h3l78zhq4m08egdx5","amount":"1"}}`, false},
		{`{"transfer":{"recipient":"OSMO1KSDGKXHJCTRGXRZ6LW232F92FQ7SR3HPLCXK65","amount":"1"}}`, false},
		{`{"send":{"contract":"` + rec + `","amount":"1","msg":""}}`, false},
		{`{"send_nft":{"contract":"` + rec + `","token_id":"1","msg":""}}`, false},
		{`{"approve":{"spender":"` + rec + `","token_id":"1"}}`, false},
		{`{"burn":{"amount":"1"}}`, false},
		{`{"transfer":{"recipient":"` + rec + `","amount":"1","memo":"x"}}`, false},
		{`{"transfer":{"recipient":"` + rec + `"}}`, false},
		{`{"transfer":{"recipient":"` + rec + `","amount":1}}`, false},
		{`{"transfer":{"recipient":"` + rec + `","amount":"-1"}}`, false},
		{`{"transfer":{"recipient":"` + rec + `","amount":""}}`, false},
		{`{"transfer_nft":{"recipient":"` + rec + `","token_id":""}}`, false},
		{`{"transfer_nft":{"recipient":"` + rec + `","token_id":7}}`, false},
		{`{"transfer":{"recipient":"` + rec + `","amount":"1"},"transfer_nft":{"recipient":"` + rec + `","token_id":"1"}}`, false},
		{`{"transfer":{"recipient":"` + rec + `","amount":"1"},"transfer":{"recipient":"x","amount":"1"}}`, false},
		{`{"transfer":{"recipient":"x","recipient":"` + rec + `","amount":"1"}}`, false},
		{`{"transfer":{"recipient":"` + rec + `","recipient":"` + rec + `","amount":"1"}}`, false},
		{`{"transfer":{"recipient":"` + rec + `","amount":"1"}} {}`, false},
		{`{"transfer":{"recipient":"` + rec + `","amount":"1"}`, false},
		{`{"transfer":[]}`, false},
		{`[]`, false},
		{`{}`, false},
		{``, false},
		// rights removal: any valid spender or operator, no expires, no extras
		{`{"decrease_allowance":{"spender":"` + other + `","amount":"340282366920938463463374607431768211455"}}`, true},
		{`{"decrease_allowance":{"spender":"` + rec + `","amount":"0"}}`, true},
		{`{"revoke":{"spender":"` + other + `","token_id":"7"}}`, true},
		{`{"revoke_all":{"operator":"` + other + `"}}`, true},
		{`{"decrease_allowance":{"spender":"` + other + `","amount":"1","expires":{"never":{}}}}`, false},
		{`{"decrease_allowance":{"spender":"` + other + `","amount":"1","expires":null}}`, false},
		{`{"decrease_allowance":{"spender":"osmo1bad","amount":"1"}}`, false},
		{`{"decrease_allowance":{"spender":"` + other + `","amount":1}}`, false},
		{`{"decrease_allowance":{"spender":"` + other + `"}}`, false},
		{`{"decrease_allowance":{"owner":"` + other + `","amount":"1"}}`, false},
		{`{"revoke":{"spender":"` + other + `"}}`, false},
		{`{"revoke":{"spender":"` + other + `","token_id":""}}`, false},
		{`{"revoke":{"spender":"` + other + `","token_id":"1","x":1}}`, false},
		{`{"revoke":{"operator":"` + other + `","token_id":"1"}}`, false},
		{`{"revoke_all":{"operator":"osmo1bad"}}`, false},
		{`{"revoke_all":{"spender":"` + other + `"}}`, false},
		{`{"revoke_all":{"operator":"` + other + `","operator":"` + other + `"}}`, false},
		{`{"revoke_all":{"operator":"` + other + `"},"revoke":{"spender":"` + other + `","token_id":"1"}}`, false},
		{`{"increase_allowance":{"spender":"` + other + `","amount":"1"}}`, false},
		{`{"approve_all":{"operator":"` + other + `"}}`, false},
		{`{"transfer_from":{"owner":"` + other + `","recipient":"` + rec + `","amount":"1"}}`, false},
	}
	for _, c := range cases {
		err := CheckContractExit([]byte(c.msg), rec, valid)
		if (err == nil) != c.ok {
			t.Errorf("%s: ok=%v err=%v", c.msg, c.ok, err)
		}
	}
}
