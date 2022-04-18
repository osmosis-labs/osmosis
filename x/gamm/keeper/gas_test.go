package keeper_test

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	balanacertypes "github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

var (
	defaultAddr       sdk.AccAddress = sdk.AccAddress([]byte("addr1---------------"))
	defaultCoins      sdk.Coins      = sdk.Coins{}
	minShareOutAmount sdk.Int        = types.OneShare.MulRaw(50)
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
	poolId := suite.prepareBalancerPool()

	poolIDFn := func(int) uint64 { return poolId }
	minShareOutAmountFn := func(int) sdk.Int { return minShareOutAmount }
	maxCoinsFn := func(int) sdk.Coins { return defaultCoins }
	startAveragingAt := 1000
	totalNumJoins := 10000

	// mint some assets to the accounts
	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, defaultAddr, sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000000)),
		sdk.NewCoin("foo", sdk.NewInt(10000000000000000)),
		sdk.NewCoin("bar", sdk.NewInt(10000000000000000)),
		sdk.NewCoin("baz", sdk.NewInt(10000000000000000)),
	))
	suite.Require().NoError(err)

	firstJoinGas := suite.measureJoinPoolGas(defaultAddr, poolId, minShareOutAmount, defaultCoins)
	suite.Assert().LessOrEqual(int(firstJoinGas), 100000)

	for i := 1; i < startAveragingAt; i++ {
		err := suite.app.GAMMKeeper.JoinPool(suite.ctx, defaultAddr, poolId, minShareOutAmount, sdk.Coins{})
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
	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, defaultAddr, sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(1000000000000000000)),
	))
	suite.Require().NoError(err)

	// number of distinct denom to test
	denomNumber := 1000

	// create pools prior to testing JoinPool using distinct denom
	coins := sdk.NewCoins(
		sdk.NewCoin("randToken1", sdk.NewInt(100)),
	)
	err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, defaultAddr, coins)
	suite.Require().NoError(err)

	defaultPoolParams := balanacertypes.PoolParams{
		SwapFee: sdk.NewDec(0),
		ExitFee: sdk.NewDec(0),
	}
	for i := 1; i <= denomNumber; i++ {
		randToken := "randToken" + strconv.Itoa(i+1)
		prevRandToken := "randToken" + strconv.Itoa(i)
		coins := sdk.NewCoins(sdk.NewCoin(randToken, sdk.NewInt(100)))

		err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, defaultAddr, coins)
		suite.Require().NoError(err)

		poolAssets := []types.PoolAsset{
			{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin(prevRandToken, sdk.NewInt(10)),
			},
			{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin(randToken, sdk.NewInt(10)),
			},
		}

		_, err = suite.app.GAMMKeeper.CreateBalancerPool(suite.ctx, defaultAddr, defaultPoolParams, poolAssets, "")
		suite.Require().NoError(err)
	}

	// test gas increase when JoinPool repeat
	initialPoolId := uint64(1)
	firstJoinGas := suite.measureJoinPoolGas(defaultAddr, initialPoolId, minShareOutAmount, defaultCoins)

	for i := 2; i < denomNumber; i++ {
		err := suite.app.GAMMKeeper.JoinPool(suite.ctx, defaultAddr, uint64(i), minShareOutAmount, sdk.Coins{})
		suite.Require().NoError(err)
	}

	lastPoolId := uint64(denomNumber)
	lastJoinGas := suite.measureJoinPoolGas(defaultAddr, lastPoolId, minShareOutAmount, defaultCoins)

	gasIncrease := lastJoinGas - firstJoinGas
	suite.Require().LessOrEqual(gasIncrease, uint64(5000))
}
