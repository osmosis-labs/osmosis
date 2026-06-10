package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v31/x/mint/types"
)

// osmoBech32Prefix is the account address prefix configured by the app at boot
// (app/params.Bech32PrefixAccAddr). The types package cannot import app/params
// without an import cycle, so it is restated here for the test.
const osmoBech32Prefix = "osmo"

// TestRestrictedAddressesParse guards the compiled-in restricted address list:
// under the chain's bech32 prefix, every entry must be valid. The curated list
// is edited by hand, so this catches a malformed entry at CI time rather than
// silently skipping it at query time (where GetRestrictedSupply tolerates a bad
// entry by skipping it).
func TestRestrictedAddressesParse(t *testing.T) {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(osmoBech32Prefix, osmoBech32Prefix+"pub")

	for _, s := range types.RestrictedAddresses {
		_, err := sdk.AccAddressFromBech32(s)
		require.NoError(t, err, "restricted address %q must be valid bech32 under the %q prefix", s, osmoBech32Prefix)
	}
}
