package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	clmodel "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

func handleCreateConcentratedLiquidityPoolsProposal(ctx sdk.Context, k Keeper, p *types.CreateConcentratedLiquidityPoolsProposal) error {
	return k.HandleCreateConcentratedLiquidityPoolsProposal(ctx, p)
}

func (k Keeper) HandleCreateConcentratedLiquidityPoolsProposal(ctx sdk.Context, p *types.CreateConcentratedLiquidityPoolsProposal) error {
	poolmanagerModuleAcc := k.accountKeeper.GetModuleAccount(ctx, poolmanagertypes.ModuleName)
	poolCreatorAddress := poolmanagerModuleAcc.GetAddress()
	for _, record := range p.PoolRecords {
		createPoolMsg := clmodel.NewMsgCreateConcentratedPool(poolCreatorAddress, record.Denom0, record.Denom1, record.TickSpacing, record.SpreadFactor)
		_, err := k.poolmanagerKeeper.CreateConcentratedPoolAsPoolManager(ctx, createPoolMsg)
		if err != nil {
			return err
		}
	}
	return nil
}

// HandleTickSpacingDecreaseProposal handles a tick spacing decrease proposal to the corresponding keeper method.
func (k Keeper) HandleTickSpacingDecreaseProposal(ctx sdk.Context, p *types.TickSpacingDecreaseProposal) error {
	return k.DecreaseConcentratedPoolTickSpacing(ctx, p.PoolIdToTickSpacingRecords)
}
