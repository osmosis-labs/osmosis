package types

import (
	"time"

	proto "github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/v8/v043_temp/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PoolI defines an interface for pools that hold tokens.
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
	// GetTotalPoolLiquidity returns the coins in the pool owned by all LPs
	GetTotalPoolLiquidity(ctx sdk.Context) sdk.Coins
	// GetTotalShares returns the total number of LP shares in the pool
	GetTotalShares() sdk.Int

	// SwapOutAmtGivenIn swaps 'tokenIn' against the pool, for tokenOutDenom, with the provided swapFee charged.
	// Balance transfers are done in the keeper, but this method updates the internal pool state.
	SwapOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error)
	// CalcOutAmtGivenIn returns how many coins SwapOutAmtGivenIn would return on these arguments.
	// This does not mutate the pool, or state.
	CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error)

	// SwapInAmtGivenOut swaps exactly enough tokensIn against the pool, to get the provided tokenOut amount out of the pool.
	// Balance transfers are done in the keeper, but this method updates the internal pool state.
	SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error)
	// CalcInAmtGivenOut returns how many coins SwapInAmtGivenOut would return on these arguments.
	// This does not mutate the pool, or state.
	CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error)

	// Returns the spot price of the 'base asset' in terms of the 'quote asset' in the pool,
	// errors if either baseAssetDenom, or quoteAssetDenom does not exist.
	// For example, if this was a UniV2 50-50 pool, with 2 ETH, and 8000 UST
	// pool.SpotPrice(ctx, "eth", "ust") = 4000.00
	SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error)

	// JoinPool joins the pool using all of the tokensIn provided.
	// The AMM swaps to the correct internal ratio should be and returns the number of shares created.
	// This function is mutative and updates the pool's internal state if there is no error.
	// It is up to pool implementation if they support LP'ing at arbitrary ratios, or a subset of ratios.
	// Pools are expected to guarantee LP'ing at the exact ratio, and single sided LP'ing.
	JoinPool(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, err error)

	// CalcJoinPoolShares returns how many LP shares JoinPool would return on these arguments.
	// This does not mutate the pool, or state.
	CalcJoinPoolShares(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error)

	// ExitPool exits #numShares LP shares from the pool, decreases its internal liquidity & LP share totals,
	// and returns the number of coins that are being returned.
	// This mutates the pool and state.
	ExitPool(ctx sdk.Context, numShares sdk.Int, exitFee sdk.Dec) (exitedCoins sdk.Coins, err error)
	// CalcExitPoolShares returns how many coins ExitPool would return on these arguments.
	// This does not mutate the pool, or state.
	CalcExitPoolShares(ctx sdk.Context, numShares sdk.Int, exitFee sdk.Dec) (exitedCoins sdk.Coins, err error)

	// PokePool determines if a pool's weights need to be updated and updates
	// them if so.
	PokePool(blockTime time.Time)
}

// PoolAmountOutExtension is an extension of the PoolI
// interface definiting an abstraction for pools that hold tokens.
// In addition, it supports JoinSwapShareAmountOut and ExitSwapExactAmountOut methods
// that allow joining with the exact amount of shares to get out, and exiting with exact
// amount of coins to get out.
// See definitions below.
type PoolAmountOutExtension interface {
	PoolI

	// CalcTokenInShareAmountOut returns the number of tokenInDenom tokens
	// that would be returned if swapped for an exact number of shares (shareOutAmount).
	// Returns error if tokenInDenom is not in the pool or if fails to approximate
	// given the shareOutAmount.
	// This method does not mutate the pool
	CalcTokenInShareAmountOut(
		ctx sdk.Context,
		tokenInDenom string,
		shareOutAmount sdk.Int,
		swapFee sdk.Dec,
	) (tokenInAmount sdk.Int, err error)

	// JoinPoolTokenInMaxShareAmountOut add liquidity to a specified pool with a maximum amount of tokens in (tokenInMaxAmount)
	// and swaps to an exact number of shares (shareOutAmount).
	JoinPoolTokenInMaxShareAmountOut(
		ctx sdk.Context,
		tokenInDenom string,
		shareOutAmount sdk.Int,
	) (tokenInAmount sdk.Int, err error)

	// ExitSwapExactAmountOut removes liquidity from a specified pool with a maximum amount of LP shares (shareInMaxAmount)
	// and swaps to an exact amount of one of the token pairs (tokenOut).
	ExitSwapExactAmountOut(
		ctx sdk.Context,
		tokenOut sdk.Coin,
		shareInMaxAmount sdk.Int,
	) (shareInAmount sdk.Int, err error)

	// IncreaseLiquidity increases the pool's liquidity by the specified sharesOut and coinsIn.
	IncreaseLiquidity(sharesOut sdk.Int, coinsIn sdk.Coins)
}

func NewPoolAddress(poolId uint64) sdk.AccAddress {
	key := append([]byte("pool"), sdk.Uint64ToBigEndian(poolId)...)
	return address.Module(ModuleName, key)
}
