package cli_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/client/cli"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

var testAddresses = osmoutils.CreateRandomAccounts(3)

type IntegrationTestSuite struct {
	suite.Suite
}

func TestNewCreatePoolCmd(t *testing.T) {
	testCases := map[string]struct {
		json      string
		expectErr bool
	}{
		"two tokens pair pool": {
			`{
			  "weights": "1node0token,3stake",
			  "initial-deposit": "100node0token,100stake",
			  "swap-fee": "0.001",
			  "exit-fee": "0.001"
			}`,
			false,
		},
		"future governor address": {
			fmt.Sprintf(`
			{
			  "%s": "1node0token,3stake",
			  "%s": "100node0token,100stake",
			  "%s": "0.001",
			  "%s": "0.001",
			  "%s": "osmo1fqlr98d45v5ysqgp6h56kpujcj4cvsjnjq9nck"
			}
			`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee, cli.PoolFileFutureGovernor),
			false,
		},
		"bad pool json - missing quotes around exit fee": {
			fmt.Sprintf(`
			{
			  "%s": "1node0token,3stake",
			  "%s": "100node0token,100stake",
			  "%s": "0.001",
			  "%s": 0.001
			}
	`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee),
			true,
		},
		"empty pool json": {
			"", true,
		},
		"smooth change params": {
			fmt.Sprintf(`
				{
					"%s": "1node0token,3stake",
					"%s": "100node0token,100stake",
					"%s": "0.001",
					"%s": "0.001",
					"%s": {
						"%s": "864h",
						"%s": "2node0token,1stake",
						"%s": "2006-01-02T15:04:05Z"
					}
				}
				`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee,
				cli.PoolFileSmoothWeightChangeParams, cli.PoolFileDuration, cli.PoolFileTargetPoolWeights, cli.PoolFileStartTime,
			),
			false,
		},
		"smooth change params - no start time": {
			fmt.Sprintf(`
				{
					"%s": "1node0token,3stake",
					"%s": "100node0token,100stake",
					"%s": "0.001",
					"%s": "0.001",
					"%s": {
						"%s": "864h",
						"%s": "2node0token,1stake"
					}
				}
				`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee,
				cli.PoolFileSmoothWeightChangeParams, cli.PoolFileDuration, cli.PoolFileTargetPoolWeights,
			),
			false,
		},
		"empty smooth change params": {
			fmt.Sprintf(`
				{
					"%s": "1node0token,3stake",
					"%s": "100node0token,100stake",
					"%s": "0.001",
					"%s": "0.001",
					"%s": {}
				}
				`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee,
				cli.PoolFileSmoothWeightChangeParams,
			),
			false,
		},
		"smooth change params wrong type": {
			fmt.Sprintf(`
				{
					"%s": "1node0token,3stake",
					"%s": "100node0token,100stake",
					"%s": "0.001",
					"%s": "0.001",
					"%s": "invalid string"
				}
				`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee,
				cli.PoolFileSmoothWeightChangeParams,
			),
			true,
		},
		"smooth change params missing duration": {
			fmt.Sprintf(`
				{
					"%s": "1node0token,3stake",
					"%s": "100node0token,100stake",
					"%s": "0.001",
					"%s": "0.001",
					"%s": {
						"%s": "2node0token,1stake"
					}
				}
				`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee,
				cli.PoolFileSmoothWeightChangeParams, cli.PoolFileTargetPoolWeights,
			),
			true,
		},
		"unknown fields in json": {
			fmt.Sprintf(`
			{
			  "%s": "1node0token",
			  "%s": "100node0token",
			  "%s": "0.001",
			  "%s": "0.001"
			  "unknown": true,
			}
			`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee),
			true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			desc := cli.NewCreatePoolCmd()
			jsonFile := testutil.WriteToNewTempFile(tt, tc.json)
			Cmd := fmt.Sprintf("--pool-file=%s --from=%s", jsonFile.Name(), testAddresses[0].String())

			txTc := osmocli.TxCliTestCase[*balancer.MsgCreateBalancerPool]{
				Cmd:                    Cmd,
				ExpectedErr:            tc.expectErr,
				OnlyCheckValidateBasic: true,
			}
			osmocli.RunTxTestCase(tt, desc, &txTc)
		})
	}
}

