package v24_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v24/app/apptesting"

	incentivestypes "github.com/osmosis-labs/osmosis/v24/x/incentives/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v24/x/protorev/types"
	twap "github.com/osmosis-labs/osmosis/v24/x/twap"
	"github.com/osmosis-labs/osmosis/v24/x/twap/types"
	twaptypes "github.com/osmosis-labs/osmosis/v24/x/twap/types"
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

	// TWAP records indexed by time should be untouched since endblocker hasn't run yet.
	twapRecords, err = osmoutils.GatherValuesFromStorePrefix(store, []byte(HistoricalTWAPTimeIndexPrefix), types.ParseTwapFromBz)
	s.Require().NoError(err)
	s.Require().Len(twapRecords, 2)

	// Run the end blocker
	s.App.EndBlocker(s.Ctx, abci.RequestEndBlock{})

	// Since the prune limit was 1, 1 TWAP record indexed by time should be completely removed, leaving one more.
	twapRecords, err = osmoutils.GatherValuesFromStorePrefix(store, []byte(HistoricalTWAPTimeIndexPrefix), types.ParseTwapFromBz)
	s.Require().NoError(err)
	s.Require().Len(twapRecords, 1)
	s.Require().Equal(twapRecord2, twapRecords[0])

	// TWAP records indexed by pool ID should be untouched.
	twapRecords, err = s.App.TwapKeeper.GetAllHistoricalPoolIndexedTWAPsForPoolId(s.Ctx, twapRecord1.PoolId)
	s.Require().NoError(err)
	s.Require().Len(twapRecords, 2)
	s.Require().Equal(twapRecord1, twapRecords[0])
	s.Require().Equal(twapRecord2, twapRecords[1])

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

	// Check that the new min value for distribution has been set
	params := s.App.IncentivesKeeper.GetParams(s.Ctx)
	s.Require().Equal(incentivestypes.DefaultMinValueForDistr, params.MinValueForDistribution)
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
