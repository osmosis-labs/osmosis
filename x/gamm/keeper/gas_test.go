package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/gamm/types"
)

var (
	defaultAddr  sdk.AccAddress = sdk.AccAddress([]byte("addr1---------------"))
	defaultCoins sdk.Coins      = sdk.Coins{}
	// defaultCoins      sdk.Coins      = sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	minShareOutAmount sdk.Int = types.OneShare.MulRaw(50)
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

// measureAvgAndMaxJoinPoolGas iterates JoinPool over designated amount of times
// to acquire average gas spent
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
func (suite *KeeperTestSuite) TestJoinPoolGas() {
	suite.SetupTest()
	poolId := suite.preparePool()

	poolIDFn := func(int) uint64 { return poolId }
	minShareOutAmountFn := func(int) sdk.Int { return minShareOutAmount }
	maxCoinsFn := func(int) sdk.Coins { return defaultCoins }
	startAveragingAt := 1000
	totalNumJoins := 10000

	err := suite.app.BankKeeper.AddCoins(suite.ctx, defaultAddr, sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000000)),
		sdk.NewCoin("foo", sdk.NewInt(10000000000000000)),
		sdk.NewCoin("bar", sdk.NewInt(10000000000000000)),
		sdk.NewCoin("baz", sdk.NewInt(10000000000000000)),
	))
	suite.Require().NoError(err)

	firstJoinGas := suite.measureJoinPoolGas(defaultAddr, poolId, types.OneShare.MulRaw(50), defaultCoins)
	suite.Assert().Equal(76608, int(firstJoinGas))

	for i := 1; i < startAveragingAt; i++ {
		err := suite.app.GAMMKeeper.JoinPool(suite.ctx, defaultAddr, poolId, types.OneShare.MulRaw(50), sdk.Coins{})
		suite.Require().NoError(err)
	}

	avgGas, maxGas := suite.measureAvgAndMaxJoinPoolGas(totalNumJoins, defaultAddr, poolIDFn, minShareOutAmountFn, maxCoinsFn)
	fmt.Printf("test deets: total %d of pools joined, begin average at %d\n", totalNumJoins, startAveragingAt)
	suite.Assert().Equal(71065, int(avgGas), "average gas / join pool")
	suite.Assert().Equal(71164, int(maxGas), "max gas / join pool")
}
