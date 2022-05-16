package cli_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/osmosis-labs/osmosis/v8/app"
	"github.com/osmosis-labs/osmosis/v8/osmoutils"
	"github.com/osmosis-labs/osmosis/v8/x/lockup/client/cli"
	lockuptestutil "github.com/osmosis-labs/osmosis/v8/x/lockup/client/testutil"
	"github.com/osmosis-labs/osmosis/v8/x/lockup/types"

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

	cfg      network.Config
	network  *network.Network
	feeCoins sdk.Coins
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.cfg = app.DefaultConfig()
	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	dayLockAmt, err := sdk.ParseCoinNormalized(fmt.Sprintf("200%s", s.network.Config.BondDenom))
	s.Require().NoError(err)
	secLockAmt, err := sdk.ParseCoinNormalized(fmt.Sprintf("11%s", s.network.Config.BondDenom))
	s.Require().NoError(err)
	thirdLockAmt, err := sdk.ParseCoinNormalized(fmt.Sprintf("12%s", s.network.Config.BondDenom))
	s.Require().NoError(err)

	s.feeCoins = sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)))

	val := s.network.Validators[0]

	// lock tokens for a day
	_, err = lockuptestutil.MsgLockTokens(val.ClientCtx, val.Address, dayLockAmt, "24h")
	s.Require().NoError(err)

	// lock tokens for a second
	_, err = lockuptestutil.MsgLockTokens(val.ClientCtx, val.Address, secLockAmt, "1s")
	s.Require().NoError(err)

	// lock tokens for a second
	_, err = lockuptestutil.MsgLockTokens(val.ClientCtx, val.Address, thirdLockAmt, "1s")
	s.Require().NoError(err)

	// begin unlock all tokens
	_, err = lockuptestutil.MsgBeginUnlocking(val.ClientCtx, val.Address)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestNewLockTokensCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewValidator",
		keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(20000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, s.feeCoins.String()),
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
			"lock 201stake tokens for 1 day",
			[]string{
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(201))).String(),
				fmt.Sprintf("--%s=%s", cli.FlagDuration, "24h"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, s.feeCoins.String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewLockTokensCmd()
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

func (s *IntegrationTestSuite) TestBeginUnlockingCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("BeginUnlockingAcc",
		keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(20000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		osmoutils.DefaultFeeString(s.cfg),
	)
	s.Require().NoError(err)

	// lock tokens for a second
	_, err = lockuptestutil.MsgLockTokens(val.ClientCtx, newAddr, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(200))), "1s")
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"begin unlocking",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, s.feeCoins.String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewBeginUnlockingCmd()
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

func (s *IntegrationTestSuite) TestNewBeginUnlockPeriodLockCmd() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("BeginUnlockPeriodLockAcc",
		keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		clientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(20000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		osmoutils.DefaultFeeString(s.cfg),
	)
	s.Require().NoError(err)

	// lock tokens for a second
	txResp := sdk.TxResponse{}
	out, err := lockuptestutil.MsgLockTokens(clientCtx,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(200))),
		"1s")
	s.Require().NoError(err)
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	// This is a hardcoded path in the events to get the lockID
	// this is incredibly brittle...
	// fmt.Println(txResp.Logs[0])
	lockID := txResp.Logs[0].Events[2].Attributes[0].Value

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"begin unlocking by id",
			[]string{
				lockID,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, s.feeCoins.String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}
	fmt.Println(testCases[0].args)

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewBeginUnlockByIDCmd()

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

func (s *IntegrationTestSuite) TestCmdAccountUnlockableCoins() {
	val := s.network.Validators[0]

	testCases := []struct {
		name  string
		args  []string
		coins sdk.Coins
	}{
		{
			"query validator account unlockable coins",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			sdk.Coins{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdAccountUnlockableCoins()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.AccountUnlockableCoinsResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(tc.coins.String(), result.Coins.String())
		})
	}
}

func (s *IntegrationTestSuite) TestCmdAccountUnlockingCoins() {
	val := s.network.Validators[0]

	testCases := []struct {
		name  string
		args  []string
		coins sdk.Coins
	}{
		{
			"query validator account unlocking coins",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			sdk.Coins{sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(200))},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdAccountUnlockingCoins()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.AccountUnlockingCoinsResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(tc.coins.String(), result.Coins.String())
		})
	}
}

func (s IntegrationTestSuite) TestCmdModuleBalance() {
	val := s.network.Validators[0]

	testCases := []struct {
		name  string
		args  []string
		coins sdk.Coins
	}{
		{
			"query module balance",
			[]string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			sdk.Coins{sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(400))},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdModuleBalance()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.ModuleBalanceResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(tc.coins.String(), result.Coins.String())
		})
	}
}

