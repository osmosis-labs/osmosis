package poolmanager_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
)

func (s *KeeperTestSuite) TestGetTakerFeeTrackerForStakersAndCommunityPool() {
	tests := map[string]struct {
		firstTakerFeeForStakers        []sdk.Coin
		secondTakerFeeForStakers       []sdk.Coin
		firstTakerFeeForCommunityPool  []sdk.Coin
		secondTakerFeeForCommunityPool []sdk.Coin
	}{
		"happy path: get updated coin with same denom coin coin": {
			firstTakerFeeForStakers:  sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(100))),
			secondTakerFeeForStakers: sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(200))),

			firstTakerFeeForCommunityPool:  sdk.NewCoins(sdk.NewCoin("usdc", osmomath.NewInt(100))),
			secondTakerFeeForCommunityPool: sdk.NewCoins(sdk.NewCoin("usdc", osmomath.NewInt(200))),
		},
		"get updated coin with different denom coins": {
			firstTakerFeeForStakers:  sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(100))),
			secondTakerFeeForStakers: sdk.NewCoins(sdk.NewCoin("usdc", osmomath.NewInt(200))),

			firstTakerFeeForCommunityPool:  sdk.NewCoins(sdk.NewCoin("usdc", osmomath.NewInt(100))),
			secondTakerFeeForCommunityPool: sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(200))),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			s.Require().Empty(s.App.PoolManagerKeeper.GetTakerFeeTrackerStartHeight(s.Ctx))
			s.Require().Empty(s.App.PoolManagerKeeper.GetTakerFeeTrackerForStakers(s.Ctx))

			for _, coin := range tc.firstTakerFeeForStakers {
				err := s.App.PoolManagerKeeper.UpdateTakerFeeTrackerForStakersByDenom(s.Ctx, coin.Denom, coin.Amount)
				s.Require().NoError(err)
			}
			for _, coin := range tc.firstTakerFeeForCommunityPool {
				err := s.App.PoolManagerKeeper.UpdateTakerFeeTrackerForCommunityPoolByDenom(s.Ctx, coin.Denom, coin.Amount)
				s.Require().NoError(err)
			}

			actualFirstTakerFeeForStakers := s.App.PoolManagerKeeper.GetTakerFeeTrackerForStakers(s.Ctx)
			actualFirstTakerFeeForCommunityPool := s.App.PoolManagerKeeper.GetTakerFeeTrackerForCommunityPool(s.Ctx)

			s.Require().Equal(tc.firstTakerFeeForStakers, actualFirstTakerFeeForStakers)
			s.Require().Equal(tc.firstTakerFeeForCommunityPool, actualFirstTakerFeeForCommunityPool)

			for _, coin := range tc.secondTakerFeeForStakers {
				err := s.App.PoolManagerKeeper.UpdateTakerFeeTrackerForStakersByDenom(s.Ctx, coin.Denom, coin.Amount)
				s.Require().NoError(err)
			}
			for _, coin := range tc.secondTakerFeeForCommunityPool {
				err := s.App.PoolManagerKeeper.UpdateTakerFeeTrackerForCommunityPoolByDenom(s.Ctx, coin.Denom, coin.Amount)
				s.Require().NoError(err)
			}

			expectedFinalTakerFeeForStakers := []sdk.Coin{}
			expectedFinalTakerFeeForCommunityPool := []sdk.Coin{}
			firstTakerFeeForStakersCoins := osmoutils.ConvertCoinArrayToCoins(actualFirstTakerFeeForStakers)
			firstTakerFeeForCommunityPoolCoins := osmoutils.ConvertCoinArrayToCoins(actualFirstTakerFeeForCommunityPool)
			secondTakerFeeForStakersCoins := osmoutils.ConvertCoinArrayToCoins(tc.secondTakerFeeForStakers)
			secondTakerFeeForCommunityPoolCoins := osmoutils.ConvertCoinArrayToCoins(tc.secondTakerFeeForCommunityPool)

			expectedFinalTakerFeeForStakers = firstTakerFeeForStakersCoins.Add(secondTakerFeeForStakersCoins...)
			expectedFinalTakerFeeForCommunityPool = firstTakerFeeForCommunityPoolCoins.Add(secondTakerFeeForCommunityPoolCoins...)

			actualSecondTakerFeeForStakers := s.App.PoolManagerKeeper.GetTakerFeeTrackerForStakers(s.Ctx)
			actualSecondTakerFeeForCommunityPool := s.App.PoolManagerKeeper.GetTakerFeeTrackerForCommunityPool(s.Ctx)

			s.Require().Equal(expectedFinalTakerFeeForStakers, actualSecondTakerFeeForStakers)
			s.Require().Equal(expectedFinalTakerFeeForCommunityPool, actualSecondTakerFeeForCommunityPool)
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

