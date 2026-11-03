package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v32/x/concentrated-liquidity/client/cli"
)

func TestGetQueryCmdRegistersTotalLiquidity(t *testing.T) {
	cmd, args, err := cli.GetQueryCmd().Find([]string{"total-liquidity"})
	require.NoError(t, err)
	require.Equal(t, "total-liquidity", cmd.Name())
	require.Empty(t, args)
	require.NotNil(t, cmd.RunE)
	require.NoError(t, cmd.ValidateArgs(nil))
	require.Error(t, cmd.ValidateArgs([]string{"1"}))
}
