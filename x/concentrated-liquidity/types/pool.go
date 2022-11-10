package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
)

type ConcentratedPoolExtension interface {
	// gammtypes.PoolI
	proto.Message

	GetAddress() sdk.AccAddress
	String() string
	GetId() uint64
	// GetSwapFee returns the pool's swap fee, based on the current state.
	// Pools may choose to make their swap fees dependent upon state
	// (prior TWAPs, network downtime, other pool states, etc.)
	// hence Context is provided as an argument.
	GetSwapFee(ctx sdk.Context) sdk.Dec
	// GetExitFee returns the pool's exit fee, based on the current state.
	// Pools may choose to make their exit fees dependent upon state.
	GetExitFee(ctx sdk.Context) sdk.Dec
	// Returns whether the pool has swaps enabled at the moment
	IsActive(ctx sdk.Context) bool
	// GetTotalShares returns the total number of LP shares in the pool
	GetTotalShares() sdk.Int

	SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error)

	// TODO: move these to separate interfaces
	GetToken0() string
	GetToken1() string
	GetCurrentSqrtPrice() sdk.Dec
	GetCurrentTick() sdk.Int
	GetLiquidity() sdk.Dec

	UpdateLiquidity(newLiquidity sdk.Dec)
	ApplySwap(ctx sdk.Context, poolId uint64, newLiquidity sdk.Dec, newCurrentTick sdk.Int, newCurrentSqrtPrice sdk.Dec) error
}