func TestNewJoinPoolCmd(t *testing.T) {
	desc, _ := cli.NewJoinPoolCmd()
	tcs := map[string]osmocli.TxCliTestCase[*types.MsgJoinPool]{
		"join pool": {
			Cmd: "--pool-id=1  --pool-id=1 --max-amounts-in=100stake --share-amount-out=100 --from=" + testAddresses[0].String(),
			ExpectedMsg: &types.MsgJoinPool{
				Sender:         testAddresses[0].String(),
				PoolId:         1,
				ShareOutAmount: sdk.NewIntFromUint64(100),
				TokenInMaxs:    sdk.NewCoins(sdk.NewInt64Coin("stake", 100)),
			},
		},
	}
	osmocli.RunTxTestCases(t, desc, tcs)
}

func TestNewExitPoolCmd(t *testing.T) {
	desc, _ := cli.NewExitPoolCmd()
	tcs := map[string]osmocli.TxCliTestCase[*types.MsgExitPool]{
		"exit pool": {
			Cmd: "--min-amounts-out=100stake --pool-id=1 --share-amount-in=10 --from=" + testAddresses[0].String(),
			ExpectedMsg: &types.MsgExitPool{
				Sender:        testAddresses[0].String(),
				PoolId:        1,
				ShareInAmount: sdk.NewIntFromUint64(10),
				TokenOutMins:  sdk.NewCoins(sdk.NewInt64Coin("stake", 100)),
			},
		},
	}
	osmocli.RunTxTestCases(t, desc, tcs)
}

func TestNewSwapExactAmountOutCmd(t *testing.T) {
	desc, _ := cli.NewSwapExactAmountOutCmd()
	tcs := map[string]osmocli.TxCliTestCase[*types.MsgSwapExactAmountOut]{
		"swap exact amount out": {
			Cmd: "10stake 20 --swap-route-pool-ids=1 --swap-route-denoms=node0token --from=" + testAddresses[0].String(),
			ExpectedMsg: &types.MsgSwapExactAmountOut{
				Sender:           testAddresses[0].String(),
				Routes:           []poolmanagertypes.SwapAmountOutRoute{{PoolId: 1, TokenInDenom: "node0token"}},
				TokenInMaxAmount: sdk.NewIntFromUint64(20),
				TokenOut:         sdk.NewInt64Coin("stake", 10),
			},
		},
	}
	osmocli.RunTxTestCases(t, desc, tcs)
}

func TestNewSwapExactAmountInCmd(t *testing.T) {
	desc, _ := cli.NewSwapExactAmountInCmd()
	tcs := map[string]osmocli.TxCliTestCase[*types.MsgSwapExactAmountIn]{
		"swap exact amount in": {
			Cmd: "10stake 3 --swap-route-pool-ids=1 --swap-route-denoms=node0token --from=" + testAddresses[0].String(),
			ExpectedMsg: &types.MsgSwapExactAmountIn{
				Sender:            testAddresses[0].String(),
				Routes:            []poolmanagertypes.SwapAmountInRoute{{PoolId: 1, TokenOutDenom: "node0token"}},
				TokenIn:           sdk.NewInt64Coin("stake", 10),
				TokenOutMinAmount: sdk.NewIntFromUint64(3),
			},
		},
	}
	osmocli.RunTxTestCases(t, desc, tcs)
}

func TestNewJoinSwapExternAmountInCmd(t *testing.T) {
	desc, _ := cli.NewJoinSwapExternAmountIn()
	tcs := map[string]osmocli.TxCliTestCase[*types.MsgJoinSwapExternAmountIn]{
		"swap exact amount in": {
			Cmd: "10stake 1 --pool-id=1 --from=" + testAddresses[0].String(),
			ExpectedMsg: &types.MsgJoinSwapExternAmountIn{
				Sender:            testAddresses[0].String(),
				PoolId:            1,
				TokenIn:           sdk.NewInt64Coin("stake", 10),
				ShareOutMinAmount: sdk.NewIntFromUint64(1),
			},
		},
	}
	osmocli.RunTxTestCases(t, desc, tcs)
}

func TestNewJoinSwapShareAmountOutCmd(t *testing.T) {
	desc, _ := cli.NewJoinSwapShareAmountOut()
	tcs := map[string]osmocli.TxCliTestCase[*types.MsgJoinSwapShareAmountOut]{
		"swap exact amount in": {
			Cmd: "stake 10 1 --pool-id=1 --from=" + testAddresses[0].String(),
			ExpectedMsg: &types.MsgJoinSwapShareAmountOut{
				Sender:           testAddresses[0].String(),
				PoolId:           1,
				TokenInDenom:     "stake",
				ShareOutAmount:   sdk.NewIntFromUint64(10),
				TokenInMaxAmount: sdk.NewIntFromUint64(1),
			},
		},
	}
	osmocli.RunTxTestCases(t, desc, tcs)
}

