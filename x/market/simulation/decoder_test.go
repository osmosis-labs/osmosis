package simulation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/osmosis-labs/osmosis/v23/x/market/keeper"
	"github.com/osmosis-labs/osmosis/v23/x/market/types"
)

func TestDecodeDistributionStore(t *testing.T) {
	cdc := keeper.MakeTestCodec(t)
	dec := NewDecodeStore(cdc)

	terraDelta := sdk.NewDecWithPrec(12, 2)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.TerraPoolDeltaKey, Value: cdc.MustMarshal(&sdk.DecProto{Dec: terraDelta})},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"TerraPoolDelta", fmt.Sprintf("%v\n%v", terraDelta, terraDelta)},
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