func (s IntegrationTestSuite) TestCmdModuleLockedAmount() {
	val := s.network.Validators[0]

	testCases := []struct {
		name  string
		args  []string
		coins sdk.Coins
	}{
		{
			"query module locked balance",
			[]string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			sdk.Coins{sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(400))},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdModuleLockedAmount()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.ModuleLockedAmountResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(tc.coins.String(), result.Coins.String())
		})
	}
}

func (s IntegrationTestSuite) TestCmdAccountLockedCoins() {
	val := s.network.Validators[0]

	testCases := []struct {
		name  string
		args  []string
		coins sdk.Coins
	}{
		{
			"query account locked coins",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			sdk.Coins{sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(200))},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdAccountLockedCoins()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.ModuleLockedAmountResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(tc.coins.String(), result.Coins.String())
		})
	}
}

func (s IntegrationTestSuite) TestCmdAccountLockedPastTime() {
	val := s.network.Validators[0]

	timestamp := time.Now().Unix()
	testCases := []struct {
		name string
		args []string
	}{
		{
			"query account locked coins past time",
			[]string{
				val.Address.String(),
				fmt.Sprintf("%d", timestamp),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdAccountLockedPastTime()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.AccountLockedPastTimeResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Len(result.Locks, 1)
		})
	}
}

func (s IntegrationTestSuite) TestCmdAccountLockedPastTimeNotUnlockingOnly() {
	val := s.network.Validators[0]

	timestamp := time.Now().Unix()
	testCases := []struct {
		name string
		args []string
	}{
		{
			"query account locked coins past time not unlocking only",
			[]string{
				val.Address.String(),
				fmt.Sprintf("%d", timestamp),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdAccountLockedPastTimeNotUnlockingOnly()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.AccountLockedPastTimeNotUnlockingOnlyResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Len(result.Locks, 0)
		})
	}
}

func (s IntegrationTestSuite) TestCmdAccountUnlockedBeforeTime() {
	val := s.network.Validators[0]

	timestamp := time.Now().Unix()
	testCases := []struct {
		name string
		args []string
	}{
		{
			"query account locked coins before time",
			[]string{
				val.Address.String(),
				fmt.Sprintf("%d", timestamp),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdAccountUnlockedBeforeTime()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.AccountUnlockedBeforeTimeResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Len(result.Locks, 0)
		})
	}
}

func (s IntegrationTestSuite) TestCmdAccountLockedPastTimeDenom() {
	val := s.network.Validators[0]

	timestamp := time.Now().Unix()
	testCases := []struct {
		name string
		args []string
	}{
		{
			"query account locked coins past time denom",
			[]string{
				val.Address.String(),
				fmt.Sprintf("%d", timestamp),
				s.cfg.BondDenom,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdAccountLockedPastTimeDenom()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.AccountLockedPastTimeDenomResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Len(result.Locks, 1)
		})
	}
}

func (s IntegrationTestSuite) TestCmdLockedByID() {
	val := s.network.Validators[0]

	testCases := []struct {
		name string
		args []string
	}{
		{
			"get lock by id",
			[]string{
				fmt.Sprintf("%d", 1),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdLockedByID()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.LockedResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(result.Lock.ID, uint64(1))
		})
	}
}

func (s IntegrationTestSuite) TestCmdAccountLockedLongerDuration() {
	val := s.network.Validators[0]

	testCases := []struct {
		name string
		args []string
	}{
		{
			"get account locked longer than duration",
			[]string{
				val.Address.String(),
				"1s",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdAccountLockedLongerDuration()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.AccountLockedLongerDurationResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Len(result.Locks, 1)
		})
	}
}

func (s IntegrationTestSuite) TestCmdAccountLockedLongerDurationNotUnlockingOnly() {
	val := s.network.Validators[0]

	testCases := []struct {
		name string
		args []string
	}{
		{
			"get account locked longer than duration not unlocking only",
			[]string{
				val.Address.String(),
				"1s",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdAccountLockedLongerDurationNotUnlockingOnly()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.AccountLockedLongerDurationNotUnlockingOnlyResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Len(result.Locks, 0)
		})
	}
}

func (s IntegrationTestSuite) TestCmdAccountLockedLongerDurationDenom() {
	val := s.network.Validators[0]

	testCases := []struct {
		name string
		args []string
	}{
		{
			"get account locked longer than duration denom",
			[]string{
				val.Address.String(),
				"1s",
				s.cfg.BondDenom,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdAccountLockedLongerDurationDenom()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.AccountLockedLongerDurationDenomResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Len(result.Locks, 1)
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
