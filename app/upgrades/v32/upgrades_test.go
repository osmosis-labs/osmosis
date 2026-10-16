package v32_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v32/app/params"
	v32 "github.com/osmosis-labs/osmosis/v32/app/upgrades/v32"
	incentivestypes "github.com/osmosis-labs/osmosis/v32/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v32/x/lockup/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v32/x/protorev/types"
)

const (
	v32UpgradeHeight = int64(10)
	// exploiterUosmo is the frozen account's only other mainnet balance: 703.055696 OSMO.
	exploiterUosmo = int64(703055696)
)

var (
	// Mainnet state of the stuck gauges, as returned by /osmosis/incentives/v1beta1/gauge_by_id/{id}.
	// Every gauge started on 2022-01-15, is paid over 180 epochs and never distributed anything.
	huahuaGaugeStartTime = time.Date(2022, 1, 15, 0, 0, 0, 0, time.UTC)
	huahuaGaugeAmounts   = map[uint64]int64{
		1954: 353_000_000_000_000,
		1955: 882_500_000_000_000,
		1956: 1_764_500_000_000_000,
		1957: 353_000_000_000_000,
		1958: 882_500_000_000_000,
		1959: 1_764_500_000_000_000,
	}
	huahuaGaugeDurations = map[uint64]time.Duration{
		1954: 24 * time.Hour, 1955: 7 * 24 * time.Hour, 1956: 14 * 24 * time.Hour,
		1957: 24 * time.Hour, 1958: 7 * 24 * time.Hour, 1959: 14 * 24 * time.Hour,
	}
	// 6,000,000,000 HUAHUA (6 decimals).
	expectedRecoveredHuahua = osmomath.NewInt(6_000_000_000_000_000)
	// The incentives module on mainnet holds 14382 uhuahua more than the gauges; it must stay there.
	huahuaModuleDust = osmomath.NewInt(14_382)

	// Upgrade time on a date after the gauges' start time.
	upgradeBlockTime = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule appmodule.HasPreBlocker
}

// TestUpgradeTestSuite runs the v32 upgrade test suite.
func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

// TestHuahuaRecovery runs the v32 upgrade on the mainnet gauge state and checks the recovered amounts,
// the finished gauges and that unrelated gauges are untouched.
func (s *UpgradeTestSuite) TestHuahuaRecovery() {
	s.setupHuahuaUpgradeTest()
	s.prepareMainnetHuahuaGauges()
	controlGaugeID := s.createControlGauge()

	recipient := sdk.MustAccAddressFromBech32(v32.HuahuaRecoveryAddress)
	moduleAddr := s.App.AccountKeeper.GetModuleAddress(incentivestypes.ModuleName)
	recipientBefore := s.App.BankKeeper.GetBalance(s.Ctx, recipient, v32.HuahuaDenom)
	activeBefore := len(s.App.IncentivesKeeper.GetActiveGauges(s.Ctx))
	controlBefore, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, controlGaugeID)
	s.Require().NoError(err)

	s.Require().NoError(s.runUpgrade())

	// 6B HUAHUA moved to the recovery address; the dust and the control gauge's HUAHUA stay in the module.
	s.Require().Equal(recipientBefore.Amount.Add(expectedRecoveredHuahua), s.App.BankKeeper.GetBalance(s.Ctx, recipient, v32.HuahuaDenom).Amount)
	s.Require().Equal(huahuaModuleDust.Add(controlBefore.Coins.AmountOf(v32.HuahuaDenom)), s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, v32.HuahuaDenom).Amount)

	// All six gauges are finished and fully accounted as distributed.
	finished := gaugeIDSet(s.App.IncentivesKeeper.GetFinishedGauges(s.Ctx))
	active := gaugeIDSet(s.App.IncentivesKeeper.GetActiveGauges(s.Ctx))
	for _, stuck := range v32.HuahuaStuckGauges {
		gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, stuck.GaugeID)
		s.Require().NoError(err)
		s.Require().Equal(gauge.Coins, gauge.DistributedCoins, "gauge %d", stuck.GaugeID)
		s.Require().Equal(uint64(180), gauge.FilledEpochs, "gauge %d", stuck.GaugeID)
		s.Require().True(gauge.IsFinishedGauge(s.Ctx.BlockTime()), "gauge %d", stuck.GaugeID)
		s.Require().Contains(finished, stuck.GaugeID)
		s.Require().NotContains(active, stuck.GaugeID)
	}
	s.Require().Len(active, activeBefore-len(v32.HuahuaStuckGauges))
	s.Require().Empty(s.denomRefs("GAMM605"))
	s.Require().Empty(s.denomRefs("GAMM606"))
	// Only the control gauge's HUAHUA is left to distribute.
	s.Require().Equal(controlBefore.Coins.AmountOf(v32.HuahuaDenom), s.App.IncentivesKeeper.GetModuleToDistributeCoins(s.Ctx).AmountOf(v32.HuahuaDenom))

	// Unrelated gauges are untouched.
	controlAfter, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, controlGaugeID)
	s.Require().NoError(err)
	s.Require().Equal(controlBefore, controlAfter)
	s.Require().Contains(active, controlGaugeID)

	// The next epoch distribution runs fine and doesn't see the recovered gauges.
	s.Require().NotPanics(func() {
		_, err := s.App.IncentivesKeeper.Distribute(s.Ctx, s.App.IncentivesKeeper.GetActiveGauges(s.Ctx))
		s.Require().NoError(err)
	})

	// Nobody can add rewards to the recovered gauges anymore.
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, "uosmo", v32.HuahuaDenom, 605)
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewInt64Coin(v32.HuahuaDenom, 1)))
	err = s.App.IncentivesKeeper.AddToGaugeRewards(s.Ctx, s.TestAccs[0], sdk.NewCoins(sdk.NewInt64Coin(v32.HuahuaDenom, 1)), 1954)
	s.Require().ErrorIs(err, incentivestypes.UnexpectedFinishedGaugeError{GaugeId: 1954})
}

