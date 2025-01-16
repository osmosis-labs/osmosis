package v24_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/osmosis-labs/osmosis/v29/x/cosmwasmpool/model"
	cwpooltypes "github.com/osmosis-labs/osmosis/v29/x/cosmwasmpool/types"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v29/app/apptesting"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"

	appparams "github.com/osmosis-labs/osmosis/v29/app/params"
	v24 "github.com/osmosis-labs/osmosis/v29/app/upgrades/v24"
	concentratedtypes "github.com/osmosis-labs/osmosis/v29/x/concentrated-liquidity/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v29/x/incentives/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v29/x/protorev/types"
	twap "github.com/osmosis-labs/osmosis/v29/x/twap"
	"github.com/osmosis-labs/osmosis/v29/x/twap/types"
	twaptypes "github.com/osmosis-labs/osmosis/v29/x/twap/types"
)

const (
	v24UpgradeHeight              = int64(10)
	HistoricalTWAPTimeIndexPrefix = "historical_time_index"
	KeySeparator                  = "|"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule appmodule.HasPreBlocker
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))

	// TWAP Setup
	//

	// Manually set up TWAP records indexed by both pool ID and time.
	twapStoreKey := s.App.GetKey(twaptypes.ModuleName)
	store := s.Ctx.KVStore(twapStoreKey)
	twapRecord1 := twaptypes.TwapRecord{
		PoolId:                      1,
		Asset0Denom:                 "foo",
		Asset1Denom:                 "bar",
		Height:                      1,
		Time:                        time.Date(2023, 0o2, 1, 0, 0, 0, 0, time.UTC),
		P0LastSpotPrice:             osmomath.OneDec(),
		P1LastSpotPrice:             osmomath.OneDec(),
		P0ArithmeticTwapAccumulator: osmomath.ZeroDec(),
		P1ArithmeticTwapAccumulator: osmomath.ZeroDec(),
		GeometricTwapAccumulator:    osmomath.ZeroDec(),
		LastErrorTime:               time.Time{}, // no previous error
	}
	twapRecord2 := twapRecord1
	twapRecord2.Time = time.Date(2023, 0o2, 2, 0, 0, 0, 0, time.UTC)
	twap.NumDeprecatedRecordsToPrunePerBlock = uint16(1)

	// Set two records
	poolIndexKey1 := types.FormatHistoricalPoolIndexTWAPKey(twapRecord1.PoolId, twapRecord1.Asset0Denom, twapRecord1.Asset1Denom, twapRecord1.Time)
	poolIndexKey2 := types.FormatHistoricalPoolIndexTWAPKey(twapRecord2.PoolId, twapRecord2.Asset0Denom, twapRecord2.Asset1Denom, twapRecord2.Time)
	osmoutils.MustSet(store, poolIndexKey1, &twapRecord1)
	osmoutils.MustSet(store, poolIndexKey2, &twapRecord2)

	// The time index key is a bit manual since we removed the old code that did this programmatically.
	var buffer bytes.Buffer
	timeS1 := osmoutils.FormatTimeString(twapRecord1.Time)
	fmt.Fprintf(&buffer, "%s%d%s%s%s%s%s%s", HistoricalTWAPTimeIndexPrefix, twapRecord1.PoolId, KeySeparator, twapRecord1.Asset0Denom, KeySeparator, twapRecord1.Asset1Denom, KeySeparator, timeS1)
	timeIndexKey1 := buffer.Bytes()
	timeS2 := osmoutils.FormatTimeString(twapRecord2.Time)
	fmt.Fprintf(&buffer, "%s%d%s%s%s%s%s%s", HistoricalTWAPTimeIndexPrefix, twapRecord2.PoolId, KeySeparator, twapRecord2.Asset0Denom, KeySeparator, twapRecord2.Asset1Denom, KeySeparator, timeS2)
	timeIndexKey2 := buffer.Bytes()
	osmoutils.MustSet(store, timeIndexKey1, &twapRecord1)
	osmoutils.MustSet(store, timeIndexKey2, &twapRecord2)

	// TWAP records indexed by time should exist
	twapRecords, err := osmoutils.GatherValuesFromStorePrefix(store, []byte(HistoricalTWAPTimeIndexPrefix), types.ParseTwapFromBz)
	s.Require().NoError(err)
	s.Require().Len(twapRecords, 2)
	s.Require().Equal(twapRecord1, twapRecords[0])
	s.Require().Equal(twapRecord2, twapRecords[1])

	// TWAP records indexed by pool ID should exist.
	twapRecords, err = s.App.TwapKeeper.GetAllHistoricalPoolIndexedTWAPsForPoolId(s.Ctx, twapRecord1.PoolId)
	s.Require().NoError(err)
	s.Require().Len(twapRecords, 2)
	s.Require().Equal(twapRecord1, twapRecords[0])
	s.Require().Equal(twapRecord2, twapRecords[1])

	// INCENTIVES Setup
	//

	concentratedPoolIDs := []uint64{}

	// Create two sets of all pool types
	allPoolsOne := s.PrepareAllSupportedPools()
	allPoolsTwo := s.PrepareAllSupportedPools()

	concentratedPoolIDs = append(concentratedPoolIDs, allPoolsOne.ConcentratedPoolID)
	concentratedPoolIDs = append(concentratedPoolIDs, allPoolsTwo.ConcentratedPoolID)

	// Update authorized quote denoms
	concentratedParams := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
	// concentratedParams.AuthorizedQuoteDenoms = append(concentratedParams.AuthorizedQuoteDenoms, apptesting.USDC)
	s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, concentratedParams)

	// Create two more concentrated pools with positions
	secondLastPoolID := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx)
	lastPoolID := secondLastPoolID + 1
	concentratedPoolIDs = append(concentratedPoolIDs, secondLastPoolID)
	concentratedPoolIDs = append(concentratedPoolIDs, lastPoolID)
	s.CreateConcentratedPoolsAndFullRangePosition([][]string{
		{"uion", appparams.BaseCoinUnit},
		{apptesting.ETH, apptesting.USDC},
	})

	lastPoolPositionID := s.App.ConcentratedLiquidityKeeper.GetNextPositionId(s.Ctx) - 1

	// Create incentive record for last pool
	incentiveCoin := sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000000))
	_, err = s.App.ConcentratedLiquidityKeeper.CreateIncentive(s.Ctx, lastPoolID, s.TestAccs[0], incentiveCoin, osmomath.OneDec(), s.Ctx.BlockTime(), concentratedtypes.DefaultAuthorizedUptimes[0])
	s.Require().NoError(err)

	// Create incentive record for second last pool
	_, err = s.App.ConcentratedLiquidityKeeper.CreateIncentive(s.Ctx, secondLastPoolID, s.TestAccs[0], incentiveCoin, osmomath.OneDec(), s.Ctx.BlockTime(), concentratedtypes.DefaultAuthorizedUptimes[0])
	s.Require().NoError(err)

	// Make 60 seconds pass
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Minute))

	err = s.App.ConcentratedLiquidityKeeper.UpdatePoolUptimeAccumulatorsToNow(s.Ctx, lastPoolID)
	s.Require().NoError(err)

	// Migrated pool claim
	migratedPoolBeforeUpgradeIncentives, _, err := s.App.ConcentratedLiquidityKeeper.GetClaimableIncentives(s.Ctx, lastPoolPositionID)
	s.Require().NoError(err)
	s.Require().NotEmpty(migratedPoolBeforeUpgradeIncentives)

	// Non-migrated pool claim
	nonMigratedPoolBeforeUpgradeIncentives, _, err := s.App.ConcentratedLiquidityKeeper.GetClaimableIncentives(s.Ctx, lastPoolPositionID-1)
	s.Require().NoError(err)
	s.Require().NotEmpty(nonMigratedPoolBeforeUpgradeIncentives)

	// Overwrite the migration list with the desired pool ID.
	concentratedtypes.MigratedIncentiveAccumulatorPoolIDsV24 = map[uint64]struct{}{}
	concentratedtypes.MigratedIncentiveAccumulatorPoolIDsV24[lastPoolID] = struct{}{}

	// PROTOREV Setup
	//

	// Set the old KVStore base denoms
	s.App.ProtoRevKeeper.DeprecatedSetBaseDenoms(s.Ctx, []protorevtypes.BaseDenom{
		{Denom: protorevtypes.OsmosisDenomination, StepSize: osmomath.NewInt(1_000_000)},
		{Denom: "atom", StepSize: osmomath.NewInt(1_000_000)},
		{Denom: "weth", StepSize: osmomath.NewInt(1_000_000)}})
	oldBaseDenoms, err := s.App.ProtoRevKeeper.DeprecatedGetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(3, len(oldBaseDenoms))
	s.Require().Equal(oldBaseDenoms[0].Denom, protorevtypes.OsmosisDenomination)
	s.Require().Equal(oldBaseDenoms[1].Denom, "atom")
	s.Require().Equal(oldBaseDenoms[2].Denom, "weth")

	// The new KVStore should be set to the default
	newBaseDenoms, err := s.App.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(protorevtypes.DefaultBaseDenoms, newBaseDenoms)

	whiteWhalePoolIds := []uint64{1463, 1462, 1461}
	for _, poolId := range whiteWhalePoolIds {
		s.App.CosmwasmPoolKeeper.SetPool(s.Ctx, &model.CosmWasmPool{
			ContractAddress: "foo",
			PoolId:          poolId,
			CodeId:          503,
			InstantiateMsg:  []byte("bar"),
		})
	}
	s.requirePoolsHaveCodeId(whiteWhalePoolIds, 503)

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
	})

	// TWAP Tests
	//

	// TWAP records indexed by time should be untouched since endblocker hasn't run yet.
	twapRecords, err = osmoutils.GatherValuesFromStorePrefix(store, []byte(HistoricalTWAPTimeIndexPrefix), types.ParseTwapFromBz)
	s.Require().NoError(err)
	s.Require().Len(twapRecords, 2)

	// Run the end blocker
	s.Require().NotPanics(func() {
		_, err := s.App.EndBlocker(s.Ctx)
		s.Require().NoError(err)
	})

	// Since the prune limit was 1, 1 TWAP record indexed by time should be completely removed, leaving one more.
	// twapRecords, err = osmoutils.GatherValuesFromStorePrefix(store, []byte(HistoricalTWAPTimeIndexPrefix), types.ParseTwapFromBz)
	// s.Require().NoError(err)
	// s.Require().Len(twapRecords, 1)
	// s.Require().Equal(twapRecord2, twapRecords[0])

	// TWAP records indexed by pool ID should be untouched.
	twapRecords, err = s.App.TwapKeeper.GetAllHistoricalPoolIndexedTWAPsForPoolId(s.Ctx, twapRecord1.PoolId)
	s.Require().NoError(err)
	s.Require().Len(twapRecords, 8)
	s.Require().Equal(twapRecord1, twapRecords[6])
	s.Require().Equal(twapRecord2, twapRecords[7])

	// PROTOREV Tests
	//

	// The new KVStore should return the old KVStore values
	newBaseDenoms, err = s.App.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(oldBaseDenoms, newBaseDenoms)

	// The old KVStore base denoms should be deleted
	oldBaseDenoms, err = s.App.ProtoRevKeeper.DeprecatedGetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)
	s.Require().Empty(oldBaseDenoms)

	// INCENTIVES Tests
	//

	// Migrated pool: ensure that the claimable incentives are the same before and after migration
	migratedPoolAfterUpgradeIncentives, _, err := s.App.ConcentratedLiquidityKeeper.GetClaimableIncentives(s.Ctx, lastPoolPositionID)
	s.Require().NoError(err)

	s.Require().Equal(migratedPoolBeforeUpgradeIncentives.String(), migratedPoolAfterUpgradeIncentives.String())

	// Non-migrated pool: ensure that the claimable incentives are the same before and after migration
	nonMigratedPoolAfterUpgradeIncentives, _, err := s.App.ConcentratedLiquidityKeeper.GetClaimableIncentives(s.Ctx, lastPoolPositionID-1)

	s.Require().NoError(err)
	s.Require().Equal(nonMigratedPoolBeforeUpgradeIncentives.String(), nonMigratedPoolAfterUpgradeIncentives.String())

	// Check that the new min value for distribution has been set
	params := s.App.IncentivesKeeper.GetParams(s.Ctx)
	s.Require().Equal(incentivestypes.DefaultMinValueForDistr, params.MinValueForDistribution)

	// Pool Migration Tests
	//

	// Test that the white whale pools have been updated
	s.requirePoolsHaveCodeId(whiteWhalePoolIds, 641)

	// TXFEES Tests
	//

	// Check that the whitelisted fee token address has been set
	whitelistedFeeTokenSetters := s.App.TxFeesKeeper.GetParams(s.Ctx).WhitelistedFeeTokenSetters
	s.Require().Len(whitelistedFeeTokenSetters, 1)
	s.Require().Equal(whitelistedFeeTokenSetters, v24.WhitelistedFeeTokenSetters)
}

func (s *UpgradeTestSuite) requirePoolsHaveCodeId(pools []uint64, codeId uint64) {
	for _, poolId := range pools {
		pool, err := s.App.CosmwasmPoolKeeper.GetPool(s.Ctx, poolId)
		s.Require().NoError(err)
		cwPool, ok := pool.(cwpooltypes.CosmWasmExtension)
		s.Require().True(ok)
		s.Require().EqualValues(codeId, cwPool.GetCodeId())
	}
}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v24UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v24", Height: v24UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: v24UpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(v24UpgradeHeight)
}
