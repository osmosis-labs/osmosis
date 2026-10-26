package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v32/x/tokenfactory/client/cli"
	"github.com/osmosis-labs/osmosis/v32/x/tokenfactory/types"
)

func TestGetQueryCmdRegistersDenomBeforeSendHook(t *testing.T) {
	cmd, args, err := cli.GetQueryCmd().Find([]string{"denom-before-send-hook"})
	require.NoError(t, err)
	require.Equal(t, "denom-before-send-hook", cmd.Name())
	require.Empty(t, args)
	require.NotNil(t, cmd.RunE)
	require.NoError(t, cmd.ValidateArgs([]string{"uatom"}))
	require.Error(t, cmd.ValidateArgs(nil))
	require.Error(t, cmd.ValidateArgs([]string{"uatom", "uosmo"}))
}

func TestGetCmdDenomAuthorityMetadata(t *testing.T) {
	desc, _ := cli.GetCmdDenomAuthorityMetadata()
	tcs := map[string]osmocli.QueryCliTestCase[*types.QueryDenomAuthorityMetadataRequest]{
		"basic test": {
			Cmd: "uatom",
			ExpectedQuery: &types.QueryDenomAuthorityMetadataRequest{
				Denom: "uatom",
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdDenomsFromCreator(t *testing.T) {
	desc, _ := cli.GetCmdDenomsFromCreator()
	tcs := map[string]osmocli.QueryCliTestCase[*types.QueryDenomsFromCreatorRequest]{
		"basic test": {
			Cmd: "osmo1test",
			ExpectedQuery: &types.QueryDenomsFromCreatorRequest{
				Creator: "osmo1test",
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}
