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
	"github.com/osmosis-labs/osmosis/v32/app/keepers"
	v33 "github.com/osmosis-labs/osmosis/v32/app/upgrades/v33"
	incentivestypes "github.com/osmosis-labs/osmosis/v32/x/incentives/types"
)

// otherAlloyDenom is the second asset of the test alloy transmuter.
const otherAlloyDenom = "nbtc"

// alloyDenoms are the tokenfactory (WBTC-like) and IBC (Noble USDC-like) denoms the migration runs for.
func alloyDenoms() []string {
	return []string{v33.WBTCDenom, v33.NobleUSDCDenom}
}

// prepareAlloyTransmuter creates an alloy transmuter holding 1e9 of denom and 1e9 of otherAlloyDenom
// with normalization factors of 1, and returns its address and alloyed denom.
func (s *UpgradeTestSuite) prepareAlloyTransmuter(denom string) (string, string) {
	pool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{denom, otherAlloyDenom}, []uint16{1, 1})
	contract := pool.GetContractAddress()
	return contract, "factory/" + contract + "/alloyed/" + apptesting.DefaultAlloyedSubDenom
}

// migrateAlloy runs the migration against transmuter with the given exclusions and unindexed holders.
func (s *UpgradeTestSuite) migrateAlloy(denom, transmuter, alloyed string, excluded, unindexed []sdk.AccAddress) v33.AlloyMigrationResult {
	result, err := v33.MigrateToAlloy(s.Ctx, &s.App.AppKeepers, v33.AlloyMigration{
		Name: denom, Transmuter: transmuter, Denom: denom, AlloyedDenom: alloyed,
		Excluded: excluded, UnindexedHolders: unindexed, MaxHolders: 100,
	})
	s.Require().NoError(err)
	return result
}

// TestMigrateToAlloy checks that user accounts are migrated 1:1, that only spendable vesting coins
// move, and that module accounts, contracts, accounts without a public key and excluded accounts
// are left alone. It runs for a tokenfactory denom and an IBC denom.
func (s *UpgradeTestSuite) TestMigrateToAlloy() {
	for _, denom := range alloyDenoms() {
		s.Run(denom, func() {
			s.SetupTest()
			transmuter, alloyed := s.prepareAlloyTransmuter(denom)
			transmuterAddr := sdk.MustAccAddressFromBech32(transmuter)
			reservesBefore := s.App.BankKeeper.GetBalance(s.Ctx, transmuterAddr, denom).Amount

			// A second contract that merely holds the denom. Its balance is not deposited.
			otherPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{"ibc/other", otherAlloyDenom}, []uint16{1, 1})
			otherContract := sdk.MustAccAddressFromBech32(otherPool.GetContractAddress())
			s.FundAcc(otherContract, sdk.NewCoins(sdk.NewInt64Coin(denom, 42)))

			userA, userB := s.newUser(), s.newUser()
			s.FundAcc(userA, sdk.NewCoins(sdk.NewInt64Coin(denom, 1_000), sdk.NewInt64Coin("uosmo", 7)))
			s.FundAcc(userB, sdk.NewCoins(sdk.NewInt64Coin(denom, 25_000_000)))

			// A delayed vesting account with 400 of its 1,000 still locked.
			vesting := s.newVestingAccount(sdk.NewInt64Coin(denom, 1_000), sdk.NewInt64Coin(denom, 400))

			moduleCoins := sdk.NewCoins(sdk.NewInt64Coin(denom, 5_000))
			s.FundModuleAcc(incentivestypes.ModuleName, moduleCoins)
			moduleAddr := s.App.AccountKeeper.GetModuleAddress(incentivestypes.ModuleName)
			moduleBefore := s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, denom)

			// An account that has never signed, and an explicitly excluded account.
			noPubKey, excluded := apptesting.CreateRandomAccounts(1)[0], s.newUser()
			s.FundAcc(noPubKey, sdk.NewCoins(sdk.NewInt64Coin(denom, 300)))
			s.FundAcc(excluded, sdk.NewCoins(sdk.NewInt64Coin(denom, 900)))

			result := s.migrateAlloy(denom, transmuter, alloyed, []sdk.AccAddress{excluded}, nil)

			s.Require().Equal(3, result.Migrated)
			s.Require().Equal(0, result.Failed)
			// The module account, the transmuter, the other contract, the account without a public
			// key and the excluded account.
			s.Require().Equal(5, result.Skipped)
			s.Require().Equal(osmomath.NewInt(1_000+25_000_000+600), result.Total)

			s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, userA, denom).IsZero())
			s.Require().Equal(osmomath.NewInt(1_000), s.App.BankKeeper.GetBalance(s.Ctx, userA, alloyed).Amount)
			s.Require().Equal(osmomath.NewInt(7), s.App.BankKeeper.GetBalance(s.Ctx, userA, "uosmo").Amount)
			s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, userB, denom).IsZero())
			s.Require().Equal(osmomath.NewInt(25_000_000), s.App.BankKeeper.GetBalance(s.Ctx, userB, alloyed).Amount)

			s.Require().Equal(osmomath.NewInt(400), s.App.BankKeeper.GetBalance(s.Ctx, vesting, denom).Amount)
			s.Require().Equal(osmomath.NewInt(600), s.App.BankKeeper.GetBalance(s.Ctx, vesting, alloyed).Amount)

			s.Require().Equal(moduleBefore, s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, denom))
			s.Require().Equal(osmomath.NewInt(42), s.App.BankKeeper.GetBalance(s.Ctx, otherContract, denom).Amount)
			s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, otherContract, alloyed).IsZero())
			s.Require().Equal(osmomath.NewInt(300), s.App.BankKeeper.GetBalance(s.Ctx, noPubKey, denom).Amount)
			s.Require().Equal(osmomath.NewInt(900), s.App.BankKeeper.GetBalance(s.Ctx, excluded, denom).Amount)
			s.Require().Equal(reservesBefore.Add(result.Total), s.App.BankKeeper.GetBalance(s.Ctx, transmuterAddr, denom).Amount)

			// The contract's own accounting includes the deposits, so holders can exit back to the denom.
			liquidity, err := s.App.CosmwasmPoolKeeper.GetTotalPoolLiquidity(s.Ctx, s.transmuterPoolID(transmuter))
			s.Require().NoError(err)
			s.Require().Equal(reservesBefore.Add(result.Total), liquidity.AmountOf(denom))
		})
	}
}

