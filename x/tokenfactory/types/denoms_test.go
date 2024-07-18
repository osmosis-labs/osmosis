package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	appparams "github.com/osmosis-labs/osmosis/v23/app/params"
	"github.com/osmosis-labs/osmosis/v23/x/tokenfactory/types"
)

func TestDeconstructDenom(t *testing.T) {
	appparams.SetAddressPrefixes()

	for _, tc := range []struct {
		desc             string
		denom            string
		expectedSubdenom string
		err              error
	}{
		{
			desc:  "empty is invalid",
			denom: "",
			err:   types.ErrInvalidDenom,
		},
		{
			desc:             "normal",
			denom:            "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/bitcoin",
			expectedSubdenom: "bitcoin",
		},
		{
			desc:             "multiple slashes in subdenom",
			denom:            "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/bitcoin/1",
			expectedSubdenom: "bitcoin/1",
		},
		{
			desc:             "no subdenom",
			denom:            "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/",
			expectedSubdenom: "",
		},
		{
			desc:  "incorrect prefix",
			denom: "ibc/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/bitcoin",
			err:   types.ErrInvalidDenom,
		},
		{
			desc:             "subdenom of only slashes",
			denom:            "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/////",
			expectedSubdenom: "////",
		},
		{
			desc:  "too long name",
			denom: "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			err:   types.ErrInvalidDenom,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			expectedCreator := "symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w"
			creator, subdenom, err := types.DeconstructDenom(tc.denom)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, expectedCreator, creator)
				require.Equal(t, tc.expectedSubdenom, subdenom)
			}
		})
	}
}

func TestGetTokenDenom(t *testing.T) {
	appparams.SetAddressPrefixes()
	for _, tc := range []struct {
		desc     string
		creator  string
		subdenom string
		valid    bool
	}{
		{
			desc:     "normal",
			creator:  "symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w",
			subdenom: "bitcoin",
			valid:    true,
		},
		{
			desc:     "multiple slashes in subdenom",
			creator:  "symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w",
			subdenom: "bitcoin/1",
			valid:    true,
		},
		{
			desc:     "no subdenom",
			creator:  "symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w",
			subdenom: "",
			valid:    true,
		},
		{
			desc:     "subdenom of only slashes",
			creator:  "symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w",
			subdenom: "/////",
			valid:    true,
		},
		{
			desc:     "too long name",
			creator:  "symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w",
			subdenom: "adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			valid:    false,
		},
		{
			desc:     "subdenom is exactly max length",
			creator:  "symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w",
			subdenom: "bitcoinfsadfsdfeadfsafwefsefsefsdfsdafasefsf",
			valid:    true,
		},
		{
			desc:     "creator is exactly max length",
			creator:  "symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7wjhgjhgkhjklhkjhkjhgjhgjgjggt",
			subdenom: "bitcoin",
			valid:    true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := types.GetTokenDenom(tc.creator, tc.subdenom)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
