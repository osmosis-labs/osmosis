package keeper_test

import (
	"testing"

	testkeeper "authenticator/testutil/keeper"

	"github.com/osmosis-labs/osmosis/v17/x/authenticator/types"

	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.AuthenticatorKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
