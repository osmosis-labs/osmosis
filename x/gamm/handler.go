package gamm

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/osmosis-labs/osmosis/v27/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/types"
)

// NewGammProposalHandler is a handler for governance proposals for the GAMM module.
func NewGammProposalHandler(k keeper.Keeper) govtypesv1.Handler {
	return func(ctx sdk.Context, content govtypesv1.Content) error {
		switch c := content.(type) {
		case *types.UpdateMigrationRecordsProposal:
			return handleUpdateMigrationRecordsProposal(ctx, k, c)
		case *types.ReplaceMigrationRecordsProposal:
			return handleReplaceMigrationRecordsProposal(ctx, k, c)
		case *types.CreateConcentratedLiquidityPoolsAndLinktoCFMMProposal:
			return handleCreatingCLPoolAndLinkToCFMMProposal(ctx, k, c)
		case *types.SetScalingFactorControllerProposal:
			return handleSetScalingFactorControllerProposal(ctx, k, c)

		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized migration record proposal content type: %T", c)
		}
	}
}

// handleReplaceMigrationRecordsProposal is a handler for replacing migration records governance proposals
func handleReplaceMigrationRecordsProposal(ctx sdk.Context, k keeper.Keeper, p *types.ReplaceMigrationRecordsProposal) error {
	return k.HandleReplaceMigrationRecordsProposal(ctx, p)
}

// handleUpdateMigrationRecordsProposal is a handler for updating migration records governance proposals
func handleUpdateMigrationRecordsProposal(ctx sdk.Context, k keeper.Keeper, p *types.UpdateMigrationRecordsProposal) error {
	return k.HandleUpdateMigrationRecordsProposal(ctx, p)
}

func handleCreatingCLPoolAndLinkToCFMMProposal(ctx sdk.Context, k keeper.Keeper, p *types.CreateConcentratedLiquidityPoolsAndLinktoCFMMProposal) error {
	for _, record := range p.PoolRecordsWithCfmmLink {
		_, err := k.CreateCanonicalConcentratedLiquidityPoolAndMigrationLink(ctx, record.BalancerPoolId, record.Denom0, record.SpreadFactor, record.TickSpacing)
		if err != nil {
			return err
		}
	}
	return nil
}

// handleSetScalingFactorControllerProposal is a handler for gov proposals to set a stableswap pool's
// scaling factor controller address
func handleSetScalingFactorControllerProposal(ctx sdk.Context, k keeper.Keeper, p *types.SetScalingFactorControllerProposal) error {
	return k.HandleSetScalingFactorControllerProposal(ctx, p)
}
