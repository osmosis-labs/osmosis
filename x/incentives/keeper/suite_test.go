package keeper_test

import (
	"crypto/rand"
	"fmt"
	"time"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	defaultLPDenom                    string        = "lptoken"
	defaultLPSyntheticDenom           string        = "lptoken/superbonding"
	defaultLPTokens                   sdk.Coins     = sdk.Coins{sdk.NewInt64Coin(defaultLPDenom, 10)}
	defaultLPTokensDoubleAmt          sdk.Coins     = sdk.Coins{sdk.NewInt64Coin(defaultLPDenom, 20)}
	defaultLPSyntheticTokens          sdk.Coins     = sdk.Coins{sdk.NewInt64Coin(defaultLPSyntheticDenom, 10)}
	defaultLPSyntheticTokensDoubleAmt sdk.Coins     = sdk.Coins{sdk.NewInt64Coin(defaultLPSyntheticDenom, 20)}
	defaultLiquidTokens               sdk.Coins     = sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}
	defaultLockDuration               time.Duration = time.Second
	oneLockupUser                     userLocks     = userLocks{
		lockDurations: []time.Duration{time.Second},
		lockAmounts:   []sdk.Coins{defaultLPTokens},
	}
	twoLockupUser userLocks = userLocks{
		lockDurations: []time.Duration{defaultLockDuration, 2 * defaultLockDuration},
		lockAmounts:   []sdk.Coins{defaultLPTokens, defaultLPTokens},
	}
	twoLockupUserDoubleAmt userLocks = userLocks{
		lockDurations: []time.Duration{defaultLockDuration, 2 * defaultLockDuration},
		lockAmounts:   []sdk.Coins{defaultLPTokensDoubleAmt, defaultLPTokensDoubleAmt},
	}
	oneSyntheticLockupUser userLocks = userLocks{
		lockDurations: []time.Duration{time.Second},
		lockAmounts:   []sdk.Coins{defaultLPSyntheticTokens},
	}
	twoSyntheticLockupUser userLocks = userLocks{
		lockDurations: []time.Duration{defaultLockDuration, 2 * defaultLockDuration},
		lockAmounts:   []sdk.Coins{defaultLPSyntheticTokens, defaultLPSyntheticTokens},
	}
	twoSyntheticLockupUserDoubleAmt userLocks = userLocks{
		lockDurations: []time.Duration{defaultLockDuration, 2 * defaultLockDuration},
		lockAmounts:   []sdk.Coins{defaultLPSyntheticTokensDoubleAmt, defaultLPSyntheticTokensDoubleAmt},
	}
	defaultRewardDenom string = "rewardDenom"
	otherDenom         string = "someOtherDenom"
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

type changeRewardReceiver struct {
	newReceiverAccIndex int
	lockId              uint64
}

// setupAddr takes a balance, prefix, and address number. Then returns the respective account address byte array.
// If prefix is left blank, it will be replaced with a random prefix.
func (s *KeeperTestSuite) setupAddr(addrNum int, prefix string, balance sdk.Coins) sdk.AccAddress {
	if prefix == "" {
		prefixBz := make([]byte, 8)
		_, _ = rand.Read(prefixBz)
		prefix = string(prefixBz)
	}

	addr := sdk.AccAddress([]byte(fmt.Sprintf("addr%s%8d", prefix, addrNum)))
	s.FundAcc(addr, balance)
	return addr
}

// SetupUserLocks takes an array of user locks, creates locks based on this array, then returns the respective array of accounts.
func (s *KeeperTestSuite) SetupUserLocks(users []userLocks) (accs []sdk.AccAddress) {
	accs = make([]sdk.AccAddress, len(users))
	for i, user := range users {
		s.Assert().Equal(len(user.lockDurations), len(user.lockAmounts))
		totalLockAmt := user.lockAmounts[0]
		for j := 1; j < len(user.lockAmounts); j++ {
			totalLockAmt = totalLockAmt.Add(user.lockAmounts[j]...)
		}
		accs[i] = s.setupAddr(i, "", totalLockAmt)
		for j := 0; j < len(user.lockAmounts); j++ {
			_, err := s.App.LockupKeeper.CreateLock(
				s.Ctx, accs[i], user.lockAmounts[j], user.lockDurations[j])
			s.Require().NoError(err)
		}
	}
	return
}

