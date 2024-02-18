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
	"github.com/osmosis-labs/osmosis/v23/app/apptesting"

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
	s.Require().Len(twapRecords, 1)
	s.Require().Equal(twap, twapRecords[0])

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
