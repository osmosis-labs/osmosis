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

func (suite *KeeperTestSuite) measureLockGas(addr sdk.AccAddress, coins sdk.Coins, dur time.Duration) uint64 {
	// fundAccount outside of gas measurement
	suite.FundAcc(addr, coins)
	// start measuring gas
	alreadySpent := suite.Ctx.GasMeter().GasConsumed()
	_, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, addr, coins, dur)
	suite.Require().NoError(err)
	newSpent := suite.Ctx.GasMeter().GasConsumed()
	spentNow := newSpent - alreadySpent
	return spentNow
}

func (suite *KeeperTestSuite) measureAvgAndMaxLockGas(
	numIterations int,
	addr sdk.AccAddress,
	coinsFn func(int) sdk.Coins,
	durFn func(int) time.Duration,
) (avg uint64, maxGas uint64) {
	runningTotal := uint64(0)
	maxGas = uint64(0)
	for i := 1; i <= numIterations; i++ {
		lockGas := suite.measureLockGas(addr, coinsFn(i), durFn(i))
		runningTotal += lockGas
		if lockGas > maxGas {
			maxGas = lockGas
			// fmt.Println(suite.Ctx.GasMeter().String())
		}
	}
	avg = runningTotal / uint64(numIterations)
	return avg, maxGas
}

// This maintains hard coded gas test vector changes,
// so we can easily track changes
func (suite *KeeperTestSuite) TestRepeatedLockTokensGas() {
	suite.SetupTest()

	coinsFn := func(int) sdk.Coins { return defaultCoins }
	durFn := func(int) time.Duration { return time.Second }
	startAveragingAt := 1000
	totalNumLocks := 10000

	firstLockGasAmount := suite.measureLockGas(defaultAddr, defaultCoins, time.Second)
	suite.Assert().LessOrEqual(int(firstLockGasAmount), 100000)

	for i := 1; i < startAveragingAt; i++ {
		suite.LockTokens(defaultAddr, defaultCoins, time.Second)
	}
	avgGas, maxGas := suite.measureAvgAndMaxLockGas(totalNumLocks-startAveragingAt, defaultAddr, coinsFn, durFn)
	fmt.Printf("test deets: total locks created %d, begin average at %d\n", totalNumLocks, startAveragingAt)
	suite.Assert().LessOrEqual(int(avgGas), 100000, "average gas / lock")
	suite.Assert().LessOrEqual(int(maxGas), 100000, "max gas / lock")
}

func (suite *KeeperTestSuite) TestRepeatedLockTokensDistinctDurationGas() {
	suite.SetupTest()

	coinsFn := func(int) sdk.Coins { return defaultCoins }
	durFn := func(i int) time.Duration { return time.Duration(i+1) * time.Second }
	totalNumLocks := 10000

	avgGas, maxGas := suite.measureAvgAndMaxLockGas(totalNumLocks, defaultAddr, coinsFn, durFn)
	fmt.Printf("test deets: total locks created %d\n", totalNumLocks)
	suite.Assert().LessOrEqual(int(avgGas), 150000, "average gas / lock")
	suite.Assert().LessOrEqual(int(maxGas), 300000, "max gas / lock")
}
