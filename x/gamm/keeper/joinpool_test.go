package keeper_test

import (
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func (suite *KeeperTestSuite) TestRandomizedJoinPoolExitPoolInvariants() {
	type testCase struct {
		initialTokensDenomIn  int64
		initialTokensDenomOut int64

		percentRatioMin int64

		joinDenomInAmt  int64
		joinDenomOutAmt int64
	}

	const (
		denomOut = "denomOut"
		denomIn  = "denomIn"
	)
	joinErrCount := 0
	exitErrCount := 0

	now := 1654939942
	rng := rand.NewSource(int64(now))
	fmt.Printf("Using random source of %d\n", now)

	// generate test case with randomized initial assets and join/exit ratio
	newCase := func() (tc *testCase) {
		tc = new(testCase)
		tc.initialTokensDenomIn = (rng.Int63() % 1_000_000) + 100
		tc.initialTokensDenomOut = (rng.Int63() % 1_000_000) + 100

		// 1%~200% of initial assets
		percentRatioIn := rng.Int63()%200 + 1
		usePerfectRatio := rng.Int63()%2 == 0
		tc.joinDenomInAmt = tc.initialTokensDenomIn * percentRatioIn / 100
		percentRatioOut := percentRatioIn
		if !usePerfectRatio {
			percentRatioOut = rng.Int63()%200 + 1
		}
		tc.joinDenomOutAmt = tc.initialTokensDenomOut * percentRatioOut / 100

		tc.percentRatioMin = percentRatioIn
		if percentRatioOut < percentRatioIn {
			tc.percentRatioMin = percentRatioOut
		}

		return tc
	}

	// create pool with randomized initial token amounts
	// and randomized ratio of join/exit
	createPool := func(tc *testCase) (poolId uint64) {
		poolAssetOut := balancer.PoolAsset{
			Token:  sdk.NewInt64Coin(denomOut, tc.initialTokensDenomOut),
			Weight: sdk.NewInt(5),
		}

		poolAssetIn := balancer.PoolAsset{
			Token:  sdk.NewInt64Coin(denomIn, tc.initialTokensDenomIn),
			Weight: sdk.NewInt(5),
		}

		return suite.PrepareBalancerPoolWithPoolAsset([]balancer.PoolAsset{poolAssetOut, poolAssetIn})
	}

	// joins with predetermined ratio
	joinPool := func(poolId uint64, tc *testCase) {
		sender := suite.TestAccs[1]
		tokensIn := sdk.Coins{
			sdk.NewInt64Coin(denomIn, tc.joinDenomInAmt),
			sdk.NewInt64Coin(denomOut, tc.joinDenomOutAmt),
		}
		suite.FundAcc(sender, tokensIn)

		initShares, _ := sdk.NewIntFromString("100000000000000000000")
		sharesWanted := initShares.Mul(sdk.NewInt(tc.percentRatioMin)).QuoRaw(100)
		ctx, writeFn := suite.Ctx.CacheContext()
		err := suite.App.AppKeepers.GAMMKeeper.JoinPoolNoSwap(ctx, sender, poolId, sharesWanted, tokensIn)
		if err != nil {
			joinErrCount += 1
		} else {
			writeFn()
		}
	}

	// exits for same amount of shares minted
	exitPool := func(poolId uint64, tc *testCase) {
		sender := suite.TestAccs[1]
		bal := suite.App.BankKeeper.GetBalance(suite.Ctx, sender, types.GetPoolShareDenom(poolId))
		ctx, writeFn := suite.Ctx.CacheContext()
		_, err := suite.App.AppKeepers.GAMMKeeper.ExitPool(ctx, sender, poolId, bal.Amount, sdk.Coins{})
		if err != nil {
			exitErrCount += 1
		} else {
			writeFn()
		}
	}

	invariantJoinExitInversePreserve := func(
		beforeCoins, afterCoins sdk.Coins,
		beforeShares, afterShares sdk.Int, tc *testCase) {
		// test token amount has been preserved
		suite.Require().True(
			!beforeCoins.IsAnyGT(afterCoins),
			"Coins has not been preserved before and after join-exit\nbefore:\t%s\nafter:\t%s",
			beforeCoins, afterCoins,
		)
		// test share amount has been preserved
		suite.Require().True(
			beforeShares.Equal(afterShares),
			"Shares has not been preserved before and after join-exit\nbefore:\t%s\nafter:\t%s",
			beforeShares, afterShares,
		)
	}

	testPoolInvariants := func() {
		tc := newCase()
		poolId := createPool(tc)
		origPool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
		suite.Require().NoError(err)
		originalCoins, originalShares := origPool.GetTotalPoolLiquidity(sdk.Context{}), origPool.GetTotalShares()
		joinPool(poolId, tc)
		exitPool(poolId, tc)
		finalPool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
		suite.Require().NoError(err)

		invariantJoinExitInversePreserve(
			originalCoins, finalPool.GetTotalPoolLiquidity(sdk.Context{}),
			originalShares, finalPool.GetTotalShares(), tc,
		)
	}

	for i := 0; i < 10000; i++ {
		testPoolInvariants()
	}

	suite.Require().Less(joinErrCount, 100)
	suite.Require().Less(exitErrCount, 100)
}
