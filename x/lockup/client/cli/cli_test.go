package cli

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/osmoutils"
	"github.com/osmosis-labs/osmosis/v13/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/lockup/types"
)

var testAddresses = osmoutils.CreateRandomAccounts(3)

func TestLockTokensCmd(t *testing.T) {
	desc, _ := NewLockTokensCmd()
	tcs := map[string]osmocli.TxCliTestCase[*types.MsgLockTokens]{
		"lock 201stake tokens for 1 day": {
			Cmd: "201uosmo --duration=24h --from=" + testAddresses[0].String(),
			ExpectedMsg: &types.MsgLockTokens{
				Owner:    testAddresses[0].String(),
				Duration: time.Hour * 24,
				Coins:    sdk.NewCoins(sdk.NewInt64Coin("uosmo", 201)),
			},
		},
	}
	osmocli.RunTxTestCases(t, desc, tcs)
}

func TestBeginUnlockingAllCmd(t *testing.T) {
	desc, _ := NewBeginUnlockingAllCmd()
	tcs := map[string]osmocli.TxCliTestCase[*types.MsgBeginUnlockingAll]{
		"basic test": {
			Cmd: "--from=" + testAddresses[0].String(),
			ExpectedMsg: &types.MsgBeginUnlockingAll{
				Owner: testAddresses[0].String(),
			},
		},
	}
	osmocli.RunTxTestCases(t, desc, tcs)
}

func TestBeginUnlockingByIDCmd(t *testing.T) {
	desc, _ := NewBeginUnlockByIDCmd()
	tcs := map[string]osmocli.TxCliTestCase[*types.MsgBeginUnlocking]{
		"basic test no coins": {
			Cmd: "10 --from=" + testAddresses[0].String(),
			ExpectedMsg: &types.MsgBeginUnlocking{
				Owner: testAddresses[0].String(),
				ID:    10,
				Coins: sdk.Coins(nil),
			},
		},
		"basic test w/ coins": {
			Cmd: "10 --amount=5uosmo --from=" + testAddresses[0].String(),
			ExpectedMsg: &types.MsgBeginUnlocking{
				Owner: testAddresses[0].String(),
				ID:    10,
				Coins: sdk.NewCoins(sdk.NewInt64Coin("uosmo", 5)),
			},
		},
	}
	osmocli.RunTxTestCases(t, desc, tcs)
}

// func (s *IntegrationTestSuite) TestCmdAccountUnlockingCoins() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name  string
// 		args  []string
// 		coins sdk.Coins
// 	}{
// 		{
// 			"query validator account unlocking coins",
// 			[]string{
// 				val.Address.String(),
// 				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
// 			},
// 			sdk.Coins{sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(200))},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdAccountUnlockingCoins()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			s.Require().NoError(err)

// 			var result types.AccountUnlockingCoinsResponse
// 			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
// 			s.Require().Equal(tc.coins.String(), result.Coins.String())
// 		})
// 	}
// }

// func (s IntegrationTestSuite) TestCmdAccountLockedCoins() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name  string
// 		args  []string
// 		coins sdk.Coins
// 	}{
// 		{
// 			"query account locked coins",
// 			[]string{
// 				val.Address.String(),
// 				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
// 			},
// 			sdk.Coins{sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(200))},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdAccountLockedCoins()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			s.Require().NoError(err)

// 			var result types.ModuleLockedAmountResponse
// 			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
// 			s.Require().Equal(tc.coins.String(), result.Coins.String())
// 		})
// 	}
// }

// func (s IntegrationTestSuite) TestCmdAccountLockedPastTime() {
// 	val := s.network.Validators[0]

// 	timestamp := time.Now().Unix()
// 	testCases := []struct {
// 		name string
// 		args []string
// 	}{
// 		{
// 			"query account locked coins past time",
// 			[]string{
// 				val.Address.String(),
// 				fmt.Sprintf("%d", timestamp),
// 				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdAccountLockedPastTime()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			s.Require().NoError(err)

// 			var result types.AccountLockedPastTimeResponse
// 			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
// 			s.Require().Len(result.Locks, 1)
// 		})
// 	}
// }