func TestNewExitSwapExternAmountOutCmd(t *testing.T) {
	desc, _ := cli.NewExitSwapExternAmountOut()
	tcs := map[string]osmocli.TxCliTestCase[*types.MsgExitSwapExternAmountOut]{
		"swap exact amount in": {
			Cmd: "10stake 1 --pool-id=1 --from=" + testAddresses[0].String(),
			ExpectedMsg: &types.MsgExitSwapExternAmountOut{
				Sender:           testAddresses[0].String(),
				PoolId:           1,
				TokenOut:         sdk.NewInt64Coin("stake", 10),
				ShareInMaxAmount: sdk.NewIntFromUint64(1),
			},
		},
	}
	osmocli.RunTxTestCases(t, desc, tcs)
}

func TestNewExitSwapShareAmountInCmd(t *testing.T) {
	desc, _ := cli.NewExitSwapShareAmountIn()
	tcs := map[string]osmocli.TxCliTestCase[*types.MsgExitSwapShareAmountIn]{
		"swap exact amount in": {
			Cmd: "stake 10 1 --pool-id=1 --from=" + testAddresses[0].String(),
			ExpectedMsg: &types.MsgExitSwapShareAmountIn{
				Sender:            testAddresses[0].String(),
				PoolId:            1,
				TokenOutDenom:     "stake",
				ShareInAmount:     sdk.NewIntFromUint64(10),
				TokenOutMinAmount: sdk.NewIntFromUint64(1),
			},
		},
	}
	osmocli.RunTxTestCases(t, desc, tcs)
}

func TestGetCmdPools(t *testing.T) {
	desc, _ := cli.GetCmdPools()
	tcs := map[string]osmocli.QueryCliTestCase[*types.QueryPoolsRequest]{
		"basic test": {
			Cmd: "--offset=2",
			ExpectedQuery: &types.QueryPoolsRequest{
				Pagination: &query.PageRequest{Key: []uint8{}, Offset: 2, Limit: 100},
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdPool(t *testing.T) {
	desc, _ := cli.GetCmdPool()
	tcs := map[string]osmocli.QueryCliTestCase[*types.QueryPoolRequest]{
		"basic test": {
			Cmd:           "1",
			ExpectedQuery: &types.QueryPoolRequest{PoolId: 1},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdSpotPrice(t *testing.T) {
	desc, _ := cli.GetCmdSpotPrice()
	tcs := map[string]osmocli.QueryCliTestCase[*types.QuerySpotPriceRequest]{
		"basic test": {
			Cmd: "1 uosmo ibc/111",
			ExpectedQuery: &types.QuerySpotPriceRequest{
				PoolId:          1,
				BaseAssetDenom:  "uosmo",
				QuoteAssetDenom: "ibc/111",
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdEstimateSwapExactAmountIn(t *testing.T) {
	desc, _ := cli.GetCmdEstimateSwapExactAmountIn()
	tcs := map[string]osmocli.QueryCliTestCase[*types.QuerySwapExactAmountInRequest]{
		"basic test": {
			Cmd: "1 osm11vmx8jtggpd9u7qr0t8vxclycz85u925sazglr7 10stake --swap-route-pool-ids=2 --swap-route-denoms=node0token",
			ExpectedQuery: &types.QuerySwapExactAmountInRequest{
				Sender:  "osm11vmx8jtggpd9u7qr0t8vxclycz85u925sazglr7",
				PoolId:  1,
				TokenIn: "10stake",
				Routes:  []poolmanagertypes.SwapAmountInRoute{{PoolId: 2, TokenOutDenom: "node0token"}},
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdEstimateSwapExactAmountOut(t *testing.T) {
	desc, _ := cli.GetCmdEstimateSwapExactAmountOut()
	tcs := map[string]osmocli.QueryCliTestCase[*types.QuerySwapExactAmountOutRequest]{
		"basic test": {
			Cmd: "1 osm11vmx8jtggpd9u7qr0t8vxclycz85u925sazglr7 10stake --swap-route-pool-ids=2 --swap-route-denoms=node0token",
			ExpectedQuery: &types.QuerySwapExactAmountOutRequest{
				Sender:   "osm11vmx8jtggpd9u7qr0t8vxclycz85u925sazglr7",
				PoolId:   1,
				TokenOut: "10stake",
				Routes:   []poolmanagertypes.SwapAmountOutRoute{{PoolId: 2, TokenInDenom: "node0token"}},
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}