// TestHuahuaRecoveryWithExtraCoinsAdded checks that coins added to a gauge before the upgrade are
// recovered too: anyone can add to an active gauge, and that must not block the recovery.
func (s *UpgradeTestSuite) TestHuahuaRecoveryWithExtraCoinsAdded() {
	s.setupHuahuaUpgradeTest()
	s.prepareMainnetHuahuaGauges()

	extra := sdk.NewCoins(sdk.NewInt64Coin(v32.HuahuaDenom, 1_000), sdk.NewInt64Coin("uosmo", 5))
	s.FundModuleAcc(incentivestypes.ModuleName, extra)
	gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, 1957)
	s.Require().NoError(err)
	gauge.Coins = gauge.Coins.Add(extra...)
	s.setGaugeRaw(gauge)

	s.Require().NoError(s.runUpgrade())

	recipient := sdk.MustAccAddressFromBech32(v32.HuahuaRecoveryAddress)
	s.Require().Equal(expectedRecoveredHuahua.AddRaw(1_000), s.App.BankKeeper.GetBalance(s.Ctx, recipient, v32.HuahuaDenom).Amount)
	s.Require().Equal(osmomath.NewInt(5), s.App.BankKeeper.GetBalance(s.Ctx, recipient, "uosmo").Amount)
}

// TestHuahuaRecoverySkippedOnUnexpectedState checks that, when the gauges are not in the expected
// state, the upgrade neither halts the chain nor applies a partial recovery.
func (s *UpgradeTestSuite) TestHuahuaRecoverySkippedOnUnexpectedState() {
	tests := map[string]func(){
		"gauge distributes to an unexpected denom": func() {
			gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, 1959)
			s.Require().NoError(err)
			gauge.DistributeTo.Denom = "gamm/pool/606"
			s.setGaugeRaw(gauge)
		},
		"gauge holds no HUAHUA": func() {
			gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, 1959)
			s.Require().NoError(err)
			gauge.Coins = sdk.NewCoins(sdk.NewInt64Coin("uosmo", 1))
			s.setGaugeRaw(gauge)
		},
		"gauge is already finished": func() {
			gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, 1959)
			s.Require().NoError(err)
			gauge.FilledEpochs = gauge.NumEpochsPaidOver
			s.setGaugeRaw(gauge)
		},
	}

	for name, corrupt := range tests {
		s.Run(name, func() {
			s.setupHuahuaUpgradeTest()
			s.prepareMainnetHuahuaGauges()
			corrupt()

			moduleAddr := s.App.AccountKeeper.GetModuleAddress(incentivestypes.ModuleName)
			moduleBefore := s.App.BankKeeper.GetAllBalances(s.Ctx, moduleAddr)
			gaugesBefore, err := s.App.IncentivesKeeper.GetGaugeFromIDs(s.Ctx, []uint64{1954, 1955, 1956, 1957, 1958})
			s.Require().NoError(err)
			activeBefore := s.App.IncentivesKeeper.GetActiveGauges(s.Ctx)

			s.Require().NoError(s.runUpgrade())

			// Nothing moved, including the gauges processed before the failing one.
			recipient := sdk.MustAccAddressFromBech32(v32.HuahuaRecoveryAddress)
			s.Require().True(s.App.BankKeeper.GetAllBalances(s.Ctx, recipient).IsZero())
			s.Require().Equal(moduleBefore, s.App.BankKeeper.GetAllBalances(s.Ctx, moduleAddr))
			gaugesAfter, err := s.App.IncentivesKeeper.GetGaugeFromIDs(s.Ctx, []uint64{1954, 1955, 1956, 1957, 1958})
			s.Require().NoError(err)
			s.Require().Equal(gaugesBefore, gaugesAfter)
			s.Require().Equal(activeBefore, s.App.IncentivesKeeper.GetActiveGauges(s.Ctx))
		})
	}
}