// TestMigrateToAlloyUnindexedHolders checks that holders missing from the bank denom index are
// migrated when listed, and only once when they are also in the index.
func (s *UpgradeTestSuite) TestMigrateToAlloyUnindexedHolders() {
	for _, denom := range alloyDenoms() {
		s.Run(denom, func() {
			s.SetupTest()
			transmuter, alloyed := s.prepareAlloyTransmuter(denom)
			indexed, unindexed, emptied := s.newUser(), s.newUser(), s.newUser()
			s.FundAcc(indexed, sdk.NewCoins(sdk.NewInt64Coin(denom, 1_000)))
			s.FundAcc(unindexed, sdk.NewCoins(sdk.NewInt64Coin(denom, 300)))
			s.unindex(unindexed, denom)

			// The index alone misses the unindexed holder.
			holders, err := v33.CollectHolders(s.Ctx, &s.App.AppKeepers, denom, nil, 100)
			s.Require().NoError(err)
			s.Require().NotContains(holders, unindexed)

			// indexed is also in the index, and emptied has none of the denom.
			result := s.migrateAlloy(denom, transmuter, alloyed, nil, []sdk.AccAddress{unindexed, indexed, emptied})

			s.Require().Equal(2, result.Migrated)
			s.Require().Equal(osmomath.NewInt(1_300), result.Total)
			s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, unindexed, denom).IsZero())
			s.Require().Equal(osmomath.NewInt(300), s.App.BankKeeper.GetBalance(s.Ctx, unindexed, alloyed).Amount)
			s.Require().Equal(osmomath.NewInt(1_000), s.App.BankKeeper.GetBalance(s.Ctx, indexed, alloyed).Amount)
			s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, emptied, alloyed).IsZero())
		})
	}
}

