package osmoutils

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var OneThird sdk.Dec = sdk.MustNewDecFromStr("3.333333333333333333")

// intended to be used with require/assert:  require.True(DecEq(...))
// TODO: Replace with function in SDK types package when we update
func DecApproxEq(t *testing.T, d1 sdk.Dec, d2 sdk.Dec, tol sdk.Dec) (*testing.T, bool, string, string, string) {
	diff := d1.Sub(d2).Abs()
	return t, diff.LTE(tol), "expected |d1 - d2| <:\t%v\ngot |d1 - d2| = \t\t%v", tol.String(), diff.String()
}
