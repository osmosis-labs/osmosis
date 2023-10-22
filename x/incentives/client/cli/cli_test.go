package cli_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	"github.com/osmosis-labs/osmosis/v15/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v15/x/incentives/client/cli"
	"github.com/osmosis-labs/osmosis/v15/x/incentives/types"
)

var testAddresses = apptesting.CreateRandomAccounts(3)

func TestGetCmdGauges(t *testing.T) {
	desc, _ := cli.GetCmdGauges()
	tcs := map[string]osmocli.QueryCliTestCase[*types.GaugesRequest]{
		"basic test": {
			Cmd: "--offset=2",
			ExpectedQuery: &types.GaugesRequest{
				Pagination: &query.PageRequest{Key: []uint8{}, Offset: 2, Limit: 100},
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdToDistributeCoins(t *testing.T) {
	desc, _ := cli.GetCmdToDistributeCoins()
	tcs := map[string]osmocli.QueryCliTestCase[*types.ModuleToDistributeCoinsRequest]{
		"basic test": {
			Cmd: "", ExpectedQuery: &types.ModuleToDistributeCoinsRequest{},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdGaugeByID(t *testing.T) {
	desc, _ := cli.GetCmdGaugeByID()
	tcs := map[string]osmocli.QueryCliTestCase[*types.GaugeByIDRequest]{
		"basic test": {
			Cmd: "1", ExpectedQuery: &types.GaugeByIDRequest{Id: 1},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdActiveGauges(t *testing.T) {
	desc, _ := cli.GetCmdActiveGauges()
	tcs := map[string]osmocli.QueryCliTestCase[*types.ActiveGaugesRequest]{
		"basic test": {
			Cmd: "--offset=2",
			ExpectedQuery: &types.ActiveGaugesRequest{
				Pagination: &query.PageRequest{Key: []uint8{}, Offset: 2, Limit: 100},
			}},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdActiveGaugesPerDenom(t *testing.T) {
	desc, _ := cli.GetCmdActiveGaugesPerDenom()
	tcs := map[string]osmocli.QueryCliTestCase[*types.ActiveGaugesPerDenomRequest]{
		"basic test": {
			Cmd: "uosmo --offset=2",
			ExpectedQuery: &types.ActiveGaugesPerDenomRequest{
				Denom:      "uosmo",
				Pagination: &query.PageRequest{Key: []uint8{}, Offset: 2, Limit: 100},
			}},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdUpcomingGauges(t *testing.T) {
	desc, _ := cli.GetCmdUpcomingGauges()
	tcs := map[string]osmocli.QueryCliTestCase[*types.UpcomingGaugesRequest]{
		"basic test": {
			Cmd: "--offset=2",
			ExpectedQuery: &types.UpcomingGaugesRequest{
				Pagination: &query.PageRequest{Key: []uint8{}, Offset: 2, Limit: 100},
			}},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdUpcomingGaugesPerDenom(t *testing.T) {
	desc, _ := cli.GetCmdUpcomingGaugesPerDenom()
	tcs := map[string]osmocli.QueryCliTestCase[*types.UpcomingGaugesPerDenomRequest]{
		"basic test": {
			Cmd: "uosmo --offset=2",
			ExpectedQuery: &types.UpcomingGaugesPerDenomRequest{
				Denom:      "uosmo",
				Pagination: &query.PageRequest{Key: []uint8{}, Offset: 2, Limit: 100},
			}},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}
