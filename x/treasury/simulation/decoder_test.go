package simulation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	core "github.com/osmosis-labs/osmosis/v23/types"
	"github.com/osmosis-labs/osmosis/v23/x/treasury/keeper"
	"github.com/osmosis-labs/osmosis/v23/x/treasury/types"
)

func TestDecodeDistributionStore(t *testing.T) {
	cdc := keeper.MakeTestCodec(t)
	dec := NewDecodeStore(cdc)

	taxRate := sdk.NewDecWithPrec(123, 2)
	rewardWeight := sdk.NewDecWithPrec(532, 2)
	taxCap := sdk.NewInt(1600)
	taxProceeds := sdk.NewCoins(sdk.NewInt64Coin(core.MicroKRWDenom, 123124), sdk.NewInt64Coin(core.MicroSDRDenom, 123124))
	epochInitialIssuance := sdk.NewCoins(sdk.NewInt64Coin(core.MicroKRWDenom, 645352342))

	TR := sdk.NewDecWithPrec(123, 2)
	SR := sdk.NewDecWithPrec(43523, 4)
	TSL := sdk.NewInt(1245213)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.TaxRateKey, Value: cdc.MustMarshal(&sdk.DecProto{Dec: taxRate})},
			{Key: types.RewardWeightKey, Value: cdc.MustMarshal(&sdk.DecProto{Dec: rewardWeight})},
			{Key: types.TaxCapKey, Value: cdc.MustMarshal(&sdk.IntProto{Int: taxCap})},
			{Key: types.TaxProceedsKey, Value: cdc.MustMarshal(&types.EpochTaxProceeds{TaxProceeds: taxProceeds})},
			{Key: types.EpochInitialIssuanceKey, Value: cdc.MustMarshal(&types.EpochInitialIssuance{Issuance: epochInitialIssuance})},
			{Key: types.TRKey, Value: cdc.MustMarshal(&sdk.DecProto{Dec: TR})},
			{Key: types.SRKey, Value: cdc.MustMarshal(&sdk.DecProto{Dec: SR})},
			{Key: types.TSLKey, Value: cdc.MustMarshal(&sdk.IntProto{Int: TSL})},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"TaxRate", fmt.Sprintf("%v\n%v", taxRate, taxRate)},
		{"RewardWeight", fmt.Sprintf("%v\n%v", rewardWeight, rewardWeight)},
		{"TaxCap", fmt.Sprintf("%v\n%v", taxCap, taxCap)},
		{"TaxProceeds", fmt.Sprintf("%v\n%v", taxProceeds, taxProceeds)},
		{"EpochInitialIssuance", fmt.Sprintf("%v\n%v", epochInitialIssuance, epochInitialIssuance)},
		{"TR", fmt.Sprintf("%v\n%v", TR, TR)},
		{"SR", fmt.Sprintf("%v\n%v", SR, SR)},
		{"TSL", fmt.Sprintf("%v\n%v", TSL, TSL)},
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
