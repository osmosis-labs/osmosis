package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
)

func (k Keeper) HandleReplaceMigrationRecordsProposal(ctx sdk.Context, p *types.ReplaceMigrationRecordsProposal) error {
	return k.ReplaceMigrationRecords(ctx, p.Records)
}

func (k Keeper) HandleUpdateMigrationRecordsProposal(ctx sdk.Context, p *types.UpdateMigrationRecordsProposal) error {
	return k.UpdateMigrationRecords(ctx, p.Records)
}
