package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/osmosis-labs/osmosis/v27/x/market/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding market type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.OsmosisPoolDeltaKey):
			var deltaA, deltaB sdk.DecProto
			cdc.MustUnmarshal(kvA.Value, &deltaA)
			cdc.MustUnmarshal(kvB.Value, &deltaB)
			return fmt.Sprintf("%v\n%v", deltaA, deltaB)
		default:
			panic(fmt.Sprintf("invalid market key prefix %X", kvA.Key[:1]))
		}
	}
}
