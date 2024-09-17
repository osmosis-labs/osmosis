package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisValidation(t *testing.T) {
	genState := DefaultGenesisState()
	require.NoError(t, ValidateGenesis(genState))

	genState.Params.BasePool = osmomath.NewDec(-1)
	require.Error(t, ValidateGenesis(genState))

	genState = DefaultGenesisState()
	genState.Params.PoolRecoveryPeriod = 0
	require.Error(t, ValidateGenesis(genState))

	genState = DefaultGenesisState()
	genState.Params.MinStabilitySpread = osmomath.NewDec(-1)
	require.Error(t, ValidateGenesis(genState))
}
