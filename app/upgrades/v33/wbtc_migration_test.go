package v33_test

import (
	"encoding/json"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

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

// migrateWBTC runs the migration against transmuter with the given exclusions and unindexed holders.
func (s *UpgradeTestSuite) migrateWBTC(transmuter, alloyed string, excluded, unindexed []sdk.AccAddress) v33.WBTCMigrationResult {
	result, err := v33.MigrateWBTCToAllBTC(s.Ctx, &s.App.AppKeepers, v33.WBTCMigration{
		Transmuter: transmuter, WBTCDenom: v33.WBTCDenom, AlloyedDenom: alloyed,
		Excluded: excluded, UnindexedHolders: unindexed,
	})
	s.Require().NoError(err)
	return result
}

// TestMigrateWBTCToAllBTC checks that user accounts are migrated 1:1, that only spendable vesting WBTC
// moves, and that module accounts, accounts without a public key and excluded accounts are left alone.
func (s *UpgradeTestSuite) TestMigrateWBTCToAllBTC() {
	transmuter, alloyed := s.prepareWBTCTransmuter()
	transmuterAddr := sdk.MustAccAddressFromBech32(transmuter)
	reservesBefore := s.App.BankKeeper.GetBalance(s.Ctx, transmuterAddr, v33.WBTCDenom).Amount

	userA, userB := s.newUser(), s.newUser()
	s.FundAcc(userA, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1_000), sdk.NewInt64Coin("uosmo", 7)))
	s.FundAcc(userB, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 25_000_000)))

	// A delayed vesting account with 400 of its 1,000 WBTC still locked.
	vesting := s.newVestingAccount(sdk.NewInt64Coin(v33.WBTCDenom, 1_000), sdk.NewInt64Coin(v33.WBTCDenom, 400))

	moduleWBTC := sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 5_000))
	s.FundModuleAcc(incentivestypes.ModuleName, moduleWBTC)
	moduleAddr := s.App.AccountKeeper.GetModuleAddress(incentivestypes.ModuleName)
	moduleBefore := s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, v33.WBTCDenom)

	// An account that has never signed, and a merchant of the issuer.
	noPubKey, merchant := apptesting.CreateRandomAccounts(1)[0], s.newUser()
	s.FundAcc(noPubKey, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 300)))
	s.FundAcc(merchant, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 900)))

	result := s.migrateWBTC(transmuter, alloyed, []sdk.AccAddress{merchant}, nil)

	s.Require().Equal(3, result.Migrated)
	s.Require().Equal(0, result.Failed)
	// The module account, the transmuter itself, the account without a public key and the merchant.
	s.Require().Equal(4, result.Skipped)
	s.Require().Equal(osmomath.NewInt(1_000+25_000_000+600), result.Total)

	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, userA, v33.WBTCDenom).IsZero())
	s.Require().Equal(osmomath.NewInt(1_000), s.App.BankKeeper.GetBalance(s.Ctx, userA, alloyed).Amount)
	s.Require().Equal(osmomath.NewInt(7), s.App.BankKeeper.GetBalance(s.Ctx, userA, "uosmo").Amount)
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, userB, v33.WBTCDenom).IsZero())
	s.Require().Equal(osmomath.NewInt(25_000_000), s.App.BankKeeper.GetBalance(s.Ctx, userB, alloyed).Amount)

	s.Require().Equal(osmomath.NewInt(400), s.App.BankKeeper.GetBalance(s.Ctx, vesting, v33.WBTCDenom).Amount)
	s.Require().Equal(osmomath.NewInt(600), s.App.BankKeeper.GetBalance(s.Ctx, vesting, alloyed).Amount)

	s.Require().Equal(moduleBefore, s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, v33.WBTCDenom))
	s.Require().Equal(osmomath.NewInt(300), s.App.BankKeeper.GetBalance(s.Ctx, noPubKey, v33.WBTCDenom).Amount)
	s.Require().Equal(osmomath.NewInt(900), s.App.BankKeeper.GetBalance(s.Ctx, merchant, v33.WBTCDenom).Amount)
	s.Require().Equal(reservesBefore.Add(result.Total), s.App.BankKeeper.GetBalance(s.Ctx, transmuterAddr, v33.WBTCDenom).Amount)

	// The contract's own accounting includes the deposits, so holders can exit back to WBTC.
	liquidity, err := s.App.CosmwasmPoolKeeper.GetTotalPoolLiquidity(s.Ctx, s.transmuterPoolID(transmuter))
	s.Require().NoError(err)
	s.Require().Equal(reservesBefore.Add(result.Total), liquidity.AmountOf(v33.WBTCDenom))
}

