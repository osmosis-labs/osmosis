package poolmanager_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestGetSetTakerFeeTrackerForStakersAndCommunityPool() {
	tests := map[string]struct {
		firstTakerFeeForStakers        sdk.Coins
		secondTakerFeeForStakers       sdk.Coins
		firstTakerFeeForCommunityPool  sdk.Coins
		secondTakerFeeForCommunityPool sdk.Coins
	}{
		"happy path: replace single coin with increased single coin": {
			firstTakerFeeForStakers:  sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100))),
			secondTakerFeeForStakers: sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(200))),

			firstTakerFeeForCommunityPool:  sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(100))),
			secondTakerFeeForCommunityPool: sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(200))),
		},
		"replace single coin with decreased single coin": {
			firstTakerFeeForStakers:  sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100))),
			secondTakerFeeForStakers: sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(50))),

			firstTakerFeeForCommunityPool:  sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(100))),
			secondTakerFeeForCommunityPool: sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(50))),
		},
		"replace single coin with different denom": {
			firstTakerFeeForStakers:  sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100))),
			secondTakerFeeForStakers: sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(100))),

			firstTakerFeeForCommunityPool:  sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(100))),
			secondTakerFeeForCommunityPool: sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100))),
		},
		"replace single coin with multiple coins": {
			firstTakerFeeForStakers:  sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100))),
			secondTakerFeeForStakers: sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100)), sdk.NewCoin("usdc", sdk.NewInt(200))),

			firstTakerFeeForCommunityPool:  sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(100))),
			secondTakerFeeForCommunityPool: sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(100)), sdk.NewCoin("eth", sdk.NewInt(200))),
		},
		"replace multiple coins with single coin": {
			firstTakerFeeForStakers:  sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100)), sdk.NewCoin("usdc", sdk.NewInt(200))),
			secondTakerFeeForStakers: sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(200))),

			firstTakerFeeForCommunityPool:  sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(100)), sdk.NewCoin("eth", sdk.NewInt(200))),
			secondTakerFeeForCommunityPool: sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(50))),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			s.Require().Empty(s.App.PoolManagerKeeper.GetTakerFeeTrackerStartHeight(s.Ctx))
			s.Require().Empty(s.App.PoolManagerKeeper.GetTakerFeeTrackerForStakers(s.Ctx))

			s.App.PoolManagerKeeper.SetTakerFeeTrackerForStakers(s.Ctx, tc.firstTakerFeeForStakers)
			s.App.PoolManagerKeeper.SetTakerFeeTrackerForCommunityPool(s.Ctx, tc.firstTakerFeeForCommunityPool)

			actualFirstTakerFeeForStakers := s.App.PoolManagerKeeper.GetTakerFeeTrackerForStakers(s.Ctx)
			actualFirstTakerFeeForCommunityPool := s.App.PoolManagerKeeper.GetTakerFeeTrackerForCommunityPool(s.Ctx)

			s.Require().Equal(tc.firstTakerFeeForStakers, actualFirstTakerFeeForStakers)
			s.Require().Equal(tc.firstTakerFeeForCommunityPool, actualFirstTakerFeeForCommunityPool)

			s.App.PoolManagerKeeper.SetTakerFeeTrackerForStakers(s.Ctx, tc.secondTakerFeeForStakers)
			s.App.PoolManagerKeeper.SetTakerFeeTrackerForCommunityPool(s.Ctx, tc.secondTakerFeeForCommunityPool)

			actualSecondTakerFeeForStakers := s.App.PoolManagerKeeper.GetTakerFeeTrackerForStakers(s.Ctx)
			actualSecondTakerFeeForCommunityPool := s.App.PoolManagerKeeper.GetTakerFeeTrackerForCommunityPool(s.Ctx)

			s.Require().Equal(tc.secondTakerFeeForStakers, actualSecondTakerFeeForStakers)
			s.Require().Equal(tc.secondTakerFeeForCommunityPool, actualSecondTakerFeeForCommunityPool)
		})
	}
}

