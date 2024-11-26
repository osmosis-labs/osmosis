package cosmwasmpool_test

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"

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
	}

	pools, err := s.App.CosmwasmPoolKeeper.GetPoolsWithWasmKeeper(s.Ctx)
	if err != nil {
		panic(err)
	}

	cosmwasmPoolPreInit := pools[0]
	oldId := cosmwasmPoolPreInit.GetId()
	oldAddress := cosmwasmPoolPreInit.GetAddress()
	oldSpreadFactor := cosmwasmPoolPreInit.GetSpreadFactor(s.Ctx)
	oldString := cosmwasmPoolPreInit.String()

	poolAnys := []*codectypes.Any{}
	for _, poolI := range pools {
		cosmwasmPool, ok := poolI.(types.CosmWasmExtension)
		if !ok {
			panic("invalid pool type")
		}
		cosmwasmPool.SetWasmKeeper(s.App.WasmKeeper)
		any, err := codectypes.NewAnyWithValue(cosmwasmPool)
		if err != nil {
			panic(err)
		}
		poolAnys = append(poolAnys, any)
	}

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

	poolStored, err := s.App.CosmwasmPoolKeeper.GetPoolById(s.Ctx, 1)
	s.Require().NoError(err)
	s.Require().Equal(oldId, poolStored.GetId())
	s.Require().Equal(oldAddress, poolStored.GetAddress())
	s.Require().Equal(oldSpreadFactor, poolStored.GetSpreadFactor(s.Ctx))
	s.Require().Equal(oldString, poolStored.String())

	_, err = s.App.CosmwasmPoolKeeper.GetPoolById(s.Ctx, 4)
	s.Require().Error(err)

	liquidity, err := s.App.CosmwasmPoolKeeper.GetTotalLiquidity(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(expectedTotalLiquidity.String(), liquidity.String())
}

func (s *PoolModuleSuite) TestExportGenesis() {
	s.Setup()

	for i := 0; i < 2; i++ {
		s.FundAcc(s.TestAccs[0], initalDefaultSupply)
		s.PrepareCustomTransmuterPool(s.TestAccs[0], defaultDenoms)
	}

	genesis := s.App.CosmwasmPoolKeeper.ExportGenesis(s.Ctx)
	s.Require().Len(genesis.Pools, 2)

	for _, pool := range genesis.Pools {
		s.Require().Equal("/osmosis.cosmwasmpool.v1beta1.CosmWasmPool", pool.GetTypeUrl())
	}
}

func (s *PoolModuleSuite) TestMarshalUnmarshalGenesis() {
	s.Setup()

	s.FundAcc(s.TestAccs[0], initalDefaultSupply)
	s.PrepareCustomTransmuterPool(s.TestAccs[0], defaultDenoms)

	genesis := s.App.CosmwasmPoolKeeper.ExportGenesis(s.Ctx)
	s.Assert().NotPanics(func() {
		s.App.CosmwasmPoolKeeper.InitGenesis(s.Ctx, genesis, s.App.AppCodec())
	})
}