// TestMigrateToAlloyKeepsIBCEscrow checks that coins escrowed for a counterparty chain stay in the
// escrow, so tokens returning from that chain can still be unescrowed.
func (s *UpgradeTestSuite) TestMigrateToAlloyKeepsIBCEscrow() {
	for _, denom := range alloyDenoms() {
		s.Run(denom, func() {
			s.SetupTest()
			transmuter, alloyed := s.prepareAlloyTransmuter(denom)

			escrow := transfertypes.GetEscrowAddress(transfertypes.PortID, "channel-0")
			escrowed := sdk.NewInt64Coin(denom, 8_692_065)
			s.FundAcc(escrow, sdk.NewCoins(escrowed))
			s.App.TransferKeeper.SetTotalEscrowForDenom(s.Ctx, escrowed)

			result := s.migrateAlloy(denom, transmuter, alloyed, nil, nil)
			s.Require().Zero(result.Migrated)
			s.Require().Equal(escrowed, s.App.BankKeeper.GetBalance(s.Ctx, escrow, denom))

			receiver := s.TestAccs[1]
			packet := channeltypes.Packet{
				Sequence:           1,
				SourcePort:         transfertypes.PortID,
				SourceChannel:      "channel-141",
				DestinationPort:    transfertypes.PortID,
				DestinationChannel: "channel-0",
			}
			data := transfertypes.NewFungibleTokenPacketData("transfer/channel-141/"+denom, escrowed.Amount.String(), "cosmos1sender", receiver.String(), "")
			s.Require().NoError(s.App.TransferKeeper.OnRecvPacket(s.Ctx, packet, data))
			s.Require().Equal(escrowed, s.App.BankKeeper.GetBalance(s.Ctx, receiver, denom))
		})
	}
}

// TestMigrateToAlloyKeepsCLSpreadRewards checks that spread rewards accrued by a CL pool stay at
// its spread rewards address, so its LPs can still fully withdraw.
func (s *UpgradeTestSuite) TestMigrateToAlloyKeepsCLSpreadRewards() {
	for _, denom := range alloyDenoms() {
		s.Run(denom, func() {
			s.SetupTest()
			transmuter, alloyed := s.prepareAlloyTransmuter(denom)

			pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], denom, "uosmo", 1, osmomath.MustNewDecFromStr("0.003"))
			positionCoins := sdk.NewCoins(sdk.NewInt64Coin(denom, 1_000_000_000), sdk.NewInt64Coin("uosmo", 1_000_000_000))
			positionID, liquidity := s.CreateFullRangePosition(pool, positionCoins)

			// A swap of denom in accrues spread rewards at the pool's spread rewards address.
			swapper := s.TestAccs[2]
			tokenIn := sdk.NewInt64Coin(denom, 10_000_000)
			s.FundAcc(swapper, sdk.NewCoins(tokenIn))
			_, err := s.App.ConcentratedLiquidityKeeper.SwapExactAmountIn(s.Ctx, swapper, pool, tokenIn, "uosmo", osmomath.OneInt(), pool.GetSpreadFactor(s.Ctx))
			s.Require().NoError(err)
			spreadRewards := s.App.BankKeeper.GetBalance(s.Ctx, pool.GetSpreadRewardsAddress(), denom)
			s.Require().True(spreadRewards.IsPositive())

			s.migrateAlloy(denom, transmuter, alloyed, nil, nil)
			s.Require().Equal(spreadRewards, s.App.BankKeeper.GetBalance(s.Ctx, pool.GetSpreadRewardsAddress(), denom))

			_, _, err = s.App.ConcentratedLiquidityKeeper.WithdrawPosition(s.Ctx, s.TestAccs[0], positionID, liquidity)
			s.Require().NoError(err)
		})
	}
}

// TestMigrateToAlloyIsolatesFailures checks that a rejected deposit is rolled back without
// affecting the other accounts.
func (s *UpgradeTestSuite) TestMigrateToAlloyIsolatesFailures() {
	for _, denom := range alloyDenoms() {
		s.Run(denom, func() {
			s.SetupTest()
			transmuter, alloyed := s.prepareAlloyTransmuter(denom)

			// The denom may not exceed 60% of the pool: a 1e8 deposit fits, a 1e10 deposit does not.
			s.registerStaticLimiter(transmuter, denom, "0.6")
			small, large := s.newUser(), s.newUser()
			s.FundAcc(small, sdk.NewCoins(sdk.NewInt64Coin(denom, 100_000_000)))
			s.FundAcc(large, sdk.NewCoins(sdk.NewInt64Coin(denom, 10_000_000_000)))

			result := s.migrateAlloy(denom, transmuter, alloyed, nil, nil)
			s.Require().Equal(1, result.Migrated)
			s.Require().Equal(1, result.Failed)

			s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, small, denom).IsZero())
			s.Require().Equal(osmomath.NewInt(100_000_000), s.App.BankKeeper.GetBalance(s.Ctx, small, alloyed).Amount)
			s.Require().Equal(osmomath.NewInt(10_000_000_000), s.App.BankKeeper.GetBalance(s.Ctx, large, denom).Amount)
			s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, large, alloyed).IsZero())
		})
	}
}

