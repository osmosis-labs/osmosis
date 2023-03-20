package apptesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/types"
)

// PrepareCosmWasmPool sets up a cosmwasm pool with the default parameters.
func (s *KeeperTestHelper) PrepareCosmWasmPool() types.CosmWasmExtension {
	return s.PrepareCustomCosmWasmPool(s.TestAccs[0])
}

// PrepareCustomConcentratedPool sets up a concentrated liquidity pool with the custom parameters.
func (s *KeeperTestHelper) PrepareCustomCosmWasmPool(owner sdk.AccAddress) types.CosmWasmExtension {
	// Mint some assets to the account.
	s.FundAcc(s.TestAccs[0], DefaultAcctFunds)

	// Create a concentrated pool via the poolmanager
	poolID, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, model.NewMsgCreateCosmWasmPool(owner))
	s.Require().NoError(err)

	// Retrieve the poolInterface via the poolID
	poolI, err := s.App.ConcentratedLiquidityKeeper.GetPool(s.Ctx, poolID)
	s.Require().NoError(err)

	// Type cast the PoolInterface to a ConcentratedPoolExtension
	pool, ok := poolI.(types.CosmWasmExtension)
	s.Require().True(ok)

	return pool
}
