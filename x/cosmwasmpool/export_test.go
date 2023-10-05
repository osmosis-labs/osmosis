package cosmwasmpool

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) DeWhitelistCodeId(ctx sdk.Context, codeId uint64) bool {
	return k.deWhiteListCodeId(ctx, codeId)
}

func (k Keeper) IsWhitelisted(ctx sdk.Context, codeId uint64) bool {
	return k.isWhitelisted(ctx, codeId)
}

func (k Keeper) UploadCodeIdAndWhitelist(ctx sdk.Context, byteCode []byte) (uint64, error) {
	return k.uploadCodeIdAndWhitelist(ctx, byteCode)
}

func (k Keeper) MigrateCosmwasmPools(ctx sdk.Context, poolIds []uint64, newCodeId uint64, uploadByteCode []byte, migrateMsg []byte) (err error) {
	return k.migrateCosmwasmPools(ctx, poolIds, newCodeId, uploadByteCode, migrateMsg)
}
