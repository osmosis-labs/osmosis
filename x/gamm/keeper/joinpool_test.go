package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

// func (suite *KeeperTestSuite) TestRandomizedJoinPoolExitPoolInvariants() {
// 	type testCase struct {
// 		initialTokensDenomIn  int64
// 		initialTokensDenomOut int64

// 		percentRatioMax int64

// 		joinDenomInAmt  int64
// 		joinDenomOutAmt int64
// 	}

// 	const (
// 		denomOut = "denomOut"
// 		denomIn  = "denomIn"
// 	)
// 	joinErrCount := 0
// 	exitErrCount := 0

// 	now := 1654939942
// 	rng := rand.NewSource(int64(now))
// 	fmt.Printf("Using random source of %d\n", now)

// 	// generate test case with randomized initial assets and join/exit ratio
// 	newCase := func() (tc *testCase) {
// 		tc = new(testCase)
// 		tc.initialTokensDenomIn = (rng.Int63() % 1_000_000) + 100
// 		tc.initialTokensDenomOut = (rng.Int63() % 1_000_000) + 100

// 		// 1%~200% of initial assets
// 		percentRatioIn := rng.Int63()%200 + 1
// 		usePerfectRatio := rng.Int63()%2 == 0
// 		tc.joinDenomInAmt = tc.initialTokensDenomIn * percentRatioIn / 100
// 		percentRatioOut := percentRatioIn
// 		if !usePerfectRatio {
// 			percentRatioOut = rng.Int63()%200 + 1
// 		}
// 		tc.joinDenomOutAmt = tc.initialTokensDenomOut * percentRatioOut / 100

// 		tc.percentRatioMax = percentRatioIn
// 		if percentRatioOut > percentRatioIn {
// 			tc.percentRatioMax = percentRatioOut
// 		}

// 		return tc
// 	}

// 	// create pool with randomized initial token amounts
// 	// and randomized ratio of join/exit
// 	createPool := func(tc *testCase) (poolId uint64) {
// 		poolAssetOut := balancer.PoolAsset{
// 			Token:  sdk.NewInt64Coin(denomOut, tc.initialTokensDenomOut),
// 			Weight: sdk.NewInt(5),
// 		}

// 		poolAssetIn := balancer.PoolAsset{
// 			Token:  sdk.NewInt64Coin(denomIn, tc.initialTokensDenomIn),
// 			Weight: sdk.NewInt(5),
// 		}

// 		return suite.PrepareBalancerPoolWithPoolAsset([]balancer.PoolAsset{poolAssetOut, poolAssetIn})
// 	}

// 	// joins with predetermined ratio
// 	joinPool := func(poolId uint64, tc *testCase) {
// 		sender := suite.TestAccs[0]
// 		tokensIn := sdk.Coins{
// 			sdk.NewInt64Coin(denomIn, tc.joinDenomInAmt),
// 			sdk.NewInt64Coin(denomOut, tc.joinDenomOutAmt),
// 		}
// 		suite.FundAcc(sender, tokensIn)

// 		initShares, _ := sdk.NewIntFromString("100000000000000000000")
// 		sharesWanted := initShares.Mul(sdk.NewInt(tc.percentRatioMax)).QuoRaw(100)
// 		err := suite.App.AppKeepers.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, sender, poolId, sharesWanted, tokensIn)
// 		if err != nil {
// 			joinErrCount += 1
// 		}
// 	}

// 	// exits for same amount of shares minted
// 	exitPool := func(poolId uint64, tc *testCase) {
// 		sender := suite.TestAccs[0]
// 		bal := suite.App.BankKeeper.GetBalance(suite.Ctx, sender, types.GetPoolShareDenom(poolId))
// 		_, err := suite.App.AppKeepers.GAMMKeeper.ExitPool(suite.Ctx, sender, poolId, bal.Amount, sdk.Coins{})
// 		if err != nil {
// 			exitErrCount += 1
// 		}
// 	}

// 	invariantJoinExitInversePreserve := func(
// 		beforeCoins, afterCoins sdk.Coins,
// 		beforeShares, afterShares sdk.Int, tc *testCase) {
// 		// test token amount has been preserved
// 		// fmt.Println("joinDenomInAmt")
// 		fmt.Printf("initialTokensDenomOut: %v. initialTokensDenomIn: %v ", tc.initialTokensDenomOut, tc.initialTokensDenomIn)
// 		fmt.Printf("joinDenomInAmt: %v. joinDenomOutAmt: %v, percentRatioMax: %v \n", tc.joinDenomInAmt, tc.joinDenomOutAmt, tc.percentRatioMax)
// 		suite.Require().True(
// 			!beforeCoins.IsAnyGT(afterCoins),
// 			"Coins has not been preserved before and after join-exit\nbefore:\t%s\nafter:\t%s",
// 			beforeCoins, afterCoins,
// 		)
// 		// test share amount has been preserved
// 		suite.Require().True(
// 			beforeShares.Equal(afterShares),
// 			"Shares has not been preserved before and after join-exit\nbefore:\t%s\nafter:\t%s",
// 			beforeShares, afterShares,
// 		)
// 	}

