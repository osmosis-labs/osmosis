package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"

	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

type ConcentratedPoolExtension interface {
	swaproutertypes.PoolI
	proto.Message

	// TODO: move these to separate interfaces
	GetToken0() string
	GetToken1() string
	GetCurrentSqrtPrice() sdk.Dec
	GetCurrentTick() sdk.Int
	GetLiquidity() sdk.Dec

	UpdateLiquidity(newLiquidity sdk.Dec)
	ApplySwap(newLiquidity sdk.Dec, newCurrentTick sdk.Int, newCurrentSqrtPrice sdk.Dec) error
	CalcActualAmounts(ctx sdk.Context, lowerTick, upperTick int64, sqrtRatioLowerTick, sqrtRatioUpperTick sdk.Dec, liquidityDelta sdk.Dec) (actualAmountDenom0 sdk.Dec, actualAmountDenom1 sdk.Dec)
	UpdateLiquidityIfActivePosition(ctx sdk.Context, lowerTick, upperTick int64, liquidityDelta sdk.Dec) bool
}
