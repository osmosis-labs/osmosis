package keeper_test

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	defaultLPDenom           string        = "lptoken"
	defaultLPSyntheticDenom  string        = "lptoken/superbonding"
	defaultLPTokens          sdk.Coins     = sdk.Coins{sdk.NewInt64Coin(defaultLPDenom, 10)}
	defaultLPSyntheticTokens sdk.Coins     = sdk.Coins{sdk.NewInt64Coin(defaultLPSyntheticDenom, 10)}
	defaultLiquidTokens      sdk.Coins     = sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}
	defaultLockDuration      time.Duration = time.Second
	oneLockupUser            userLocks     = userLocks{
		lockDurations: []time.Duration{time.Second},
		lockAmounts:   []sdk.Coins{defaultLPTokens},
	}
	twoLockupUser userLocks = userLocks{
		lockDurations: []time.Duration{defaultLockDuration, 2 * defaultLockDuration},
		lockAmounts:   []sdk.Coins{defaultLPTokens, defaultLPTokens},
	}
	oneSyntheticLockupUser userLocks = userLocks{
		lockDurations: []time.Duration{time.Second},
		lockAmounts:   []sdk.Coins{defaultLPSyntheticTokens},
	}
	twoSyntheticLockupUser userLocks = userLocks{
		lockDurations: []time.Duration{defaultLockDuration, 2 * defaultLockDuration},
		lockAmounts:   []sdk.Coins{defaultLPSyntheticTokens, defaultLPSyntheticTokens},
	}
	defaultRewardDenom string = "rewardDenom"
)

// TODO: Switch more code to use userLocks and perpGaugeDesc
// TODO: Create issue for the above.
type userLocks struct {
	lockDurations []time.Duration
	lockAmounts   []sdk.Coins
}

type perpGaugeDesc struct {
	lockDenom    string
	lockDuration time.Duration
	rewardAmount sdk.Coins
}

// setupAddr takes a balance, prefix, and address number. Then returns the respective account address byte array.
// If prefix is left blank, it will be replaced with a random prefix.
func (suite *KeeperTestSuite) setupAddr(addrNum int, prefix string, balance sdk.Coins) sdk.AccAddress {
	if prefix == "" {
		prefixBz := make([]byte, 8)
		_, _ = rand.Read(prefixBz)
		prefix = string(prefixBz)
	} else {
		prefix = fmt.Sprintf("%8.8s", prefix)
	}

	addr := sdk.AccAddress([]byte(fmt.Sprintf("addr%s%8d", prefix, addrNum)))
	suite.FundAcc(addr, balance)
	return addr
}

// SetupUserLocks takes an array of user locks, creates locks based on this array, then returns the respective account address byte array.
func (suite *KeeperTestSuite) SetupUserLocks(users []userLocks) (accs []sdk.AccAddress) {
	accs = make([]sdk.AccAddress, len(users))
	for i, user := range users {
		suite.Assert().Equal(len(user.lockDurations), len(user.lockAmounts))
		totalLockAmt := user.lockAmounts[0]
		for j := 1; j < len(user.lockAmounts); j++ {
			totalLockAmt = totalLockAmt.Add(user.lockAmounts[j]...)
		}
		accs[i] = suite.setupAddr(i, "", totalLockAmt)
		for j := 0; j < len(user.lockAmounts); j++ {
			_, err := suite.App.LockupKeeper.CreateLock(
				suite.Ctx, accs[i], user.lockAmounts[j], user.lockDurations[j])
			suite.Require().NoError(err)
		}
	}
	return
}

// SetupUserSyntheticLocks takes an array of user locks creates synthetic locks based on this array, then returns the respective account address byte array.
func (suite *KeeperTestSuite) SetupUserSyntheticLocks(users []userLocks) (accs []sdk.AccAddress) {
	accs = make([]sdk.AccAddress, len(users))
	coins := sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}
	var lockupID uint64
	lockupID = 1
	for i, user := range users {
		suite.Assert().Equal(len(user.lockDurations), len(user.lockAmounts))
		totalLockAmt := user.lockAmounts[0]
		for j := 1; j < len(user.lockAmounts); j++ {
			totalLockAmt = totalLockAmt.Add(user.lockAmounts[j]...)
		}
		accs[i] = suite.setupAddr(i, "", totalLockAmt)
		for j := 0; j < len(user.lockAmounts); j++ {
			suite.LockTokens(accs[i], coins, user.lockDurations[j])
			err := suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, lockupID, "lptoken/superbonding", user.lockDurations[j], false)
			lockupID++
			suite.Require().NoError(err)
		}
	}
	return
}

// SetupGauges takes an array of perpGaugeDesc structs. Then returns the corresponding array of Gauge structs.
func (suite *KeeperTestSuite) SetupGauges(gaugeDescriptors []perpGaugeDesc, denom string) []types.Gauge {
	gauges := make([]types.Gauge, len(gaugeDescriptors))
	perpetual := true
	for i, desc := range gaugeDescriptors {
		_, gaugePtr, _, _ := suite.setupNewGaugeWithDuration(perpetual, desc.rewardAmount, desc.lockDuration, denom)
		gauges[i] = *gaugePtr
	}
	return gauges
}

