package v33_test

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/app/apptesting"
	v33 "github.com/osmosis-labs/osmosis/v32/app/upgrades/v33"
	incentivestypes "github.com/osmosis-labs/osmosis/v32/x/incentives/types"
)

// otherBTCDenom is the second asset of the test alloy transmuter.
const otherBTCDenom = "nbtc"

// prepareWBTCTransmuter creates an alloy transmuter holding 1e9 of WBTC and 1e9 of otherBTCDenom
// with normalization factors of 1, and returns its address and alloyed denom.
func (s *UpgradeTestSuite) prepareWBTCTransmuter() (string, string) {
	pool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{v33.WBTCDenom, otherBTCDenom}, []uint16{1, 1})
	contract := pool.GetContractAddress()
	return contract, "factory/" + contract + "/alloyed/" + apptesting.DefaultAlloyedSubDenom
}

// TestMigrateWBTCToAllBTC checks that user accounts are migrated 1:1, that only spendable vesting WBTC
// moves, and that module accounts and contracts are left alone.
func (s *UpgradeTestSuite) TestMigrateWBTCToAllBTC() {
	transmuter, alloyed := s.prepareWBTCTransmuter()
	transmuterAddr := sdk.MustAccAddressFromBech32(transmuter)
	reservesBefore := s.App.BankKeeper.GetBalance(s.Ctx, transmuterAddr, v33.WBTCDenom).Amount

	userA, userB := s.TestAccs[1], s.TestAccs[2]
	s.FundAcc(userA, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1_000), sdk.NewInt64Coin("uosmo", 7)))
	s.FundAcc(userB, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 25_000_000)))

	// A delayed vesting account with 400 of its 1,000 WBTC still locked.
	vesting := s.newVestingAccount(sdk.NewInt64Coin(v33.WBTCDenom, 1_000), sdk.NewInt64Coin(v33.WBTCDenom, 400))

	moduleWBTC := sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 5_000))
	s.FundModuleAcc(incentivestypes.ModuleName, moduleWBTC)
	moduleAddr := s.App.AccountKeeper.GetModuleAddress(incentivestypes.ModuleName)
	moduleBefore := s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, v33.WBTCDenom)

	result, err := v33.MigrateWBTCToAllBTC(s.Ctx, &s.App.AppKeepers, transmuter, v33.WBTCDenom, alloyed)
	s.Require().NoError(err)

	s.Require().Equal(3, result.Migrated)
	s.Require().Equal(0, result.Failed)
	// The module account and the transmuter itself.
	s.Require().Equal(2, result.Skipped)
	s.Require().Equal(osmomath.NewInt(1_000+25_000_000+600), result.Total)

	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, userA, v33.WBTCDenom).IsZero())
	s.Require().Equal(osmomath.NewInt(1_000), s.App.BankKeeper.GetBalance(s.Ctx, userA, alloyed).Amount)
	s.Require().Equal(osmomath.NewInt(7), s.App.BankKeeper.GetBalance(s.Ctx, userA, "uosmo").Amount)
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, userB, v33.WBTCDenom).IsZero())
	s.Require().Equal(osmomath.NewInt(25_000_000), s.App.BankKeeper.GetBalance(s.Ctx, userB, alloyed).Amount)

	s.Require().Equal(osmomath.NewInt(400), s.App.BankKeeper.GetBalance(s.Ctx, vesting, v33.WBTCDenom).Amount)
	s.Require().Equal(osmomath.NewInt(600), s.App.BankKeeper.GetBalance(s.Ctx, vesting, alloyed).Amount)

	s.Require().Equal(moduleBefore, s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, v33.WBTCDenom))
	s.Require().Equal(reservesBefore.Add(result.Total), s.App.BankKeeper.GetBalance(s.Ctx, transmuterAddr, v33.WBTCDenom).Amount)

	// The contract's own accounting includes the deposits, so holders can exit back to WBTC.
	liquidity, err := s.App.CosmwasmPoolKeeper.GetTotalPoolLiquidity(s.Ctx, s.transmuterPoolID(transmuter))
	s.Require().NoError(err)
	s.Require().Equal(reservesBefore.Add(result.Total), liquidity.AmountOf(v33.WBTCDenom))
}

// TestMigrateWBTCToAllBTCIsolatesFailures checks that a rejected deposit is rolled back without
// affecting the other accounts.
func (s *UpgradeTestSuite) TestMigrateWBTCToAllBTCIsolatesFailures() {
	transmuter, alloyed := s.prepareWBTCTransmuter()

	// WBTC may not exceed 60% of the pool: a 1e8 deposit fits, a 1e10 deposit does not.
	s.registerStaticLimiter(transmuter, v33.WBTCDenom, "0.6")
	small, large := s.TestAccs[1], s.TestAccs[2]
	s.FundAcc(small, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 100_000_000)))
	s.FundAcc(large, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 10_000_000_000)))

	result, err := v33.MigrateWBTCToAllBTC(s.Ctx, &s.App.AppKeepers, transmuter, v33.WBTCDenom, alloyed)
	s.Require().NoError(err)
	s.Require().Equal(1, result.Migrated)
	s.Require().Equal(1, result.Failed)

	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, small, v33.WBTCDenom).IsZero())
	s.Require().Equal(osmomath.NewInt(100_000_000), s.App.BankKeeper.GetBalance(s.Ctx, small, alloyed).Amount)
	s.Require().Equal(osmomath.NewInt(10_000_000_000), s.App.BankKeeper.GetBalance(s.Ctx, large, v33.WBTCDenom).Amount)
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, large, alloyed).IsZero())
}

