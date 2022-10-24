package cli_test

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app"
	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	gammtestutil "github.com/osmosis-labs/osmosis/v12/x/gamm/client/testutil"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/client/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.cfg = app.DefaultConfig()
	s.cfg.GenesisState = gammtestutil.UpdateTxFeeDenom(s.cfg.Codec, s.cfg.BondDenom)

	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	val := s.network.Validators[0]

	// create a new pool
	_, err = gammtestutil.MsgCreatePool(s.T(), val.ClientCtx, val.Address, "5stake,5node0token", "100stake,100node0token", "0.01", "0.01", "")
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
	suite.Run(t, new(IntegrationTestSuite))
}

func (s IntegrationTestSuite) TestNewSwapExactAmountOutCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewSwapExactAmountOut",
		keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 20000), sdk.NewInt64Coin("node0token", 20000)), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		osmoutils.DefaultFeeString(s.cfg),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name string
		args []string

		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"swap exact amount out", // osmosisd tx swaprouter swap-exact-amount-out 10stake 20 --swap-route-pool-ids=1 --swap-route-denoms=node0token --from=validator --keyring-backend=test --chain-id=testing --yes
			[]string{
				"10stake", "20",
				fmt.Sprintf("--%s=%d", cli.FlagSwapRoutePoolIds, 1),
				fmt.Sprintf("--%s=%s", cli.FlagSwapRouteDenoms, "node0token"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewSwapExactAmountOutCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

// func (s *IntegrationTestSuite) TestGetCmdEstimateSwapExactAmountIn() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expectErr bool
// 	}{
// 		{
// 			"query pool estimate swap exact amount in", // osmosisd query gamm estimate-swap-exact-amount-in 1 cosmos1n8skk06h3kyh550ad9qketlfhc2l5dsdevd3hq 10.0stake --swap-route-pool-ids=1 --swap-route-denoms=node0token
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
// 			"query pool estimate swap exact amount in", // osmosisd query gamm estimate-swap-exact-amount-in 1 cosmos1n8skk06h3kyh550ad9qketlfhc2l5dsdevd3hq 10.0stake --swap-route-pool-ids=1 --swap-route-denoms=node0token
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

func (s IntegrationTestSuite) TestNewSwapExactAmountInCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewSwapExactAmountIn",
		keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 20000), sdk.NewInt64Coin("node0token", 20000)), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		osmoutils.DefaultFeeString(s.cfg),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name string
		args []string

		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"swap exact amount in", // osmosisd tx swaprouter swap-exact-amount-in 10stake 3 --swap-route-pool-ids=1 --swap-route-denoms=node0token --from=validator --keyring-backend=test --chain-id=testing --yes
			[]string{
				"10stake", "3",
				fmt.Sprintf("--%s=%d", cli.FlagSwapRoutePoolIds, 1),
				fmt.Sprintf("--%s=%s", cli.FlagSwapRouteDenoms, "node0token"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewSwapExactAmountInCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}
