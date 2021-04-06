package cli_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/c-osmosis/osmosis/app"
	"github.com/c-osmosis/osmosis/app/params"
	"github.com/c-osmosis/osmosis/x/claim/client/cli"
	"github.com/c-osmosis/osmosis/x/claim/types"
	claimtypes "github.com/c-osmosis/osmosis/x/claim/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tm-db"
)

var addr1 sdk.AccAddress
var mnemonic1 = "fork thank one athlete focus observe cupboard leaf unusual sad town just sunny jaguar world adult dinosaur price parade tiger jelly around kingdom false"
var pwd1 = ""
var addr2 sdk.AccAddress

func init() {
	params.SetBech32Prefixes()
	addr1, _ = sdk.AccAddressFromBech32("osm11l2ptnxekj836wxmc38gus7yem9ak5khrj47t04")
	addr2, _ = sdk.AccAddressFromBech32("osm11xwjpumw85skkh04xst8zyp6dn8mdvfvm9zvamv")
}

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {

	s.T().Log("setting up integration test suite")
	encCfg := app.MakeEncodingConfig()

	genState := app.ModuleBasics.DefaultGenesis(encCfg.Marshaler)
	claimGenState := claimtypes.DefaultGenesis()
	claimGenState.AirdropAmount = sdk.NewInt(30)
	claimGenState.Claimables = []banktypes.Balance{
		{
			Address: addr1.String(),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)),
		},
		{
			Address: addr2.String(),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 20)),
		},
	}
	claimGenStateBz := encCfg.Marshaler.MustMarshalJSON(claimGenState)
	genState[claimtypes.ModuleName] = claimGenStateBz

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
		GenesisState:    genState,
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
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestNewClaimCmd() {
	val := s.network.Validators[0]

	info, err := val.ClientCtx.Keyring.NewAccount("addr1", mnemonic1, pwd1, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)
	s.Require().Equal(addr1.String(), info.GetAddress().String())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		addr1,
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
			"claim 20stake from addr1",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr1.String()),
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
			cmd := cli.NewCmdClaim()
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

func (s *IntegrationTestSuite) TestCmdQueryClaimable() {
	val := s.network.Validators[0]

	testCases := []struct {
		name  string
		args  []string
		coins sdk.Coins
	}{
		{
			"query claimable amount",
			[]string{
				addr2.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			sdk.Coins{sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(20))},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryClaimable()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.ClaimableResponse
			s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(tc.coins.String(), sdk.Coins(result.Coins).String())
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
