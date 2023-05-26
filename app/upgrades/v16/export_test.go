package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/osmosis-labs/osmosis/v15/app/keepers"
	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v15/x/tokenfactory/keeper"
)

var (
	AuthorizedQuoteDenoms = authorizedQuoteDenoms
	AuthorizedUptimes     = authorizedUptimes
)

func CreateConcentratedPoolFromCFMM(ctx sdk.Context, cfmmPoolIdToLinkWith uint64, desiredDenom0 string, accountKeeper authkeeper.AccountKeeper, gammKeeper gammkeeper.Keeper, poolmanagerKeeper poolmanager.Keeper) (poolmanagertypes.PoolI, error) {
	return createConcentratedPoolFromCFMM(ctx, cfmmPoolIdToLinkWith, desiredDenom0, accountKeeper, gammKeeper, poolmanagerKeeper)
}

func CreateCanonicalConcentratedLiquidityPoolAndMigrationLink(ctx sdk.Context, cfmmPoolId uint64, desiredDenom0 string, keepers *keepers.AppKeepers) (poolmanagertypes.PoolI, error) {
	return createCanonicalConcentratedLiquidityPoolAndMigrationLink(ctx, cfmmPoolId, desiredDenom0, keepers)
}

func UpdateTokenFactoryParams(ctx sdk.Context, tokenFactoryKeeper *tokenfactorykeeper.Keeper) {
	updateTokenFactoryParams(ctx, tokenFactoryKeeper)
}