// TestMigrateWBTCToAllBTCUnindexedHolders checks that holders missing from the bank denom index are
// migrated when listed, and only once when they are also in the index.
func (s *UpgradeTestSuite) TestMigrateWBTCToAllBTCUnindexedHolders() {
	transmuter, alloyed := s.prepareWBTCTransmuter()
	indexed, unindexed, emptied := s.newUser(), s.newUser(), s.newUser()
	s.FundAcc(indexed, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1_000)))
	s.FundAcc(unindexed, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 300)))
	s.unindex(unindexed)

	// The index alone misses the unindexed holder.
	holders, err := v33.CollectHolders(s.Ctx, &s.App.AppKeepers, v33.WBTCDenom, nil, v33.MaxWBTCHoldersToMigrate)
	s.Require().NoError(err)
	s.Require().NotContains(holders, unindexed)

	// indexed is also in the index, and emptied has no WBTC.
	result := s.migrateWBTC(transmuter, alloyed, nil, []sdk.AccAddress{unindexed, indexed, emptied})

	s.Require().Equal(2, result.Migrated)
	s.Require().Equal(osmomath.NewInt(1_300), result.Total)
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, unindexed, v33.WBTCDenom).IsZero())
	s.Require().Equal(osmomath.NewInt(300), s.App.BankKeeper.GetBalance(s.Ctx, unindexed, alloyed).Amount)
	s.Require().Equal(osmomath.NewInt(1_000), s.App.BankKeeper.GetBalance(s.Ctx, indexed, alloyed).Amount)
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, emptied, alloyed).IsZero())
}

// TestMigrateWBTCToAllBTCKeepsIBCEscrow checks that WBTC escrowed for a counterparty chain stays in
// the escrow, so tokens returning from that chain can still be unescrowed.
func (s *UpgradeTestSuite) TestMigrateWBTCToAllBTCKeepsIBCEscrow() {
	transmuter, alloyed := s.prepareWBTCTransmuter()

	escrow := transfertypes.GetEscrowAddress(transfertypes.PortID, "channel-0")
	escrowed := sdk.NewInt64Coin(v33.WBTCDenom, 8_692_065)
	s.FundAcc(escrow, sdk.NewCoins(escrowed))
	s.App.TransferKeeper.SetTotalEscrowForDenom(s.Ctx, escrowed)

	result := s.migrateWBTC(transmuter, alloyed, nil, nil)
	s.Require().Zero(result.Migrated)
	s.Require().Equal(escrowed, s.App.BankKeeper.GetBalance(s.Ctx, escrow, v33.WBTCDenom))

	receiver := s.TestAccs[1]
	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         transfertypes.PortID,
		SourceChannel:      "channel-141",
		DestinationPort:    transfertypes.PortID,
		DestinationChannel: "channel-0",
	}
	data := transfertypes.NewFungibleTokenPacketData("transfer/channel-141/"+v33.WBTCDenom, escrowed.Amount.String(), "cosmos1sender", receiver.String(), "")
	s.Require().NoError(s.App.TransferKeeper.OnRecvPacket(s.Ctx, packet, data))
	s.Require().Equal(escrowed, s.App.BankKeeper.GetBalance(s.Ctx, receiver, v33.WBTCDenom))
}

// TestMigrateWBTCToAllBTCKeepsCLSpreadRewards checks that WBTC spread rewards accrued by a CL pool
// stay at its spread rewards address, so its LPs can still fully withdraw.
func (s *UpgradeTestSuite) TestMigrateWBTCToAllBTCKeepsCLSpreadRewards() {
	transmuter, alloyed := s.prepareWBTCTransmuter()

	pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], v33.WBTCDenom, "uosmo", 1, osmomath.MustNewDecFromStr("0.003"))
	positionCoins := sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1_000_000_000), sdk.NewInt64Coin("uosmo", 1_000_000_000))
	positionID, liquidity := s.CreateFullRangePosition(pool, positionCoins)

	// A swap of WBTC in accrues WBTC spread rewards at the pool's spread rewards address.
	swapper := s.TestAccs[2]
	tokenIn := sdk.NewInt64Coin(v33.WBTCDenom, 10_000_000)
	s.FundAcc(swapper, sdk.NewCoins(tokenIn))
	_, err := s.App.ConcentratedLiquidityKeeper.SwapExactAmountIn(s.Ctx, swapper, pool, tokenIn, "uosmo", osmomath.OneInt(), pool.GetSpreadFactor(s.Ctx))
	s.Require().NoError(err)
	spreadRewards := s.App.BankKeeper.GetBalance(s.Ctx, pool.GetSpreadRewardsAddress(), v33.WBTCDenom)
	s.Require().True(spreadRewards.IsPositive())

	s.migrateWBTC(transmuter, alloyed, nil, nil)
	s.Require().Equal(spreadRewards, s.App.BankKeeper.GetBalance(s.Ctx, pool.GetSpreadRewardsAddress(), v33.WBTCDenom))

	_, _, err = s.App.ConcentratedLiquidityKeeper.WithdrawPosition(s.Ctx, s.TestAccs[0], positionID, liquidity)
	s.Require().NoError(err)
}

