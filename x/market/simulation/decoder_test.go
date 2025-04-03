package simulation

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/kv"

	simapp "github.com/osmosis-labs/osmosis/v27/app"

	"github.com/osmosis-labs/osmosis/v27/x/mint/simulation"
)

func TestDecodeDistributionStore(t *testing.T) {
	cdc, _ := simapp.MakeCodecs()
	dec := simulation.NewDecodeStore(cdc)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
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
