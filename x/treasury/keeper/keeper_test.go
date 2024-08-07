package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/treasury/types"
)

func TestTaxRate(t *testing.T) {
	input := CreateTestInput(t)

	// See that we can get and set tax rate
	for i := int64(0); i < 10; i++ {
		input.TreasuryKeeper.SetTaxRate(input.Ctx, sdk.NewDecWithPrec(i, 2))
		require.Equal(t, sdk.NewDecWithPrec(i, 2), input.TreasuryKeeper.GetTaxRate(input.Ctx))
	}
}

func TestParams(t *testing.T) {
	input := CreateTestInput(t)

	defaultParams := types.DefaultParams()
	input.TreasuryKeeper.SetParams(input.Ctx, defaultParams)

	retrievedParams := input.TreasuryKeeper.GetParams(input.Ctx)
	require.Equal(t, defaultParams, retrievedParams)
}
