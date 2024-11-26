package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/txfees/types"
)

var (
	testBaseDenom = appparams.BaseCoinUnit
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
	testWhitelistAddrs = []string{"osmo106x8q2nv7xsg7qrec2zgdf3vvq0t3gn49zvaha", "osmo105l5r3rjtynn7lg362r2m9hkpfvmgmjtkglsn9"}
)

func (s *KeeperTestSuite) TestInitGenesis() {
	s.SetupTest(false)
	s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000000000000000000)), sdk.NewCoin("uion", osmomath.NewInt(1000000000000000000)))...)
	s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000000000000000000)), sdk.NewCoin("wbtc", osmomath.NewInt(1000000000000000000)))...)

	s.App.TxFeesKeeper.InitGenesis(s.Ctx, types.GenesisState{
		Basedenom: testBaseDenom,
		Feetokens: testFeeTokens,
		Params: types.Params{
			WhitelistedFeeTokenSetters: testWhitelistAddrs,
		},
	})

	actualBaseDenom, err := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(testBaseDenom, actualBaseDenom)
	s.Require().Equal(testFeeTokens, s.App.TxFeesKeeper.GetFeeTokens(s.Ctx))

	actualParams := s.App.TxFeesKeeper.GetParams(s.Ctx)
	s.Require().Equal(testWhitelistAddrs, actualParams.WhitelistedFeeTokenSetters)
}

func (s *KeeperTestSuite) TestExportGenesis() {
	s.SetupTest(false)
	s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000000000000000000)), sdk.NewCoin("uion", osmomath.NewInt(1000000000000000000)))...)
	s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000000000000000000)), sdk.NewCoin("wbtc", osmomath.NewInt(1000000000000000000)))...)

	s.App.TxFeesKeeper.InitGenesis(s.Ctx, types.GenesisState{
		Basedenom: testBaseDenom,
		Feetokens: testFeeTokens,
		Params: types.Params{
			WhitelistedFeeTokenSetters: testWhitelistAddrs,
		},
	})

	genesis := s.App.TxFeesKeeper.ExportGenesis(s.Ctx)
	s.Require().Equal(testBaseDenom, genesis.Basedenom)
	s.Require().Equal(testFeeTokens, genesis.Feetokens)
	s.Require().Equal(testWhitelistAddrs, genesis.Params.WhitelistedFeeTokenSetters)
}
