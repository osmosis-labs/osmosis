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

	// Error - tax_rate range error
	genState.TaxRate = genState.Params.MaxFeeMultiplier.Add(sdk.NewDecWithPrec(1, 1))
	require.Error(t, ValidateGenesis(genState))

	// Valid
	genState.TaxRate = sdk.NewDecWithPrec(1, 2)
	require.NoError(t, ValidateGenesis(genState))

	// Valid
	require.NoError(t, ValidateGenesis(genState))
}
