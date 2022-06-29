package cli_test

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/app"
	gammtestutil "github.com/osmosis-labs/osmosis/v7/x/gamm/client/testutil"
	"github.com/osmosis-labs/osmosis/v7/x/incentives/client/cli"
	"github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	lockuptestutil "github.com/osmosis-labs/osmosis/v7/x/lockup/client/testutil"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.cfg = app.DefaultConfig()

	// modification to pay pool creation fee with test bond denom "stake"
	genesisState := app.ModuleBasics.DefaultGenesis(s.cfg.Codec)
	gammGen := gammtypes.DefaultGenesis()
	gammGen.Params.PoolCreationFee = sdk.Coins{sdk.NewInt64Coin(s.cfg.BondDenom, 1000000)}
	gammGenJson := s.cfg.Codec.MustMarshalJSON(gammGen)
	genesisState[gammtypes.ModuleName] = gammGenJson
	s.cfg.GenesisState = genesisState

	s.network = network.New(s.T(), s.cfg)

	val := s.network.Validators[0]

	// create a pool to receive gamm tokens
	_, err := gammtestutil.MsgCreatePool(s.T(), val.ClientCtx, val.Address, "5stake,5node0token", "100stake,100node0token", "0.01", "0.01", "")
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	lockAmt, err := sdk.ParseCoinNormalized(fmt.Sprint("100000gamm/pool/1"))
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	// lock gamm pool tokens to create lock ID 1
	lockuptestutil.MsgLockTokens(val.ClientCtx, val.Address, lockAmt, "24h")

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	secondLockAmt, err := sdk.ParseCoinNormalized(fmt.Sprint("100000gamm/pool/1"))
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	// lock gamm pool tokens to create lock ID 2
	lockuptestutil.MsgLockTokens(val.ClientCtx, val.Address, secondLockAmt, "168h")

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestGetCmdGauges() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		expectErr bool
		args      []string
		respType  proto.Message
	}{
		{
			"query gauges",
			false,
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			&types.GaugesResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdGauges()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdToDistributeCoins() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query to distribute coins",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			&types.ModuleToDistributeCoinsResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdToDistributeCoins()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdDistributedCoins() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query to distribute coins",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			&types.ModuleDistributedCoinsResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdDistributedCoins()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdGaugeByID() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query gauge by id",
			[]string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			&types.GaugeByIDResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdGaugeByID()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdActiveGauges() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query active gauges",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			&types.ActiveGaugesResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdActiveGauges()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdActiveGaugesPerDenom() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query active gauges per denom",
			[]string{s.cfg.BondDenom, fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			&types.ActiveGaugesPerDenomResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdActiveGaugesPerDenom()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdUpcomingGauges() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query upcoming gauges",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			&types.UpcomingGaugesResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdUpcomingGauges()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdUpcomingGaugesPerDenom() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query upcoming gauges per denom",
			[]string{s.cfg.BondDenom, fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			&types.UpcomingGaugesPerDenomResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdUpcomingGaugesPerDenom()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdRewardsEst() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query rewards estimation by owner",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagOwner, val.Address.String()),
				fmt.Sprintf("--%s=100", cli.FlagEndEpoch),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			&types.RewardsEstResponse{},
		},
		{
			"query rewards estimation by lock id",
			[]string{
				fmt.Sprintf("--%s=1,2", cli.FlagLockIds),
				fmt.Sprintf("--%s=100", cli.FlagEndEpoch),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			&types.RewardsEstResponse{},
		},
		{
			"query rewards estimation with empty end epoch",
			[]string{
				fmt.Sprintf("--%s=1,2", cli.FlagLockIds),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			&types.RewardsEstResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdRewardsEst()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
