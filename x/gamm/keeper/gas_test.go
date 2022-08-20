package keeper_test

import (
	"fmt"
	"strconv"

	"github.com/osmosis-labs/osmosis/v11/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v11/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v11/x/gamm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	defaultAddr       sdk.AccAddress = sdk.AccAddress([]byte("addr1---------------"))
	defaultCoins      sdk.Coins      = sdk.Coins{}
	minShareOutAmount sdk.Int        = types.OneShare.MulRaw(50)
)

func (suite *KeeperTestSuite) measureJoinPoolGas(
	addr sdk.AccAddress,
	poolID uint64,
	shareOutAmountMax sdk.Int, maxCoins sdk.Coins,
) uint64 {
	alreadySpent := suite.Ctx.GasMeter().GasConsumed()
	_, _, err := suite.App.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, addr, poolID, shareOutAmountMax, maxCoins)
	suite.Require().NoError(err)
	newSpent := suite.Ctx.GasMeter().GasConsumed()
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
	maxCoinsFn func(int) sdk.Coins,
) (avg uint64, maxGas uint64) {
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
	poolId := suite.PrepareBalancerPool()

	poolIDFn := func(int) uint64 { return poolId }
	minShareOutAmountFn := func(int) sdk.Int { return minShareOutAmount }
	maxCoinsFn := func(int) sdk.Coins { return defaultCoins }
	startAveragingAt := 1000
	totalNumJoins := 10000

	// mint some assets to the accounts
	suite.FundAcc(defaultAddr, sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000000)),
		sdk.NewCoin("foo", sdk.NewInt(10000000000000000)),
		sdk.NewCoin("bar", sdk.NewInt(10000000000000000)),
		sdk.NewCoin("baz", sdk.NewInt(10000000000000000)),
	))

	firstJoinGas := suite.measureJoinPoolGas(defaultAddr, poolId, minShareOutAmount, defaultCoins)
	suite.Assert().LessOrEqual(int(firstJoinGas), 100000)

	for i := 1; i < startAveragingAt; i++ {
		_, _, err := suite.App.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, defaultAddr, poolId, minShareOutAmount, sdk.Coins{})
		suite.Require().NoError(err)
	}

	avgGas, maxGas := suite.measureAvgAndMaxJoinPoolGas(totalNumJoins, defaultAddr, poolIDFn, minShareOutAmountFn, maxCoinsFn)
	fmt.Printf("test deets: total %d of pools joined, begin average at %d\n", totalNumJoins, startAveragingAt)
	suite.Assert().LessOrEqual(int(avgGas), 100000, "average gas / join pool")
	suite.Assert().LessOrEqual(int(maxGas), 100000, "max gas / join pool")
}

func (suite *KeeperTestSuite) TestRepeatedJoinPoolDistinctDenom() {
	suite.SetupTest()

	// mint some usomo to account
	suite.FundAcc(defaultAddr, sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(1000000000000000000)),
	))

	// number of distinct denom to test
	denomNumber := 1000

	// create pools prior to testing JoinPool using distinct denom
	coins := sdk.NewCoins(
		sdk.NewCoin("randToken1", sdk.NewInt(100)),
	)
	suite.FundAcc(defaultAddr, coins)
	defaultPoolParams := balancertypes.PoolParams{
		SwapFee: sdk.NewDec(0),
		ExitFee: sdk.NewDec(0),
	}
	for i := 1; i <= denomNumber; i++ {
		randToken := "randToken" + strconv.Itoa(i+1)
		prevRandToken := "randToken" + strconv.Itoa(i)
		coins := sdk.NewCoins(sdk.NewCoin(randToken, sdk.NewInt(100)))

		suite.FundAcc(defaultAddr, coins)

		poolAssets := []balancertypes.PoolAsset{
			{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin(prevRandToken, sdk.NewInt(10)),
			},
			{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin(randToken, sdk.NewInt(10)),
			},
		}
		msg := balancer.NewMsgCreateBalancerPool(defaultAddr, defaultPoolParams, poolAssets, "")
		_, err := suite.App.GAMMKeeper.CreatePool(suite.Ctx, msg)
		suite.Require().NoError(err)
	}

	// test gas increase when JoinPool repeat
	initialPoolId := uint64(1)
	firstJoinGas := suite.measureJoinPoolGas(defaultAddr, initialPoolId, minShareOutAmount, defaultCoins)

	for i := 2; i < denomNumber; i++ {
		_, _, err := suite.App.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, defaultAddr, uint64(i), minShareOutAmount, sdk.Coins{})
		suite.Require().NoError(err)
	}

	lastPoolId := uint64(denomNumber)
	lastJoinGas := suite.measureJoinPoolGas(defaultAddr, lastPoolId, minShareOutAmount, defaultCoins)

	gasIncrease := lastJoinGas - firstJoinGas
	suite.Require().LessOrEqual(gasIncrease, uint64(5000))
}
