package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v21/x/txfees/types"
)

var (
	testBaseDenom = "uosmo"
	testFeeTokens = []types.FeeToken{
		{
			Denom:  "uion",
			PoolID: 1,
		},
		{
			Denom:  "wbtc",
			PoolID: 2,
		},
	}

	testTxFeesTracker = types.TxFeesTracker{
		TxFees:                     sdk.Coins{sdk.NewCoin("uosmo", sdk.NewInt(1000))},
		HeightAccountingStartsFrom: 100,
	}
)

func (s *KeeperTestSuite) TestInitGenesis() {
	s.SetupTest(false)
	s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000000000000000000)), sdk.NewCoin("uion", sdk.NewInt(1000000000000000000)))...)
	s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000000000000000000)), sdk.NewCoin("wbtc", sdk.NewInt(1000000000000000000)))...)

	s.App.TxFeesKeeper.InitGenesis(s.Ctx, types.GenesisState{
		Basedenom:     testBaseDenom,
		Feetokens:     testFeeTokens,
		TxFeesTracker: &testTxFeesTracker,
	})

	actualBaseDenom, err := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
	s.Require().NoError(err)

	s.Require().Equal(testBaseDenom, actualBaseDenom)
	s.Require().Equal(testFeeTokens, s.App.TxFeesKeeper.GetFeeTokens(s.Ctx))
	s.Require().Equal(testTxFeesTracker.TxFees, s.App.TxFeesKeeper.GetTxFeesTrackerValue(s.Ctx))
	s.Require().Equal(testTxFeesTracker.HeightAccountingStartsFrom, s.App.TxFeesKeeper.GetTxFeesTrackerStartHeight(s.Ctx))
}

func (s *KeeperTestSuite) TestExportGenesis() {
	s.SetupTest(false)
	s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000000000000000000)), sdk.NewCoin("uion", sdk.NewInt(1000000000000000000)))...)
	s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000000000000000000)), sdk.NewCoin("wbtc", sdk.NewInt(1000000000000000000)))...)

	s.App.TxFeesKeeper.InitGenesis(s.Ctx, types.GenesisState{
		Basedenom:     testBaseDenom,
		Feetokens:     testFeeTokens,
		TxFeesTracker: &testTxFeesTracker,
	})

	genesis := s.App.TxFeesKeeper.ExportGenesis(s.Ctx)
	s.Require().Equal(testBaseDenom, genesis.Basedenom)
	s.Require().Equal(testFeeTokens, genesis.Feetokens)
	s.Require().Equal(testTxFeesTracker.TxFees, genesis.TxFeesTracker.TxFees)
	s.Require().Equal(testTxFeesTracker.HeightAccountingStartsFrom, genesis.TxFeesTracker.HeightAccountingStartsFrom)
}
