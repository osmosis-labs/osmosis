package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v19/x/tokenfactory/keeper"
)

var (
	AuthorizedQuoteDenoms = authorizedQuoteDenoms
	AuthorizedUptimes     = authorizedUptimes
)

func UpdateTokenFactoryParams(ctx sdk.Context, tokenFactoryKeeper *tokenfactorykeeper.Keeper) {
	updateTokenFactoryParams(ctx, tokenFactoryKeeper)
}
