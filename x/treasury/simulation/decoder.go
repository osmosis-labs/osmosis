package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/osmosis-labs/osmosis/v27/x/treasury/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding treasury type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.TaxRateKey):
			var taxRateA, taxRateB sdk.DecProto
			cdc.MustUnmarshal(kvA.Value, &taxRateA)
			cdc.MustUnmarshal(kvB.Value, &taxRateB)
			return fmt.Sprintf("%v\n%v", taxRateA, taxRateB)
		default:
			panic(fmt.Sprintf("invalid oracle key prefix %X", kvA.Key[:1]))
		}
	}
}