// TestMigrateToAlloyPreflight checks that a failed pre-flight changes nothing.
func (s *UpgradeTestSuite) TestMigrateToAlloyPreflight() {
	for _, denom := range alloyDenoms() {
		s.Run(denom, func() {
			s.SetupTest()
			holder := s.newUser()
			s.FundAcc(holder, sdk.NewCoins(sdk.NewInt64Coin(denom, 1_000)))

			s.Run("denom not listed", func() {
				pool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{"ibc/other", otherAlloyDenom}, []uint16{1, 1})
				s.expectPreflightSkip(holder, denom, pool.GetContractAddress(), "factory/"+pool.GetContractAddress()+"/alloyed/"+apptesting.DefaultAlloyedSubDenom, "does not list")
			})
			s.Run("normalization mismatch", func() {
				pool := s.PrepareCustomTransmuterPoolV3WithNormalization(s.TestAccs[0], []string{denom, otherAlloyDenom}, []string{"1000", "1"}, []uint16{1, 1})
				contract := pool.GetContractAddress()
				s.expectPreflightSkip(holder, denom, contract, "factory/"+contract+"/alloyed/"+apptesting.DefaultAlloyedSubDenom, "normalization factor")
			})
			s.Run("inactive pool", func() {
				transmuter, alloyed := s.prepareAlloyTransmuter(denom)
				s.execTransmuter(transmuter, map[string]any{"set_active_status": map[string]any{"active": false}})
				s.expectPreflightSkip(holder, denom, transmuter, alloyed, "not active")
			})
			s.Run("missing contract", func() {
				missing := apptesting.CreateRandomAccounts(1)[0].String()
				s.expectPreflightSkip(holder, denom, missing, "factory/"+missing+"/alloyed/allBTC", "not a contract")
			})
			s.Run("corrupted denom", func() {
				transmuter, alloyed := s.prepareAlloyTransmuter(denom)
				s.execTransmuter(transmuter, map[string]any{"mark_corrupted_assets": map[string]any{"denoms": []string{denom}}})
				s.expectPreflightSkip(holder, denom, transmuter, alloyed, "corrupted")
			})
			s.Run("alloyed denom not listed", func() {
				transmuter, _ := s.prepareAlloyTransmuter(denom)
				s.expectPreflightSkip(holder, denom, transmuter, "factory/"+transmuter+"/alloyed/not-the-share", "does not list")
			})
		})
	}
}

// TestMigrateToAlloyHolderCap checks that exceeding MaxHolders changes nothing.
func (s *UpgradeTestSuite) TestMigrateToAlloyHolderCap() {
	transmuter, alloyed := s.prepareAlloyTransmuter(v33.WBTCDenom)
	holder := s.newUser()
	s.FundAcc(holder, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1_000)))
	before := s.App.BankKeeper.GetAllBalances(s.Ctx, holder)

	_, err := v33.MigrateToAlloy(s.Ctx, &s.App.AppKeepers, v33.AlloyMigration{
		Name: "WBTC to allBTC", Transmuter: transmuter, Denom: v33.WBTCDenom, AlloyedDenom: alloyed, MaxHolders: 1,
	})
	s.Require().ErrorContains(err, "more than 1 holders")
	s.Require().Equal(before, s.App.BankKeeper.GetAllBalances(s.Ctx, holder))
}

// expectPreflightSkip runs the migration and checks that it fails without moving the holder's funds.
func (s *UpgradeTestSuite) expectPreflightSkip(holder sdk.AccAddress, denom, transmuter, alloyed, errContains string) {
	before := s.App.BankKeeper.GetAllBalances(s.Ctx, holder)
	_, err := v33.MigrateToAlloy(s.Ctx, &s.App.AppKeepers, v33.AlloyMigration{
		Name: denom, Transmuter: transmuter, Denom: denom, AlloyedDenom: alloyed, MaxHolders: 100,
	})
	s.Require().ErrorContains(err, errContains)
	s.Require().Equal(before, s.App.BankKeeper.GetAllBalances(s.Ctx, holder))
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, holder, denom).IsPositive())
}