// prepareMainnetHuahuaGauges writes gauges 1954-1959 exactly as they are on mainnet, including the
// missing denom index for GAMM606, and funds the incentives module accordingly.
func (s *UpgradeTestSuite) prepareMainnetHuahuaGauges() {
	total := osmomath.ZeroInt()
	for _, stuck := range v32.HuahuaStuckGauges {
		amount := osmomath.NewInt(huahuaGaugeAmounts[stuck.GaugeID])
		gauge := incentivestypes.Gauge{
			Id:          stuck.GaugeID,
			IsPerpetual: false,
			DistributeTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         stuck.Denom,
				Duration:      huahuaGaugeDurations[stuck.GaugeID],
				Timestamp:     time.Unix(0, 0).UTC(),
			},
			Coins:             sdk.NewCoins(sdk.NewCoin(v32.HuahuaDenom, amount)),
			StartTime:         huahuaGaugeStartTime,
			NumEpochsPaidOver: 180,
			FilledEpochs:      0,
			DistributedCoins:  sdk.NewCoins(),
		}
		s.Require().NoError(s.App.IncentivesKeeper.SetGaugeWithRefKey(s.Ctx, &gauge))
		total = total.Add(amount)
	}
	s.Require().Equal(expectedRecoveredHuahua, total)
	s.App.IncentivesKeeper.SetLastGaugeID(s.Ctx, 1959)

	// On mainnet GAMM605 still indexes 1954-1956, while GAMM606 indexes nothing.
	s.removeDenomRef("GAMM606")
	s.Require().Len(s.denomRefs("GAMM605"), 3)
	s.Require().Empty(s.denomRefs("GAMM606"))

	s.FundModuleAcc(incentivestypes.ModuleName, sdk.NewCoins(sdk.NewCoin(v32.HuahuaDenom, total.Add(huahuaModuleDust))))

	active := gaugeIDSet(s.App.IncentivesKeeper.GetActiveGauges(s.Ctx))
	for _, stuck := range v32.HuahuaStuckGauges {
		s.Require().Contains(active, stuck.GaugeID)
	}
}

// createControlGauge creates a regular active gauge that the upgrade must not touch.
func (s *UpgradeTestSuite) createControlGauge() uint64 {
	coins := sdk.NewCoins(sdk.NewInt64Coin(v32.HuahuaDenom, 1_000), sdk.NewInt64Coin("uion", 1_000))
	gauge := incentivestypes.Gauge{
		Id:                2000,
		DistributeTo:      lockuptypes.QueryCondition{LockQueryType: lockuptypes.ByDuration, Denom: "gamm/pool/605", Duration: 24 * time.Hour},
		Coins:             coins,
		StartTime:         huahuaGaugeStartTime,
		NumEpochsPaidOver: 10_000,
		DistributedCoins:  sdk.NewCoins(),
	}
	s.Require().NoError(s.App.IncentivesKeeper.SetGaugeWithRefKey(s.Ctx, &gauge))
	s.FundModuleAcc(incentivestypes.ModuleName, coins)
	return gauge.Id
}

// The helpers below access the incentives store directly, mirroring the unexported keys of x/incentives/keeper.

// incentivesStore returns the incentives module KV store.
func (s *UpgradeTestSuite) incentivesStore() storetypes.KVStore {
	return s.Ctx.KVStore(s.App.GetKey(incentivestypes.StoreKey))
}

// combineKeys joins two keys with the incentives key separator.
func combineKeys(a, b []byte) []byte {
	key := append(append([]byte{}, a...), incentivestypes.KeyIndexSeparator...)
	return append(key, b...)
}

// gaugeStoreKey returns the store key of the gauge with the given ID.
func gaugeStoreKey(id uint64) []byte {
	return combineKeys(incentivestypes.KeyPrefixPeriodGauge, sdk.Uint64ToBigEndian(id))
}

// gaugeDenomStoreKey returns the store key of the gauge index for the given lock denom.
func gaugeDenomStoreKey(denom string) []byte {
	return combineKeys(incentivestypes.KeyPrefixGaugesByDenom, []byte(denom))
}

