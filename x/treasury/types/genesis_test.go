package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGenesisValidation(t *testing.T) {
	genState := DefaultGenesisState()
	require.NoError(t, ValidateGenesis(genState))

	// Error - tax_rate range error
	genState.TaxRate = sdk.NewDec(-1)
	require.Error(t, ValidateGenesis(genState))

	// Valid
	genState.TaxRate = sdk.NewDecWithPrec(1, 2)
	require.NoError(t, ValidateGenesis(genState))

	// Error - reward_weight range error
	genState.RewardWeight = sdk.NewDec(-1)
	require.Error(t, ValidateGenesis(genState))

	// Valid
	genState.RewardWeight = sdk.NewDecWithPrec(5, 2)
	require.NoError(t, ValidateGenesis(genState))

	dummyDec := sdk.NewDec(10)
	dummyInt := sdk.NewInt(10)

	genState.EpochStates = []EpochState{
		{
			Epoch:             0,
			TaxReward:         dummyDec,
			SeigniorageReward: dummyDec,
			TotalStakedLuna:   dummyInt,
		},
		{
			Epoch:             1,
			TaxReward:         dummyDec,
			SeigniorageReward: dummyDec,
			TotalStakedLuna:   dummyInt,
		},
	}

	// Valid
	require.NoError(t, ValidateGenesis(genState))
}
