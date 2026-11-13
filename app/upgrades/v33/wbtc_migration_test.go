package v33_test

import (
	"encoding/json"
	"math/big"
	"time"

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

// migrateWBTC runs the migration against transmuter, whose admin is s.TestAccs[0], with the given
// exclusions and unindexed holders.
func (s *UpgradeTestSuite) migrateWBTC(transmuter, alloyed string, excluded, unindexed []sdk.AccAddress) v33.WBTCMigrationResult {
	result, err := v33.MigrateWBTCToAllBTC(s.Ctx, &s.App.AppKeepers, s.wbtcMigration(transmuter, alloyed, excluded, unindexed))
	s.Require().NoError(err)
	return result
}

// wbtcMigration returns the migration configuration for transmuter, whose admin is s.TestAccs[0].
func (s *UpgradeTestSuite) wbtcMigration(transmuter, alloyed string, excluded, unindexed []sdk.AccAddress) v33.WBTCMigration {
	return v33.WBTCMigration{
		Transmuter: transmuter, Admin: s.TestAccs[0], WBTCDenom: v33.WBTCDenom, AlloyedDenom: alloyed,
		Excluded: excluded, UnindexedHolders: unindexed,
	}
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
	holders, err := v33.CollectHolders(s.Ctx, &s.App.AppKeepers, v33.WBTCDenom, nil)
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

	// The transmuter rejects funds above the largest Uint128, so the deposit of large fails.
	tooLarge := osmomath.NewIntFromBigInt(new(big.Int).Lsh(big.NewInt(1), 128))
	small, large := s.newUser(), s.newUser()
	s.FundAcc(small, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 100_000_000)))
	s.FundAcc(large, sdk.NewCoins(sdk.NewCoin(v33.WBTCDenom, tooLarge)))

	result := s.migrateWBTC(transmuter, alloyed, nil, nil)
	s.Require().Equal(1, result.Migrated)
	s.Require().Equal(1, result.Failed)

	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, small, v33.WBTCDenom).IsZero())
	s.Require().Equal(osmomath.NewInt(100_000_000), s.App.BankKeeper.GetBalance(s.Ctx, small, alloyed).Amount)
	s.Require().Equal(tooLarge, s.App.BankKeeper.GetBalance(s.Ctx, large, v33.WBTCDenom).Amount)
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, large, alloyed).IsZero())
}

// TestMigrateWBTCToAllBTCLiftsLimiters checks that deposits the WBTC limiters would reject are
// migrated, and that the limiters enforce their configured bounds again afterwards. As on
// osmosis-1, WBTC is at its static limiter before the migration, so any deposit would be rejected.
func (s *UpgradeTestSuite) TestMigrateWBTCToAllBTCLiftsLimiters() {
	transmuter, alloyed := s.prepareWBTCTransmuter()
	s.registerStaticLimiter(transmuter, v33.WBTCDenom, "0.5")
	s.registerChangeLimiter(transmuter, v33.WBTCDenom, "0.001")
	s.registerStaticLimiter(transmuter, otherBTCDenom, "0.9")
	limitersBefore := s.listLimiters(transmuter)

	// Give the change limiter a moving average to enforce, as on a pool that has seen swaps.
	s.joinTransmuter(transmuter, s.TestAccs[0], sdk.NewInt64Coin(otherBTCDenom, 1_000_000))
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(10 * time.Minute))
	s.Require().ErrorContains(s.tryJoinTransmuter(transmuter, sdk.NewInt64Coin(v33.WBTCDenom, 10_000_000)), "Upper limit exceeded")

	userA, userB := s.newUser(), s.newUser()
	s.FundAcc(userA, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1_000_000_000)))
	s.FundAcc(userB, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 500_000_000)))

	result := s.migrateWBTC(transmuter, alloyed, nil, nil)
	s.Require().Equal(2, result.Migrated)
	s.Require().Equal(0, result.Failed)
	s.Require().Equal(osmomath.NewInt(1_000_000_000), s.App.BankKeeper.GetBalance(s.Ctx, userA, alloyed).Amount)
	s.Require().Equal(osmomath.NewInt(500_000_000), s.App.BankKeeper.GetBalance(s.Ctx, userB, alloyed).Amount)

	// The limiters are back to their configured bounds and reject WBTC deposits again.
	s.Require().Equal(boundsOf(limitersBefore), boundsOf(s.listLimiters(transmuter)))
	s.Require().ErrorContains(s.tryJoinTransmuter(transmuter, sdk.NewInt64Coin(v33.WBTCDenom, 10_000_000)), "Upper limit exceeded")
}