// setGaugeRaw overwrites a gauge in the store without touching its reference keys.
func (s *UpgradeTestSuite) setGaugeRaw(gauge *incentivestypes.Gauge) {
	bz, err := proto.Marshal(gauge)
	s.Require().NoError(err)
	s.incentivesStore().Set(gaugeStoreKey(gauge.Id), bz)
}

// removeDenomRef deletes the gauge index for the given lock denom.
func (s *UpgradeTestSuite) removeDenomRef(denom string) {
	s.incentivesStore().Delete(gaugeDenomStoreKey(denom))
}

// denomRefs returns the gauge IDs indexed under the given lock denom.
func (s *UpgradeTestSuite) denomRefs(denom string) []uint64 {
	ids := []uint64{}
	if bz := s.incentivesStore().Get(gaugeDenomStoreKey(denom)); bz != nil {
		s.Require().NoError(json.Unmarshal(bz, &ids))
	}
	return ids
}

// gaugeIDSet returns the IDs of the given gauges.
func gaugeIDSet(gauges []incentivestypes.Gauge) []uint64 {
	ids := make([]uint64, len(gauges))
	for i, g := range gauges {
		ids[i] = g.Id
	}
	return ids
}

func (s *UpgradeTestSuite) TestMoveProtorevNonBaseDenomsToSubDAO() {
	// Non-mainnet chain ID so the allBTC seizure step is skipped and this test only covers the protorev sweep.
	s.setupUpgradeTestWithChainID("osmo-test-a")

	recipient := sdk.MustAccAddressFromBech32(v32.LiquiditySubDAOAddress)
	protorevAddr := s.App.AccountKeeper.GetModuleAddress(protorevtypes.ModuleName)

	baseDenoms, err := s.App.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(baseDenoms)
	baseDenom := baseDenoms[0].Denom

	baseCoin := sdk.NewCoin(baseDenom, osmomath.NewInt(1_000))
	stuckCoins := sdk.NewCoins(
		sdk.NewCoin("ibc/STUCKA", osmomath.NewInt(5_000)),
		sdk.NewCoin("ibc/STUCKB", osmomath.NewInt(7_000)),
	)
	s.FundModuleAcc(protorevtypes.ModuleName, stuckCoins.Add(baseCoin))

	recipientBefore := s.App.BankKeeper.GetAllBalances(s.Ctx, recipient)

	s.Require().NoError(s.runUpgrade())

	// Base denom stays in protorev for the regular epoch distribution; everything else is gone.
	s.Require().Equal(sdk.NewCoins(baseCoin), s.App.BankKeeper.GetAllBalances(s.Ctx, protorevAddr))
	s.Require().Equal(recipientBefore.Add(stuckCoins...), s.App.BankKeeper.GetAllBalances(s.Ctx, recipient))
}

func (s *UpgradeTestSuite) TestSeizeFrozenAllBTC() {
	s.setupUpgradeTest()

	exploiter := sdk.MustAccAddressFromBech32(v32.ExploiterAddress)
	subDAO := sdk.MustAccAddressFromBech32(v32.LiquiditySubDAOAddress)
	expected := expectedAllBTC(s)

	s.FundAcc(exploiter, sdk.NewCoins(
		sdk.NewCoin(v32.AllBTCDenom, expected),
		sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(exploiterUosmo)),
	))

	s.Require().NoError(s.runUpgrade())

	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, exploiter, v32.AllBTCDenom).IsZero())
	s.Require().Equal(
		sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(exploiterUosmo)),
		s.App.BankKeeper.GetBalance(s.Ctx, exploiter, appparams.BaseCoinUnit),
	)
	s.Require().Equal(sdk.NewCoin(v32.AllBTCDenom, expected), s.App.BankKeeper.GetBalance(s.Ctx, subDAO, v32.AllBTCDenom))
}

func (s *UpgradeTestSuite) TestSeizeFrozenAllBTCRejectsUnexpectedBalance() {
	exploiter := sdk.MustAccAddressFromBech32(v32.ExploiterAddress)
	subDAO := sdk.MustAccAddressFromBech32(v32.LiquiditySubDAOAddress)

	for _, amount := range []osmomath.Int{osmomath.ZeroInt(), osmomath.NewInt(1)} {
		s.setupUpgradeTest()

		coins := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(exploiterUosmo)))
		if !amount.IsZero() {
			coins = coins.Add(sdk.NewCoin(v32.AllBTCDenom, amount))
		}
		s.FundAcc(exploiter, coins)

		err := s.runUpgrade()
		s.Require().Error(err)
		s.Require().ErrorContains(err, "seizing frozen allBTC")

		s.Require().Equal(sdk.NewCoin(v32.AllBTCDenom, amount), s.App.BankKeeper.GetBalance(s.Ctx, exploiter, v32.AllBTCDenom))
		s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, subDAO, v32.AllBTCDenom).IsZero())
		s.Require().Equal(
			sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(exploiterUosmo)),
			s.App.BankKeeper.GetBalance(s.Ctx, exploiter, appparams.BaseCoinUnit),
		)
	}
}