func (s *KeeperTestSuite) TestGetSetTakerFeeTrackerStartHeight() {
	tests := map[string]struct {
		firstTakerFeeTrackerStartHeight  int64
		secondTakerFeeTrackerStartHeight int64
	}{
		"replace tracker height with a higher height": {
			firstTakerFeeTrackerStartHeight:  100,
			secondTakerFeeTrackerStartHeight: 5000,
		},
		"replace tracker height with a lower height": {
			firstTakerFeeTrackerStartHeight:  100,
			secondTakerFeeTrackerStartHeight: 50,
		},
		"replace tracker height back to zero": {
			firstTakerFeeTrackerStartHeight:  100,
			secondTakerFeeTrackerStartHeight: 0,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			s.Require().Empty(s.App.PoolManagerKeeper.GetTakerFeeTrackerStartHeight(s.Ctx))

			s.App.PoolManagerKeeper.SetTakerFeeTrackerStartHeight(s.Ctx, tc.firstTakerFeeTrackerStartHeight)
			actualFirstTakerFeeTrackerStartHeight := s.App.PoolManagerKeeper.GetTakerFeeTrackerStartHeight(s.Ctx)
			s.Require().Equal(tc.firstTakerFeeTrackerStartHeight, actualFirstTakerFeeTrackerStartHeight)

			s.App.PoolManagerKeeper.SetTakerFeeTrackerStartHeight(s.Ctx, tc.secondTakerFeeTrackerStartHeight)
			actualSecondTakerFeeTrackerStartHeight := s.App.PoolManagerKeeper.GetTakerFeeTrackerStartHeight(s.Ctx)
			s.Require().Equal(tc.secondTakerFeeTrackerStartHeight, actualSecondTakerFeeTrackerStartHeight)
		})
	}
}

func (s *KeeperTestSuite) TestIncreaseTakerFeeTrackerForStakersAndCommunityPool() {
	tests := map[string]struct {
		initialTakerFeeForStakers       sdk.Coins
		initialTakerFeeForCommunityPool sdk.Coins

		increaseTakerFeeForStakersBy       sdk.Coin
		increaseTakerFeeForCommunityPoolBy sdk.Coin
	}{
		"happy path: increase single denom tracker": {
			initialTakerFeeForStakers:       sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100))),
			initialTakerFeeForCommunityPool: sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(100))),

			increaseTakerFeeForStakersBy:       sdk.NewCoin("eth", sdk.NewInt(50)),
			increaseTakerFeeForCommunityPoolBy: sdk.NewCoin("usdc", sdk.NewInt(50)),
		},
		"increase multi denom tracker": {
			initialTakerFeeForStakers:       sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100)), sdk.NewCoin("usdc", sdk.NewInt(200))),
			initialTakerFeeForCommunityPool: sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100)), sdk.NewCoin("usdc", sdk.NewInt(200))),

			increaseTakerFeeForStakersBy:       sdk.NewCoin("eth", sdk.NewInt(50)),
			increaseTakerFeeForCommunityPoolBy: sdk.NewCoin("usdc", sdk.NewInt(50)),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			s.Require().Empty(s.App.PoolManagerKeeper.GetTakerFeeTrackerStartHeight(s.Ctx))
			s.Require().Empty(s.App.PoolManagerKeeper.GetTakerFeeTrackerForStakers(s.Ctx))

			s.App.PoolManagerKeeper.SetTakerFeeTrackerForStakers(s.Ctx, tc.initialTakerFeeForStakers)
			s.App.PoolManagerKeeper.SetTakerFeeTrackerForCommunityPool(s.Ctx, tc.initialTakerFeeForCommunityPool)

			actualInitialTakerFeeForStakers := s.App.PoolManagerKeeper.GetTakerFeeTrackerForStakers(s.Ctx)
			actualInitialTakerFeeForCommunityPool := s.App.PoolManagerKeeper.GetTakerFeeTrackerForCommunityPool(s.Ctx)

			s.Require().Equal(tc.initialTakerFeeForStakers, actualInitialTakerFeeForStakers)
			s.Require().Equal(tc.initialTakerFeeForCommunityPool, actualInitialTakerFeeForCommunityPool)

			s.App.PoolManagerKeeper.IncreaseTakerFeeTrackerForStakers(s.Ctx, tc.increaseTakerFeeForStakersBy)
			s.App.PoolManagerKeeper.IncreaseTakerFeeTrackerForCommunityPool(s.Ctx, tc.increaseTakerFeeForCommunityPoolBy)

			takerFeeForStakersAfterIncrease := s.App.PoolManagerKeeper.GetTakerFeeTrackerForStakers(s.Ctx)
			takerFeeForCommunityPoolAfterIncrease := s.App.PoolManagerKeeper.GetTakerFeeTrackerForCommunityPool(s.Ctx)

			s.Require().Equal(tc.initialTakerFeeForStakers.Add(sdk.NewCoins(tc.increaseTakerFeeForStakersBy)...), takerFeeForStakersAfterIncrease)
			s.Require().Equal(tc.initialTakerFeeForCommunityPool.Add(sdk.NewCoins(tc.increaseTakerFeeForCommunityPoolBy)...), takerFeeForCommunityPoolAfterIncrease)
		})
	}
}
