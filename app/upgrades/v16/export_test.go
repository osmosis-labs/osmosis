package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v32/x/tokenfactory/keeper"
)

// AuthorizedQuoteDenoms = authorizedQuoteDenoms
var AuthorizedUptimes = authorizedUptimes

func UpdateTokenFactoryParams(ctx sdk.Context, tokenFactoryKeeper *tokenfactorykeeper.Keeper) {
	updateTokenFactoryParams(ctx, tokenFactoryKeeper)
}
