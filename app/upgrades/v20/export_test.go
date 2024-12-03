package v20

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v28/app/keepers"
)

func CreateGroupsForIncentivePairs(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	return createGroupsForIncentivePairs(ctx, keepers)
}