// TestMigrateWBTCToAllBTCRejectsUnexpectedTransmuter checks that nothing moves when the transmuter is
// missing or does not accept WBTC.
func (s *UpgradeTestSuite) TestMigrateWBTCToAllBTCRejectsUnexpectedTransmuter() {
	nonWBTCPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{"ibc/other", otherBTCDenom}, []uint16{1, 1})
	tests := map[string]string{
		"transmuter does not accept WBTC": nonWBTCPool.GetContractAddress(),
		"transmuter is not a contract":    apptesting.CreateRandomAccounts(1)[0].String(),
	}
	holder := s.TestAccs[1]
	s.FundAcc(holder, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1_000)))

	for name, transmuter := range tests {
		s.Run(name, func() {
			before := s.App.BankKeeper.GetAllBalances(s.Ctx, holder)
			_, err := v33.MigrateWBTCToAllBTC(s.Ctx, &s.App.AppKeepers, transmuter, v33.WBTCDenom, "factory/"+transmuter+"/alloyed/allBTC")
			s.Require().Error(err)
			s.Require().Equal(before, s.App.BankKeeper.GetAllBalances(s.Ctx, holder))
		})
	}
}

// TestCollectHoldersCap checks that the holder list is bounded.
func (s *UpgradeTestSuite) TestCollectHoldersCap() {
	accs := apptesting.CreateRandomAccounts(3)
	for _, acc := range accs {
		s.FundAcc(acc, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1)))
	}

	holders, err := v33.CollectHolders(s.Ctx, &s.App.AppKeepers, v33.WBTCDenom, 3)
	s.Require().NoError(err)
	s.Require().ElementsMatch(accs, holders)

	_, err = v33.CollectHolders(s.Ctx, &s.App.AppKeepers, v33.WBTCDenom, 2)
	s.Require().ErrorContains(err, "more than 2 holders")
}

// TestUpgradeWBTCMigrationGating checks that the upgrade only migrates on mainnet, and that a mainnet
// upgrade without the allBTC transmuter neither halts nor moves WBTC, while the HUAHUA recovery still runs.
func (s *UpgradeTestSuite) TestUpgradeWBTCMigrationGating() {
	for _, chainID := range []string{"osmosis-testnet", v33.MainnetChainID} {
		s.Run(chainID, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithChainID(chainID)
			s.prepareMainnetHuahuaGauges()
			holder := s.TestAccs[1]
			s.FundAcc(holder, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1_000)))

			s.runUpgrade()

			s.Require().Equal(osmomath.NewInt(1_000), s.App.BankKeeper.GetBalance(s.Ctx, holder, v33.WBTCDenom).Amount)
			recipient := sdk.MustAccAddressFromBech32(v33.HuahuaRecoveryAddress)
			s.Require().Equal(expectedRecoveredHuahua, s.App.BankKeeper.GetBalance(s.Ctx, recipient, v33.HuahuaDenom).Amount)
		})
	}
}

// newVestingAccount creates a funded delayed vesting account whose locked coins are still vesting.
func (s *UpgradeTestSuite) newVestingAccount(funds, locked sdk.Coin) sdk.AccAddress {
	addr := apptesting.CreateRandomAccounts(1)[0]
	base := authtypes.NewBaseAccountWithAddress(addr)
	acc, err := vestingtypes.NewDelayedVestingAccount(base, sdk.NewCoins(locked), s.Ctx.BlockTime().Add(365*24*3600*1e9).Unix())
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, s.App.AccountKeeper.NewAccount(s.Ctx, acc))
	s.FundAcc(addr, sdk.NewCoins(funds))
	s.Require().Equal(funds.Sub(locked), s.App.BankKeeper.SpendableCoin(s.Ctx, addr, funds.Denom))
	return addr
}

// registerStaticLimiter registers a static limiter on the transmuter, executed by its admin.
func (s *UpgradeTestSuite) registerStaticLimiter(transmuter, denom, upperLimit string) {
	msg := map[string]any{"register_limiter": map[string]any{
		"denom":          denom,
		"label":          "static",
		"limiter_params": map[string]any{"static_limiter": map[string]any{"upper_limit": upperLimit}},
	}}
	bz, err := json.Marshal(msg)
	s.Require().NoError(err)
	_, err = s.App.ContractKeeper.Execute(s.Ctx, sdk.MustAccAddressFromBech32(transmuter), s.TestAccs[0], bz, nil)
	s.Require().NoError(err)
}

// transmuterPoolID returns the cosmwasm pool ID of the given contract.
func (s *UpgradeTestSuite) transmuterPoolID(contract string) uint64 {
	pools, err := s.App.CosmwasmPoolKeeper.GetPoolsWithWasmKeeper(s.Ctx)
	s.Require().NoError(err)
	for _, pool := range pools {
		if pool.GetAddress().String() == contract {
			return pool.GetId()
		}
	}
	s.FailNow("pool not found", contract)
	return 0
}