func (s *KeeperTestSuite) SetupChangeRewardReceiver(changeRewardReceivers []changeRewardReceiver, accs []sdk.AccAddress) {
	for _, changeRewardReceiver := range changeRewardReceivers {
		lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, changeRewardReceiver.lockId)
		s.Require().NoError(err)

		err = s.App.LockupKeeper.SetLockRewardReceiverAddress(s.Ctx, changeRewardReceiver.lockId, lock.OwnerAddress(), accs[changeRewardReceiver.newReceiverAccIndex].String())
		s.Require().NoError(err)
	}
}

// SetupUserSyntheticLocks takes an array of user locks and creates synthetic locks based on this array, then returns the respective account address byte array.
func (s *KeeperTestSuite) SetupUserSyntheticLocks(users []userLocks) (accs []sdk.AccAddress) {
	accs = make([]sdk.AccAddress, len(users))
	lockupID := uint64(1)
	for i, user := range users {
		s.Assert().Equal(len(user.lockDurations), len(user.lockAmounts))
		totalLockAmt := user.lockAmounts[0]
		for j := 1; j < len(user.lockAmounts); j++ {
			totalLockAmt = totalLockAmt.Add(user.lockAmounts[j]...)
		}
		accs[i] = s.setupAddr(i, "", totalLockAmt)
		for j := 0; j < len(user.lockAmounts); j++ {
			coins := sdk.Coins{sdk.NewInt64Coin("lptoken", user.lockAmounts[j][0].Amount.Int64())}
			s.LockTokens(accs[i], coins, user.lockDurations[j])
			err := s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, lockupID, "lptoken/superbonding", user.lockDurations[j], false)
			lockupID++
			s.Require().NoError(err)
		}
	}
	return
}

// SetupGauges takes an array of perpGaugeDesc structs. Then returns the corresponding array of Gauge structs.
func (s *KeeperTestSuite) SetupGauges(gaugeDescriptors []perpGaugeDesc, denom string) []types.Gauge {
	gauges := make([]types.Gauge, len(gaugeDescriptors))
	perpetual := true
	for i, desc := range gaugeDescriptors {
		_, gaugePtr, _, _ := s.setupNewGaugeWithDuration(perpetual, desc.rewardAmount, desc.lockDuration, denom)
		gauges[i] = *gaugePtr
	}
	return gauges
}

// CreateGauge creates a gauge struct given the required params.
func (s *KeeperTestSuite) CreateGauge(isPerpetual bool, addr sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpoch uint64) (uint64, *types.Gauge) {
	// Since this test creates or adds to a gauge, we need to ensure a route exists in protorev hot routes.
	// The pool doesn't need to actually exist for this test, so we can just ensure the denom pair has some entry.
	for _, coin := range coins {
		s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, coin.Denom, 9999)
	}

	s.FundAcc(addr, coins)
	gaugeID, err := s.App.IncentivesKeeper.CreateGauge(s.Ctx, isPerpetual, addr, coins, distrTo, startTime, numEpoch, 0)
	s.Require().NoError(err)
	gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID)
	s.Require().NoError(err)
	return gaugeID, gauge
}

// AddToGauge adds coins to the specified gauge.
func (s *KeeperTestSuite) AddToGauge(coins sdk.Coins, gaugeID uint64) uint64 {
	// Since this test creates or adds to a gauge, we need to ensure a route exists in protorev hot routes.
	// The pool doesn't need to actually exist for this test, so we can just ensure the denom pair has some entry.
	for _, coin := range coins {
		s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, coin.Denom, 9999)
	}

	addr := sdk.AccAddress([]byte("addrx---------------"))
	s.FundAcc(addr, coins)
	err := s.App.IncentivesKeeper.AddToGaugeRewards(s.Ctx, addr, coins, gaugeID)
	s.Require().NoError(err)
	return gaugeID
}

