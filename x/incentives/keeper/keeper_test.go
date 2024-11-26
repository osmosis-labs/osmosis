package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	incentivetypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	querier keeper.Querier
}

// SetupTest sets incentives parameters from the suite's context
func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.querier = keeper.NewQuerier(*s.App.IncentivesKeeper)
	lockableDurations := s.App.IncentivesKeeper.GetLockableDurations(s.Ctx)
	lockableDurations = append(lockableDurations, 2*time.Second)
	s.App.IncentivesKeeper.SetLockableDurations(s.Ctx, lockableDurations)
	s.App.IncentivesKeeper.SetParam(s.Ctx, incentivetypes.KeyMinValueForDistr, sdk.NewCoin("stake", osmomath.NewInt(1)))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupGroupGauge(clPoolId uint64, lockOwner sdk.AccAddress, numOfNoLockGauges uint64, numOfLockGauges uint64) []uint64 {
	internalGauges := s.setupNoLockInternalGauge(clPoolId, numOfNoLockGauges)

	for i := uint64(1); i <= numOfLockGauges; i++ {
		// setup lock
		s.LockTokens(lockOwner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Hour*7)

		// create gauge
		gaugeID, _, _, _ := s.SetupNewGauge(true, sdk.NewCoins())
		internalGauges = append(internalGauges, gaugeID)
	}

	return internalGauges
}

// setupNoLockInternalGauge create no lock perp internal gauges.
func (s *KeeperTestSuite) setupNoLockInternalGauge(poolId uint64, numberOfExistingGauges uint64) []uint64 {
	var internalGauges []uint64
	for i := uint64(1); i <= numberOfExistingGauges; i++ {
		internalGauge := s.CreateNoLockExternalGauges(poolId, sdk.NewCoins(), s.TestAccs[1], uint64(1))
		internalGauges = append(internalGauges, internalGauge)
	}

	return internalGauges
}

// ValidateDistributedGauge checks that the gauge is updated as expected after distribution
func (s *KeeperTestSuite) ValidateDistributedGauge(gaugeID uint64, expectedFilledEpoch uint64, expectedDistributions sdk.Coins) {
	// Check that filled epcohs is not updated
	gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)
	s.Require().Equal(expectedFilledEpoch, gauge.FilledEpochs)
	// Check that distributed coins is not updated

	s.Require().Equal(expectedDistributions, gauge.DistributedCoins)
}

// ValidateNotDistributedGauge checks that the gauge is not updated after distribution
func (s *KeeperTestSuite) ValidateNotDistributedGauge(gaugeID uint64) {
	s.ValidateDistributedGauge(gaugeID, 0, sdk.Coins(nil))
}

func (s *KeeperTestSuite) ValidateIncentiveRecord(poolId uint64, remainingCoin sdk.Coin, incentiveRecord cltypes.IncentiveRecord) {
	epochInfo := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx)
	distributedDecCoin := sdk.NewDecCoinFromCoin(remainingCoin)
	emissionRateForPoolClPool := distributedDecCoin.Amount.QuoTruncate(osmomath.NewDec(epochInfo.Duration.Milliseconds()).QuoInt(osmomath.NewInt(1000)))

	s.Require().Equal(poolId, incentiveRecord.PoolId)
	s.Require().Equal(emissionRateForPoolClPool, incentiveRecord.GetIncentiveRecordBody().EmissionRate)
	s.Require().Equal(types.DefaultConcentratedUptime, incentiveRecord.MinUptime)
	s.Require().Equal(distributedDecCoin, incentiveRecord.GetIncentiveRecordBody().RemainingCoin)
}
