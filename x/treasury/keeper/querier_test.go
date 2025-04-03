package keeper

import (
	"testing"

	"github.com/osmosis-labs/osmosis/v27/x/treasury/types"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestQueryParams(t *testing.T) {
	input := CreateTestInput(t)
	ctx := sdk.WrapSDKContext(input.Ctx)

	querier := NewQuerier(input.TreasuryKeeper)
	res, err := querier.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)

	require.Equal(t, input.TreasuryKeeper.GetParams(input.Ctx), res.Params)
}

func TestQueryTaxRate(t *testing.T) {
	input := CreateTestInput(t)
	ctx := sdk.WrapSDKContext(input.Ctx)

	querier := NewQuerier(input.TreasuryKeeper)
	res, err := querier.TaxRate(ctx, &types.QueryTaxRateRequest{})
	require.NoError(t, err)

	require.Equal(t, input.TreasuryKeeper.GetTaxRate(input.Ctx), res.TaxRate)
}
