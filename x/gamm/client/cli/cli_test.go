package cli_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/c-osmosis/osmosis/app"
	"github.com/c-osmosis/osmosis/x/gamm/client/cli"
	gammtestutil "github.com/c-osmosis/osmosis/x/gamm/client/testutil"
	"github.com/c-osmosis/osmosis/x/gamm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tm-db"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	encCfg := app.MakeEncodingConfig()

	s.cfg = network.Config{
		Codec:             encCfg.Marshaler,
		TxConfig:          encCfg.TxConfig,
		LegacyAmino:       encCfg.Amino,
		InterfaceRegistry: encCfg.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor: func(val network.Validator) servertypes.Application {
			return app.NewOsmosisApp(
				val.Ctx.Logger, dbm.NewMemDB(), nil, true, make(map[int64]bool), val.Ctx.Config.RootDir, 0,
				encCfg,
				simapp.EmptyAppOptions{},
				baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
			)
		},
		GenesisState:    app.ModuleBasics.DefaultGenesis(encCfg.Marshaler),
		TimeoutCommit:   2 * time.Second,
		ChainID:         "osmosis-1",
		NumValidators:   1,
		BondDenom:       sdk.DefaultBondDenom,
		MinGasPrices:    fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom),
		AccountTokens:   sdk.TokensFromConsensusPower(1000),
		StakingTokens:   sdk.TokensFromConsensusPower(500),
		BondedTokens:    sdk.TokensFromConsensusPower(100),
		PruningStrategy: storetypes.PruningOptionNothing,
		CleanupDir:      true,
		SigningAlgo:     string(hd.Secp256k1Type),
		KeyringOptions:  []keyring.Option{},
	}

	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	val := s.network.Validators[0]

	// create a new pool
	_, err = gammtestutil.MsgCreatePool(val.ClientCtx, val.Address, "5stake,5node0token", "100stake,100node0token", "0.01", "0.01")
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestNewCreatePoolCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewCreatePoolAddr", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 20000), sdk.NewInt64Coin("node0token", 20000)), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err)

	okJSON := testutil.WriteToNewTempFile(s.T(), fmt.Sprintf(`
	{
	  "%s": "1node0token,3stake",
	  "%s": "100node0token,100stake",
	  "%s": "0.001",
	  "%s": "0.001"
	}
	`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee))

	badJSON := testutil.WriteToNewTempFile(s.T(), "bad json")

	// this badJSON is missing quotes around the FlagExitFee value
	badJSON2 := testutil.WriteToNewTempFile(s.T(), fmt.Sprintf(`
	{
	  "%s": "1node0token,3stake",
	  "%s": "100node0token,100stake",
	  "%s": "0.001",
	  "%s": 0.001
	}
	`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee))

	testCases := []struct {
		name         string
		json         string
		args         []string
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
			`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee))),
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 4,
		},
		{
			"two tokens pair pool",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagWeights, "1node0token,3stake"),
				fmt.Sprintf("--%s=%s", cli.FlagInitialDeposit, "100node0token,100stake"),
				fmt.Sprintf("--%s=%s", cli.FlagSwapFee, "0.001"),
				fmt.Sprintf("--%s=%s", cli.FlagExitFee, "0.001"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{ // --record-tokens=100.0stake2 --record-tokens=100.0stake --record-tokens-weight=5 --record-tokens-weight=5 --swap-fee=0.01 --exit-fee=0.01 --from=validator --keyring-backend=test --chain-id=testing --yes
			"three tokens pair pool - insufficient balance check",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagWeights, "1node0token,1stake,2btc"),
				fmt.Sprintf("--%s=%s", cli.FlagInitialDeposit, "100node0token,100stake,100btc"),
				fmt.Sprintf("--%s=%s", cli.FlagSwapFee, "0.001"),
				fmt.Sprintf("--%s=%s", cli.FlagExitFee, "0.001"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 5,
		},
		{
			"future governor address",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagWeights, "1node0token,3stake"),
				fmt.Sprintf("--%s=%s", cli.FlagInitialDeposit, "100node0token,100stake"),
				fmt.Sprintf("--%s=%s", cli.FlagSwapFee, "0.001"),
				fmt.Sprintf("--%s=%s", cli.FlagExitFee, "0.001"),
				fmt.Sprintf("--%s=%s", cli.FlagFutureGovernor, "cosmos1fqlr98d45v5ysqgp6h56kpujcj4cvsjn6mkrwy"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"future governor time",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagWeights, "1node0token,3stake"),
				fmt.Sprintf("--%s=%s", cli.FlagInitialDeposit, "100node0token,100stake"),
				fmt.Sprintf("--%s=%s", cli.FlagSwapFee, "0.001"),
				fmt.Sprintf("--%s=%s", cli.FlagExitFee, "0.001"),
				fmt.Sprintf("--%s=%s", cli.FlagFutureGovernor, "2h"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"future governor token + time",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagWeights, "1node0token,3stake"),
				fmt.Sprintf("--%s=%s", cli.FlagInitialDeposit, "100node0token,100stake"),
				fmt.Sprintf("--%s=%s", cli.FlagSwapFee, "0.001"),
				fmt.Sprintf("--%s=%s", cli.FlagExitFee, "0.001"),
				fmt.Sprintf("--%s=%s", cli.FlagFutureGovernor, "token,1000h"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"invalid future governor",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagWeights, "1node0token,3stake"),
				fmt.Sprintf("--%s=%s", cli.FlagInitialDeposit, "100node0token,100stake"),
				fmt.Sprintf("--%s=%s", cli.FlagSwapFee, "0.001"),
				fmt.Sprintf("--%s=%s", cli.FlagExitFee, "0.001"),
				fmt.Sprintf("--%s=%s", cli.FlagFutureGovernor, "validdenom,invalidtime"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 7,
		},
		{
			"pool json",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagPoolFile, okJSON.Name()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"bad pool json",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagPoolFile, badJSON.Name()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"bad pool json 2",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagPoolFile, badJSON2.Name()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"nonexistant pool json",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagPoolFile, "fileDoesNotExist"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"incompatible flags with --pool-file",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagPoolFile, okJSON.Name()),
				fmt.Sprintf("--%s=%s", cli.FlagSwapFee, "0.001"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCreatePoolCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s IntegrationTestSuite) TestNewJoinPoolCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewJoinPoolAddr", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 20000), sdk.NewInt64Coin("node0token", 20000)), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"join pool with insufficient balance",
			[]string{
				fmt.Sprintf("--%s=%d", cli.FlagPoolId, 1),
				fmt.Sprintf("--%s=%s", cli.FlagMaxAmountsIn, "100stake"),
				fmt.Sprintf("--%s=%s", cli.FlagShareAmountOut, "1000000000"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 6,
		},
		{
			"join pool with sufficient balance",
			[]string{ // join-pool --pool-id=1 --max-amounts-in=100stake --share-amount-out=100 --from=validator --keyring-backend=test --chain-id=testing --yes
				fmt.Sprintf("--%s=%d", cli.FlagPoolId, 1),
				fmt.Sprintf("--%s=%s", cli.FlagMaxAmountsIn, "100stake"),
				fmt.Sprintf("--%s=%s", cli.FlagShareAmountOut, "10000000"),
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
			cmd := cli.NewJoinPoolCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s IntegrationTestSuite) TestNewExitPoolCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"ask too much when exit",
			[]string{ // --min-amounts-out=100stake --pool-id=1 --share-amount-in=10 --from=validator --keyring-backend=test --chain-id=testing --yes
				fmt.Sprintf("--%s=%d", cli.FlagPoolId, 1),
				fmt.Sprintf("--%s=%s", cli.FlagShareAmountIn, "20000000"),
				fmt.Sprintf("--%s=%s", cli.FlagMinAmountsOut, "20stake"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 7,
		},
		{
			"ask enough when exit",
			[]string{ // --min-amounts-out=100stake --pool-id=1 --share-amount-in=10 --from=validator --keyring-backend=test --chain-id=testing --yes
				fmt.Sprintf("--%s=%d", cli.FlagPoolId, 1),
				fmt.Sprintf("--%s=%s", cli.FlagShareAmountIn, "20000000"),
				fmt.Sprintf("--%s=%s", cli.FlagMinAmountsOut, "10stake"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
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
			cmd := cli.NewExitPoolCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s IntegrationTestSuite) TestNewSwapExactAmountOutCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewSwapExactAmountOut", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 20000), sdk.NewInt64Coin("node0token", 20000)), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
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
			"swap exact amount out", // osmosisd tx gamm swap-exact-amount-out 10stake 20 --swap-route-pool-ids=1 --swap-route-denoms=node0token --from=validator --keyring-backend=test --chain-id=testing --yes
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
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s IntegrationTestSuite) TestNewJoinSwapExternAmountInCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewJoinSwapExternAmountIn", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(20000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"join swap extern amount in", // osmosisd tx gamm join-swap-extern-amount-in --pool-id=1 10stake 1 --from=validator --keyring-backend=test --chain-id=testing --yes
			[]string{
				"10stake", "1",
				fmt.Sprintf("--%s=%d", cli.FlagPoolId, 1),
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
			cmd := cli.NewJoinSwapExternAmountIn()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s IntegrationTestSuite) TestNewExitSwapExternAmountOutCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"exit swap extern amount out", // osmosisd tx gamm exit-swap-extern-amount-out --pool-id=1 10stake 1 --from=validator --keyring-backend=test --chain-id=testing --yes
			[]string{
				"10stake", "10000000",
				fmt.Sprintf("--%s=%d", cli.FlagPoolId, 1),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
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
			cmd := cli.NewExitSwapExternAmountOut()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s IntegrationTestSuite) TestNewJoinSwapShareAmountOutCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewJoinSwapShareAmountOutAddr", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(20000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"join swap share amount out", // osmosisd tx gamm join-swap-share-amount-out --pool-id=1 stake 10 1 --from=validator --keyring-backend=test --chain-id=testing --yes
			[]string{
				"stake", "50", "5000000",
				fmt.Sprintf("--%s=%d", cli.FlagPoolId, 1),
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
			cmd := cli.NewJoinSwapShareAmountOut()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s IntegrationTestSuite) TestNewExitSwapShareAmountInCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"exit swap share amount in", // osmosisd tx gamm exit-swap-share-amount-in --pool-id=1 stake 10 1 --from=validator --keyring-backend=test --chain-id=testing --yes
			[]string{
				"stake", "10000000", "1",
				fmt.Sprintf("--%s=%d", cli.FlagPoolId, 1),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
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
			cmd := cli.NewExitSwapShareAmountIn()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdPools() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query pools",
			[]string{
				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdPools() // osmosisd query gamm pools
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				resp := types.QueryPoolsResponse{}
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &resp), out.String())

				s.Require().Greater(len(resp.Pools), 0, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdPool() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query pool by id", // osmosisd query gamm pool 1
			[]string{
				"1",
				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdPool()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())

				resp := types.QueryPoolResponse{}
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &resp), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdPoolParams() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query pool params by id", // osmosisd query gamm pool-params 1
			[]string{
				"1",
				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdPoolParams()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				resp := types.QueryPoolParamsResponse{}
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &resp), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdRecords() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query pool records by id", // osmosisd query gamm records 1
			[]string{
				"1",
				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdRecords()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				resp := types.QueryPoolAssetsResponse{}
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &resp), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdTotalShare() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query pool total share by id", // osmosisd query gamm total-share 1
			[]string{
				"1",
				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdTotalShare()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				resp := types.QueryTotalShareResponse{}
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &resp), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdSpotPrice() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query pool spot price", // osmosisd query gamm spot-price 1 stake node0token
			[]string{
				"1", "stake", "node0token",
				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdSpotPrice()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				resp := types.QuerySpotPriceResponse{}
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &resp), out.String())
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
// 				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &resp), out.String())
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
// 				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &resp), out.String())
// 			}
// 		})
// 	}
// }

func (s IntegrationTestSuite) TestNewSwapExactAmountInCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewSwapExactAmountIn", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 20000), sdk.NewInt64Coin("node0token", 20000)), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
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
			"swap exact amount in", // osmosisd tx gamm swap-exact-amount-in 10stake 3 --swap-route-pool-ids=1 --swap-route-denoms=node0token --from=validator --keyring-backend=test --chain-id=testing --yes
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
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
