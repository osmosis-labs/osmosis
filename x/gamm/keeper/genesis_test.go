package keeper_test

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	gammmigration "github.com/osmosis-labs/osmosis/v16/x/gamm/types/migration"
)

var DefaultMigrationRecords = gammmigration.MigrationRecords{BalancerToConcentratedPoolLinks: []gammmigration.BalancerToConcentratedPoolLink{
	{BalancerPoolId: 1, ClPoolId: 4},
	{BalancerPoolId: 2, ClPoolId: 5},
	{BalancerPoolId: 3, ClPoolId: 6},
}}

func (s *KeeperTestSuite) TestGammInitGenesis() {
	s.SetupTest()

	for i := 0; i < 3; i++ {
		s.PrepareBalancerPool()
	}
	for i := 0; i < 3; i++ {
		s.PrepareConcentratedPool()
	}

	pools, err := s.App.GAMMKeeper.GetPoolsAndPoke(s.Ctx)
	if err != nil {
		panic(err)
	}

	balancerPoolPreInit := pools[0]

	poolAnys := []*codectypes.Any{}
	for _, poolI := range pools {
		any, err := codectypes.NewAnyWithValue(poolI)
		if err != nil {
			panic(err)
		}
		poolAnys = append(poolAnys, any)
	}

	// Reset the testing env so that we can see if the pools get re-initialized via init genesis
	s.SetupTest()

	// Check if the pools were reset
	_, err = s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, 1)
	s.Require().Error(err)

	s.App.GAMMKeeper.InitGenesis(s.Ctx, types.GenesisState{
		Pools:          poolAnys,
		NextPoolNumber: 7,
		Params: types.Params{
			PoolCreationFee: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000_000_000)},
		},
		MigrationRecords: &DefaultMigrationRecords,
	}, s.App.AppCodec())

	poolStored, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, 1)
	s.Require().NoError(err)
	s.Require().Equal(balancerPoolPreInit.GetId(), poolStored.GetId())
	s.Require().Equal(balancerPoolPreInit.GetAddress(), poolStored.GetAddress())
	s.Require().Equal(balancerPoolPreInit.GetSpreadFactor(s.Ctx), poolStored.GetSpreadFactor(s.Ctx))
	s.Require().Equal(balancerPoolPreInit.GetExitFee(s.Ctx), poolStored.GetExitFee(s.Ctx))
	s.Require().Equal(balancerPoolPreInit.GetTotalShares(), poolStored.GetTotalShares())
	s.Require().Equal(balancerPoolPreInit.String(), poolStored.String())

	_, err = s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, 7)
	s.Require().Error(err)

	liquidity, err := s.App.GAMMKeeper.GetTotalLiquidity(s.Ctx)
	s.Require().NoError(err)
	expectedLiquidity := sdk.NewCoins(sdk.NewInt64Coin("bar", 15000000), sdk.NewInt64Coin("baz", 15000000), sdk.NewInt64Coin("foo", 15000000), sdk.NewInt64Coin("uosmo", 15000000))
	s.Require().Equal(expectedLiquidity.String(), liquidity.String())

	postInitGenMigrationRecords, err := s.App.GAMMKeeper.GetAllMigrationInfo(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(DefaultMigrationRecords, postInitGenMigrationRecords)
}

func (s *KeeperTestSuite) TestGammExportGenesis() {
	s.SetupTest()
	ctx := s.Ctx

	acc1 := s.TestAccs[0]
	err := simapp.FundAccount(s.App.BankKeeper, ctx, acc1, sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
		sdk.NewInt64Coin("foo", 100000),
		sdk.NewInt64Coin("bar", 100000),
	))
	s.Require().NoError(err)

	msg := balancer.NewMsgCreateBalancerPool(acc1, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}, []balancer.PoolAsset{{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}, {
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
	}}, "")
	_, err = s.App.PoolManagerKeeper.CreatePool(ctx, msg)
	s.Require().NoError(err)

	msg = balancer.NewMsgCreateBalancerPool(acc1, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}, []balancer.PoolAsset{{
		Weight: sdk.NewInt(70),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}, {
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
	}}, "")
	_, err = s.App.PoolManagerKeeper.CreatePool(ctx, msg)
	s.Require().NoError(err)

	s.App.GAMMKeeper.SetMigrationRecords(ctx, DefaultMigrationRecords)

	genesis := s.App.GAMMKeeper.ExportGenesis(ctx)
	s.Require().Len(genesis.Pools, 2)
	s.Require().Equal(&DefaultMigrationRecords, genesis.MigrationRecords)
}

func (s *KeeperTestSuite) TestMarshalUnmarshalGenesis() {
	s.SetupTest()
	ctx := s.Ctx

	acc1 := s.TestAccs[0]
	err := simapp.FundAccount(s.App.BankKeeper, ctx, acc1, sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
		sdk.NewInt64Coin("foo", 100000),
		sdk.NewInt64Coin("bar", 100000),
	))
	s.Require().NoError(err)

	msg := balancer.NewMsgCreateBalancerPool(acc1, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}, []balancer.PoolAsset{{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}, {
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
	}}, "")
	_, err = s.App.PoolManagerKeeper.CreatePool(ctx, msg)
	s.Require().NoError(err)

	s.App.GAMMKeeper.SetMigrationRecords(ctx, DefaultMigrationRecords)
	s.Require().NoError(err)

	genesis := s.App.GAMMKeeper.ExportGenesis(ctx)
	s.Assert().NotPanics(func() {
		s.App.GAMMKeeper.InitGenesis(s.Ctx, *genesis, s.App.AppCodec())
	})
}
