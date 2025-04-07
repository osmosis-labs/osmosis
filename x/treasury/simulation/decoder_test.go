package simulation

import (
	"fmt"
	"github.com/osmosis-labs/osmosis/osmomath"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/osmosis-labs/osmosis/v27/x/treasury/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/treasury/types"
)

func TestDecodeDistributionStore(t *testing.T) {
	cdc := keeper.MakeTestCodec(t)
	dec := NewDecodeStore(cdc)

	taxRate := osmomath.NewDecWithPrec(123, 2)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.TaxRateKey, Value: cdc.MustMarshal(&sdk.DecProto{Dec: taxRate})},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"TaxRate", fmt.Sprintf("%v\n%v", taxRate, taxRate)},
		{"other", ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
