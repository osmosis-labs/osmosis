package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	defaultAddr  sdk.AccAddress = sdk.AccAddress([]byte("addr1---------------"))
	defaultCoins sdk.Coins      = sdk.Coins{sdk.NewInt64Coin("stake", 10)}
)

func (s *KeeperTestSuite) measureLockGas(addr sdk.AccAddress, coins sdk.Coins, dur time.Duration) uint64 {
	// fundAccount outside of gas measurement. TODO: Do one batch fundAcc.
	s.FundAcc(addr, coins)
	// start measuring gas
	alreadySpent := s.Ctx.GasMeter().GasConsumed()
	_, err := s.App.LockupKeeper.CreateLock(s.Ctx, addr, coins, dur)
	s.Require().NoError(err)
	newSpent := s.Ctx.GasMeter().GasConsumed()
	spentNow := newSpent - alreadySpent
	return spentNow
}

func (s *KeeperTestSuite) measureAvgAndMaxLockGas(
	numIterations int,
	addr sdk.AccAddress,
	coinsFn func(int) sdk.Coins,
	durFn func(int) time.Duration,
) (avg uint64, maxGas uint64) {
	runningTotal := uint64(0)
	maxGas = uint64(0)
	for i := 1; i <= numIterations; i++ {
		lockGas := s.measureLockGas(addr, coinsFn(i), durFn(i))
		runningTotal += lockGas
		if lockGas > maxGas {
			maxGas = lockGas
			// fmt.Println(s.Ctx.GasMeter().String())
		}
	}
	avg = runningTotal / uint64(numIterations)
	return avg, maxGas
}

// This maintains hard coded gas test vector changes,
// so we can easily track changes
func (s *KeeperTestSuite) TestRepeatedLockTokensGas() {
	s.SetupTest()

	coinsFn := func(int) sdk.Coins { return defaultCoins }
	durFn := func(int) time.Duration { return time.Second }
	startAveragingAt := 500
	totalNumLocks := 4000

	firstLockGasAmount := s.measureLockGas(defaultAddr, defaultCoins, time.Second)
	s.Assert().LessOrEqual(int(firstLockGasAmount), 100000)

	for i := 1; i < startAveragingAt; i++ {
		s.LockTokens(defaultAddr, defaultCoins, time.Second)
	}
	avgGas, maxGas := s.measureAvgAndMaxLockGas(totalNumLocks-startAveragingAt, defaultAddr, coinsFn, durFn)
	fmt.Printf("test deets: total locks created %d, begin average at %d\n", totalNumLocks, startAveragingAt)
	s.Assert().LessOrEqual(int(avgGas), 60000, "average gas / lock")
	s.Assert().LessOrEqual(int(maxGas), 100000, "max gas / lock")
}

func (s *KeeperTestSuite) TestRepeatedLockTokensDistinctDurationGas() {
	s.SetupTest()

	coinsFn := func(int) sdk.Coins { return defaultCoins }
	durFn := func(i int) time.Duration { return time.Duration(i+1) * time.Second }
	totalNumLocks := 4000

	avgGas, maxGas := s.measureAvgAndMaxLockGas(totalNumLocks, defaultAddr, coinsFn, durFn)
	fmt.Printf("test deets: total locks created %d\n", totalNumLocks)
	s.Assert().LessOrEqual(int(avgGas), 105000, "average gas / lock")
	s.Assert().LessOrEqual(int(maxGas), 250000, "max gas / lock")
}
