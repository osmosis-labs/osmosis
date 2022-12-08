package cli_test

import (
	"testing"

	"github.com/osmosis-labs/osmosis/v13/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/mint/client/cli"
	"github.com/osmosis-labs/osmosis/v13/x/mint/types"
)

func TestGetCmdQueryParams(t *testing.T) {
	desc, _ := cli.GetCmdQueryParams()
	tcs := map[string]osmocli.QueryCliTestCase[*types.QueryParamsRequest]{
		"basic test case": {
			Cmd:           "params",
			ExpectedQuery: &types.QueryParamsRequest{},
			ExpectedErr:   false,
		},
	}

	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdQueryEpochProvisions(t *testing.T) {
	desc, _ := cli.GetCmdQueryEpochProvisions()
	tcs := map[string]osmocli.QueryCliTestCase[*types.QueryEpochProvisionsRequest]{
		"basic test case": {
			Cmd:           "params",
			ExpectedQuery: &types.QueryEpochProvisionsRequest{},
			ExpectedErr:   false,
		},
	}

	osmocli.RunQueryTestCases(t, desc, tcs)
}
