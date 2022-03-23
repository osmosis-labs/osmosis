package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	proto "github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/v7/v043_temp/address"
)

// PoolI defines an interface for pools that hold tokens.
type PoolI interface {
	proto.Message

	GetAddress() sdk.AccAddress
	String() string
	GetId() uint64
	// Todo: may have to generalize this later
	// Also todo: come up with better name here
	GetSwapFee(ctx sdk.Context) sdk.Dec
	GetExitFee(ctx sdk.Context) sdk.Dec
	// Todo: These may not be needed
	NumAssets() int
	IsActive(curBlockTime time.Time) bool
	// Returns the coins in the pool owned by all LP shareholders
	GetTotalLpBalances(ctx sdk.Context) sdk.Coins
	// TODO: Add ctx here
	GetTotalShares() sdk.Int

	CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.DecCoin, err error)
	CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.DecCoin, err error)

	// TODO: Ensure this can only be called via gamm
	// TODO: Think through the API guarantees this is providing in conjunction with the caller being
	// expected to Set the pool into state as well.
	ApplySwap(ctx sdk.Context, tokenIn sdk.Coins, tokenOut sdk.Coins) error

	SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error)

	// JoinPool joins the pool, and uses all of the tokensIn provided.
	// The AMM swaps to whatever the ratio should be and returns the number of shares created.
	// Internally the pool updates its count for the number of shares in this function.
	// If the function errors, or should not be mutative, then state must be reverted after this call.
	JoinPool(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, err error)
	ExitPool(ctx sdk.Context, numShares sdk.Int, exitFee sdk.Dec) (exitedCoins sdk.Coins, err error)
}

var (
	MaxUserSpecifiedWeight    sdk.Int = sdk.NewIntFromUint64(1 << 20)
	GuaranteedWeightPrecision int64   = 1 << 30
)

func NewPoolAddress(poolId uint64) sdk.AccAddress {
	key := append([]byte("pool"), sdk.Uint64ToBigEndian(poolId)...)
	return address.Module(ModuleName, key)
}

func ValidateUserSpecifiedWeight(weight sdk.Int) error {
	if !weight.IsPositive() {
		return sdkerrors.Wrap(ErrNotPositiveWeight, weight.String())
	}

	if weight.GTE(MaxUserSpecifiedWeight) {
		return sdkerrors.Wrap(ErrWeightTooLarge, weight.String())
	}
	return nil
}
