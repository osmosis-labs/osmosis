package data

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetIBCAliasesMap(t *testing.T) {
	ibcAliases := GetIBCAliasesMap()
	require.NotEmpty(t, ibcAliases)

	// a few denom checks
	require.Equal(t, ibcAliases["uatom"], "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2")
	require.Equal(t, ibcAliases["uosmo"], "uosmo")

	// check that subsequent calls return the same map
	secondIBCAliases := GetIBCAliasesMap()
	require.Equal(t, ibcAliases, secondIBCAliases)

	require.False(t, func() bool {
		if _, err := os.Stat("assetlist.json"); err == nil {
			// file exists
			return true
		} else {
			// file does not exist
			return false
		}
	}())
}
