package cosmwasm_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/cosmwasm"
)

var testContractAddress = sdk.AccAddress("contract____________").String()

// recordingContractKeeper is a contract keeper that records the gas limit of the context a sudo call was made
// on and consumes a fixed amount of gas from it.
type recordingContractKeeper struct {
	gasToConsume storetypes.Gas
	seenGasLimit storetypes.Gas
}

func (k *recordingContractKeeper) Instantiate(sdk.Context, uint64, sdk.AccAddress, sdk.AccAddress, []byte, string, sdk.Coins) (sdk.AccAddress, []byte, error) {
	return nil, nil, nil
}

func (k *recordingContractKeeper) Execute(sdk.Context, sdk.AccAddress, sdk.AccAddress, []byte, sdk.Coins) ([]byte, error) {
	return nil, nil
}

func (k *recordingContractKeeper) Sudo(ctx sdk.Context, _ sdk.AccAddress, _ []byte) ([]byte, error) {
	k.seenGasLimit = ctx.GasMeter().Limit()
	ctx.GasMeter().ConsumeGas(k.gasToConsume, "contract call")
	return nil, nil
}

// TestSudoGasLimit tests that a sudo call gets a gas meter limited to what the caller has left, capped at
// DefaultContractCallGasLimit, and that the gas the call consumes is charged to the caller.
func TestSudoGasLimit(t *testing.T) {
	tests := map[string]struct {
		gasMeter             storetypes.GasMeter
		consumedBeforeCall   storetypes.Gas
		gasToConsume         storetypes.Gas
		expectedGasLimit     storetypes.Gas
		expectedErr          bool
		expectedOutOfGas     bool
		expectedConsumedPost storetypes.Gas
	}{
		"limit is the gas remaining in the caller's meter, not its total limit": {
			gasMeter:             storetypes.NewGasMeter(1_000_000),
			consumedBeforeCall:   900_000,
			gasToConsume:         50_000,
			expectedGasLimit:     100_000,
			expectedConsumedPost: 950_000,
		},
		"limit is capped at the default contract call gas limit": {
			gasMeter:             storetypes.NewGasMeter(100_000_000),
			consumedBeforeCall:   1_000_000,
			gasToConsume:         50_000,
			expectedGasLimit:     cosmwasm.DefaultContractCallGasLimit,
			expectedConsumedPost: 1_050_000,
		},
		"infinite gas meter gets the default contract call gas limit": {
			gasMeter:             storetypes.NewInfiniteGasMeter(),
			consumedBeforeCall:   1_000_000,
			gasToConsume:         50_000,
			expectedGasLimit:     cosmwasm.DefaultContractCallGasLimit,
			expectedConsumedPost: 1_050_000,
		},
		"call that needs more than the caller has left runs out of gas at what is left": {
			gasMeter:           storetypes.NewGasMeter(1_000_000),
			consumedBeforeCall: 900_000,
			gasToConsume:       500_000,
			expectedGasLimit:   100_000,
			expectedOutOfGas:   true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.NewContext(nil, cmtproto.Header{}, false, log.NewNopLogger()).WithGasMeter(tc.gasMeter)
			ctx.GasMeter().ConsumeGas(tc.consumedBeforeCall, "before call")
			contractKeeper := &recordingContractKeeper{gasToConsume: tc.gasToConsume}

			call := func() {
				_, err := cosmwasm.Sudo[struct{}, struct{}](ctx, contractKeeper, testContractAddress, struct{}{})
				require.NoError(t, err)
			}

			if tc.expectedOutOfGas {
				// The call's consumption is charged to the caller, whose meter then runs out.
				require.PanicsWithValue(t, storetypes.ErrorOutOfGas{Descriptor: "Track contract call gas"}, call)
			} else {
				call()
				require.Equal(t, tc.expectedConsumedPost, ctx.GasMeter().GasConsumed())
			}
			require.Equal(t, tc.expectedGasLimit, contractKeeper.seenGasLimit)
		})
	}
}