// TestMigrateWBTCToAllBTCIsolatesFailures checks that a rejected deposit is rolled back without
// affecting the other accounts.
func (s *UpgradeTestSuite) TestMigrateWBTCToAllBTCIsolatesFailures() {
	transmuter, alloyed := s.prepareWBTCTransmuter()

	// WBTC may not exceed 60% of the pool: a 1e8 deposit fits, a 1e10 deposit does not.
	s.registerStaticLimiter(transmuter, v33.WBTCDenom, "0.6")
	small, large := s.newUser(), s.newUser()
	s.FundAcc(small, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 100_000_000)))
	s.FundAcc(large, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 10_000_000_000)))

	result := s.migrateWBTC(transmuter, alloyed, nil, nil)
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
	holder := s.newUser()
	s.FundAcc(holder, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1_000)))

	for name, transmuter := range tests {
		s.Run(name, func() {
			before := s.App.BankKeeper.GetAllBalances(s.Ctx, holder)
			_, err := v33.MigrateWBTCToAllBTC(s.Ctx, &s.App.AppKeepers, v33.WBTCMigration{
				Transmuter: transmuter, WBTCDenom: v33.WBTCDenom, AlloyedDenom: "factory/" + transmuter + "/alloyed/allBTC",
			})
			s.Require().Error(err)
			s.Require().Equal(before, s.App.BankKeeper.GetAllBalances(s.Ctx, holder))
		})
	}
}

// TestCollectHoldersCap checks that the holder list is bounded and counts listed unindexed holders.
func (s *UpgradeTestSuite) TestCollectHoldersCap() {
	accs := apptesting.CreateRandomAccounts(3)
	for _, acc := range accs {
		s.FundAcc(acc, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1)))
	}
	s.unindex(accs[2])

	holders, err := v33.CollectHolders(s.Ctx, &s.App.AppKeepers, v33.WBTCDenom, accs[2:], 3)
	s.Require().NoError(err)
	s.Require().ElementsMatch(accs, holders)

	_, err = v33.CollectHolders(s.Ctx, &s.App.AppKeepers, v33.WBTCDenom, accs[2:], 2)
	s.Require().ErrorContains(err, "more than 2 holders")
}

// TestWBTCUnindexedHoldersList checks that the embedded list parses and fits under the holder cap.
func (s *UpgradeTestSuite) TestWBTCUnindexedHoldersList() {
	addrs, err := v33.ParseAddresses(v33.WBTCUnindexedHolders, v33.MaxWBTCHoldersToMigrate)
	s.Require().NoError(err)
	s.Require().Len(addrs, 483)

	seen := map[string]bool{}
	for _, addr := range addrs {
		s.Require().False(seen[addr.String()], "duplicate %s", addr)
		seen[addr.String()] = true
	}
}

// TestUpgradeWBTCMigrationGating checks that the upgrade only migrates on mainnet, and that a mainnet
// upgrade without the WBTC issuer or allBTC transmuter neither halts nor moves WBTC, while the HUAHUA
// recovery still runs.
func (s *UpgradeTestSuite) TestUpgradeWBTCMigrationGating() {
	for _, chainID := range []string{"osmosis-testnet", v33.MainnetChainID} {
		s.Run(chainID, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithChainID(chainID)
			s.prepareMainnetHuahuaGauges()
			holder := s.newUser()
			s.FundAcc(holder, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1_000)))

			s.runUpgrade()

			s.Require().Equal(osmomath.NewInt(1_000), s.App.BankKeeper.GetBalance(s.Ctx, holder, v33.WBTCDenom).Amount)
			recipient := sdk.MustAccAddressFromBech32(v33.HuahuaRecoveryAddress)
			s.Require().Equal(expectedRecoveredHuahua, s.App.BankKeeper.GetBalance(s.Ctx, recipient, v33.HuahuaDenom).Amount)
		})
	}
}

// newUser creates an account that has signed a transaction, so it has a public key.
func (s *UpgradeTestSuite) newUser() sdk.AccAddress {
	pub := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pub.Address())
	acc := s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, addr)
	s.Require().NoError(acc.SetPubKey(pub))
	s.App.AccountKeeper.SetAccount(s.Ctx, acc)
	return addr
}

// unindex removes the WBTC balance of addr from the bank denom index, as on mainnet for balances
// unchanged since the index was added back.
func (s *UpgradeTestSuite) unindex(addr sdk.AccAddress) {
	s.Require().NoError(s.App.BankKeeper.Balances.Indexes.Denom.Unreference(s.Ctx, collections.Join(addr, v33.WBTCDenom), nil))
}

// newVestingAccount creates a funded delayed vesting account with a public key whose locked coins are
// still vesting.
func (s *UpgradeTestSuite) newVestingAccount(funds, locked sdk.Coin) sdk.AccAddress {
	pub := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pub.Address())
	base := authtypes.NewBaseAccount(addr, pub, 0, 0)
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