// CreateGauge creates a gauge struct given the required params.
func (suite *KeeperTestSuite) CreateGauge(isPerpetual bool, addr sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpoch uint64) (uint64, *types.Gauge) {
	suite.FundAcc(addr, coins)
	gaugeID, err := suite.App.IncentivesKeeper.CreateGauge(suite.Ctx, isPerpetual, addr, coins, distrTo, startTime, numEpoch)
	suite.Require().NoError(err)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	return gaugeID, gauge
}

// AddToGauge adds coins to the specified gauge.
func (suite *KeeperTestSuite) AddToGauge(coins sdk.Coins, gaugeID uint64) uint64 {
	addr := sdk.AccAddress([]byte("addrx---------------"))
	suite.FundAcc(addr, coins)
	err := suite.App.IncentivesKeeper.AddToGaugeRewards(suite.Ctx, addr, coins, gaugeID)
	suite.Require().NoError(err)
	return gaugeID
}

// LockTokens locks tokens for the specified duration
func (suite *KeeperTestSuite) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) {
	suite.FundAcc(addr, coins)
	_, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, addr, coins, duration)
	suite.Require().NoError(err)
}

// setupNewGaugeWithDuration creates a gauge with the specified duration.
func (suite *KeeperTestSuite) setupNewGaugeWithDuration(isPerpetual bool, coins sdk.Coins, duration time.Duration, denom string) (
	uint64, *types.Gauge, sdk.Coins, time.Time,
) {
	addr := sdk.AccAddress([]byte("Gauge_Creation_Addr_"))
	startTime2 := time.Now()
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         denom,
		Duration:      duration,
	}

	// mints coins so supply exists on chain
	mintCoins := sdk.Coins{sdk.NewInt64Coin(distrTo.Denom, 200)}
	suite.FundAcc(addr, mintCoins)

	numEpochsPaidOver := uint64(2)
	if isPerpetual {
		numEpochsPaidOver = uint64(1)
	}
	gaugeID, gauge := suite.CreateGauge(isPerpetual, addr, coins, distrTo, startTime2, numEpochsPaidOver)
	return gaugeID, gauge, coins, startTime2
}

// TODO: Delete all usages of this method
// TODO: Why delete all usages of this method?

// SetupNewGauge creates a gauge with the default lock duration.
func (suite *KeeperTestSuite) SetupNewGauge(isPerpetual bool, coins sdk.Coins) (uint64, *types.Gauge, sdk.Coins, time.Time) {
	return suite.setupNewGaugeWithDuration(isPerpetual, coins, defaultLockDuration, "lptoken")
}

// setupNewGaugeWithDenom creates a gauge with the specified duration and denom.
func (suite *KeeperTestSuite) setupNewGaugeWithDenom(isPerpetual bool, coins sdk.Coins, duration time.Duration, denom string) (
	uint64, *types.Gauge, sdk.Coins, time.Time,
) {
	addr := sdk.AccAddress([]byte("Gauge_Creation_Addr_"))
	startTime2 := time.Now()
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         denom,
		Duration:      duration,
	}

	// mints coins so supply exists on chain
	mintCoins := sdk.Coins{sdk.NewInt64Coin(distrTo.Denom, 200)}
	suite.FundAcc(addr, mintCoins)

	numEpochsPaidOver := uint64(2)
	if isPerpetual {
		numEpochsPaidOver = uint64(1)
	}
	gaugeID, gauge := suite.CreateGauge(isPerpetual, addr, coins, distrTo, startTime2, numEpochsPaidOver)
	return gaugeID, gauge, coins, startTime2
}

// SetupNewGaugeWithDenom creates a gauge with the specified duration and denom.
func (suite *KeeperTestSuite) SetupNewGaugeWithDenom(isPerpetual bool, coins sdk.Coins, denom string) (uint64, *types.Gauge, sdk.Coins, time.Time) {
	return suite.setupNewGaugeWithDenom(isPerpetual, coins, defaultLockDuration, denom)
}

// SetupManyLocks creates as many locks as the user defines.
func (suite *KeeperTestSuite) SetupManyLocks(numLocks int, liquidBalance sdk.Coins, coinsPerLock sdk.Coins,
	lockDuration time.Duration,
) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, 0, numLocks)
	randPrefix := make([]byte, 8)
	_, _ = rand.Read(randPrefix)

	bal := liquidBalance.Add(coinsPerLock...)
	for i := 0; i < numLocks; i++ {
		addr := suite.setupAddr(i, string(randPrefix), bal)
		_, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, addr, coinsPerLock, lockDuration)
		suite.Require().NoError(err)
		addrs = append(addrs, addr)
	}
	return addrs
}

// SetupLockAndGauge creates both a lock and a gauge.
func (suite *KeeperTestSuite) SetupLockAndGauge(isPerpetual bool) (sdk.AccAddress, uint64, sdk.Coins, time.Time) {
	// create a gauge and locks
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	suite.LockTokens(lockOwner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)

	// create gauge
	gaugeID, _, gaugeCoins, startTime := suite.SetupNewGauge(isPerpetual, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	return lockOwner, gaugeID, gaugeCoins, startTime
}
