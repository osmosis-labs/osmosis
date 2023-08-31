package testutil

import (
	"fmt"
	"time"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	osmoapp "github.com/osmosis-labs/osmosis/v19/app"
	"github.com/osmosis-labs/osmosis/v19/x/gov/client/cli"
	"github.com/osmosis-labs/osmosis/v19/x/gov/types"
)

type DepositTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
	fees    string
}

func NewDepositTestSuite(cfg network.Config) *DepositTestSuite {
	return &DepositTestSuite{cfg: cfg}
}

func (s *DepositTestSuite) SetupSuite() {
	s.T().Log("setting up test suite")

	app := osmoapp.Setup(false)
	s.cfg = network.DefaultConfig()
	encCfg := osmoapp.MakeEncodingConfig()
	s.cfg.Codec = encCfg.Marshaler
	s.cfg.TxConfig = encCfg.TxConfig
	s.cfg.LegacyAmino = encCfg.Amino
	s.cfg.InterfaceRegistry = encCfg.InterfaceRegistry

	// export the state and import it into a new app
	govGenState := types.DefaultGenesisState()
	genesisState := osmoapp.ModuleBasics.DefaultGenesis(app.AppCodec())
	govGenState.VotingParams.VotingPeriod = time.Second * 5
	govGenState.DepositParams.MaxDepositPeriod = time.Second * 5
	genesisState[types.ModuleName] = app.AppCodec().MustMarshalJSON(govGenState)

	s.cfg.GenesisState = genesisState
	s.cfg.AppConstructor = NewAppConstructor(encCfg)
	s.cfg.NumValidators = 1
	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

}

func (s *DepositTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	s.network.Cleanup()
}

func (s *DepositTestSuite) TestQueryDepositsInitialDeposit() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	initialDeposit := sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Sub(sdk.NewInt(20))).String()

	// create a proposal with deposit
	_, err := MsgSubmitProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 1", "Where is the title!?", types.ProposalTypeText,
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, initialDeposit))
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	// deposit more amount
	_, err = MsgDeposit(clientCtx, val.Address.String(), "1", sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(50)).String())
	s.Require().NoError(err)

	// query deposits
	deposits := s.queryDeposits(val, "1", false)
	s.Require().Equal(len(deposits.Deposits), 1)
	// verify initial deposit
	s.Require().Equal(deposits.Deposits[0].Amount.String(), sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10000030)).String())
}

func (s *DepositTestSuite) TestQueryDepositsWithoutInitialDeposit() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	// create a proposal without deposit
	_, err := MsgSubmitProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 2", "Where is the title!?", types.ProposalTypeText)
	s.Require().NoError(err)

	// deposit amount
	_, err = MsgDeposit(clientCtx, val.Address.String(), "2", sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Add(sdk.NewInt(50))).String())
	s.Require().NoError(err)

	// query deposit
	deposit := s.queryDeposit(val, "2", false)
	s.Require().Equal(deposit.Amount.String(), sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Add(sdk.NewInt(50))).String())

	// query deposits
	deposits := s.queryDeposits(val, "2", false)
	s.Require().Equal(len(deposits.Deposits), 1)
	// verify initial deposit
	s.Require().Equal(deposits.Deposits[0].Amount.String(), sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Add(sdk.NewInt(50))).String())
}

func (s *DepositTestSuite) TestQueryProposalNotEnoughDeposits() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	initialDeposit := sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Sub(sdk.NewInt(2000))).String()

	// create a proposal with deposit
	_, err := MsgSubmitProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 3", "Where is the title!?", types.ProposalTypeText,
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, initialDeposit))
	s.Require().NoError(err)

	// wait for voting period to end
	time.Sleep(13 * time.Second)

	// query proposal
	args := []string{"3", fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	cmd := cli.GetCmdQueryProposal()
	resp, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	fmt.Println(err)
	fmt.Println(resp)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "proposal 3 doesn't exist")
}

func (s *DepositTestSuite) queryDeposits(val *network.Validator, proposalID string, exceptErr bool) types.QueryDepositsResponse {
	args := []string{proposalID, fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	var depositsRes types.QueryDepositsResponse
	cmd := cli.GetCmdQueryDeposits()
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	fmt.Println(err)
	fmt.Println(out)
	if exceptErr {
		s.Require().Error(err)
		return types.QueryDepositsResponse{}
	}
	s.Require().NoError(err)

	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &depositsRes))
	return depositsRes
}

func (s *DepositTestSuite) queryDeposit(val *network.Validator, proposalID string, exceptErr bool) types.Deposit {
	args := []string{proposalID, val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	var depositRes types.Deposit
	cmd := cli.GetCmdQueryDeposit()
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	fmt.Println(err)
	fmt.Println(out)
	if exceptErr {
		s.Require().Error(err)
		return types.Deposit{}
	}
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &depositRes))
	return depositRes
}
