package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGenesisValidation(t *testing.T) {
	genState := DefaultGenesisState()
	require.NoError(t, ValidateGenesis(genState))

	genState.Params.BasePool = sdk.NewDec(-1)
	require.Error(t, ValidateGenesis(genState))

	genState = DefaultGenesisState()
	genState.Params.PoolRecoveryPeriod = 0
	require.Error(t, ValidateGenesis(genState))

	genState = DefaultGenesisState()
	genState.Params.MinStabilitySpread = sdk.NewDec(-1)
	require.Error(t, ValidateGenesis(genState))
}
