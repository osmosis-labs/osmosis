package cli_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/app"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/cli"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/queryproto"
	poolmanagertestutil "github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/testutil"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

var testAddresses = osmoutils.CreateRandomAccounts(3)

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.cfg = app.DefaultConfig()
	s.cfg.GenesisState = poolmanagertestutil.UpdateTxFeeDenom(s.cfg.Codec, s.cfg.BondDenom)

	net, err := network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)
	s.network = net

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	val := s.network.Validators[0]

	// create a new pool
	_, err = poolmanagertestutil.MsgCreatePool(s.T(), val.ClientCtx, val.Address, "5stake,5node0token", "100stake,100node0token", "0.01", "0.01", "")
	s.Require().NoError(err)

	s.T().Log("test")

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func TestIntegrationTestSuite(t *testing.T) {
	// TODO: re-enable this once poolmanager is fully merged.
	t.SkipNow()

	suite.Run(t, new(IntegrationTestSuite))
}

func TestNewSwapExactAmountOutCmd(t *testing.T) {
	desc, _ := cli.NewSwapExactAmountOutCmd()
	tcs := map[string]osmocli.TxCliTestCase[*types.MsgSwapExactAmountOut]{
		"swap exact amount out": {
			Cmd: "10stake 20 --swap-route-pool-ids=1 --swap-route-denoms=node0token --from=" + testAddresses[0].String(),
			ExpectedMsg: &types.MsgSwapExactAmountOut{
				Sender:           testAddresses[0].String(),
				Routes:           []types.SwapAmountOutRoute{{PoolId: 1, TokenInDenom: "node0token"}},
				TokenInMaxAmount: osmomath.NewIntFromUint64(20),
				TokenOut:         sdk.NewInt64Coin("stake", 10),
			},
		},
	}
	osmocli.RunTxTestCases(t, desc, tcs)
}

// func (s *IntegrationTestSuite) TestGetCmdEstimateSwapExactAmountIn() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expectErr bool
// 	}{
// 		{
// 			"query pool estimate swap exact amount in", // osmosisd query poolmanager estimate-swap-exact-amount-in 1 cosmos1n8skk06h3kyh550ad9qketlfhc2l5dsdevd3hq 10.0stake --swap-route-pool-ids=1 --swap-route-denoms=node0token
// 			[]string{
// 				"1",
// 				"cosmos1n8skk06h3kyh550ad9qketlfhc2l5dsdevd3hq",
// 				"10.0stake",
// 				fmt.Sprintf("--%s=%d", cli.FlagSwapRoutePoolIds, 1),
// 				fmt.Sprintf("--%s=%s", cli.FlagSwapRouteDenoms, "node0token"),
// 				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
// 			},
// 			false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdEstimateSwapExactAmountIn()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			if tc.expectErr {
// 				s.Require().Error(err)
// 			} else {
// 				resp := types.QuerySwapExactAmountInResponse{}
// 				s.Require().NoError(err, out.String())
// 				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
// 			}
// 		})
// 	}
// }

// func (s *IntegrationTestSuite) TestGetCmdEstimateSwapExactAmountOut() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expectErr bool
// 	}{
// 		{
// 			"query pool estimate swap exact amount in", // osmosisd query poolmanager estimate-swap-exact-amount-in 1 cosmos1n8skk06h3kyh550ad9qketlfhc2l5dsdevd3hq 10.0stake --swap-route-pool-ids=1 --swap-route-denoms=node0token
// 			[]string{
// 				"1",
// 				val.Address.String(),
// 				"10.0stake",
// 				fmt.Sprintf("--%s=%d", cli.FlagSwapRoutePoolIds, 1),
// 				fmt.Sprintf("--%s=%s", cli.FlagSwapRouteDenoms, "node0token"),
// 				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
// 			},
// 			false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdEstimateSwapExactAmountOut()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			if tc.expectErr {
// 				s.Require().Error(err)
// 			} else {
// 				resp := types.QuerySwapExactAmountOutResponse{}
// 				s.Require().NoError(err, out.String())
// 				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
// 			}
// 		})
// 	}
// }

func TestNewSwapExactAmountInCmd(t *testing.T) {
	desc, _ := cli.NewSwapExactAmountInCmd()
	tcs := map[string]osmocli.TxCliTestCase[*types.MsgSwapExactAmountIn]{
		"swap exact amount in": {
			Cmd: "10stake 3 --swap-route-pool-ids=1 --swap-route-denoms=node0token --from=" + testAddresses[0].String(),
			ExpectedMsg: &types.MsgSwapExactAmountIn{
				Sender:            testAddresses[0].String(),
				Routes:            []types.SwapAmountInRoute{{PoolId: 1, TokenOutDenom: "node0token"}},
				TokenIn:           sdk.NewInt64Coin("stake", 10),
				TokenOutMinAmount: osmomath.NewIntFromUint64(3),
			},
		},
	}
	osmocli.RunTxTestCases(t, desc, tcs)
}

