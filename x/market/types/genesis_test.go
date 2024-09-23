package types

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisValidation(t *testing.T) {
	genState := DefaultGenesisState()
	require.NoError(t, ValidateGenesis(genState))

	genState.Params.ExchangePool = osmomath.NewDec(-1)
	require.Error(t, ValidateGenesis(genState))
}
