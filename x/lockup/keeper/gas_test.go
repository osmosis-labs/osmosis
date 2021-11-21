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
	alreadySpent := suite.ctx.GasMeter().GasConsumed()
	suite.LockTokens(addr, coins, dur)
	newSpent := suite.ctx.GasMeter().GasConsumed()
	spentNow := newSpent - alreadySpent
	return spentNow
}

func (suite *KeeperTestSuite) measureAvgAndMaxLockGas(
	numIterations int,
	addr sdk.AccAddress,
	coinsFn func(int) sdk.Coins,
	durFn func(int) time.Duration) (avg uint64, maxGas uint64) {
	runningTotal := uint64(0)
	maxGas = uint64(0)
	for i := 1; i <= numIterations; i++ {
		lockGas := suite.measureLockGas(addr, coinsFn(i), durFn(i))
		runningTotal += lockGas
		if lockGas > maxGas {
			maxGas = lockGas
			// fmt.Println(suite.ctx.GasMeter().String())
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
	suite.Assert().Equal(93703, int(firstLockGasAmount))

	for i := 1; i < startAveragingAt; i++ {
		suite.LockTokens(defaultAddr, defaultCoins, time.Second)
	}
	avgGas, maxGas := suite.measureAvgAndMaxLockGas(totalNumLocks-startAveragingAt, defaultAddr, coinsFn, durFn)
	fmt.Printf("test deets: total locks created %d, begin average at %d\n", totalNumLocks, startAveragingAt)
	suite.Assert().Equal(75618, int(avgGas), "average gas / lock")
	suite.Assert().Equal(75708, int(maxGas), "max gas / lock")
}

func (suite *KeeperTestSuite) TestRepeatedLockTokensDistinctDurationGas() {
	suite.SetupTest()

	coinsFn := func(int) sdk.Coins { return defaultCoins }
	durFn := func(i int) time.Duration { return time.Duration(i+1) * time.Second }
	totalNumLocks := 10000

	avgGas, maxGas := suite.measureAvgAndMaxLockGas(totalNumLocks, defaultAddr, coinsFn, durFn)
	fmt.Printf("test deets: total locks created %d\n", totalNumLocks)
	suite.Assert().EqualValues(122316, int(avgGas), "average gas / lock")
	suite.Assert().EqualValues(242903, int(maxGas), "max gas / lock")
}
