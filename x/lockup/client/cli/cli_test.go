package cli

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"
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
				Coins:    sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseCoinUnit, 201)),
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
				Coins: sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseCoinUnit, 5)),
			},
		},
	}
	osmocli.RunTxTestCases(t, desc, tcs)
}

func TestModuleBalanceCmd(t *testing.T) {
	desc, _ := GetCmdModuleBalance()
	tcs := map[string]osmocli.QueryCliTestCase[*types.ModuleBalanceRequest]{
		"basic test": {
			Cmd:           "",
			ExpectedQuery: &types.ModuleBalanceRequest{},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestAccountUnlockingCoinsCmd(t *testing.T) {
	desc, _ := GetCmdAccountUnlockingCoins()
	tcs := map[string]osmocli.QueryCliTestCase[*types.AccountUnlockingCoinsRequest]{
		"basic test": {
			Cmd: testAddresses[0].String(),
			ExpectedQuery: &types.AccountUnlockingCoinsRequest{
				Owner: testAddresses[0].String(),
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestCmdAccountLockedPastTime(t *testing.T) {
	desc, _ := GetCmdAccountLockedPastTime()
	tcs := map[string]osmocli.QueryCliTestCase[*types.AccountLockedPastTimeRequest]{
		"basic test": {
			Cmd: testAddresses[0].String() + " 1670431012",
			ExpectedQuery: &types.AccountLockedPastTimeRequest{
				Owner:     testAddresses[0].String(),
				Timestamp: time.Unix(1670431012, 0),
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestCmdAccountLockedPastTimeNotUnlockingOnly(t *testing.T) {
	desc, _ := GetCmdAccountLockedPastTimeNotUnlockingOnly()
	tcs := map[string]osmocli.QueryCliTestCase[*types.AccountLockedPastTimeNotUnlockingOnlyRequest]{
		"basic test": {
			Cmd: testAddresses[0].String() + " 1670431012",
			ExpectedQuery: &types.AccountLockedPastTimeNotUnlockingOnlyRequest{
				Owner:     testAddresses[0].String(),
				Timestamp: time.Unix(1670431012, 0),
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestCmdTotalLockedByDenom(t *testing.T) {
	desc, _ := GetCmdTotalLockedByDenom()
	tcs := map[string]osmocli.QueryCliTestCase[*types.LockedDenomRequest]{
		"basic test": {
			Cmd: "uosmo --min-duration=1s",
			ExpectedQuery: &types.LockedDenomRequest{
				Denom:    appparams.BaseCoinUnit,
				Duration: time.Second,
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}