// TestCollectHoldersCap checks that the holder list is bounded and counts listed unindexed holders.
func (s *UpgradeTestSuite) TestCollectHoldersCap() {
	accs := apptesting.CreateRandomAccounts(3)
	for _, acc := range accs {
		s.FundAcc(acc, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1)))
	}
	s.unindex(accs[2], v33.WBTCDenom)

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

// TestNobleUSDCUnindexedHoldersPlaceholder checks that the USDC list is the empty scan placeholder.
func (s *UpgradeTestSuite) TestNobleUSDCUnindexedHoldersPlaceholder() {
	s.Require().Contains(v33.NobleUSDCUnindexedHolders, "TODO(scan)")
	s.Require().NotContains(v33.NobleUSDCUnindexedHolders, "osmo1")
	addrs, err := v33.ParseAddresses(v33.NobleUSDCUnindexedHolders, v33.MaxNobleUSDCHoldersToMigrate)
	s.Require().NoError(err)
	s.Require().Empty(addrs)

	m, err := v33.MainnetNobleUSDCMigration(s.Ctx, &s.App.AppKeepers)
	s.Require().NoError(err)
	s.Require().Equal("Noble USDC to allUSDC", m.Name)
	s.Require().Equal(v33.NobleUSDCDenom, m.Denom)
	s.Require().Equal(v33.AllUSDCDenom, m.AlloyedDenom)
	s.Require().Equal(v33.AllUSDCTransmuterAddress, m.Transmuter)
	s.Require().Equal(v33.MaxNobleUSDCHoldersToMigrate, m.MaxHolders)
	s.Require().Empty(m.Excluded)
	s.Require().Empty(m.UnindexedHolders)
}

// TestUpgradeSkipsAlloyMigrationsOffMainnet checks that a non-mainnet upgrade does not run either
// migration, and that HUAHUA recovery still runs.
func (s *UpgradeTestSuite) TestUpgradeSkipsAlloyMigrationsOffMainnet() {
	s.Ctx = s.Ctx.WithChainID("osmosis-testnet")
	s.prepareMainnetHuahuaGauges()
	transmuter, alloyed := s.prepareAlloyTransmuter(v33.WBTCDenom)
	holder := s.newUser()
	s.FundAcc(holder, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1_000)))
	usdcHolder := s.newUser()
	s.FundAcc(usdcHolder, sdk.NewCoins(sdk.NewInt64Coin(v33.NobleUSDCDenom, 500)))

	called := 0
	restore := v33.SetAlloyMigrationBuildersForTest([]v33.AlloyMigrationBuilder{
		func(sdk.Context, *keepers.AppKeepers) (v33.AlloyMigration, error) {
			called++
			return v33.AlloyMigration{
				Name: "WBTC to allBTC", Transmuter: transmuter, Denom: v33.WBTCDenom,
				AlloyedDenom: alloyed, MaxHolders: 100,
			}, nil
		},
		func(sdk.Context, *keepers.AppKeepers) (v33.AlloyMigration, error) {
			called++
			return v33.AlloyMigration{
				Name: "Noble USDC to allUSDC", Transmuter: transmuter, Denom: v33.NobleUSDCDenom,
				AlloyedDenom: alloyed, MaxHolders: 100,
			}, nil
		},
	})
	defer restore()

	s.runUpgrade()
	s.Require().Zero(called)
	s.Require().Equal(osmomath.NewInt(1_000), s.App.BankKeeper.GetBalance(s.Ctx, holder, v33.WBTCDenom).Amount)
	s.Require().Equal(osmomath.NewInt(500), s.App.BankKeeper.GetBalance(s.Ctx, usdcHolder, v33.NobleUSDCDenom).Amount)
	s.assertHuahuaRecovered()
}