func TestGetCmdNumPools(t *testing.T) {
	desc, _ := cli.GetCmdNumPools()
	tcs := map[string]osmocli.QueryCliTestCase[*queryproto.NumPoolsRequest]{
		"basic test": {
			Cmd:           "",
			ExpectedQuery: &queryproto.NumPoolsRequest{},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdEstimateSwapExactAmountIn(t *testing.T) {
	desc, _ := cli.GetCmdEstimateSwapExactAmountIn()
	tcs := map[string]osmocli.QueryCliTestCase[*queryproto.EstimateSwapExactAmountInRequest]{
		"basic test": {
			Cmd: "1 10stake --swap-route-pool-ids=2 --swap-route-denoms=node0token",
			ExpectedQuery: &queryproto.EstimateSwapExactAmountInRequest{
				PoolId:  1,
				TokenIn: "10stake",
				Routes:  []types.SwapAmountInRoute{{PoolId: 2, TokenOutDenom: "node0token"}},
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdEstimateSwapExactAmountOut(t *testing.T) {
	desc, _ := cli.GetCmdEstimateSwapExactAmountOut()
	tcs := map[string]osmocli.QueryCliTestCase[*queryproto.EstimateSwapExactAmountOutRequest]{
		"basic test": {
			Cmd: "1 10stake --swap-route-pool-ids=2 --swap-route-denoms=node0token",
			ExpectedQuery: &queryproto.EstimateSwapExactAmountOutRequest{
				PoolId:   1,
				TokenOut: "10stake",
				Routes:   []types.SwapAmountOutRoute{{PoolId: 2, TokenInDenom: "node0token"}},
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdEstimateSinglePoolSwapExactAmountIn(t *testing.T) {
	desc, _ := cli.GetCmdEstimateSinglePoolSwapExactAmountIn()
	tcs := map[string]osmocli.QueryCliTestCase[*queryproto.EstimateSinglePoolSwapExactAmountInRequest]{
		"basic test": {
			Cmd: "1 10stake node0token",
			ExpectedQuery: &queryproto.EstimateSinglePoolSwapExactAmountInRequest{
				PoolId:        1,
				TokenIn:       "10stake",
				TokenOutDenom: "node0token",
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdEstimateSinglePoolSwapExactAmountOut(t *testing.T) {
	desc, _ := cli.GetCmdEstimateSinglePoolSwapExactAmountOut()
	tcs := map[string]osmocli.QueryCliTestCase[*queryproto.EstimateSinglePoolSwapExactAmountOutRequest]{
		"basic test": {
			Cmd: "1 node0token 10stake",
			ExpectedQuery: &queryproto.EstimateSinglePoolSwapExactAmountOutRequest{
				PoolId:       1,
				TokenInDenom: "node0token",
				TokenOut:     "10stake",
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func (s *IntegrationTestSuite) TestNewCreatePoolCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewCreatePoolAddr",
		keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	pubkey, err := info.GetPubKey()
	newAddr := sdk.AccAddress(pubkey.Address())

	_, err = clitestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 200000000), sdk.NewInt64Coin("node0token", 20000)),
		addresscodec.NewBech32Codec("osmo"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		osmoutils.DefaultFeeString(s.cfg),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		json         string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"one token pair pool",
			fmt.Sprintf(`
			{
			  "%s": "1node0token",
			  "%s": "100node0token",
			  "%s": "0.001",
			  "%s": "0.001"
			}
			`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee),
			true, &sdk.TxResponse{}, 4,
		},
		{
			"two tokens pair pool",
			fmt.Sprintf(`
			{
			  "%s": "1node0token,3stake",
			  "%s": "100node0token,100stake",
			  "%s": "0.001",
			  "%s": "0.001"
			}
			`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"change order of json fields",
			fmt.Sprintf(`
			{
			  "%s": "100node0token,100stake",
			  "%s": "0.001",
			  "%s": "1node0token,3stake",
			  "%s": "0.001"
			}
			`, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileWeights, cli.PoolFileExitFee),
			false, &sdk.TxResponse{}, 0,
		},
		{ // --record-tokens=100.0stake2 --record-tokens=100.0stake --record-tokens-weight=5 --record-tokens-weight=5 --swap-fee=0.01 --exit-fee=0.01 --from=validator --keyring-backend=test --chain-id=testing --yes
			"three tokens pair pool - insufficient balance check",
			fmt.Sprintf(`
			{
			  "%s": "1node0token,1stake,2btc",
			  "%s": "100node0token,100stake,100btc",
			  "%s": "0.001",
			  "%s": "0.001"
			}
			`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee),
			false, &sdk.TxResponse{}, 5,
		},
		{
			"future governor address",
			fmt.Sprintf(`
			{
			  "%s": "1node0token,3stake",
			  "%s": "100node0token,100stake",
			  "%s": "0.001",
			  "%s": "0.001",
			  "%s": "osmo1fqlr98d45v5ysqgp6h56kpujcj4cvsjnjq9nck"
			}
			`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee, cli.PoolFileFutureGovernor),
			false, &sdk.TxResponse{}, 0,
		},
		// Due to CI time concerns, we leave these CLI tests commented out, and instead guaranteed via
		// the logic tests.
		// {
		// 	"future governor time",
		// 	fmt.Sprintf(`
		// 	{
		// 	  "%s": "1node0token,3stake",
		// 	  "%s": "100node0token,100stake",
		// 	  "%s": "0.001",
		// 	  "%s": "0.001",
		// 	  "%s": "2h"
		// 	}
		// 	`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee, cli.PoolFileFutureGovernor),
		// 	false, &sdk.TxResponse{}, 0,
		// },
		// {
		// 	"future governor token + time",
		// 	fmt.Sprintf(`
		// 	{
		// 	  "%s": "1node0token,3stake",
		// 	  "%s": "100node0token,100stake",
		// 	  "%s": "0.001",
		// 	  "%s": "0.001",
		// 	  "%s": "token,1000h"
		// 	}
		// 	`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee, cli.PoolFileFutureGovernor),
		// 	false, &sdk.TxResponse{}, 0,
		// },
		// {
		// 	"invalid future governor",
		// 	fmt.Sprintf(`
		// 	{
		// 	  "%s": "1node0token,3stake",
		// 	  "%s": "100node0token,100stake",
		// 	  "%s": "0.001",
		// 	  "%s": "0.001",
		// 	  "%s": "validdenom,invalidtime"
		// 	}
		// 	`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee, cli.PoolFileFutureGovernor),
		// 	true, &sdk.TxResponse{}, 7,
		// },
		{
			"not valid json",
			"bad json",
			true, &sdk.TxResponse{}, 0,
		},
		{
			"bad pool json - missing quotes around exit fee",
			fmt.Sprintf(`
			{
			  "%s": "1node0token,3stake",
			  "%s": "100node0token,100stake",
			  "%s": "0.001",
			  "%s": 0.001
			}
	`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee),
			true, &sdk.TxResponse{}, 0,
		},
		{
			"empty pool json",
			"", true, &sdk.TxResponse{}, 0,
		},
		{
			"smooth change params",
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
			false, &sdk.TxResponse{}, 0,
		},
		{
			"smooth change params - no start time",
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
			false, &sdk.TxResponse{}, 0,
		},
		{
			"empty smooth change params",
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
			false, &sdk.TxResponse{}, 0,
		},
		{
			"smooth change params wrong type",
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
			true, &sdk.TxResponse{}, 0,
		},
		{
			"smooth change params missing duration",
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
			true, &sdk.TxResponse{}, 0,
		},
		{
			"unknown fields in json",
			fmt.Sprintf(`
			{
			  "%s": "1node0token",
			  "%s": "100node0token",
			  "%s": "0.001",
			  "%s": "0.001"
			  "unknown": true,
			}
			`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee),
			true, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCreatePoolCmd()
			clientCtx := val.ClientCtx

			jsonFile := testutil.WriteToNewTempFile(s.T(), tc.json)

			args := []string{
				fmt.Sprintf("--%s=%s", cli.FlagPoolFile, jsonFile.Name()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				osmoutils.DefaultFeeString(s.cfg),
				fmt.Sprintf("--%s=%s", flags.FlagGas, fmt.Sprint(400000)),
			}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType)
				s.Require().NoError(err, out.String())

				var txResp *sdk.TxResponse
				switch resp := tc.respType.(type) {
				case *sdk.TxResponse:
					txResp = resp
				default:
					s.T().Fatalf("unexpected response type: %T", tc.respType)
				}
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func TestEstimateTradeBasedOnPriceImpact(t *testing.T) {
	desc, _ := cli.GetCmdEstimateTradeBasedOnPriceImpact()
	tcs := map[string]osmocli.QueryCliTestCase[*queryproto.EstimateTradeBasedOnPriceImpactRequest]{
		"basic test": {
			Cmd: "100node0token stake 1 0.01 0.02",
			ExpectedQuery: &queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin: sdk.Coin{
					Denom:  "node0token",
					Amount: osmomath.NewInt(100),
				},
				ToCoinDenom:    "stake",
				PoolId:         1,
				MaxPriceImpact: osmomath.MustNewDecFromStr("0.01"), // equivalent to 0.01
				ExternalPrice:  osmomath.MustNewDecFromStr("0.02"), // equivalent to 0.02
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}