// func (s IntegrationTestSuite) TestCmdAccountLockedPastTimeNotUnlockingOnly() {
// 	val := s.network.Validators[0]

// 	timestamp := time.Now().Unix()
// 	testCases := []struct {
// 		name string
// 		args []string
// 	}{
// 		{
// 			"query account locked coins past time not unlocking only",
// 			[]string{
// 				val.Address.String(),
// 				fmt.Sprintf("%d", timestamp),
// 				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdAccountLockedPastTimeNotUnlockingOnly()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			s.Require().NoError(err)

// 			var result types.AccountLockedPastTimeNotUnlockingOnlyResponse
// 			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
// 			s.Require().Len(result.Locks, 0)
// 		})
// 	}
// }

// func (s IntegrationTestSuite) TestCmdAccountUnlockedBeforeTime() {
// 	val := s.network.Validators[0]

// 	timestamp := time.Now().Unix()
// 	testCases := []struct {
// 		name string
// 		args []string
// 	}{
// 		{
// 			"query account locked coins before time",
// 			[]string{
// 				val.Address.String(),
// 				fmt.Sprintf("%d", timestamp),
// 				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdAccountUnlockedBeforeTime()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			s.Require().NoError(err)

// 			var result types.AccountUnlockedBeforeTimeResponse
// 			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
// 			s.Require().Len(result.Locks, 0)
// 		})
// 	}
// }

// func (s IntegrationTestSuite) TestCmdAccountLockedPastTimeDenom() {
// 	val := s.network.Validators[0]

// 	timestamp := time.Now().Unix()
// 	testCases := []struct {
// 		name string
// 		args []string
// 	}{
// 		{
// 			"query account locked coins past time denom",
// 			[]string{
// 				val.Address.String(),
// 				fmt.Sprintf("%d", timestamp),
// 				s.cfg.BondDenom,
// 				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdAccountLockedPastTimeDenom()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			s.Require().NoError(err)

// 			var result types.AccountLockedPastTimeDenomResponse
// 			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
// 			s.Require().Len(result.Locks, 1)
// 		})
// 	}
// }

// func (s IntegrationTestSuite) TestCmdLockedByID() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name string
// 		args []string
// 	}{
// 		{
// 			"get lock by id",
// 			[]string{
// 				fmt.Sprintf("%d", 1),
// 				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdLockedByID()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			s.Require().NoError(err)

// 			var result types.LockedResponse
// 			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
// 			s.Require().Equal(result.Lock.ID, uint64(1))
// 		})
// 	}
// }

// func (s IntegrationTestSuite) TestCmdAccountLockedLongerDuration() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name string
// 		args []string
// 	}{
// 		{
// 			"get account locked longer than duration",
// 			[]string{
// 				val.Address.String(),
// 				"1s",
// 				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdAccountLockedLongerDuration()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			s.Require().NoError(err)

// 			var result types.AccountLockedLongerDurationResponse
// 			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
// 			s.Require().Len(result.Locks, 1)
// 		})
// 	}
// }

// func (s IntegrationTestSuite) TestCmdAccountLockedLongerDurationNotUnlockingOnly() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name string
// 		args []string
// 	}{
// 		{
// 			"get account locked longer than duration not unlocking only",
// 			[]string{
// 				val.Address.String(),
// 				"1s",
// 				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdAccountLockedLongerDurationNotUnlockingOnly()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			s.Require().NoError(err)

// 			var result types.AccountLockedLongerDurationNotUnlockingOnlyResponse
// 			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
// 			s.Require().Len(result.Locks, 0)
// 		})
// 	}
// }

// func (s IntegrationTestSuite) TestCmdAccountLockedLongerDurationDenom() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name string
// 		args []string
// 	}{
// 		{
// 			"get account locked longer than duration denom",
// 			[]string{
// 				val.Address.String(),
// 				"1s",
// 				s.cfg.BondDenom,
// 				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdAccountLockedLongerDurationDenom()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			s.Require().NoError(err)

// 			var result types.AccountLockedLongerDurationDenomResponse
// 			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
// 			s.Require().Len(result.Locks, 1)
// 		})
// 	}
// }

// func TestIntegrationTestSuite(t *testing.T) {
// 	suite.Run(t, new(IntegrationTestSuite))
// }