// LockTokens locks tokens for the specified duration
func (s *KeeperTestSuite) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) {
	s.FundAcc(addr, coins)
	_, err := s.App.LockupKeeper.CreateLock(s.Ctx, addr, coins, duration)
	s.Require().NoError(err)
}

// setupNewGaugeWithDuration creates a gauge with the specified duration.
func (s *KeeperTestSuite) setupNewGaugeWithDuration(isPerpetual bool, coins sdk.Coins, duration time.Duration, denom string) (
	uint64, *types.Gauge, sdk.Coins, time.Time,
) {
	addr := sdk.AccAddress([]byte("Gauge_Creation_Addr_"))
	startTime2 := time.Now()
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         denom,
		Duration:      duration,
	}

	// Since this test creates or adds to a gauge, we need to ensure a route exists in protorev hot routes.
	// The pool doesn't need to actually exist for this test, so we can just ensure the denom pair has some entry.
	for _, coin := range coins {
		s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, coin.Denom, 9999)
	}

	// mints coins so supply exists on chain
	mintCoins := sdk.Coins{sdk.NewInt64Coin(distrTo.Denom, 200)}
	s.FundAcc(addr, mintCoins)

	numEpochsPaidOver := uint64(2)
	if isPerpetual {
		numEpochsPaidOver = uint64(1)
	}
	gaugeID, gauge := s.CreateGauge(isPerpetual, addr, coins, distrTo, startTime2, numEpochsPaidOver)
	return gaugeID, gauge, coins, startTime2
}

// SetupNewGauge creates a gauge with the default lock duration.
func (s *KeeperTestSuite) SetupNewGauge(isPerpetual bool, coins sdk.Coins) (uint64, *types.Gauge, sdk.Coins, time.Time) {
	return s.setupNewGaugeWithDuration(isPerpetual, coins, defaultLockDuration, "lptoken")
}

// setupNewGaugeWithDenom creates a gauge with the specified duration and denom.
func (s *KeeperTestSuite) setupNewGaugeWithDenom(isPerpetual bool, coins sdk.Coins, duration time.Duration, denom string) (
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
	s.FundAcc(addr, mintCoins)

	numEpochsPaidOver := uint64(2)
	if isPerpetual {
		numEpochsPaidOver = uint64(1)
	}
	gaugeID, gauge := s.CreateGauge(isPerpetual, addr, coins, distrTo, startTime2, numEpochsPaidOver)
	return gaugeID, gauge, coins, startTime2
}

// SetupNewGaugeWithDenom creates a gauge with the specified duration and denom.
func (s *KeeperTestSuite) SetupNewGaugeWithDenom(isPerpetual bool, coins sdk.Coins, denom string) (uint64, *types.Gauge, sdk.Coins, time.Time) {
	return s.setupNewGaugeWithDenom(isPerpetual, coins, defaultLockDuration, denom)
}

// SetupManyLocks creates as many locks as the user defines.
func (s *KeeperTestSuite) SetupManyLocks(numLocks int, liquidBalance sdk.Coins, coinsPerLock sdk.Coins,
	lockDuration time.Duration,
) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, 0, numLocks)
	randPrefix := make([]byte, 8)
	_, _ = rand.Read(randPrefix)

	bal := liquidBalance.Add(coinsPerLock...)
	for i := 0; i < numLocks; i++ {
		addr := s.setupAddr(i, string(randPrefix), bal)
		_, err := s.App.LockupKeeper.CreateLock(s.Ctx, addr, coinsPerLock, lockDuration)
		s.Require().NoError(err)
		addrs = append(addrs, addr)
	}
	return addrs
}

// SetupLockAndGauge creates both a lock and a gauge.
func (s *KeeperTestSuite) SetupLockAndGauge(isPerpetual bool) (sdk.AccAddress, uint64, sdk.Coins, time.Time) {
	// create a gauge and locks
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	s.LockTokens(lockOwner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)

	// create gauge
	gaugeID, _, gaugeCoins, startTime := s.SetupNewGauge(isPerpetual, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	return lockOwner, gaugeID, gaugeCoins, startTime
}