// TestMigrateWBTCToAllBTCRequiresAdmin checks that nothing moves when the configured admin is not the
// transmuter's admin, since the WBTC limiters could not be lifted.
func (s *UpgradeTestSuite) TestMigrateWBTCToAllBTCRequiresAdmin() {
	transmuter, alloyed := s.prepareWBTCTransmuter()
	holder := s.newUser()
	s.FundAcc(holder, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1_000)))

	m := s.wbtcMigration(transmuter, alloyed, nil, nil)
	m.Admin = s.TestAccs[1]
	_, err := v33.MigrateWBTCToAllBTC(s.Ctx, &s.App.AppKeepers, m)
	s.Require().ErrorContains(err, "expected "+s.TestAccs[1].String())
	s.Require().Equal(osmomath.NewInt(1_000), s.App.BankKeeper.GetBalance(s.Ctx, holder, v33.WBTCDenom).Amount)
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
			m := s.wbtcMigration(transmuter, "factory/"+transmuter+"/alloyed/allBTC", nil, nil)
			_, err := v33.MigrateWBTCToAllBTC(s.Ctx, &s.App.AppKeepers, m)
			s.Require().Error(err)
			s.Require().Equal(before, s.App.BankKeeper.GetAllBalances(s.Ctx, holder))
		})
	}
}

// TestCollectHolders checks that the holder list includes listed unindexed holders.
func (s *UpgradeTestSuite) TestCollectHolders() {
	accs := apptesting.CreateRandomAccounts(3)
	for _, acc := range accs {
		s.FundAcc(acc, sdk.NewCoins(sdk.NewInt64Coin(v33.WBTCDenom, 1)))
	}
	s.unindex(accs[2])

	holders, err := v33.CollectHolders(s.Ctx, &s.App.AppKeepers, v33.WBTCDenom, accs[2:])
	s.Require().NoError(err)
	s.Require().ElementsMatch(accs, holders)
}

// TestWBTCUnindexedHoldersList checks that the embedded list parses and has no duplicates.
func (s *UpgradeTestSuite) TestWBTCUnindexedHoldersList() {
	addrs, err := v33.ParseAddresses(v33.WBTCUnindexedHolders)
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

// registerChangeLimiter registers a change limiter with a one hour window on the transmuter,
// executed by its admin.
func (s *UpgradeTestSuite) registerChangeLimiter(transmuter, denom, boundaryOffset string) {
	msg := map[string]any{"register_limiter": map[string]any{
		"denom": denom,
		"label": "change",
		"limiter_params": map[string]any{"change_limiter": map[string]any{
			"window_config":   map[string]any{"window_size": "3600000000000", "division_count": "6"},
			"boundary_offset": boundaryOffset,
		}},
	}}
	bz, err := json.Marshal(msg)
	s.Require().NoError(err)
	_, err = s.App.ContractKeeper.Execute(s.Ctx, sdk.MustAccAddressFromBech32(transmuter), s.TestAccs[0], bz, nil)
	s.Require().NoError(err)
}

// listLimiters returns the transmuter's limiters by denom and label.
func (s *UpgradeTestSuite) listLimiters(transmuter string) map[[2]string]map[string]map[string]any {
	bz, err := s.App.WasmKeeper.QuerySmart(s.Ctx, sdk.MustAccAddressFromBech32(transmuter), []byte(`{"list_limiters": {}}`))
	s.Require().NoError(err)
	var resp struct {
		Limiters [][2]json.RawMessage `json:"limiters"`
	}
	s.Require().NoError(json.Unmarshal(bz, &resp))
	limiters := make(map[[2]string]map[string]map[string]any, len(resp.Limiters))
	for _, entry := range resp.Limiters {
		var key [2]string
		var limiter map[string]map[string]any
		s.Require().NoError(json.Unmarshal(entry[0], &key))
		s.Require().NoError(json.Unmarshal(entry[1], &limiter))
		limiters[key] = limiter
	}
	return limiters
}

// boundsOf returns the type and configured bound of each limiter, leaving out the moving average
// state of change limiters, which deposits update.
func boundsOf(limiters map[[2]string]map[string]map[string]any) map[[2]string]string {
	bounds := make(map[[2]string]string, len(limiters))
	for key, limiter := range limiters {
		if static, ok := limiter["static_limiter"]; ok {
			bounds[key] = "static " + static["upper_limit"].(string)
		} else {
			bounds[key] = "change " + limiter["change_limiter"]["boundary_offset"].(string)
		}
	}
	return bounds
}

// joinTransmuter funds sender with coin and deposits it into the transmuter.
func (s *UpgradeTestSuite) joinTransmuter(transmuter string, sender sdk.AccAddress, coin sdk.Coin) {
	s.FundAcc(sender, sdk.NewCoins(coin))
	_, err := s.App.ContractKeeper.Execute(s.Ctx, sdk.MustAccAddressFromBech32(transmuter), sender, []byte(`{"join_pool": {}}`), sdk.NewCoins(coin))
	s.Require().NoError(err)
}

// tryJoinTransmuter deposits coin into the transmuter from a new account in a discarded cache
// context, and returns the error of the deposit.
func (s *UpgradeTestSuite) tryJoinTransmuter(transmuter string, coin sdk.Coin) error {
	ctx := s.Ctx
	defer func() { s.Ctx = ctx }()
	s.Ctx, _ = ctx.CacheContext()
	sender := apptesting.CreateRandomAccounts(1)[0]
	s.FundAcc(sender, sdk.NewCoins(coin))
	_, err := s.App.ContractKeeper.Execute(s.Ctx, sdk.MustAccAddressFromBech32(transmuter), sender, []byte(`{"join_pool": {}}`), sdk.NewCoins(coin))
	return err
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
