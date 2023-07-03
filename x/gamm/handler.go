package gamm

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
)

// NewMigrationRecordHandler is a handler for governance proposals on new migration records.
func NewMigrationRecordHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.UpdateMigrationRecordsProposal:
			return handleUpdateMigrationRecordsProposal(ctx, k, c)
		case *types.ReplaceMigrationRecordsProposal:
			return handleReplaceMigrationRecordsProposal(ctx, k, c)

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
