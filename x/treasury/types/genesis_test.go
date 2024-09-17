package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisValidation(t *testing.T) {
	genState := DefaultGenesisState()
	require.NoError(t, ValidateGenesis(genState))

	// Error - tax_rate range error
	genState.TaxRate = osmomath.NewDec(-1)
	require.Error(t, ValidateGenesis(genState))

	// Error - tax_rate range error
	genState.TaxRate = genState.Params.MaxFeeMultiplier.Add(osmomath.NewDecWithPrec(1, 1))
	require.Error(t, ValidateGenesis(genState))

	// Valid
	genState.TaxRate = osmomath.NewDecWithPrec(1, 2)
	require.NoError(t, ValidateGenesis(genState))

	// Valid
	require.NoError(t, ValidateGenesis(genState))
}
