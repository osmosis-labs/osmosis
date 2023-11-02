package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestGetSetTxFeesTrackerValue() {
	tests := map[string]struct {
		firstTxFeesValue  sdk.Coins
		secondTxFeesValue sdk.Coins
	}{
		"happy path: replace single coin with increased single coin": {
			firstTxFeesValue:  sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100))),
			secondTxFeesValue: sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(200))),
		},
		"replace single coin with decreased single coin": {
			firstTxFeesValue:  sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100))),
			secondTxFeesValue: sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(50))),
		},
		"replace single coin with different denom": {
			firstTxFeesValue:  sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100))),
			secondTxFeesValue: sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(100))),
		},
		"replace single coin with multiple coins": {
			firstTxFeesValue:  sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100))),
			secondTxFeesValue: sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100)), sdk.NewCoin("usdc", sdk.NewInt(200))),
		},
		"replace multiple coins with single coin": {
			firstTxFeesValue:  sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100)), sdk.NewCoin("usdc", sdk.NewInt(200))),
			secondTxFeesValue: sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(200))),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest(false)

			s.Require().Empty(s.App.TxFeesKeeper.GetTxFeesTrackerValue(s.Ctx))

			s.App.TxFeesKeeper.SetTxFeesTrackerValue(s.Ctx, tc.firstTxFeesValue)

			actualFirstTxFeesValue := s.App.TxFeesKeeper.GetTxFeesTrackerValue(s.Ctx)

			s.Require().Equal(tc.firstTxFeesValue, actualFirstTxFeesValue)

			s.App.TxFeesKeeper.SetTxFeesTrackerValue(s.Ctx, tc.secondTxFeesValue)

			actualSecondTxFeesValue := s.App.TxFeesKeeper.GetTxFeesTrackerValue(s.Ctx)

			s.Require().Equal(tc.secondTxFeesValue, actualSecondTxFeesValue)
		})
	}
}

func (s *KeeperTestSuite) TestGetSetTxFeesTrackerStartHeight() {
	tests := map[string]struct {
		firstTxFeesTrackerStartHeight  int64
		secondTxFeesTrackerStartHeight int64
	}{
		"replace tracker height with a higher height": {
			firstTxFeesTrackerStartHeight:  100,
			secondTxFeesTrackerStartHeight: 5000,
		},
		"replace tracker height with a lower height": {
			firstTxFeesTrackerStartHeight:  100,
			secondTxFeesTrackerStartHeight: 50,
		},
		"replace tracker height back to zero": {
			firstTxFeesTrackerStartHeight:  100,
			secondTxFeesTrackerStartHeight: 0,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest(false)

			s.Require().Empty(s.App.TxFeesKeeper.GetTxFeesTrackerStartHeight(s.Ctx))

			s.App.TxFeesKeeper.SetTxFeesTrackerStartHeight(s.Ctx, tc.firstTxFeesTrackerStartHeight)
			actualFirstTxFeesTrackerStartHeight := s.App.TxFeesKeeper.GetTxFeesTrackerStartHeight(s.Ctx)
			s.Require().Equal(tc.firstTxFeesTrackerStartHeight, actualFirstTxFeesTrackerStartHeight)

			s.App.TxFeesKeeper.SetTxFeesTrackerStartHeight(s.Ctx, tc.secondTxFeesTrackerStartHeight)
			actualSecondTxFeesTrackerStartHeight := s.App.TxFeesKeeper.GetTxFeesTrackerStartHeight(s.Ctx)
			s.Require().Equal(tc.secondTxFeesTrackerStartHeight, actualSecondTxFeesTrackerStartHeight)
		})
	}
}

func (s *KeeperTestSuite) TestIncreaseTxFeesTracker() {
	tests := map[string]struct {
		initialTxFeesValue    sdk.Coins
		increaseTxFeesValueBy sdk.Coin
	}{
		"happy path: increase single denom tracker": {
			initialTxFeesValue:    sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100))),
			increaseTxFeesValueBy: sdk.NewCoin("eth", sdk.NewInt(50)),
		},
		"increase multi denom tracker": {
			initialTxFeesValue:    sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100)), sdk.NewCoin("usdc", sdk.NewInt(200))),
			increaseTxFeesValueBy: sdk.NewCoin("eth", sdk.NewInt(50)),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest(false)

			s.Require().Empty(s.App.TxFeesKeeper.GetTxFeesTrackerStartHeight(s.Ctx))

			s.App.TxFeesKeeper.SetTxFeesTrackerValue(s.Ctx, tc.initialTxFeesValue)
			actualInitialTxFeesValue := s.App.TxFeesKeeper.GetTxFeesTrackerValue(s.Ctx)
			s.Require().Equal(tc.initialTxFeesValue, actualInitialTxFeesValue)

			s.App.TxFeesKeeper.IncreaseTxFeesTracker(s.Ctx, tc.increaseTxFeesValueBy)
			txFeesValueAfterIncrease := s.App.TxFeesKeeper.GetTxFeesTrackerValue(s.Ctx)
			s.Require().Equal(tc.initialTxFeesValue.Add(sdk.NewCoins(tc.increaseTxFeesValueBy)...), txFeesValueAfterIncrease)
		})
	}
}