// 	testPoolInvariants := func() {
// 		tc := newCase()
// 		poolId := createPool(tc)
// 		origPool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
// 		suite.Require().NoError(err)
// 		originalCoins, originalShares := origPool.GetTotalPoolLiquidity(sdk.Context{}), origPool.GetTotalShares()
// 		joinPool(poolId, tc)
// 		exitPool(poolId, tc)
// 		finalPool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
// 		suite.Require().NoError(err)
// 		invariantJoinExitInversePreserve(
// 			originalCoins, finalPool.GetTotalPoolLiquidity(sdk.Context{}),
// 			originalShares, finalPool.GetTotalShares(), tc,
// 		)
// 	}

// 	for i := 0; i < 10000; i++ {
// 		testPoolInvariants()
// 	}
// }

func (suite *KeeperTestSuite) TestRandomizedJoinPoolExitPoolInvariants111() {
	type testCase struct {
		initialTokensDenomIn  int64
		initialTokensDenomOut int64

		percentRatioMax int64

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
	// rng := rand.NewSource(int64(now))
	fmt.Printf("Using random source of %d\n", now)

	// generate test case with randomized initial assets and join/exit ratio
	newCase := func() (tc *testCase) {
		tc = new(testCase)
		// tc.initialTokensDenomIn = (rng.Int63() % 1_000_000) + 100
		// tc.initialTokensDenomOut = (rng.Int63() % 1_000_000) + 100

		// // 1%~200% of initial assets
		// percentRatioIn := rng.Int63()%200 + 1
		// usePerfectRatio := rng.Int63()%2 == 0
		// tc.joinDenomInAmt = tc.initialTokensDenomIn * percentRatioIn / 100
		// percentRatioOut := percentRatioIn
		// if !usePerfectRatio {
		// 	percentRatioOut = rng.Int63()%200 + 1
		// }
		// tc.joinDenomOutAmt = tc.initialTokensDenomOut * percentRatioOut / 100

		// tc.percentRatioMax = percentRatioIn
		// if percentRatioOut > percentRatioIn {
		// 	tc.percentRatioMax = percentRatioOut
		// }

		tc.initialTokensDenomOut = 707791
		tc.initialTokensDenomIn = 935340
		tc.joinDenomInAmt = 1870680
		tc.joinDenomOutAmt = 1415582
		tc.percentRatioMax = 200

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
		sender := suite.TestAccs[0]
		tokensIn := sdk.Coins{
			sdk.NewInt64Coin(denomIn, tc.joinDenomInAmt),
			sdk.NewInt64Coin(denomOut, tc.joinDenomOutAmt),
		}
		suite.FundAcc(sender, tokensIn)

		initShares, _ := sdk.NewIntFromString("100000000000000000000")
		sharesWanted := initShares.Mul(sdk.NewInt(tc.percentRatioMax)).QuoRaw(100)
		err := suite.App.AppKeepers.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, sender, poolId, sharesWanted, tokensIn)
		if err != nil {
			joinErrCount += 1
		}
	}

	// exits for same amount of shares minted
	exitPool := func(poolId uint64, tc *testCase) {
		sender := suite.TestAccs[0]
		bal := suite.App.BankKeeper.GetBalance(suite.Ctx, sender, types.GetPoolShareDenom(poolId))
		_, err := suite.App.AppKeepers.GAMMKeeper.ExitPool(suite.Ctx, sender, poolId, bal.Amount, sdk.Coins{})
		if err != nil {
			exitErrCount += 1
		}
	}

	invariantJoinExitInversePreserve := func(
		beforeCoins, afterCoins sdk.Coins,
		beforeShares, afterShares sdk.Int, tc *testCase) {
		// test token amount has been preserved
		// fmt.Println("joinDenomInAmt")
		fmt.Printf("initialTokensDenomOut: %v. initialTokensDenomIn: %v\n", tc.initialTokensDenomOut, tc.initialTokensDenomIn)
		fmt.Printf("joinDenomInAmt: %v. joinDenomOutAmt: %v, percentRatioMax: %v", tc.joinDenomInAmt, tc.joinDenomOutAmt, tc.percentRatioMax)
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
}