// TestUpgradeMainnetWithoutTransmuters checks that a mainnet upgrade without the issuer or the
// transmuters neither halts nor moves either denom, while HUAHUA recovery still runs.
func (s *UpgradeTestSuite) TestUpgradeMainnetWithoutTransmuters() {
	s.Ctx = s.Ctx.WithChainID(v33.MainnetChainID)
	s.prepareMainnetHuahuaGauges()
	wbtcHolder, usdcHolder := s.newUser(), s.newUser()
	s.FundAcc(wbtcHolder, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1_000)))
	s.FundAcc(usdcHolder, sdk.NewCoins(sdk.NewInt64Coin(v33.NobleUSDCDenom, 500)))

	s.runUpgrade()

	s.Require().Equal(osmomath.NewInt(1_000), s.App.BankKeeper.GetBalance(s.Ctx, wbtcHolder, v33.WBTCDenom).Amount)
	s.Require().Equal(osmomath.NewInt(500), s.App.BankKeeper.GetBalance(s.Ctx, usdcHolder, v33.NobleUSDCDenom).Amount)
	s.assertHuahuaRecovered()
}

// TestUpgradeAlloyMigrationsAreIndependent checks that one migration's pre-flight failure does not
// stop the other migration inside the upgrade handler.
func (s *UpgradeTestSuite) TestUpgradeAlloyMigrationsAreIndependent() {
	s.Ctx = s.Ctx.WithChainID(v33.MainnetChainID)
	s.prepareMainnetHuahuaGauges()
	transmuter, alloyed := s.prepareAlloyTransmuter(v33.NobleUSDCDenom)
	usdcHolder, wbtcHolder := s.newUser(), s.newUser()
	s.FundAcc(usdcHolder, sdk.NewCoins(sdk.NewInt64Coin(v33.NobleUSDCDenom, 2_000)))
	s.FundAcc(wbtcHolder, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1_000)))

	restore := v33.SetAlloyMigrationBuildersForTest([]v33.AlloyMigrationBuilder{
		func(sdk.Context, *keepers.AppKeepers) (v33.AlloyMigration, error) {
			return v33.AlloyMigration{
				Name: "WBTC to allBTC", Transmuter: apptesting.CreateRandomAccounts(1)[0].String(),
				Denom: v33.WBTCDenom, AlloyedDenom: v33.AllBTCDenom, MaxHolders: 100,
			}, nil
		},
		func(sdk.Context, *keepers.AppKeepers) (v33.AlloyMigration, error) {
			return v33.AlloyMigration{
				Name: "Noble USDC to allUSDC", Transmuter: transmuter, Denom: v33.NobleUSDCDenom,
				AlloyedDenom: alloyed, MaxHolders: 100,
			}, nil
		},
	})
	defer restore()

	s.runUpgrade()

	s.Require().Equal(osmomath.NewInt(1_000), s.App.BankKeeper.GetBalance(s.Ctx, wbtcHolder, v33.WBTCDenom).Amount)
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, usdcHolder, v33.NobleUSDCDenom).IsZero())
	s.Require().Equal(osmomath.NewInt(2_000), s.App.BankKeeper.GetBalance(s.Ctx, usdcHolder, alloyed).Amount)
	s.assertHuahuaRecovered()
}

// assertHuahuaRecovered checks that the v33 HUAHUA recovery paid the ecosystem fund.
func (s *UpgradeTestSuite) assertHuahuaRecovered() {
	recipient := sdk.MustAccAddressFromBech32(v33.HuahuaRecoveryAddress)
	s.Require().Equal(expectedRecoveredHuahua, s.App.BankKeeper.GetBalance(s.Ctx, recipient, v33.HuahuaDenom).Amount)
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

// unindex removes the balance of addr from the bank denom index, as on mainnet for balances
// unchanged since the index was added back.
func (s *UpgradeTestSuite) unindex(addr sdk.AccAddress, denom string) {
	s.Require().NoError(s.App.BankKeeper.Balances.Indexes.Denom.Unreference(s.Ctx, collections.Join(addr, denom), nil))
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
	s.execTransmuter(transmuter, map[string]any{"register_limiter": map[string]any{
		"denom":          denom,
		"label":          "static",
		"limiter_params": map[string]any{"static_limiter": map[string]any{"upper_limit": upperLimit}},
	}})
}

// execTransmuter executes msg on the transmuter as its admin.
func (s *UpgradeTestSuite) execTransmuter(transmuter string, msg any) {
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
