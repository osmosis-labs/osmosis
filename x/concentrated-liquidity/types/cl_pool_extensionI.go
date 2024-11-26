package types

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
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
	GetLiquidity() sdkmath.LegacyDec
	GetLastLiquidityUpdate() time.Time
	SetCurrentSqrtPrice(newSqrtPrice osmomath.BigDec)
	SetCurrentTick(newTick int64)
	SetTickSpacing(newTickSpacing uint64)
	SetLastLiquidityUpdate(newTime time.Time)

	UpdateLiquidity(newLiquidity osmomath.Dec)
	ApplySwap(newLiquidity osmomath.Dec, newCurrentTick int64, newCurrentSqrtPrice osmomath.BigDec) error
	CalcActualAmounts(ctx sdk.Context, lowerTick, upperTick int64, liquidityDelta osmomath.Dec) (actualAmountDenom0 osmomath.Dec, actualAmountDenom1 osmomath.Dec, err error)
	UpdateLiquidityIfActivePosition(ctx sdk.Context, lowerTick, upperTick int64, liquidityDelta osmomath.Dec) bool
}
