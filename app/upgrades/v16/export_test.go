package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/osmosis-labs/osmosis/v15/app/keepers"
	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	incentiveskeeper "github.com/osmosis-labs/osmosis/v15/x/incentives/keeper"
	incentivestypes "github.com/osmosis-labs/osmosis/v15/x/incentives/types"
	poolincentiveskeeper "github.com/osmosis-labs/osmosis/v15/x/pool-incentives/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

func GetGaugesForCFMMPool(ctx sdk.Context, incentivesKeeper incentiveskeeper.Keeper, poolIncentivesKeeper poolincentiveskeeper.Keeper, poolId uint64) ([]incentivestypes.Gauge, error) {
	return getGaugesForCFMMPool(ctx, incentivesKeeper, poolIncentivesKeeper, poolId)
}

func CreateConcentratedPoolFromCFMM(ctx sdk.Context, cfmmPoolIdToLinkWith uint64, desiredDenom0 string, accountKeeper authkeeper.AccountKeeper, gammKeeper gammkeeper.Keeper, poolmanagerKeeper poolmanager.Keeper) (poolmanagertypes.PoolI, error) {
	return createConcentratedPoolFromCFMM(ctx, cfmmPoolIdToLinkWith, desiredDenom0, accountKeeper, gammKeeper, poolmanagerKeeper)
}

func CreateCanonicalConcentratedLiuqidityPoolAndMigrationLink(ctx sdk.Context, cfmmPoolId uint64, desiredDenom0 string, keepers *keepers.AppKeepers) error {
	return createCanonicalConcentratedLiuqidityPoolAndMigrationLink(ctx, cfmmPoolId, desiredDenom0, keepers)
}
