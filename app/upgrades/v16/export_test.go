package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v25/x/tokenfactory/keeper"
)

var AuthorizedUptimes = authorizedUptimes

func UpdateTokenFactoryParams(ctx sdk.Context, tokenFactoryKeeper *tokenfactorykeeper.Keeper) {
	updateTokenFactoryParams(ctx, tokenFactoryKeeper)
}
