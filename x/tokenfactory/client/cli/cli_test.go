package cli_test

import (
	"testing"

	"github.com/osmosis-labs/osmosis/v13/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/tokenfactory/client/cli"
	"github.com/osmosis-labs/osmosis/v13/x/tokenfactory/types"
)

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
