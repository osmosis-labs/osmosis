package keeper_test

import (
	"fmt"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

func (suite *KeeperTestSuite) CreateGauge(isPerpetual bool, addr sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpoch uint64) (uint64, *types.Gauge) {
	suite.app.BankKeeper.SetBalances(suite.ctx, addr, coins)
	gaugeID, err := suite.app.IncentivesKeeper.CreateGauge(suite.ctx, isPerpetual, addr, coins, distrTo, startTime, numEpoch)
	suite.Require().NoError(err)
	gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	return gaugeID, gauge
}

func (suite *KeeperTestSuite) AddToGauge(coins sdk.Coins, gaugeID uint64) uint64 {
	addr := sdk.AccAddress([]byte("addrx---------------"))
	suite.app.BankKeeper.SetBalances(suite.ctx, addr, coins)
	err := suite.app.IncentivesKeeper.AddToGaugeRewards(suite.ctx, addr, coins, gaugeID)
	suite.Require().NoError(err)
	return gaugeID
}

func (suite *KeeperTestSuite) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) {
	suite.app.BankKeeper.SetBalances(suite.ctx, addr, coins)
	_, err := suite.app.LockupKeeper.LockTokens(suite.ctx, addr, coins, duration)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) SetupNewGauge(isPerpetual bool, coins sdk.Coins) (uint64, *types.Gauge, sdk.Coins, time.Time) {
	addr2 := sdk.AccAddress([]byte("addr1---------------"))
	startTime2 := time.Now()
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	numEpochsPaidOver := uint64(2)
	if isPerpetual {
		numEpochsPaidOver = uint64(1)
	}
	gaugeID, gauge := suite.CreateGauge(isPerpetual, addr2, coins, distrTo, startTime2, numEpochsPaidOver)
	return gaugeID, gauge, coins, startTime2
}

func (suite *KeeperTestSuite) SetupManyLocks(numLocks int, liquidBalance sdk.Coins, coinsPerLock sdk.Coins,
	lockDuration time.Duration) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, 0, numLocks)
	randPrefix := make([]byte, 8)
	_, _ = rand.Read(randPrefix)

	bal := liquidBalance.Add(coinsPerLock...)
	for i := 0; i < numLocks; i++ {
		addr := sdk.AccAddress([]byte(fmt.Sprintf("addr%s%8d", string(randPrefix), i)))
		addrs = append(addrs, addr)
		err := suite.app.BankKeeper.SetBalances(suite.ctx, addr, bal)
		suite.Require().NoError(err)
		_, err = suite.app.LockupKeeper.LockTokens(suite.ctx, addr, coinsPerLock, lockDuration)
		suite.Require().NoError(err)
	}
	return addrs
}

func (suite *KeeperTestSuite) SetupLockAndGauge(isPerpetual bool) (sdk.AccAddress, uint64, sdk.Coins, time.Time) {
	// create a gauge and locks
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	suite.LockTokens(lockOwner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)

	// create gauge
	gaugeID, _, gaugeCoins, startTime := suite.SetupNewGauge(isPerpetual, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	return lockOwner, gaugeID, gaugeCoins, startTime
}
