package cosmwasmpool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

// Whitelists the code id.
func (k Keeper) WhitelistCodeId(ctx sdk.Context, codeId uint64) {
	params := k.GetParams(ctx)
	if !osmoutils.Contains(params.CodeIdWhitelist, codeId) {
		params.CodeIdWhitelist = append(params.CodeIdWhitelist, codeId)
		k.SetParams(ctx, params)
	}
}

// deWhitelistCodeId removes the code id from the whitelist.
// Returns true if the code id was in the whitelist and was removed.
// Returns false if the code id was not in the whitelist.
// nolint: unused
func (k Keeper) deWhiteListCodeId(ctx sdk.Context, codeId uint64) bool {
	params := k.GetParams(ctx)
	whitelist := params.CodeIdWhitelist
	indexToRemove := len(whitelist)

	for i, id := range whitelist {
		if id == codeId {
			indexToRemove = i
			break
		}
	}

	foundCodeId := indexToRemove < len(whitelist)
	if foundCodeId {
		whitelist = append(whitelist[:indexToRemove], whitelist[indexToRemove+1:]...)
		params.CodeIdWhitelist = whitelist
		k.SetParams(ctx, params)
	}

	return foundCodeId
}

// isWhitelisted returns true if the code id is in the whitelist.
func (k Keeper) isWhitelisted(ctx sdk.Context, codeId uint64) bool {
	whitelist := k.GetParams(ctx).CodeIdWhitelist
	return osmoutils.Contains(whitelist, codeId)
}
