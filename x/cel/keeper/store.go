package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetCell(ctx sdk.Context, cellID uint64) (res *types.CellI, err error) {
	bz := ctx.KVStore(k.storeKey).Get(types.KeyCell(cellID))
	err = k.cdc.UnmarshalBinaryBare(bz, &res)
	return
}

func (k Keeper) SetCell(ctx sdk.Context, cellID uint64, cell *types.CellI) {
	bz := k.cdc.MustMarshalBinaryBare(cell)
	ctx.KVStore(k.storeKey).Set(types.KeyCell(cellID), bz)
}

func (k Keeper) GetExpr(ctx sdk.Context, cellID, exprID uint64) (res *types.Expr, err error) {
	bz := ctx.KVStore(k.storeKey).Get(types.KeyExpr(cellID, exprID))
	err = k.cdc.UnmarshalBinaryBare(bz, &res)
	return
}

func (k Keeper) SetExpr(ctx sdk.Context, cellID, exprID uint64, expr *types.Expr) {
	bz := k.cdc.MustMarshalBinaryBare(expr)
	ctx.KVStore(k.storeKey).Set(types.KeyExpr(cellID, exprID), bz)
}