func (s *KeeperTestSuite) TestUpdateTakerFeeTrackerForStakersAndCommunityPool() {
	tests := map[string]struct {
		initialTakerFeeForStakers       []sdk.Coin
		initialTakerFeeForCommunityPool []sdk.Coin

		increaseTakerFeeForStakersBy       sdk.Coin
		increaseTakerFeeForCommunityPoolBy sdk.Coin
	}{
		"happy path: increase single denom tracker": {
			initialTakerFeeForStakers:       sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(100))),
			initialTakerFeeForCommunityPool: sdk.NewCoins(sdk.NewCoin("usdc", osmomath.NewInt(100))),

			increaseTakerFeeForStakersBy:       sdk.NewCoin("eth", osmomath.NewInt(50)),
			increaseTakerFeeForCommunityPoolBy: sdk.NewCoin("usdc", osmomath.NewInt(50)),
		},
		"increase multi denom tracker": {
			initialTakerFeeForStakers:       sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(100)), sdk.NewCoin("usdc", osmomath.NewInt(200))),
			initialTakerFeeForCommunityPool: sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(100)), sdk.NewCoin("usdc", osmomath.NewInt(200))),

			increaseTakerFeeForStakersBy:       sdk.NewCoin("eth", osmomath.NewInt(50)),
			increaseTakerFeeForCommunityPoolBy: sdk.NewCoin("usdc", osmomath.NewInt(50)),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			s.Require().Empty(s.App.PoolManagerKeeper.GetTakerFeeTrackerStartHeight(s.Ctx))
			s.Require().Empty(s.App.PoolManagerKeeper.GetTakerFeeTrackerForStakers(s.Ctx))

			for _, coin := range tc.initialTakerFeeForStakers {
				err := s.App.PoolManagerKeeper.UpdateTakerFeeTrackerForStakersByDenom(s.Ctx, coin.Denom, coin.Amount)
				s.Require().NoError(err)
			}
			for _, coin := range tc.initialTakerFeeForCommunityPool {
				err := s.App.PoolManagerKeeper.UpdateTakerFeeTrackerForCommunityPoolByDenom(s.Ctx, coin.Denom, coin.Amount)
				s.Require().NoError(err)
			}

			actualInitialTakerFeeForStakers := s.App.PoolManagerKeeper.GetTakerFeeTrackerForStakers(s.Ctx)
			actualInitialTakerFeeForCommunityPool := s.App.PoolManagerKeeper.GetTakerFeeTrackerForCommunityPool(s.Ctx)

			s.Require().Equal(tc.initialTakerFeeForStakers, actualInitialTakerFeeForStakers)
			s.Require().Equal(tc.initialTakerFeeForCommunityPool, actualInitialTakerFeeForCommunityPool)

			err := s.App.PoolManagerKeeper.UpdateTakerFeeTrackerForStakersByDenom(s.Ctx, tc.increaseTakerFeeForStakersBy.Denom, tc.increaseTakerFeeForStakersBy.Amount)
			s.Require().NoError(err)
			err = s.App.PoolManagerKeeper.UpdateTakerFeeTrackerForCommunityPoolByDenom(s.Ctx, tc.increaseTakerFeeForCommunityPoolBy.Denom, tc.increaseTakerFeeForCommunityPoolBy.Amount)
			s.Require().NoError(err)

			expectedFinalTakerFeeForStakers := []sdk.Coin{}
			expectedFinalTakerFeeForCommunityPool := []sdk.Coin{}
			initialTakerFeeForStakersCoins := osmoutils.ConvertCoinArrayToCoins(tc.initialTakerFeeForStakers)
			initialTakerFeeForCommunityPoolCoins := osmoutils.ConvertCoinArrayToCoins(tc.initialTakerFeeForCommunityPool)

			expectedFinalTakerFeeForStakers = initialTakerFeeForStakersCoins.Add(sdk.NewCoins(tc.increaseTakerFeeForStakersBy)...)
			expectedFinalTakerFeeForCommunityPool = initialTakerFeeForCommunityPoolCoins.Add(sdk.NewCoins(tc.increaseTakerFeeForCommunityPoolBy)...)

			takerFeeForStakersAfterIncrease := s.App.PoolManagerKeeper.GetTakerFeeTrackerForStakers(s.Ctx)
			takerFeeForCommunityPoolAfterIncrease := s.App.PoolManagerKeeper.GetTakerFeeTrackerForCommunityPool(s.Ctx)

			s.Require().Equal(expectedFinalTakerFeeForStakers, takerFeeForStakersAfterIncrease)
			s.Require().Equal(expectedFinalTakerFeeForCommunityPool, takerFeeForCommunityPoolAfterIncrease)
		})
	}
}
