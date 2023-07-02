package cosmwasmpool_test

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/cosmwasmpool/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
)

func (s *PoolModuleSuite) TestInitGenesis() {
	s.Setup()

	expectedTotalLiquidity := sdk.Coins{}
	for i := 0; i < 3; i++ {
		s.FundAcc(s.TestAccs[0], initalDefaultSupply)
		pool := s.PrepareCustomTransmuterPool(s.TestAccs[0], defaultDenoms)
		s.JoinTransmuterPool(s.TestAccs[0], pool.GetId(), initalDefaultSupply)
		expectedTotalLiquidity = expectedTotalLiquidity.Add(initalDefaultSupply...)
		fmt.Println("poolsss", pool.GetTotalPoolLiquidity(s.Ctx))
	}

	pool, _ := s.App.CosmwasmPoolKeeper.GetPoolById(s.Ctx, 1)
	fmt.Println("pool liq", pool.GetTotalPoolLiquidity(s.Ctx))

	pools, err := s.App.CosmwasmPoolKeeper.GetPools(s.Ctx)
	if err != nil {
		panic(err)
	}

	cosmwasmPool, _ := pools[0].(types.CosmWasmExtension)
	fmt.Println("cosmwasmPool", cosmwasmPool.GetTotalPoolLiquidity(s.Ctx))

	cosmwasmPoolPreInit := pools[0]

	sf := cosmwasmPoolPreInit.GetSpreadFactor(s.Ctx)
	fmt.Println("spread factor", sf)

	poolAnys := []*codectypes.Any{}
	for _, poolI := range pools {
		cosmwasmPool, ok := poolI.(types.CosmWasmExtension)
		if !ok {
			panic("invalid pool type")
		}
		sf := cosmwasmPool.GetSpreadFactor(s.Ctx)
		fmt.Println("spread factor 2", sf)
		cosmwasmPool.SetWasmKeeper(s.App.WasmKeeper)
		any, err := codectypes.NewAnyWithValue(cosmwasmPool)
		if err != nil {
			panic(err)
		}
		poolAnys = append(poolAnys, any)
	}

	fmt.Println("pools liq", pools[0].(types.CosmWasmExtension).GetTotalPoolLiquidity(s.Ctx))

	liquiditytotal, err := s.App.CosmwasmPoolKeeper.GetTotalLiquidity(s.Ctx)
	s.Require().NoError(err)
	fmt.Println("TOTAL PRE liquidity", liquiditytotal.String())

	// We need to export the wasm module and reimport it after we reset the test environment
	// This is because cosmwasmpools point to a contract address, and if we don't export this as well
	// this test will fail.

	wasmGenState := wasmkeeper.ExportGenesis(s.Ctx, s.App.WasmKeeper)

	// Reset the testing env so that we can see if the pools get re-initialized via init genesis
	s.Setup()

	// Check if the pools were reset
	_, err = s.App.CosmwasmPoolKeeper.GetPool(s.Ctx, 1)
	s.Require().Error(err)

	_, err = wasmkeeper.InitGenesis(s.Ctx, s.App.WasmKeeper, *wasmGenState)
	s.Require().NoError(err)

	s.App.CosmwasmPoolKeeper.InitGenesis(s.Ctx, &types.GenesisState{
		Params: types.DefaultParams(),
		Pools:  poolAnys,
	}, s.App.AppCodec())

	poolStored, err := s.App.CosmwasmPoolKeeper.GetPool(s.Ctx, 1)
	s.Require().NoError(err)
	s.Require().Equal(cosmwasmPoolPreInit.GetId(), poolStored.GetId())
	s.Require().Equal(cosmwasmPoolPreInit.GetAddress(), poolStored.GetAddress())
	s.Require().Equal(cosmwasmPoolPreInit.GetSpreadFactor(s.Ctx), poolStored.GetSpreadFactor(s.Ctx))
	s.Require().Equal(cosmwasmPoolPreInit.String(), poolStored.String())

	_, err = s.App.CosmwasmPoolKeeper.GetPoolById(s.Ctx, 4)
	s.Require().Error(err)

	liquidity, err := s.App.CosmwasmPoolKeeper.GetTotalLiquidity(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(expectedTotalLiquidity.String(), liquidity.String())
}

func (s *PoolModuleSuite) TestExportGenesis() {
	s.Setup()
	//validTransmuterCodeId := uint64(1)

	// acc1 := s.TestAccs[0]
	// err := simapp.FundAccount(s.App.BankKeeper, s.Ctx, acc1, sdk.NewCoins(
	// 	sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
	// 	sdk.NewInt64Coin("foo", 100000),
	// 	sdk.NewInt64Coin("bar", 100000),
	// ))
	// s.Require().NoError(err)

	// msg := model.NewMsgCreateCosmWasmPool(validTransmuterCodeId, s.TestAccs[0], s.GetDefaultTransmuterInstantiateMsgBytes())
	// _, err = s.App.PoolManagerKeeper.CreatePool(s.Ctx, msg)
	// s.Require().NoError(err)
	s.FundAcc(s.TestAccs[0], initalDefaultSupply)
	s.PrepareCustomTransmuterPool(s.TestAccs[0], defaultDenoms)

	// msg = model.NewMsgCreateCosmWasmPool(validTransmuterCodeId, s.TestAccs[0], s.GetDefaultTransmuterInstantiateMsgBytes())
	// _, err = s.App.PoolManagerKeeper.CreatePool(s.Ctx, msg)
	// s.Require().NoError(err)
	s.FundAcc(s.TestAccs[0], initalDefaultSupply)
	s.PrepareCustomTransmuterPool(s.TestAccs[0], defaultDenoms)

	genesis := s.App.CosmwasmPoolKeeper.ExportGenesis(s.Ctx)
	s.Require().Len(genesis.Pools, 2)
}

func (s *PoolModuleSuite) TestMarshalUnmarshalGenesis() {
	s.Setup()
	//validTransmuterCodeId := uint64(1)

	acc1 := s.TestAccs[0]
	err := simapp.FundAccount(s.App.BankKeeper, s.Ctx, acc1, sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
		sdk.NewInt64Coin("foo", 100000),
		sdk.NewInt64Coin("bar", 100000),
	))
	s.Require().NoError(err)

	// msg := model.NewMsgCreateCosmWasmPool(validTransmuterCodeId, s.TestAccs[0], s.GetDefaultTransmuterInstantiateMsgBytes())
	// _, err = s.App.PoolManagerKeeper.CreatePool(s.Ctx, msg)
	// s.Require().NoError(err)
	s.FundAcc(s.TestAccs[0], initalDefaultSupply)
	s.PrepareCustomTransmuterPool(s.TestAccs[0], defaultDenoms)

	genesis := s.App.CosmwasmPoolKeeper.ExportGenesis(s.Ctx)
	s.Assert().NotPanics(func() {
		s.App.CosmwasmPoolKeeper.InitGenesis(s.Ctx, genesis, s.App.AppCodec())
	})
}
