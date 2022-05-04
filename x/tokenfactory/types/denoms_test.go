package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	appparams "github.com/osmosis-labs/osmosis/v7/app/params"
	"github.com/osmosis-labs/osmosis/v7/x/tokenfactory/types"
)

func TestDecomposeDenoms(t *testing.T) {
	appparams.SetAddressPrefixes()
	for _, tc := range []struct {
		desc  string
		denom string
		valid bool
	}{
		{
			desc:  "empty is invalid",
			denom: "",
			valid: false,
		},
		{
			desc:  "normal",
			denom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/bitcoin",
			valid: true,
		},
		{
			desc:  "multiple slashes in nonce",
			denom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/bitcoin/1",
			valid: true,
		},
		{
			desc:  "no nonce",
			denom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/",
			valid: true,
		},
		{
			desc:  "incorrect prefix",
			denom: "ibc/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/bitcoin",
			valid: false,
		},
		{
			desc:  "nonce of only slashes",
			denom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/////",
			valid: true,
		},
		{
			desc:  "too long name",
			denom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			valid: false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, _, err := types.DeconstructDenom(tc.denom)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
