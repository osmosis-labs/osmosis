package cli_test

import (
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v10/app"
	"github.com/osmosis-labs/osmosis/v10/x/pool-incentives/client/cli"
	"github.com/osmosis-labs/osmosis/v10/x/pool-incentives/types"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

//SetupSuite sets up an integration test suite.
func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.cfg = app.DefaultConfig()

	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

// TearDownSuite tears down the integration test suite.
func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

// TestGetCmdGaugeIds tests gaugeID CLI query commands
func (s *IntegrationTestSuite) TestGetCmdGaugeIds() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query gauge-ids",
			false, &types.QueryGaugeIdsResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdGaugeIds()
			clientCtx := val.ClientCtx

			args := []string{"1"}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

// TestGetCmdDistrInfo tests distribution info CLI query commands
func (s *IntegrationTestSuite) TestGetCmdDistrInfo() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query distr-info",
			false, &types.QueryDistrInfoResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdDistrInfo()
			clientCtx := val.ClientCtx

			args := []string{}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

// TestGetCmdParams tests module params CLI query commands
func (s *IntegrationTestSuite) TestGetCmdParams() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query module params",
			false, &types.QueryParamsResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdParams()
			clientCtx := val.ClientCtx

			args := []string{}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

// TestGetCmdLockableDurations tests lockable duration CLI query commands
func (s *IntegrationTestSuite) TestGetCmdLockableDurations() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query lockable durations",
			false, &types.QueryLockableDurationsResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdLockableDurations()
			clientCtx := val.ClientCtx

			args := []string{}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

// TestGetCmdIncentivizedPools tests incentivized pools CLI query commands
func (s *IntegrationTestSuite) TestGetCmdIncentivizedPools() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query incentivized pools",
			false, &types.QueryIncentivizedPoolsResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdIncentivizedPools()
			clientCtx := val.ClientCtx

			args := []string{}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

// TestGetCmdExternalIncentiveGauges tests incentivized pools CLI query commands
func (s *IntegrationTestSuite) TestGetCmdExternalIncentiveGauges() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query external incentivized pools",
			false, &types.QueryExternalIncentiveGaugesResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdExternalIncentiveGauges()
			clientCtx := val.ClientCtx

			args := []string{}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}
