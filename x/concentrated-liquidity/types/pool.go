package types

import (
	proto "github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PoolI interface {
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

	// TODO: move these to separate interfaces
	CalcOutAmtGivenIn(ctx sdk.Context,
		poolTickKVStore types.KVStore,
		tokenInMin sdk.Coin, tokenOutDenom string,
		swapFee sdk.Dec, priceLimit sdk.Dec,
		poolId uint64) (tokenIn, tokenOut sdk.Coin, updatedTick sdk.Int, updatedLiquidity, updatedSqrtPrice sdk.Dec, err error)
	SwapOutAmtGivenIn(ctx sdk.Context,
		poolTickKVStore types.KVStore,
		tokenIn sdk.Coin, tokenOutDenom string,
		swapFee sdk.Dec, priceLimit sdk.Dec,
		poolId uint64) (tokenOut sdk.Coin, err error)

	CalcInAmtGivenOut(ctx sdk.Context,
		poolTickKVStore types.KVStore,
		tokenOutMin sdk.Coin, tokenInDenom string,
		swapFee sdk.Dec, priceLimit sdk.Dec,
		poolId uint64) (tokenIn, tokenOut sdk.Coin, updatedTick sdk.Int, updatedLiquidity, updatedSqrtPrice sdk.Dec, err error)
	SwapInAmtGivenOut(ctx sdk.Context,
		poolTickKVStore types.KVStore,
		tokenOut sdk.Coin, tokenInDenom string,
		swapFee sdk.Dec, priceLimit sdk.Dec,
		poolId uint64) (tokenIn sdk.Coin, err error)

	// TODO: move these to separate interfaces
	UpdateLiquidity(newLiquidity sdk.Dec)
}
