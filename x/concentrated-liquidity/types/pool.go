package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

type ConcentratedPoolExtension interface {
	poolmanagertypes.PoolI

	GetIncentivesAddress() sdk.AccAddress
	GetToken0() string
	GetToken1() string
	GetCurrentSqrtPrice() sdk.Dec
	GetCurrentTick() sdk.Int
	GetExponentAtPriceOne() sdk.Int
	GetTickSpacing() uint64
	GetLiquidity() sdk.Dec
	GetLastLiquidityUpdate() time.Time
	SetCurrentSqrtPrice(newSqrtPrice sdk.Dec)
	SetCurrentTick(newTick sdk.Int)
	SetTickSpacing(newTickSpacing uint64)
	SetLastLiquidityUpdate(newTime time.Time)

	UpdateLiquidity(newLiquidity sdk.Dec)
	ApplySwap(newLiquidity sdk.Dec, newCurrentTick sdk.Int, newCurrentSqrtPrice sdk.Dec) error
	CalcActualAmounts(ctx sdk.Context, lowerTick, upperTick int64, liquidityDelta sdk.Dec) (actualAmountDenom0 sdk.Dec, actualAmountDenom1 sdk.Dec, err error)
	UpdateLiquidityIfActivePosition(ctx sdk.Context, lowerTick, upperTick int64, liquidityDelta sdk.Dec) bool
}
