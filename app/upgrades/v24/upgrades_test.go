package v24_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v23/app/apptesting"

	concentratedtypes "github.com/osmosis-labs/osmosis/v23/x/concentrated-liquidity/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v23/x/protorev/types"
	"github.com/osmosis-labs/osmosis/v23/x/twap/types"
	twaptypes "github.com/osmosis-labs/osmosis/v23/x/twap/types"
)

const (
	v24UpgradeHeight              = int64(10)
	HistoricalTWAPTimeIndexPrefix = "historical_time_index"
	KeySeparator                  = "|"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()

	// TWAP Setup
	//

	// Manually set up TWAP records indexed by both pool ID and time.
	twapStoreKey := s.App.GetKey(twaptypes.ModuleName)
	store := s.Ctx.KVStore(twapStoreKey)
	twap := twaptypes.TwapRecord{
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
	poolIndexKey := types.FormatHistoricalPoolIndexTWAPKey(twap.PoolId, twap.Asset0Denom, twap.Asset1Denom, twap.Time)
	osmoutils.MustSet(store, poolIndexKey, &twap)

	// The time index key is a bit manual since we removed the old code that did this programmatically.
	var buffer bytes.Buffer
	timeS := osmoutils.FormatTimeString(twap.Time)
	fmt.Fprintf(&buffer, "%s%d%s%s%s%s%s%s", HistoricalTWAPTimeIndexPrefix, twap.PoolId, KeySeparator, twap.Asset0Denom, KeySeparator, twap.Asset1Denom, KeySeparator, timeS)
	timeIndexKey := buffer.Bytes()
	osmoutils.MustSet(store, timeIndexKey, &twap)

	// TWAP records indexed by time should exist
	twapRecords, err := osmoutils.GatherValuesFromStorePrefix(store, []byte(HistoricalTWAPTimeIndexPrefix), types.ParseTwapFromBz)
	s.Require().NoError(err)
	s.Require().Len(twapRecords, 1)
	s.Require().Equal(twap, twapRecords[0])

	// TWAP records indexed by pool ID should exist.
	twapRecords, err = s.App.TwapKeeper.GetAllHistoricalPoolIndexedTWAPsForPoolId(s.Ctx, twap.PoolId)
	s.Require().NoError(err)
	s.Require().Len(twapRecords, 1)
	s.Require().Equal(twap, twapRecords[0])

	// INCENTIVES Setup
	//

	concentratedPoolIDs := []uint64{}

	// Create two sets of all pools
	allPoolsOne := s.PrepareAllSupportedPools()
	allPoolsTwo := s.PrepareAllSupportedPools()

	concentratedPoolIDs = append(concentratedPoolIDs, allPoolsOne.ConcentratedPoolID)
	concentratedPoolIDs = append(concentratedPoolIDs, allPoolsTwo.ConcentratedPoolID)

	// Update authorized quote denoms
	concentratedParams := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
	concentratedParams.AuthorizedQuoteDenoms = append(concentratedParams.AuthorizedQuoteDenoms, apptesting.USDC)
	s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, concentratedParams)

	// Create two more concentrated pools with positions
	secondLastPoolID := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx)
	lastPoolID := secondLastPoolID + 1
	concentratedPoolIDs = append(concentratedPoolIDs, secondLastPoolID)
	concentratedPoolIDs = append(concentratedPoolIDs, lastPoolID)
	s.CreateConcentratedPoolsAndFullRangePosition([][]string{
		{"uion", "uosmo"},
		{apptesting.ETH, apptesting.USDC},
	})

	lastPoolPositionID := s.App.ConcentratedLiquidityKeeper.GetNextPositionId(s.Ctx) - 1

	// Create incentive record for last pool
	incentiveCoin := sdk.NewCoin("uosmo", sdk.NewInt(1000000))
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
	//oldMigrationList := concentratedtypes.MigratedIncentiveAccumulatorPoolIDs
	concentratedtypes.MigratedIncentiveAccumulatorPoolIDs = map[uint64]struct{}{}
	concentratedtypes.MigratedIncentiveAccumulatorPoolIDs[lastPoolID] = struct{}{}

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

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})
	})

	// TWAP Tests
	//

	// TWAP records indexed by time should be completely removed.
	twapRecords, err = osmoutils.GatherValuesFromStorePrefix(store, []byte(HistoricalTWAPTimeIndexPrefix), types.ParseTwapFromBz)
	s.Require().NoError(err)
	s.Require().Len(twapRecords, 0)

	// TWAP records indexed by pool ID should be untouched.
	twapRecords, err = s.App.TwapKeeper.GetAllHistoricalPoolIndexedTWAPsForPoolId(s.Ctx, twap.PoolId)
	s.Require().NoError(err)
	s.Require().Len(twapRecords, 3)
	s.Require().Equal(twap, twapRecords[2])

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
}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v24UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v24", Height: v24UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, exists := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().True(exists)

	s.Ctx = s.Ctx.WithBlockHeight(v24UpgradeHeight)
}
