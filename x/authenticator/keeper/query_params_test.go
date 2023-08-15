package keeper_test

import (
	"testing"

	testkeeper "authenticator/testutil/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v17/x/authenticator/types"
	"github.com/stretchr/testify/require"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.AuthenticatorKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
