package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// NewMigrationRecordHandler is a handler for governance proposals on new migration records.
func NewConcentratedLiquidityProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.CreateConcentratedLiquidityPoolProposal:
			return handleCreateConcentratedLiquidityPoolProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized migration record proposal content type: %T", c)
		}
	}
}

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
