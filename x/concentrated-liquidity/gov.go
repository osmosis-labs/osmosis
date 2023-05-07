package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

func handleCreateConcentratedLiquidityPoolProposal(ctx sdk.Context, k Keeper, p *types.CreateConcentratedLiquidityPoolProposal) error {
	return k.HandleCreateConcentratedLiquidityPoolProposal(ctx, p)
}

func (k Keeper) HandleCreateConcentratedLiquidityPoolProposal(ctx sdk.Context, p *types.CreateConcentratedLiquidityPoolProposal) error {
	poolmanagerModuleAcc := k.accountKeeper.GetModuleAccount(ctx, poolmanagertypes.ModuleName)
	poolCreatorAddress := poolmanagerModuleAcc.GetAddress()
	createPoolMsg := clmodel.NewMsgCreateConcentratedPool(poolCreatorAddress, p.Denom0, p.Denom1, p.TickSpacing, p.SwapFee)
	_, err := k.poolmanagerKeeper.CreateConcentratedPoolAsPoolManager(ctx, createPoolMsg)
	return err
}

// HandleTickSpacingDecreaseProposal handles a tick spacing decrease proposal to the corresponding keeper method.
func (k Keeper) HandleTickSpacingDecreaseProposal(ctx sdk.Context, p *types.TickSpacingDecreaseProposal) error {
	return k.DecreaseConcentratedPoolTickSpacing(ctx, p.PoolIdToTickSpacingRecords)
}
