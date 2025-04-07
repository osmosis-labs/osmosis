package concentrated_liquidity

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	clmodel "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

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

func NewConcentratedLiquidityProposalHandler(k Keeper) govtypesv1.Handler {
	return func(ctx sdk.Context, content govtypesv1.Content) error {
		switch c := content.(type) {
		case *types.TickSpacingDecreaseProposal:
			return k.HandleTickSpacingDecreaseProposal(ctx, c)
		case *types.CreateConcentratedLiquidityPoolsProposal:
			return k.HandleCreateConcentratedLiquidityPoolsProposal(ctx, c)
		default:
			return fmt.Errorf("unrecognized concentrated liquidity proposal content type: %T", c)
		}
	}
}
