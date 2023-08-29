package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

type ConcentratedPoolExtension interface {
	poolmanagertypes.PoolI

	IsCurrentTickInRange(lowerTick, upperTick int64) bool
	GetIncentivesAddress() sdk.AccAddress
	GetSpreadRewardsAddress() sdk.AccAddress
	GetToken0() string
	GetToken1() string
	GetCurrentSqrtPrice() osmomath.BigDec
	GetCurrentTick() int64
	GetExponentAtPriceOne() int64
	GetTickSpacing() uint64
	GetLiquidity() sdk.Dec
	GetLastLiquidityUpdate() time.Time
	SetCurrentSqrtPrice(newSqrtPrice osmomath.BigDec)
	SetCurrentTick(newTick int64)
	SetTickSpacing(newTickSpacing uint64)
	SetLastLiquidityUpdate(newTime time.Time)

	UpdateLiquidity(newLiquidity sdk.Dec)
	ApplySwap(newLiquidity sdk.Dec, newCurrentTick int64, newCurrentSqrtPrice osmomath.BigDec) error
	CalcActualAmounts(ctx sdk.Context, lowerTick, upperTick int64, liquidityDelta sdk.Dec) (actualAmountDenom0 sdk.Dec, actualAmountDenom1 sdk.Dec, err error)
	UpdateLiquidityIfActivePosition(ctx sdk.Context, lowerTick, upperTick int64, liquidityDelta sdk.Dec) bool
}
