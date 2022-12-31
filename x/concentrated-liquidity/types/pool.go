package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

type ConcentratedPoolExtension interface {
	swaproutertypes.PoolI

	// TODO: move these to separate interfaces
	GetToken0() string
	GetToken1() string
	GetCurrentSqrtPrice() sdk.Dec
	GetCurrentTick() sdk.Int
	GetTickSpacing() uint64
	GetLiquidity() sdk.Dec
	GetTimeOfCreation() time.Time
	GetGlobalSecondsPerLiquidity() sdk.Dec
	SetCurrentSqrtPrice(newSqrtPrice sdk.Dec)
	SetCurrentTick(newTick sdk.Int)
	SetGlobalSecondsPerLiquidity(newGlobalSecondsPerLiquidity sdk.Dec)

	UpdateLiquidity(newLiquidity sdk.Dec)
	ApplySwap(newLiquidity sdk.Dec, newCurrentTick sdk.Int, newCurrentSqrtPrice sdk.Dec) error
	CalcActualAmounts(ctx sdk.Context, lowerTick, upperTick int64, sqrtRatioLowerTick, sqrtRatioUpperTick sdk.Dec, liquidityDelta sdk.Dec) (actualAmountDenom0 sdk.Dec, actualAmountDenom1 sdk.Dec)
	UpdateLiquidityIfActivePosition(ctx sdk.Context, lowerTick, upperTick int64, liquidityDelta sdk.Dec) bool
}