func (s *UpgradeTestSuite) TestSeizeFrozenAllBTCLeavesSurplus() {
	s.setupUpgradeTest()

	exploiter := sdk.MustAccAddressFromBech32(v32.ExploiterAddress)
	subDAO := sdk.MustAccAddressFromBech32(v32.LiquiditySubDAOAddress)
	authorized := expectedAllBTC(s)
	surplus := osmomath.NewInt(1)

	s.FundAcc(exploiter, sdk.NewCoins(
		sdk.NewCoin(v32.AllBTCDenom, authorized.Add(surplus)),
		sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(exploiterUosmo)),
	))

	s.Require().NoError(s.runUpgrade())

	s.Require().Equal(sdk.NewCoin(v32.AllBTCDenom, surplus), s.App.BankKeeper.GetBalance(s.Ctx, exploiter, v32.AllBTCDenom))
	s.Require().Equal(
		sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(exploiterUosmo)),
		s.App.BankKeeper.GetBalance(s.Ctx, exploiter, appparams.BaseCoinUnit),
	)
	s.Require().Equal(sdk.NewCoin(v32.AllBTCDenom, authorized), s.App.BankKeeper.GetBalance(s.Ctx, subDAO, v32.AllBTCDenom))
}

func (s *UpgradeTestSuite) TestSeizeFrozenAllBTCAddsToExistingSubDAOBalance() {
	s.setupUpgradeTest()

	exploiter := sdk.MustAccAddressFromBech32(v32.ExploiterAddress)
	subDAO := sdk.MustAccAddressFromBech32(v32.LiquiditySubDAOAddress)
	expected := expectedAllBTC(s)
	existing := osmomath.NewInt(100)

	s.FundAcc(subDAO, sdk.NewCoins(sdk.NewCoin(v32.AllBTCDenom, existing)))
	s.FundAcc(exploiter, sdk.NewCoins(sdk.NewCoin(v32.AllBTCDenom, expected)))

	s.Require().NoError(s.runUpgrade())

	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, exploiter, v32.AllBTCDenom).IsZero())
	s.Require().Equal(sdk.NewCoin(v32.AllBTCDenom, existing.Add(expected)), s.App.BankKeeper.GetBalance(s.Ctx, subDAO, v32.AllBTCDenom))
}

func (s *UpgradeTestSuite) TestSkipsSeizureOnNonMainnetChain() {
	s.setupUpgradeTestWithChainID("osmo-test-a")

	exploiter := sdk.MustAccAddressFromBech32(v32.ExploiterAddress)
	subDAO := sdk.MustAccAddressFromBech32(v32.LiquiditySubDAOAddress)

	s.Require().NoError(s.runUpgrade())

	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, exploiter, v32.AllBTCDenom).IsZero())
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, subDAO, v32.AllBTCDenom).IsZero())
}

func (s *UpgradeTestSuite) setupUpgradeTest() {
	s.setupUpgradeTestWithChainID(v32.MainnetChainID)
}

func (s *UpgradeTestSuite) setupUpgradeTestWithChainID(chainID string) {
	s.SetupWithCustomChainId(chainID)
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))
}

// setupHuahuaUpgradeTest sets up a non-mainnet app (the allBTC seizure is mainnet-only) with the block
// time moved past the stuck gauges' start time, so they are active when the upgrade runs.
func (s *UpgradeTestSuite) setupHuahuaUpgradeTest() {
	s.setupUpgradeTestWithChainID("osmo-test-a")
	s.Ctx = s.Ctx.WithBlockTime(upgradeBlockTime)
}

func (s *UpgradeTestSuite) runUpgrade() error {
	s.Ctx = s.Ctx.WithBlockHeight(v32UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: v32.UpgradeName, Height: v32UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: v32UpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(v32UpgradeHeight)
	_, err = s.preModule.PreBlock(s.Ctx)
	return err
}

func expectedAllBTC(s *UpgradeTestSuite) osmomath.Int {
	amount, ok := osmomath.NewIntFromString(v32.ExpectedExploiterAllBTCAmount)
	s.Require().True(ok)
	return amount
}
