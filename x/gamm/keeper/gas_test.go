package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	defaultAddr       sdk.AccAddress = sdk.AccAddress([]byte("addr1---------------"))
	defaultCoins      sdk.Coins      = sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	minShareOutAmount sdk.Int        = sdk.OneInt()
)

func (suite *KeeperTestSuite) measureJoinPoolGas(
	addr sdk.AccAddress,
	poolID uint64,
	shareOutAmountMax sdk.Int, maxCoins sdk.Coins) uint64 {
	alreadySpent := suite.ctx.GasMeter().GasConsumed()
	err := suite.app.GAMMKeeper.JoinPool(suite.ctx, addr, poolID, shareOutAmountMax, maxCoins)
	suite.Require().NoError(err)
	newSpent := suite.ctx.GasMeter().GasConsumed()
	spentNow := newSpent - alreadySpent
	return spentNow
}

func (suite *KeeperTestSuite) measureAvgAndMaxJoinPoolGas(
	numIterations int,
	addr sdk.AccAddress,
	poolIDFn func(int) uint64,
	shareOutAmountMaxFn func(int) sdk.Int,
	maxCoinsFn func(int) sdk.Coins) (avg uint64, maxGas uint64) {
	runningTotal := uint64(0)
	maxGas = uint64(0)
	for i := 1; i <= numIterations; i++ {
		lockGas := suite.measureJoinPoolGas(addr, poolIDFn(i), shareOutAmountMaxFn(i), maxCoinsFn(i))
		runningTotal += lockGas
		if lockGas > maxGas {
			maxGas = lockGas
		}
	}
	avg = runningTotal / uint64(numIterations)
	return avg, maxGas
}

// This maintains hard coded gas test vector changes,
// so we can easily track changes
// func (suite *KeeperTestSuite) TestJoinPoolGas() {
// 	suite.SetupTest()

// 	poolIDFn := func(int) uint64 { return 1 }
// 	minShareOutAmountFn := func(int) sdk.Int { return minShareOutAmount }
// 	maxCoinsFn := func(int) sdk.Coins { return defaultCoins }
// 	totalNumJoins := 10000

// 	suite.preparePool()
// 	firstJoinGas := suite.measureJoinPoolGas(defaultAddr, defaultCoins, time.Second)
// 	suite.Assert().Equal(73686, int(firstJoinGas))

// 	avgGas, maxGas := suite.measureAvgAndMaxJoinPoolGas(totalNumJoins, defaultAddr, poolIDFn, minShareOutAmountFn, maxCoinsFn)
// 	fmt.Printf("test deets: total locks created %d, begin average at %d\n", totalNumLocks, startAveragingAt)
// 	suite.Assert().Equal(64028, int(avgGas), "average gas / lock")
// 	suite.Assert().Equal(64118, int(maxGas), "max gas / lock")
// }

// func (suite *KeeperTestSuite) TestJoinPoolDistinctDenomGas() {
// 	suite.SetupTest()

// 	coinsFn := func(int) sdk.Coins { return defaultCoins }
// 	durFn := func(i int) time.Duration { return time.Duration(i+1) * time.Second }
// 	totalNumLocks := 10000

// 	suite.LockTokens(defaultAddr, defaultCoins, time.Second)
// 	firstLockGasAmount := suite.ctx.GasMeter().GasConsumed()
// 	suite.Assert().Equal(firstLockGasAmount, uint64(73686))

// 	avgGas, maxGas := suite.measureAvgAndMaxLockGas(totalNumLocks, defaultAddr, coinsFn, durFn)
// 	fmt.Printf("test deets: total locks created %d\n", totalNumLocks)
// 	suite.Assert().EqualValues(110729, int(avgGas), "average gas / lock")
// 	suite.Assert().EqualValues(231313, int(maxGas), "max gas / lock")
// }
