package cli_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v13/osmoutils"
	"github.com/osmosis-labs/osmosis/v13/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/client/cli"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var testAddresses = osmoutils.CreateRandomAccounts(3)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
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
				Routes:           []types.SwapAmountOutRoute{{PoolId: 1, TokenInDenom: "node0token"}},
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
				Routes:            []types.SwapAmountInRoute{{PoolId: 1, TokenOutDenom: "node0token"}},
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

// func (s *IntegrationTestSuite) TestGetCmdPools() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expectErr bool
// 	}{
// 		{
// 			"query pools",
// 			[]string{
// 				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
// 			},
// 			false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdPools() // osmosisd query gamm pools
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			if tc.expectErr {
// 				s.Require().Error(err)
// 			} else {
// 				resp := types.QueryPoolsResponse{}
// 				s.Require().NoError(err, out.String())
// 				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())

// 				s.Require().Greater(len(resp.Pools), 0, out.String())
// 			}
// 		})
// 	}
// }

// func (s *IntegrationTestSuite) TestGetCmdPool() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expectErr bool
// 	}{
// 		{
// 			"query pool by id", // osmosisd query gamm pool 1
// 			[]string{
// 				"1",
// 				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
// 			},
// 			false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdPool()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			if tc.expectErr {
// 				s.Require().Error(err)
// 			} else {
// 				s.Require().NoError(err, out.String())

// 				resp := types.QueryPoolResponse{}
// 				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
// 			}
// 		})
// 	}
// }

// func (s *IntegrationTestSuite) TestGetCmdTotalShares() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expectErr bool
// 	}{
// 		{
// 			"query pool total share by id", // osmosisd query gamm total-share 1
// 			[]string{
// 				"1",
// 				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
// 			},
// 			false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdTotalShares()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			if tc.expectErr {
// 				s.Require().Error(err)
// 			} else {
// 				resp := types.QueryTotalSharesResponse{}
// 				s.Require().NoError(err, out.String())
// 				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
// 			}
// 		})
// 	}
// }

// func (s *IntegrationTestSuite) TestGetCmdSpotPrice() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expectErr bool
// 	}{
// 		{
// 			"query pool spot price", // osmosisd query gamm spot-price 1 stake node0token
// 			[]string{
// 				"1", "stake", "node0token",
// 				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
// 			},
// 			false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdSpotPrice()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			if tc.expectErr {
// 				s.Require().Error(err)
// 			} else {
// 				resp := types.QuerySpotPriceResponse{}
// 				s.Require().NoError(err, out.String())
// 				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
// 			}
// 		})
// 	}
// }

// // func (s *IntegrationTestSuite) TestGetCmdEstimateSwapExactAmountIn() {
// // 	val := s.network.Validators[0]

// // 	testCases := []struct {
// // 		name      string
// // 		args      []string
// // 		expectErr bool
// // 	}{
// // 		{
// // 			"query pool estimate swap exact amount in", // osmosisd query gamm estimate-swap-exact-amount-in 1 cosmos1n8skk06h3kyh550ad9qketlfhc2l5dsdevd3hq 10.0stake --swap-route-pool-ids=1 --swap-route-denoms=node0token
// // 			[]string{
// // 				"1",
// // 				"cosmos1n8skk06h3kyh550ad9qketlfhc2l5dsdevd3hq",
// // 				"10.0stake",
// // 				fmt.Sprintf("--%s=%d", cli.FlagSwapRoutePoolIds, 1),
// // 				fmt.Sprintf("--%s=%s", cli.FlagSwapRouteDenoms, "node0token"),
// // 				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
// // 			},
// // 			false,
// // 		},
// // 	}

// // 	for _, tc := range testCases {
// // 		tc := tc

// // 		s.Run(tc.name, func() {
// // 			cmd := cli.GetCmdEstimateSwapExactAmountIn()
// // 			clientCtx := val.ClientCtx

// // 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// // 			if tc.expectErr {
// // 				s.Require().Error(err)
// // 			} else {
// // 				resp := types.QuerySwapExactAmountInResponse{}
// // 				s.Require().NoError(err, out.String())
// // 				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
// // 			}
// // 		})
// // 	}
// // }

// // func (s *IntegrationTestSuite) TestGetCmdEstimateSwapExactAmountOut() {
// // 	val := s.network.Validators[0]

// // 	testCases := []struct {
// // 		name      string
// // 		args      []string
// // 		expectErr bool
// // 	}{
// // 		{
// // 			"query pool estimate swap exact amount in", // osmosisd query gamm estimate-swap-exact-amount-in 1 cosmos1n8skk06h3kyh550ad9qketlfhc2l5dsdevd3hq 10.0stake --swap-route-pool-ids=1 --swap-route-denoms=node0token
// // 			[]string{
// // 				"1",
// // 				val.Address.String(),
// // 				"10.0stake",
// // 				fmt.Sprintf("--%s=%d", cli.FlagSwapRoutePoolIds, 1),
// // 				fmt.Sprintf("--%s=%s", cli.FlagSwapRouteDenoms, "node0token"),
// // 				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
// // 			},
// // 			false,
// // 		},
// // 	}

// // 	for _, tc := range testCases {
// // 		tc := tc

// // 		s.Run(tc.name, func() {
// // 			cmd := cli.GetCmdEstimateSwapExactAmountOut()
// // 			clientCtx := val.ClientCtx

// // 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// // 			if tc.expectErr {
// // 				s.Require().Error(err)
// // 			} else {
// // 				resp := types.QuerySwapExactAmountOutResponse{}
// // 				s.Require().NoError(err, out.String())
// // 				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
// // 			}
// // 		})
// // 	}
// // }
